package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log"

	"github.com/waldirborbajr/kvstok/internal/must"
)

// CheckError logs the error message if the error is not nil.
func CheckError(e error) {
	if e != nil {
		fmt.Printf("%s", e.Error())
	}
}

// RSAGenerateKey generates a new RSA key pair with the given bit size.
func RSAGenerateKey(bits int) (*rsa.PrivateKey, *rsa.PublicKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	must.Must(err, "RSAGenerateKey() - generating key pair.")
	return privateKey, &privateKey.PublicKey
}

// PrivateKeyToBytes private key to bytes
func PrivateKeyToBytes(priv *rsa.PrivateKey) []byte {
	privBytes := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(priv),
		},
	)
	return privBytes
}

// PublicKeyToBytes public key to bytes
func PublicKeyToBytes(pub *rsa.PublicKey) []byte {
	pubASN1, err := x509.MarshalPKIXPublicKey(pub)
	must.Must(err, "PublicKeyToBytes() - converting public key to bytes.")
	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubASN1,
	})
	return pubBytes
}

// BytesToPrivateKey bytes to private key
func BytesToPrivateKey(priv []byte) *rsa.PrivateKey {
	block, _ := pem.Decode(priv)
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	must.Must(err, "BytesToPrivateKey()[ParsePKCS1PrivateKey] - converting private key")
	return key
}

// BytesToPublicKey bytes to public key
func BytesToPublicKey(pub []byte) *rsa.PublicKey {
	block, _ := pem.Decode(pub)
	ifc, err := x509.ParsePKIXPublicKey(block.Bytes)
	must.Must(err, "BytesToPublicKey()[ParsePKIXPublicKey] - converting to public key.")
	key, ok := ifc.(*rsa.PublicKey)
	if !ok {
		log.Fatal("not ok")
	}
	return key
}

// EncryptWithPublicKey encrypts data with public key
func EncryptWithPublicKey(msg []byte, pub *rsa.PublicKey) []byte {
	hash := sha512.New()
	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, pub, msg, nil)
	must.Must(err, "EncryptWithPublicKey() - encrypting with public key.")
	return ciphertext
}

// DecryptWithPrivateKey decrypts data with private key
func DecryptWithPrivateKey(ciphertext []byte, priv *rsa.PrivateKey) []byte {
	hash := sha512.New()
	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, priv, ciphertext, nil)
	must.Must(err, "DecryptWithPrivateKey() - decrypting with private key.")
	return plaintext
}

// RSAOAEPEncrypt encrypts a secret message using RSA-OAEP with SHA-256.
func RSAOAEPEncrypt(secretMessage string, key rsa.PublicKey) string {
	label := []byte("OAEP Encrypted")
	rng := rand.Reader
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rng, &key, []byte(secretMessage), label)
	CheckError(err)
	return base64.StdEncoding.EncodeToString(ciphertext)
}

// RSAOAEPDecrypt decrypts a base64-encoded ciphertext using RSA-OAEP with SHA-256.
func RSAOAEPDecrypt(cipherText string, privKey rsa.PrivateKey) string {
	ct, _ := base64.StdEncoding.DecodeString(cipherText)
	label := []byte("OAEP Encrypted")
	rng := rand.Reader
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rng, &privKey, ct, label)
	CheckError(err)
	return string(plaintext)
}
