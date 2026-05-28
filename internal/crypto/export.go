package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"io"
)

type ExportEncryptor struct {
	masterKey []byte // Derivada de passphrase via Argon2
}

func (e *ExportEncryptor) EncryptExport(data map[string]string) ([]byte, error) {
	block, _ := aes.NewCipher(e.masterKey)
	gcm, _ := cipher.NewGCM(block)

	nonce := make([]byte, gcm.NonceSize())
	io.ReadFull(rand.Reader, nonce)

	plaintext, _ := json.Marshal(data)
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	return ciphertext, nil
}

func (e *ExportEncryptor) DecryptExport(data []byte) (map[string]string, error) {
	block, _ := aes.NewCipher(e.masterKey)
	gcm, _ := cipher.NewGCM(block)

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, _ := gcm.Open(nil, nonce, ciphertext, nil)

	var result map[string]string
	json.Unmarshal(plaintext, &result)

	return result, nil
}
