package security

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
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

// GenerateEd25519Key generates a new Ed25519 key pair.
func GenerateEd25519Key() (ed25519.PublicKey, ed25519.PrivateKey) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	must.Must(err, "GenerateEd25519Key() - generating key pair.")
	return publicKey, privateKey
}

// PrivateKeyToBytes converts an Ed25519 private key to PEM bytes.
func PrivateKeyToBytes(priv ed25519.PrivateKey) []byte {
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	must.Must(err, "PrivateKeyToBytes() - converting private key to bytes.")

	return pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	})
}

// PublicKeyToBytes converts an Ed25519 public key to PEM bytes.
func PublicKeyToBytes(pub ed25519.PublicKey) []byte {
	pubASN1, err := x509.MarshalPKIXPublicKey(pub)
	must.Must(err, "PublicKeyToBytes() - converting public key to bytes.")

	return pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	})
}

// BytesToPrivateKey converts PEM bytes to an Ed25519 private key.
func BytesToPrivateKey(priv []byte) ed25519.PrivateKey {
	block, _ := pem.Decode(priv)
	if block == nil {
		log.Fatal("failed to decode private key PEM")
	}

	keyIfc, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	must.Must(err, "BytesToPrivateKey()[ParsePKCS8PrivateKey] - converting private key")
	key, ok := keyIfc.(ed25519.PrivateKey)
	if !ok {
		log.Fatal("not ok")
	}

	return key
}

// BytesToPublicKey converts PEM bytes to an Ed25519 public key.
func BytesToPublicKey(pub []byte) ed25519.PublicKey {
	block, _ := pem.Decode(pub)
	if block == nil {
		log.Fatal("failed to decode public key PEM")
	}

	ifc, err := x509.ParsePKIXPublicKey(block.Bytes)
	must.Must(err, "BytesToPublicKey()[ParsePKIXPublicKey] - converting to public key.")
	key, ok := ifc.(ed25519.PublicKey)
	if !ok {
		log.Fatal("not ok")
	}

	return key
}
