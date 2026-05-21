// internal/security/master.go
package security

import (
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
)

const (
	// Configuração Argon2id (recomendada para senhas)
	argonTime    uint32 = 4
	argonMemory  uint32 = 64 * 1024 // 64 MB
	argonThreads uint8  = 4
	saltSize            = 16
	keySize             = 32 // 256 bits
)

var (
	ErrInvalidMasterPassword = errors.New("senha mestra inválida")
	ErrMasterNotSet          = errors.New("master password não foi configurada")
)

// MasterKey gerencia a chave derivada da senha mestra
type MasterKey struct {
	key  []byte
	salt []byte
	mu   sync.RWMutex
}

var masterInstance *MasterKey
var once sync.Once

// GetMasterKey retorna a instância singleton (thread-safe)
func GetMasterKey() *MasterKey {
	once.Do(func() {
		masterInstance = &MasterKey{}
	})
	return masterInstance
}

// SetMasterPassword define e deriva a chave da senha mestra
func (m *MasterKey) SetMasterPassword(password string) error {
	if password == "" {
		return errors.New("senha mestra não pode ser vazia")
	}

	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return err
	}

	key := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, keySize)

	m.mu.Lock()
	m.key = key
	m.salt = salt
	m.mu.Unlock()

	return nil
}

// VerifyMasterPassword verifica se a senha está correta
func (m *MasterKey) VerifyMasterPassword(password string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.key == nil || m.salt == nil {
		return ErrMasterNotSet
	}

	expectedKey := argon2.IDKey([]byte(password), m.salt, argonTime, argonMemory, argonThreads, keySize)

	if subtle.ConstantTimeCompare(m.key, expectedKey) != 1 {
		return ErrInvalidMasterPassword
	}
	return nil
}

// Encrypt criptografa dados usando XChaCha20-Poly1305
func (m *MasterKey) Encrypt(plaintext []byte) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.key == nil {
		return nil, ErrMasterNotSet
	}

	aead, err := chacha20poly1305.NewX(m.key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := aead.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt descriptografa os dados
func (m *MasterKey) Decrypt(ciphertext []byte) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.key == nil {
		return nil, ErrMasterNotSet
	}

	aead, err := chacha20poly1305.NewX(m.key)
	if err != nil {
		return nil, err
	}

	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext muito curto")
	}

	nonce := ciphertext[:nonceSize]
	ciphertext = ciphertext[nonceSize:]

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("falha ao descriptografar: senha incorreta ou dados corrompidos")
	}

	return plaintext, nil
}

// SaveSalt salva o salt em um arquivo (para persistir entre execuções)
func (m *MasterKey) SaveSalt(path string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.salt == nil {
		return ErrMasterNotSet
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	return os.WriteFile(path, m.salt, 0600)
}

// LoadSalt carrega o salt salvo
func (m *MasterKey) LoadSalt(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(data) != saltSize {
		return errors.New("salt inválido")
	}

	m.mu.Lock()
	m.salt = data
	m.mu.Unlock()

	return nil
}