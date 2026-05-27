package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"

	"golang.org/x/crypto/argon2"
)

// internal/security/key_protection.go - NOVO

func ProtectPrivateKey(privateKey *rsa.PrivateKey, passphrase string) []byte {
	// 1. Gerar salt aleatório
	salt := make([]byte, 16)
	rand.Read(salt)

	// 2. Derivar key usando Argon2id
	key := argon2.IDKey(
		[]byte(passphrase),
		salt,
		3, 64*1024, 4, 32,
	)

	// 3. Encriptar private key com AES-256-GCM
	privPEM := x509.MarshalPKCS1PrivateKey(privateKey)

	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)

	nonce := make([]byte, gcm.NonceSize())
	rand.Read(nonce)

	ciphertext := gcm.Seal(nonce, nonce, privPEM, nil)

	// 4. Combinar salt + nonce + ciphertext
	return append(salt, append(nonce, ciphertext...)...)
}

func UnprotectPrivateKey(protected []byte, passphrase string) *rsa.PrivateKey {
	// 1. Extrair salt (primeiros 16 bytes)
	salt := protected[:16]

	// 2. Derivar key usando Argon2id com o mesmo salt
	key := argon2.IDKey(
		[]byte(passphrase),
		salt,
		3, 64*1024, 4, 32,
	)

	// 3. Extrair nonce e ciphertext
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)
	nonceSize := gcm.NonceSize()

	nonce := protected[16 : 16+nonceSize]
	ciphertext := protected[16+nonceSize:]

	// 4. Descriptografar
	plaintext, _ := gcm.Open(nil, nonce, ciphertext, nil)

	// 5. Desserializar PKCS1 private key
	privateKey, _ := x509.ParsePKCS1PrivateKey(plaintext)

	return privateKey
}
