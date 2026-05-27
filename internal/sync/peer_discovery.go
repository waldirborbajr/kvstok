// internal/sync/peer_discovery.go

package sync

import (
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/grandcat/zeroconf"
)

// Peer represents a remote kvstok peer on the network
type Peer struct {
	ID         string // Public key hash as identifier
	DeviceName string
	PublicKey  *rsa.PublicKey
	Addresses  []net.IP
	Port       int
	LastSeen   time.Time
	TrustLevel int // 0-100, incremented on successful syncs
	HashedKey  string
}

// PeerManager handles peer discovery and management via mDNS
type PeerManager struct {
	deviceName  string
	publicKey   *rsa.PublicKey
	mdnsServer  *zeroconf.Server
	peers       map[string]*Peer // ID -> Peer
	peersMu     sync.RWMutex
	serviceType string // _kvstok._tcp
	domain      string // local.
	port        int
	stopChan    chan struct{}
}

const (
	ServiceType           = "_kvstok._tcp"
	Domain                = "local."
	DefaultPort           = 9999
	PeerDiscoveryInterval = 30 * time.Second
	PeerTimeoutDuration   = 5 * time.Minute
)

// NewPeerManager creates a new peer discovery manager
func NewPeerManager(deviceName string, publicKey *rsa.PublicKey, port int) (*PeerManager, error) {
	if port == 0 {
		port = DefaultPort
	}

	pm := &PeerManager{
		deviceName:  deviceName,
		publicKey:   publicKey,
		serviceType: ServiceType,
		domain:      Domain,
		port:        port,
		peers:       make(map[string]*Peer),
		stopChan:    make(chan struct{}),
	}

	return pm, nil
}

// Start begins advertising this device as a kvstok peer and discovering others
func (p *PeerManager) Start() error {
	// Create mDNS entry for this device
	info := []string{
		fmt.Sprintf("device=%s", p.deviceName),
		fmt.Sprintf("pubkey=%s", hashPublicKey(p.publicKey)),
	}

	server, err := zeroconf.Register(
		p.deviceName,
		p.serviceType,
		p.domain,
		p.port,
		info,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register mDNS service: %w", err)
	}

	p.mdnsServer = server

	// Start discovery goroutine
	go p.discoverPeersRoutine()

	return nil
}

// Stop stops the mDNS service and discovery
func (p *PeerManager) Stop() error {
	close(p.stopChan)

	if p.mdnsServer != nil {
		p.mdnsServer.Shutdown()
	}

	return nil
}

// discoverPeersRoutine periodically discovers peers on the network
func (p *PeerManager) discoverPeersRoutine() {
	ticker := time.NewTicker(PeerDiscoveryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopChan:
			return
		case <-ticker.C:
			entries := make(chan *zeroconf.ServiceEntry)
			go func(results <-chan *zeroconf.ServiceEntry) {
				for entry := range results {
					p.processPeerEntry(entry)
				}
			}(entries)

			// Query for kvstok services on the local network
			ctx, cancel := zeroconf.NewResolver(nil)
			if ctx == nil {
				continue
			}

			err := ctx.Browse("_kvstok._tcp", "local.", entries, p.stopChan)
			cancel()

			if err != nil {
				continue // Silently continue on discovery errors
			}

			// Clean up stale peers
			p.pruneStalePeers()
		}
	}
}

// processPeerEntry processes a discovered mDNS service entry
func (p *PeerManager) processPeerEntry(entry *zeroconf.ServiceEntry) {
	if entry.Instance == p.deviceName {
		return // Skip self
	}

	if len(entry.AddrIPv4) == 0 {
		return // No addresses
	}

	// Extract public key hash from TXT record
	pubkeyHash := extractTXTField(entry.Text, "pubkey")
	if pubkeyHash == "" {
		return
	}

	peerID := pubkeyHash
	addresses := make([]net.IP, len(entry.AddrIPv4))
	copy(addresses, entry.AddrIPv4)

	peer := &Peer{
		ID:         peerID,
		DeviceName: entry.Instance,
		Addresses:  addresses,
		Port:       entry.Port,
		LastSeen:   time.Now(),
		TrustLevel: 0,
		HashedKey:  pubkeyHash,
	}

	p.peersMu.Lock()
	p.peers[peerID] = peer
	p.peersMu.Unlock()
}

// DiscoverPeers returns all currently discovered peers
func (p *PeerManager) DiscoverPeers() []Peer {
	p.peersMu.RLock()
	defer p.peersMu.RUnlock()

	peers := make([]Peer, 0, len(p.peers))
	for _, peer := range p.peers {
		peers = append(peers, *peer)
	}

	return peers
}

// GetPeer returns a specific peer by ID
func (p *PeerManager) GetPeer(peerID string) (*Peer, error) {
	p.peersMu.RLock()
	defer p.peersMu.RUnlock()

	peer, exists := p.peers[peerID]
	if !exists {
		return nil, fmt.Errorf("peer not found: %s", peerID)
	}

	return peer, nil
}

// UpdatePeerTrustLevel increments a peer's trust level after successful interaction
func (p *PeerManager) UpdatePeerTrustLevel(peerID string, delta int) error {
	p.peersMu.Lock()
	defer p.peersMu.Unlock()

	peer, exists := p.peers[peerID]
	if !exists {
		return fmt.Errorf("peer not found: %s", peerID)
	}

	peer.TrustLevel += delta
	if peer.TrustLevel > 100 {
		peer.TrustLevel = 100
	}
	if peer.TrustLevel < 0 {
		peer.TrustLevel = 0
	}

	peer.LastSeen = time.Now()
	return nil
}

// pruneStalePeers removes peers that haven't been seen in a while
func (p *PeerManager) pruneStalePeers() {
	p.peersMu.Lock()
	defer p.peersMu.Unlock()

	now := time.Now()
	for peerID, peer := range p.peers {
		if now.Sub(peer.LastSeen) > PeerTimeoutDuration {
			delete(p.peers, peerID)
		}
	}
}

// hashPublicKey returns the hex-encoded SHA256 hash of a public key
func hashPublicKey(pubkey *rsa.PublicKey) string {
	if pubkey == nil {
		return ""
	}

	data := pubkey.N.Bytes()
	hash := hashBytes(data)
	return hex.EncodeToString(hash[:])
}

// hashBytes computes SHA256 hash of bytes
func hashBytes(data []byte) [32]byte {
	return sha256.Sum256(data)
}

// extractTXTField extracts a key=value pair from mDNS TXT records
func extractTXTField(txtRecords []string, key string) string {
	for _, record := range txtRecords {
		if len(record) > 0 && record[0] == key[0] {
			// Simple parsing: key=value
			for i := 0; i < len(record); i++ {
				if record[i] == '=' {
					if record[:i] == key {
						return record[i+1:]
					}
					break
				}
			}
		}
	}
	return ""
}
