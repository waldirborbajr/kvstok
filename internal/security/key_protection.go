package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"

	"golang.org/x/crypto/argon2"
)

// internal/security/key_protection.go

func ProtectPrivateKey(privateKey *rsa.PrivateKey, passphrase string) []byte {
	// 1. Generate a random salt
	salt := make([]byte, 16)
	rand.Read(salt)

	// 2. Derive a key using Argon2id
	key := argon2.IDKey(
		[]byte(passphrase),
		salt,
		3, 64*1024, 4, 32,
	)

	// 3. Encrypt the private key with AES-256-GCM
	privPEM := x509.MarshalPKCS1PrivateKey(privateKey)

	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)

	nonce := make([]byte, gcm.NonceSize())
	rand.Read(nonce)

	ciphertext := gcm.Seal(nonce, nonce, privPEM, nil)

	// 4. Combine salt + nonce + ciphertext
	return append(salt, append(nonce, ciphertext...)...)
}

func UnprotectPrivateKey(protected []byte, passphrase string) *rsa.PrivateKey {
	// 1. Extract the salt (first 16 bytes)
	salt := protected[:16]

	// 2. Derive the key using Argon2id with the same salt
	key := argon2.IDKey(
		[]byte(passphrase),
		salt,
		3, 64*1024, 4, 32,
	)

	// 3. Extract nonce and ciphertext
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)
	nonceSize := gcm.NonceSize()

	nonce := protected[16 : 16+nonceSize]
	ciphertext := protected[16+nonceSize:]

	// 4. Decrypt
	plaintext, _ := gcm.Open(nil, nonce, ciphertext, nil)

	// 5. Parse the PKCS1 private key
	privateKey, _ := x509.ParsePKCS1PrivateKey(plaintext)

	return privateKey
}
