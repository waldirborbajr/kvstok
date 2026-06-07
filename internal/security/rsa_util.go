package security

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"

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

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func backupFile(path string) error {
	if !fileExists(path) {
		return nil
	}

	backupPath := path + ".old"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return os.Rename(path, backupPath)
	}

	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s.old%d", path, i)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return os.Rename(path, candidate)
		}
	}
}

func isRSAPublicKey(pemBytes []byte) bool {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return false
	}

	ifc, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false
	}

	_, ok := ifc.(*rsa.PublicKey)
	return ok
}

func isRSAPrivateKey(pemBytes []byte) bool {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return false
	}

	_, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	return err == nil
}

func isEd25519PublicKey(pemBytes []byte) bool {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return false
	}

	ifc, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false
	}

	_, ok := ifc.(ed25519.PublicKey)
	return ok
}

func isEd25519PrivateKey(pemBytes []byte) bool {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return false
	}

	keyIfc, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return false
	}

	_, ok := keyIfc.(ed25519.PrivateKey)
	return ok
}

// MigrateRSAKeysToEd25519 migrates legacy RSA key files to Ed25519.
// Legacy RSA files are backed up before a new Ed25519 key pair is generated.
func MigrateRSAKeysToEd25519(pubPath, privPath string) (bool, error) {
	pubExists := fileExists(pubPath)
	privExists := fileExists(privPath)

	if !pubExists && !privExists {
		return false, nil
	}

	var pubBytes, privBytes []byte
	var err error

	if pubExists {
		pubBytes, err = os.ReadFile(pubPath)
		if err != nil {
			return false, fmt.Errorf("failed to read public key file: %w", err)
		}
	}

	if privExists {
		privBytes, err = os.ReadFile(privPath)
		if err != nil {
			return false, fmt.Errorf("failed to read private key file: %w", err)
		}
	}

	if pubExists && privExists && isEd25519PublicKey(pubBytes) && isEd25519PrivateKey(privBytes) {
		return false, nil
	}

	if (pubExists && !isRSAPublicKey(pubBytes) && !isEd25519PublicKey(pubBytes)) ||
		(privExists && !isRSAPrivateKey(privBytes) && !isEd25519PrivateKey(privBytes)) {
		return false, fmt.Errorf("existing key files are in an unsupported format")
	}

	if err := backupFile(pubPath); err != nil {
		return false, err
	}
	if err := backupFile(privPath); err != nil {
		return false, err
	}

	publicKey, privateKey := GenerateEd25519Key()

	if err := os.WriteFile(pubPath, PublicKeyToBytes(publicKey), 0600); err != nil {
		return false, err
	}
	if err := os.WriteFile(privPath, PrivateKeyToBytes(privateKey), 0600); err != nil {
		return false, err
	}

	return true, nil
}
