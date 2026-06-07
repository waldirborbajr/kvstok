package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"errors"

	"golang.org/x/crypto/argon2"
)

// ProtectPrivateKey encrypts an Ed25519 private key using a passphrase.
func ProtectPrivateKey(privateKey ed25519.PrivateKey, passphrase string) ([]byte, error) {
	if len(privateKey) == 0 {
		return nil, errors.New("private key cannot be nil")
	}

	if passphrase == "" {
		return nil, errors.New("passphrase cannot be empty")
	}

	// 1. Generate a random salt
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	// 2. Derive a key using Argon2id
	key := argon2.IDKey(
		[]byte(passphrase),
		salt,
		3, 64*1024, 4, 32,
	)

	// 3. Encrypt the private key with AES-256-GCM
	privDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, err
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

	ciphertext := gcm.Seal(nonce, nonce, privDER, nil)

	// 4. Combine salt + nonce + ciphertext
	return append(salt, append(nonce, ciphertext...)...), nil
}

// UnprotectPrivateKey decrypts an Ed25519 private key using a passphrase.
func UnprotectPrivateKey(protected []byte, passphrase string) (ed25519.PrivateKey, error) {
	if len(protected) < 16 {
		return nil, errors.New("protected data too short")
	}

	if passphrase == "" {
		return nil, errors.New("passphrase cannot be empty")
	}

	// 1. Extract the salt (first 16 bytes)
	salt := protected[:16]

	// 2. Derive the key using Argon2id with the same salt
	key := argon2.IDKey(
		[]byte(passphrase),
		salt,
		3, 64*1024, 4, 32,
	)

	// 3. Extract nonce and ciphertext
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(protected) < 16+nonceSize {
		return nil, errors.New("protected data too short")
	}

	nonce := protected[16 : 16+nonceSize]
	ciphertext := protected[16+nonceSize:]

	// 4. Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	// 5. Parse the PKCS8 private key
	privateKeyIfc, err := x509.ParsePKCS8PrivateKey(plaintext)
	if err != nil {
		return nil, err
	}

	privateKey, ok := privateKeyIfc.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("unexpected private key type")
	}

	return privateKey, nil
}
