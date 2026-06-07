// internal/security/migration.go
// RSA to Ed25519 migration - key rotation only
// (Data re-encryption happens in database package)

package security

import (
	"fmt"
	"log"
	"os"
)

// MigrationStatus tracks the outcome of key migration
type MigrationStatus struct {
	Migrated      bool
	KeysGenerated bool
	BackupDir     string
	Error         error
}

// PerformRSAtoEd25519Migration detects old RSA keys and migrates to Ed25519
// Only handles KEY migration - data re-encryption is delegated to caller
// This runs automatically on first startup if old RSA keys are detected
func PerformRSAtoEd25519Migration(pubKeyPath, privKeyPath string) (*MigrationStatus, error) {
	status := &MigrationStatus{}

	// Step 1: Check if migration is needed
	pubBytes, privBytes, err := readKeyFiles(pubKeyPath, privKeyPath)
	if err != nil {
		return status, fmt.Errorf("failed to read key files: %w", err)
	}

	// No keys exist = fresh install, no migration needed
	if len(pubBytes) == 0 && len(privBytes) == 0 {
		return status, nil
	}

	// Already Ed25519 = nothing to do
	if isEd25519PublicKey(pubBytes) && isEd25519PrivateKey(privBytes) {
		return status, nil
	}

	// Not RSA and not Ed25519 = error state
	if !isRSAPublicKey(pubBytes) || !isRSAPrivateKey(privBytes) {
		return status, fmt.Errorf("existing keys are in unsupported format")
	}

	// Step 2: Backup existing key files
	log.Println("🔄 Detected RSA keys. Starting migration to Ed25519...")

	if err := backupFile(pubKeyPath); err != nil {
		return status, fmt.Errorf("failed to backup public key: %w", err)
	}
	if err := backupFile(privKeyPath); err != nil {
		return status, fmt.Errorf("failed to backup private key: %w", err)
	}

	status.BackupDir = pubKeyPath + ".old"
	log.Printf("✅ Keys backed up to: %s\n", status.BackupDir)

	// Step 3: Generate new Ed25519 keys
	publicKey, privateKey := GenerateEd25519Key()
	if err := os.WriteFile(pubKeyPath, PublicKeyToBytes(publicKey), 0600); err != nil {
		return status, fmt.Errorf("failed to write new public key: %w", err)
	}
	if err := os.WriteFile(privKeyPath, PrivateKeyToBytes(privateKey), 0600); err != nil {
		return status, fmt.Errorf("failed to write new private key: %w", err)
	}

	status.KeysGenerated = true
	status.Migrated = true
	log.Println("✅ New Ed25519 keys generated")

	return status, nil

// readKeyFiles safely reads both key files
func readKeyFiles(pubPath, privPath string) ([]byte, []byte, error) {
	var pubBytes, privBytes []byte

	if fileExists(pubPath) {
		var err error
		pubBytes, err = os.ReadFile(pubPath)
		if err != nil {
			return nil, nil, err
		}
	}

	if fileExists(privPath) {
		var err error
		privBytes, err = os.ReadFile(privPath)
		if err != nil {
			return nil, nil, err
		}
	}

	return pubBytes, privBytes, nil
}
