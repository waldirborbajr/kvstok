// internal/security/encrypt.go
package security

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Value representa um valor que será armazenado de forma segura
type Value struct {
	Data []byte `json:"data"`
}

// SecureEncrypt is a high-level wrapper for secure encryption usage across the project
type SecureEncrypt struct {
	master *MasterKey
}

// NewSecureEncrypt cria uma nova instância
func NewSecureEncrypt() *SecureEncrypt {
	return &SecureEncrypt{
		master: GetMasterKey(),
	}
}

// EncryptString encrypts a string value.
func (s *SecureEncrypt) EncryptString(plaintext string) ([]byte, error) {
	if plaintext == "" {
		return nil, errors.New("value cannot be empty")
	}
	return s.master.Encrypt([]byte(plaintext))
}

// DecryptString decrypts ciphertext and returns the plain string
func (s *SecureEncrypt) DecryptString(ciphertext []byte) (string, error) {
	plaintext, err := s.master.Decrypt(ciphertext)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// EncryptJSON encrypts any struct or map as JSON
func (s *SecureEncrypt) EncryptJSON(v any) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize JSON: %w", err)
	}
	return s.master.Encrypt(data)
}

// DecryptJSON decrypts ciphertext and unmarshals into the struct
func (s *SecureEncrypt) DecryptJSON(ciphertext []byte, v any) error {
	plaintext, err := s.master.Decrypt(ciphertext)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(plaintext, v); err != nil {
		return fmt.Errorf("failed to deserialize JSON: %w", err)
	}
	return nil
}

// EncryptBytes encrypts raw bytes
func (s *SecureEncrypt) EncryptBytes(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("data cannot be empty")
	}
	return s.master.Encrypt(data)
}

// DecryptBytes decrypts ciphertext and returns bytes
func (s *SecureEncrypt) DecryptBytes(ciphertext []byte) ([]byte, error) {
	return s.master.Decrypt(ciphertext)
}

// IsMasterPasswordSet verifica se já existe uma master password configurada
func (s *SecureEncrypt) IsMasterPasswordSet() bool {
	m := GetMasterKey()
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.key != nil
}

// RequireMasterPassword checks whether the master password is set and returns a user-friendly error
func (s *SecureEncrypt) RequireMasterPassword() error {
	if !s.IsMasterPasswordSet() {
		return errors.New("master password is not configured. Run: kvstok init")
	}
	return nil
}
