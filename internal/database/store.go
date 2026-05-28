// internal/database/store.go
package database

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/nutsdb/nutsdb"
	"github.com/waldirborbajr/kvstok/internal/security"
)

const (
	DBName  = ".6B7673"
	Bucket  = "kvstok"
	SaltKey = "master_salt"
)

var (
	DB             *nutsdb.DB
	ErrKeyNotFound = errors.New("key not found")
	ErrKeyExpired  = errors.New("key expired")
)

// Store is the main encrypted database access layer
type Store struct {
	db     *nutsdb.DB
	sec    *security.SecureEncrypt
	dbPath string
	mu     sync.RWMutex
}

// NewStore cria uma nova instância do Store
func defaultDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "kvstok", DBName), nil
}

func NewStore(path string) (*Store, error) {
	if path == "" {
		var err error
		path, err = defaultDBPath()
		if err != nil {
			return nil, err
		}
	}

	if err := os.MkdirAll(path, 0700); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	opts := nutsdb.DefaultOptions
	opts.Dir = path
	opts.EntryMaxSize = 1024 * 1024 * 8 // 8MB

	db, err := nutsdb.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	DB = db

	s := &Store{
		db:     db,
		sec:    security.NewSecureEncrypt(),
		dbPath: path,
	}

	return s, nil
}

// Close closes the database
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Close()
}

// SetMasterPassword initializes or derives the master key using the loaded salt.
func (s *Store) SetMasterPassword(password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.sec.GetMasterKey().SetMasterPassword(password); err != nil {
		return err
	}

	// Save the salt for future sessions
	saltPath := filepath.Join(s.dbPath, "master.salt")
	return s.sec.GetMasterKey().SaveSalt(saltPath)
}

// LoadMasterSalt loads the saved salt for the master password.
func (s *Store) LoadMasterSalt() error {
	saltPath := filepath.Join(s.dbPath, "master.salt")
	return s.sec.GetMasterKey().LoadSalt(saltPath)
}

func (s *Store) IsMasterPasswordSet() bool {
	saltPath := filepath.Join(s.dbPath, "master.salt")
	_, err := os.Stat(saltPath)
	return err == nil
}

func (s *Store) DB() *nutsdb.DB {
	return s.db
}

// Put inserts or updates a key with encryption
func (s *Store) Put(key string, value string, ttl uint32, tags []string) error {
	if err := s.sec.RequireMasterPassword(); err != nil {
		return err
	}

	entry := SecretEntry{
		Value:     value,
		TTL:       ttl,
		Tags:      tags,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Encrypt the value
	encrypted, err := s.sec.EncryptString(value)
	if err != nil {
		return err
	}
	entry.Value = string(encrypted) // armazenamos como string base64 ou raw

	data, err := s.sec.EncryptJSON(entry)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	err = s.db.Update(func(tx *nutsdb.Tx) error {
		return tx.Put(Bucket, []byte(key), data, ttl)
	})

	return err
}

// Get returns the decrypted value
func (s *Store) Get(key string) (string, error) {
	if err := s.sec.RequireMasterPassword(); err != nil {
		return "", err
	}

	var data []byte

	s.mu.RLock()
	err := s.db.View(func(tx *nutsdb.Tx) error {
		val, err := tx.Get(Bucket, []byte(key))
		if err != nil {
			return err
		}
		data = append([]byte{}, val...) // copia
		return nil
	})
	s.mu.RUnlock()

	if err != nil {
		if err == nutsdb.ErrKeyNotFound {
			return "", ErrKeyNotFound
		}
		return "", err
	}

	// Decrypt the entry
	var entry SecretEntry
	if err := s.sec.DecryptJSON(data, &entry); err != nil {
		return "", err
	}

	// Verifica TTL manualmente se necessário
	if entry.TTL > 0 && time.Since(entry.UpdatedAt) > time.Duration(entry.TTL)*time.Second {
		return "", ErrKeyExpired
	}

	return entry.Value, nil
}

// GetRaw retorna valor + metadados
func (s *Store) GetRaw(key string) (value string, entry *SecretEntry, err error) {
	if err = s.sec.RequireMasterPassword(); err != nil {
		return "", nil, err
	}

	var data []byte

	s.mu.RLock()
	err = s.db.View(func(tx *nutsdb.Tx) error {
		data, err = tx.Get(Bucket, []byte(key))
		return err
	})
	s.mu.RUnlock()

	if err != nil {
		if err == nutsdb.ErrKeyNotFound {
			return "", nil, ErrKeyNotFound
		}
		return "", nil, err
	}

	var se SecretEntry
	if err = s.sec.DecryptJSON(data, &se); err != nil {
		return "", nil, fmt.Errorf("decryption failed: %w", err)
	}

	// Check TTL
	if se.TTL > 0 && time.Since(se.UpdatedAt) > time.Duration(se.TTL)*time.Second {
		_ = s.Delete(key) // remove expired key
	}

	return se.Value, &se, nil
}

// ListAll returns all existing keys (key names only, without decryption)
func (s *Store) ListAll() ([]string, error) {
	if err := s.sec.RequireMasterPassword(); err != nil {
		return nil, err
	}

	var keys []string

	s.mu.RLock()
	err := s.db.View(func(tx *nutsdb.Tx) error {
		return tx.ForEach(Bucket, func(key, value []byte) bool {
			keys = append(keys, string(key))
			return true
		})
	})
	s.mu.RUnlock()

	if err != nil {
		if err == nutsdb.ErrBucketNotFound {
			return []string{}, nil
		}
		return nil, err
	}

	return keys, nil
}

// List returns all entries with decrypted values (full view)
func (s *Store) List() (map[string]SecretEntry, error) {
	if err := s.sec.RequireMasterPassword(); err != nil {
		return nil, err
	}

	result := make(map[string]SecretEntry)

	s.mu.RLock()
	err := s.db.View(func(tx *nutsdb.Tx) error {
		return tx.ForEach(Bucket, func(k, v []byte) bool {
			keyStr := string(k)

			var entry SecretEntry
			if decErr := s.sec.DecryptJSON(v, &entry); decErr == nil {
				// Verifica TTL
				if entry.TTL > 0 && time.Since(entry.UpdatedAt) > time.Duration(entry.TTL)*time.Second {
					return true // ignora expirada
				}
				result[keyStr] = entry
			}
			return true
		})
	})
	s.mu.RUnlock()

	if err != nil && err != nutsdb.ErrBucketNotFound {
		return nil, err
	}

	return result, nil
}

// Delete removes a key
func (s *Store) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.db.Update(func(tx *nutsdb.Tx) error {
		return tx.Delete(Bucket, []byte(key))
	})
}

// DeleteMultiple removes multiple keys
func (s *Store) DeleteMultiple(keys []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.db.Update(func(tx *nutsdb.Tx) error {
		for _, key := range keys {
			_ = tx.Delete(Bucket, []byte(key)) // ignora erros individuais
		}
		return nil
	})
}

// Search finds keys by prefix or tag
func (s *Store) Search(query string) (map[string]SecretEntry, error) {
	if err := s.sec.RequireMasterPassword(); err != nil {
		return nil, err
	}

	if query == "" {
		return s.List()
	}

	result := make(map[string]SecretEntry)
	queryLower := strings.ToLower(query)

	s.mu.RLock()
	err := s.db.View(func(tx *nutsdb.Tx) error {
		return tx.ForEach(Bucket, func(k, v []byte) bool {
			keyStr := string(k)
			keyLower := strings.ToLower(keyStr)

			var entry SecretEntry
			if decErr := s.sec.DecryptJSON(v, &entry); decErr != nil {
				return true
			}

			// Verifica TTL
			if entry.TTL > 0 && time.Since(entry.UpdatedAt) > time.Duration(entry.TTL)*time.Second {
				return true
			}

			// Search in the key name
			if strings.Contains(keyLower, queryLower) {
				result[keyStr] = entry
				return true
			}

			// Search in the tags
			for _, tag := range entry.Tags {
				if strings.Contains(strings.ToLower(tag), queryLower) {
					result[keyStr] = entry
					break
				}
			}

			return true
		})
	})
	s.mu.RUnlock()

	if err != nil && err != nutsdb.ErrBucketNotFound {
		return nil, err
	}

	return result, nil
}
