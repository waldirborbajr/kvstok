// internal/database/migration.go
// Data re-encryption during RSA→Ed25519 migration
// This lives in database package to avoid circular imports

package database

import (
	"fmt"
	"log"

	"github.com/waldirborbajr/kvstok/internal/security"
)

// MigrationResult tracks data re-encryption outcome
type MigrationResult struct {
	Migrated int
	Failed   int
}

// MigrateDataRSAtoEd25519 re-encrypts all database values from RSA to Ed25519
// Call this AFTER security.PerformRSAtoEd25519Migration() completes key migration
// Must be called with valid master password already set
func MigrateDataRSAtoEd25519(db *DB) (*MigrationResult, error) {
	result := &MigrationResult{}

	// Verify encryption is available
	se := security.NewSecureEncrypt()
	if !se.IsMasterPasswordSet() {
		return result, fmt.Errorf("master password required for data migration")
	}

	// Get all entries - your implementation below
	entries, err := db.getAllStoredKeys()
	if err != nil {
		return result, fmt.Errorf("failed to read database entries: %w", err)
	}

	if len(entries) == 0 {
		log.Println("✅ No data to migrate (fresh database)")
		return result, nil
	}

	log.Printf("🔄 Re-encrypting %d records...\n", len(entries))

	// Re-encrypt each entry
	for _, entry := range entries {
		plaintext, err := decryptWithLegacyRSAKey(entry.Value)
		if err != nil {
			log.Printf("⚠️  Failed to decrypt key '%s': %v\n", entry.Key, err)
			result.Failed++
			continue
		}

		// Re-encrypt with new Ed25519 key
		newEncrypted, err := se.EncryptBytes(plaintext)
		if err != nil {
			log.Printf("⚠️  Failed to re-encrypt key '%s': %v\n", entry.Key, err)
			result.Failed++
			continue
		}

		// Update in database
		if err := db.updateStoredValue(entry.Key, newEncrypted); err != nil {
			log.Printf("⚠️  Failed to update key '%s': %v\n", entry.Key, err)
			result.Failed++
			continue
		}

		result.Migrated++
	}

	if result.Failed > 0 {
		log.Printf("⚠️  Migration partial: %d succeeded, %d failed\n", result.Migrated, result.Failed)
		return result, fmt.Errorf("migration completed with %d failures", result.Failed)
	}

	log.Printf("✅ Data migration complete: %d records re-encrypted\n", result.Migrated)
	return result, nil
}

// ============================================================================
// IMPLEMENTATION: Add these to your existing internal/database/db.go
// ============================================================================

// StoredEntry represents a key-value pair in the database
type StoredEntry struct {
	Key   string
	Value []byte // encrypted
}

// getAllStoredKeys returns all keys and their encrypted values
// IMPLEMENT THIS based on your NutsDB structure:
func (db *DB) getAllStoredKeys() ([]StoredEntry, error) {
	// Example implementation (adjust bucket name and structure to match yours):
	/*
	var entries []StoredEntry

	err := db.nuts.View(func(tx *nutsdb.Tx) error {
		bucket := tx.GetBucket("secrets") // What's your bucket name?
		if bucket == nil {
			return nil
		}
		return bucket.ForEach(func(key, value []byte) error {
			entries = append(entries, StoredEntry{
				Key:   string(key),
				Value: append([]byte{}, value...), // Copy value
			})
			return nil
		})
	})
	return entries, err
	*/
	return nil, fmt.Errorf("implement based on your DB bucket structure")
}

// updateStoredValue updates a single encrypted value in the database
// IMPLEMENT THIS based on your NutsDB structure:
func (db *DB) updateStoredValue(key string, newEncrypted []byte) error {
	// Example implementation:
	/*
	return db.nuts.Update(func(tx *nutsdb.Tx) error {
		bucket := tx.GetBucket("secrets") // Same bucket name as above
		return bucket.Put([]byte(key), newEncrypted)
	})
	*/
	return fmt.Errorf("implement based on your DB bucket structure")
}

// ============================================================================
// RSA Decryption: Implement this helper in your database or security package
// ============================================================================

// decryptWithLegacyRSAKey decrypts data encrypted with the old RSA key
// YOU IMPLEMENT THIS - examples below
func decryptWithLegacyRSAKey(ciphertext []byte) ([]byte, error) {
	// Option 1: If you have an RSADecrypt function in security package:
	// return security.DecryptRSALegacy(ciphertext)

	// Option 2: If RSA decrypt is in database package (where it belongs):
	// rsaKey := loadRSAPrivateKeyFromBackup()
	// return rsaKey.Decrypt(ciphertext)

	// Option 3: Call your existing decrypt function by a different name

	return nil, fmt.Errorf("implement RSA decryption - load backup key and decrypt")
}

// Example: How to load the RSA backup key (you write this):
/*
func loadRSAPrivateKeyFromBackup(backupPath string) (*rsa.PrivateKey, error) {
	// Read the .old backup file
	keyBytes, err := os.ReadFile(backupPath + ".old")
	if err != nil {
		return nil, err
	}

	// Parse PEM
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM")
	}

	// Parse PKCS1 private key
	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privKey, nil
}
*/
