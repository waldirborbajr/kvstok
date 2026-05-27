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
	DBName  = ".6B7673" // .kvs
	Bucket  = "kvstok"
	SaltKey = "master_salt" // chave especial para armazenar o salt
)

var (
	ErrKeyNotFound = errors.New("chave não encontrada")
	ErrKeyExpired  = errors.New("chave expirada")
)

// Store é a camada principal de acesso ao banco com criptografia
type Store struct {
	db     *nutsdb.DB
	sec    *security.SecureEncrypt
	dbPath string
	mu     sync.RWMutex
}

// NewStore cria uma nova instância do Store
func NewStore(path string) (*Store, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(home, DBName)
	}

	opts := nutsdb.DefaultOptions
	opts.Dir = path
	opts.EntryMaxSize = 1024 * 1024 * 8 // 8MB

	db, err := nutsdb.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("falha ao abrir banco de dados: %w", err)
	}

	s := &Store{
		db:     db,
		sec:    security.NewSecureEncrypt(),
		dbPath: path,
	}

	return s, nil
}

// Close fecha o banco
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Close()
}

// SetMasterPassword configura a senha mestra (usado no init)
func (s *Store) SetMasterPassword(password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.sec.GetMasterKey().SetMasterPassword(password); err != nil {
		return err
	}

	// Salva o salt
	saltPath := filepath.Join(s.dbPath, "master.salt")
	return s.sec.GetMasterKey().SaveSalt(saltPath)
}

// LoadMasterSalt carrega o salt salvo
func (s *Store) LoadMasterSalt() error {
	saltPath := filepath.Join(s.dbPath, "master.salt")
	return s.sec.GetMasterKey().LoadSalt(saltPath)
}

// Put insere ou atualiza uma chave com criptografia
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

	// Criptografa o valor
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

// Get retorna o valor descriptografado
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

	// Descriptografa o entry
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
		return "", nil, fmt.Errorf("descriptografia falhou: %w", err)
	}

	// Verifica TTL
	if se.TTL > 0 && time.Since(se.UpdatedAt) > time.Duration(se.TTL)*time.Second {
		_ = s.Delete(key) // limpa chave expirada
		return "", nil, ErrKeyExpired
	}

	return se.Value, &se, nil
}

// ListAll retorna todas as chaves existentes (apenas as keys, sem descriptografar)
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

// List retorna todas as entradas com valor descriptografado (mais completo)
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

// Delete remove uma chave
func (s *Store) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.db.Update(func(tx *nutsdb.Tx) error {
		return tx.Delete(Bucket, []byte(key))
	})
}

// DeleteMultiple remove várias chaves
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

// Search busca chaves por prefixo ou por tag
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

			// Busca no nome da chave
			if strings.Contains(keyLower, queryLower) {
				result[keyStr] = entry
				return true
			}

			// Busca nas tags
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
