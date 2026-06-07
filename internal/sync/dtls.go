// Package sync handles DTLS-based peer-to-peer synchronization for kvstok.
package sync

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/pion/dtls/v2"
	"github.com/waldirborbajr/kvstok/internal/database"
)

// Message represents a message sent between peers
type Message struct {
	Type             string     `json:"type"` // "key_delta", "ack", "heartbeat"
	Timestamp        time.Time  `json:"timestamp"`
	SenderID         string     `json:"sender_id"`
	Keys             []KeyDelta `json:"keys,omitempty"`
	Signature        []byte     `json:"signature,omitempty"`
	EncryptedPayload []byte     `json:"encrypted_payload,omitempty"`
}

// KeyDelta represents a single key change for delta sync
type KeyDelta struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Tags      []string  `json:"tags"`
	TTL       uint32    `json:"ttl"`
	UpdatedAt time.Time `json:"updated_at"`
	Deleted   bool      `json:"deleted"` // True if key was deleted
}

// DTLSSync handles DTLS connection and synchronization
type DTLSSync struct {
	peerManager     *PeerManager
	store           *database.Store
	localPublicKey  ed25519.PublicKey
	localPrivateKey ed25519.PrivateKey
	connections     map[string]*dtls.Conn // peerID -> connection
	connMu          sync.RWMutex
}

// NewDTLSSync creates a new DTLS synchronization handler
func NewDTLSSync(pm *PeerManager, store *database.Store, pubkey ed25519.PublicKey, privkey ed25519.PrivateKey) *DTLSSync {
	return &DTLSSync{
		peerManager:     pm,
		store:           store,
		localPublicKey:  pubkey,
		localPrivateKey: privkey,
		connections:     make(map[string]*dtls.Conn),
	}
}

// SyncWithPeer establishes a DTLS connection to a peer and syncs delta keys
func (d *DTLSSync) SyncWithPeer(peer Peer) error {
	// Check if already connected
	d.connMu.RLock()
	_, exists := d.connections[peer.ID]
	d.connMu.RUnlock()

	if exists {
		return d.sendDeltaSync(peer.ID)
	}

	// Establish DTLS connection
	conn, err := d.establishDTLSConnection(&peer)
	if err != nil {
		return fmt.Errorf("failed to establish DTLS connection to %s: %w", peer.DeviceName, err)
	}

	d.connMu.Lock()
	d.connections[peer.ID] = conn
	d.connMu.Unlock()

	// Perform initial key sync
	if err := d.sendDeltaSync(peer.ID); err != nil {
		conn.Close()
		d.connMu.Lock()
		delete(d.connections, peer.ID)
		d.connMu.Unlock()
		return fmt.Errorf("failed to sync with peer: %w", err)
	}

	// Update peer trust level on successful sync.
	if err := d.peerManager.UpdatePeerTrustLevel(peer.ID, 5); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to update peer trust level: %v\n", err)
	}

	return nil
}

// establishDTLSConnection creates a DTLS connection to a peer
func (d *DTLSSync) establishDTLSConnection(peer *Peer) (*dtls.Conn, error) {
	if len(peer.Addresses) == 0 {
		return nil, fmt.Errorf("peer has no addresses")
	}

	// Use first available address
	addr := net.UDPAddr{
		Port: peer.Port,
		IP:   peer.Addresses[0],
	}

	// DTLS configuration
	config := &dtls.Config{
		Certificates:       []tls.Certificate{}, // In production, use proper certificates
		ClientAuth:         dtls.NoClientCert,
		InsecureSkipVerify: false,
		PSK: func(_ []byte) ([]byte, error) {
			// Use pre-shared key derived from peer's public key
			return deriveSharedSecret(d.localPublicKey, peer.PublicKey), nil
		},
		PSKIdentityHint: []byte(hashPublicKey(d.localPublicKey)),
	}

	conn, err := dtls.Dial("udp", &addr, config)
	if err != nil {
		return nil, fmt.Errorf("DTLS dial failed: %w", err)
	}

	return conn, nil
}

// sendDeltaSync computes delta and sends to peer
func (d *DTLSSync) sendDeltaSync(peerID string) error {
	// Get all current keys from store
	allKeys, err := d.store.List()
	if err != nil {
		return fmt.Errorf("failed to list keys: %w", err)
	}

	// Convert to KeyDelta
	deltas := make([]KeyDelta, 0)
	for key, entry := range allKeys {
		deltas = append(deltas, KeyDelta{
			Key:       key,
			Value:     entry.Value,
			Tags:      entry.Tags,
			TTL:       entry.TTL,
			UpdatedAt: entry.UpdatedAt,
			Deleted:   false,
		})
	}

	// Create sync message
	msg := Message{
		Type:      "key_delta",
		Timestamp: time.Now(),
		SenderID:  hashPublicKey(d.localPublicKey),
		Keys:      deltas,
	}

	// Encrypt with Signal Protocol
	encryptedPayload, err := d.encryptMessageSignalProtocol(peerID, msg)
	if err != nil {
		return fmt.Errorf("failed to encrypt message: %w", err)
	}

	msg.EncryptedPayload = encryptedPayload
	msg.Keys = nil // Clear plaintext keys

	// Sign the message with local private key
	msg.Signature, err = signMessage(d.localPrivateKey, msg)
	if err != nil {
		return fmt.Errorf("failed to sign message: %w", err)
	}

	// Send through DTLS connection
	return d.sendMessage(peerID, msg)
}

// encryptMessageSignalProtocol encrypts a message using Signal Protocol
func (d *DTLSSync) encryptMessageSignalProtocol(peerID string, msg Message) ([]byte, error) {
	// Serialize message
	plaintext, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	// In production, use proper Signal Protocol implementation
	// For now, use AES-256-GCM as a placeholder key derivation.
	key := make([]byte, 32)
	peerBytes := []byte(peerID)
	if len(peerBytes) >= len(key) {
		copy(key, peerBytes[:len(key)])
	} else {
		copy(key, peerBytes)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decryptMessageSignalProtocol decrypts a message encrypted with Signal Protocol
func (d *DTLSSync) decryptMessageSignalProtocol(peerID string, ciphertext []byte) ([]byte, error) {
	key := make([]byte, 32)
	peerBytes := []byte(peerID)
	if len(peerBytes) >= len(key) {
		copy(key, peerBytes[:len(key)])
	} else {
		copy(key, peerBytes)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce := ciphertext[:nonceSize]
	plaintext, err := gcm.Open(nil, nonce, ciphertext[nonceSize:], nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// sendMessage sends a message through DTLS connection
func (d *DTLSSync) sendMessage(peerID string, msg Message) error {
	d.connMu.RLock()
	conn, exists := d.connections[peerID]
	d.connMu.RUnlock()

	if !exists {
		return fmt.Errorf("no connection to peer %s", peerID)
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	_, err = conn.Write(data)
	return err
}

// ReceiveMessage receives and processes messages from a peer
func (d *DTLSSync) ReceiveMessage(peerID string, msgData []byte) error {
	var msg Message
	if err := json.Unmarshal(msgData, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// Verify signature
	if !d.verifyMessageSignature(peerID, msg.SenderID, msg.Signature, msg) {
		return fmt.Errorf("invalid message signature from peer %s", peerID)
	}

	// Decrypt payload if present
	if len(msg.EncryptedPayload) > 0 {
		plaintext, err := d.decryptMessageSignalProtocol(peerID, msg.EncryptedPayload)
		if err != nil {
			return fmt.Errorf("failed to decrypt message: %w", err)
		}

		if err := json.Unmarshal(plaintext, &msg); err != nil {
			return fmt.Errorf("failed to unmarshal decrypted message: %w", err)
		}
	}

	// Process based on message type
	switch msg.Type {
	case "key_delta":
		return d.processDeltaSync(peerID, msg.Keys)
	case "ack":
		return nil // Just acknowledge
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// processDeltaSync applies received key deltas to local store
func (d *DTLSSync) processDeltaSync(_ string, deltas []KeyDelta) error {
	for _, delta := range deltas {
		if delta.Deleted {
			if err := d.store.Delete(delta.Key); err != nil {
				return err
			}
			continue
		}

		if err := d.store.Put(delta.Key, delta.Value, delta.TTL, delta.Tags); err != nil {
			return err
		}
	}

	return nil
}

// CloseConnection closes the connection to a peer
func (d *DTLSSync) CloseConnection(peerID string) error {
	d.connMu.Lock()
	defer d.connMu.Unlock()

	conn, exists := d.connections[peerID]
	if !exists {
		return fmt.Errorf("no connection to peer %s", peerID)
	}

	delete(d.connections, peerID)
	return conn.Close()
}

func signedMessageBytes(msg Message) ([]byte, error) {
	copyMsg := msg
	copyMsg.Signature = nil
	return json.Marshal(copyMsg)
}

// deriveSharedSecret derives a pre-shared key from two Ed25519 public keys
func deriveSharedSecret(key1, key2 ed25519.PublicKey) []byte {
	if len(key1) == 0 || len(key2) == 0 {
		return make([]byte, 32)
	}

	combined := append(key1, key2...)
	hash := sha256.Sum256(combined)
	return hash[:]
}

// signMessage signs a message with an Ed25519 private key
func signMessage(priv ed25519.PrivateKey, msg Message) ([]byte, error) {
	msgBytes, err := signedMessageBytes(msg)
	if err != nil {
		return nil, err
	}

	return ed25519.Sign(priv, msgBytes), nil
}

// verifyMessageSignature verifies a message signature for a peer
func (d *DTLSSync) verifyMessageSignature(peerID string, senderID string, signature []byte, msg Message) bool {
	if senderID != peerID {
		return false
	}

	peer, err := d.peerManager.GetPeer(peerID)
	if err != nil || len(peer.PublicKey) == 0 {
		return false
	}

	msgBytes, err := signedMessageBytes(msg)
	if err != nil {
		return false
	}

	return ed25519.Verify(peer.PublicKey, msgBytes, signature)
}
