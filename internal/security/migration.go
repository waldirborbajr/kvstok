// internal/security/migration.go
// RSA to Ed25519 migration for existing kvstok data

package security

import (
	"fmt"
	"log"
	"os"

	"github.com/waldirborbajr/kvstok/internal/database"
)

// MigrationStatus tracks the outcome of migration
type MigrationStatus struct {
	Migrated      bool
	KeysGenerated bool
	DataRestored  int
	DataFailed    int
	BackupDir     string
	Error         error
}

// PerformRSAtoEd25519Migration detects old RSA keys and migrates everything to Ed25519
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
	log.Println("✅ New Ed25519 keys generated")

	// Step 4: Re-encrypt all database values
	// This requires the master password to be set
	se := NewSecureEncrypt()
	if !se.IsMasterPasswordSet() {
		return status, fmt.Errorf("master password required for data migration")
	}

	db := database.GetDB()
	if db == nil {
		return status, fmt.Errorf("database not initialized")
	}

	migratedCount, failedCount, err := reEncryptAllData(db, se)
	if err != nil {
		log.Printf("⚠️  Data migration encountered errors: %v\n", err)
		status.DataFailed = failedCount
	}

	status.DataRestored = migratedCount
	status.Migrated = true

	if failedCount > 0 {
		log.Printf("⚠️  Migration partial: %d succeeded, %d failed\n", migratedCount, failedCount)
		return status, fmt.Errorf("migration completed with %d failures", failedCount)
	}

	log.Printf("✅ Migration complete: %d records re-encrypted\n", migratedCount)
	return status, nil
}

// reEncryptAllData iterates through all database entries and re-encrypts them
// This uses the newly loaded MasterKey (Ed25519 based)
func reEncryptAllData(db *database.DB, se *SecureEncrypt) (int, int, error) {
	succeeded := 0
	failed := 0

	// Iterate through all keys in the database
	// This depends on your database structure - adjust Bucket/Prefix as needed
	entries, err := db.GetAllEntries()
	if err != nil {
		return succeeded, failed, fmt.Errorf("failed to read database entries: %w", err)
	}

	log.Printf("🔄 Re-encrypting %d records...\n", len(entries))

	for _, entry := range entries {
		// Each entry contains: key (string) and encrypted value ([]byte)
		// We need to:
		// 1. Decrypt with old MasterKey (RSA)
		// 2. Re-encrypt with new MasterKey (Ed25519)

		plaintext, err := decryptWithLegacyRSAKey(entry.EncryptedValue)
		if err != nil {
			log.Printf("⚠️  Failed to decrypt key '%s': %v\n", entry.Key, err)
			failed++
			continue
		}

		// Re-encrypt with new Ed25519 key (now active in MasterKey)
		newEncrypted, err := se.EncryptBytes(plaintext)
		if err != nil {
			log.Printf("⚠️  Failed to re-encrypt key '%s': %v\n", entry.Key, err)
			failed++
			continue
		}

		// Write back to database
		if err := db.UpdateEntry(entry.Key, newEncrypted); err != nil {
			log.Printf("⚠️  Failed to update key '%s': %v\n", entry.Key, err)
			failed++
			continue
		}

		succeeded++
	}

	return succeeded, failed, nil
}

// decryptWithLegacyRSAKey decrypts data that was encrypted with the old RSA key
// IMPORTANT: This assumes the RSA private key is still available in memory
// You'll need to implement this based on how RSA was originally used
func decryptWithLegacyRSAKey(ciphertext []byte) ([]byte, error) {
	// This function depends on your RSA implementation
	// You may need to:
	// 1. Load the RSA private key from backup
	// 2. Decrypt the ciphertext
	// 3. Return the plaintext

	// Placeholder - implement based on your RSA crypto logic
	// Example pseudo-code:
	// rsaKey := LoadRSAPrivateKeyFromBackup()
	// return rsaKey.Decrypt(ciphertext)

	return nil, fmt.Errorf("RSA decryption not yet implemented - add your RSA decrypt logic here")
}

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
