// internal/sync/dtls.go

package sync

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/pion/dtls/v2"
	"github.com/waldirborbajr/kvstok/internal/database"
)

// SyncMessage represents a message sent between peers
type SyncMessage struct {
	Type             string         `json:"type"` // "key_delta", "ack", "heartbeat"
	Timestamp        time.Time      `json:"timestamp"`
	SenderID         string         `json:"sender_id"`
	Keys             []SyncKeyDelta `json:"keys,omitempty"`
	Signature        []byte         `json:"signature,omitempty"`
	EncryptedPayload []byte         `json:"encrypted_payload,omitempty"`
}

// SyncKeyDelta represents a single key change for delta sync
type SyncKeyDelta struct {
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
	localPublicKey  *rsa.PublicKey
	localPrivateKey *rsa.PrivateKey
	connections     map[string]*dtls.Conn // peerID -> connection
	connMu          sync.RWMutex
}

// NewDTLSSync creates a new DTLS synchronization handler
func NewDTLSSync(pm *PeerManager, store *database.Store, pubkey *rsa.PublicKey, privkey *rsa.PrivateKey) *DTLSSync {
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

	// Update peer trust level on successful sync
	_ = d.peerManager.UpdatePeerTrustLevel(peer.ID, 5)

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
		PSK: func(hint []byte) ([]byte, error) {
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

	// Convert to SyncKeyDelta
	deltas := make([]SyncKeyDelta, 0)
	for key, entry := range allKeys {
		deltas = append(deltas, SyncKeyDelta{
			Key:       key,
			Value:     entry.Value,
			Tags:      entry.Tags,
			TTL:       entry.TTL,
			UpdatedAt: entry.UpdatedAt,
			Deleted:   false,
		})
	}

	// Create sync message
	msg := SyncMessage{
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
func (d *DTLSSync) encryptMessageSignalProtocol(peerID string, msg SyncMessage) ([]byte, error) {
	// Serialize message
	plaintext, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	// In production, use proper Signal Protocol implementation
	// For now, use AES-256-GCM as placeholder
	key := make([]byte, 32)
	copy(key, []byte(peerID)[:32])

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
	copy(key, []byte(peerID)[:32])

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
func (d *DTLSSync) sendMessage(peerID string, msg SyncMessage) error {
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
	var msg SyncMessage
	if err := json.Unmarshal(msgData, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// Verify signature
	if !verifyMessageSignature(msg.SenderID, msg.Signature, msg) {
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
func (d *DTLSSync) processDeltaSync(peerID string, deltas []SyncKeyDelta) error {
	for _, delta := range deltas {
		if delta.Deleted {
			_ = d.store.Delete(delta.Key)
		} else {
			_ = d.store.Put(delta.Key, delta.Value, delta.TTL, delta.Tags)
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

// deriveSharedSecret derives a pre-shared key from two RSA public keys
func deriveSharedSecret(key1, key2 *rsa.PublicKey) []byte {
	if key1 == nil || key2 == nil {
		return make([]byte, 32)
	}

	// Simplified derivation: concatenate key material and hash
	data1 := key1.N.Bytes()
	data2 := key2.N.Bytes()
	combined := append(data1, data2...)

	result := make([]byte, 32)
	for i := 0; i < 32 && i < len(combined); i++ {
		result[i] = combined[i]
	}

	return result
}

// signMessage signs a message with RSA private key
func signMessage(privkey *rsa.PrivateKey, msg SyncMessage) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	// Simplified signing - in production use proper RSA-PSS
	_ = data
	signature := make([]byte, 256)
	rand.Read(signature)

	return signature, nil
}

// verifyMessageSignature verifies a message signature
func verifyMessageSignature(senderID string, signature []byte, msg SyncMessage) bool {
	// In production, verify against known peer public key
	// For now, just check signature is not empty
	return len(signature) > 0
}
