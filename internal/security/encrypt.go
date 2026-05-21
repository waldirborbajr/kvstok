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

// SecureEncrypt é um wrapper de alto nível para facilitar o uso em todo o projeto
type SecureEncrypt struct {
	master *MasterKey
}

// NewSecureEncrypt cria uma nova instância
func NewSecureEncrypt() *SecureEncrypt {
	return &SecureEncrypt{
		master: GetMasterKey(),
	}
}

// EncryptString criptografa uma string (mais comum no seu caso)
func (s *SecureEncrypt) EncryptString(plaintext string) ([]byte, error) {
	if plaintext == "" {
		return nil, errors.New("valor não pode ser vazio")
	}
	return s.master.Encrypt([]byte(plaintext))
}

// DecryptString descriptografa e retorna string
func (s *SecureEncrypt) DecryptString(ciphertext []byte) (string, error) {
	plaintext, err := s.master.Decrypt(ciphertext)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// EncryptJSON criptografa qualquer struct/map como JSON
func (s *SecureEncrypt) EncryptJSON(v any) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("falha ao serializar JSON: %w", err)
	}
	return s.master.Encrypt(data)
}

// DecryptJSON descriptografa e converte para struct
func (s *SecureEncrypt) DecryptJSON(ciphertext []byte, v any) error {
	plaintext, err := s.master.Decrypt(ciphertext)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(plaintext, v); err != nil {
		return fmt.Errorf("falha ao deserializar JSON: %w", err)
	}
	return nil
}

// EncryptBytes criptografa bytes crus
func (s *SecureEncrypt) EncryptBytes(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("dados não podem estar vazios")
	}
	return s.master.Encrypt(data)
}

// DecryptBytes descriptografa e retorna bytes
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

// RequireMasterPassword verifica se a master password está definida, retorna erro amigável
func (s *SecureEncrypt) RequireMasterPassword() error {
	if !s.IsMasterPasswordSet() {
		return errors.New("master password não configurada. Execute: kvstok init")
	}
	return nil
}
