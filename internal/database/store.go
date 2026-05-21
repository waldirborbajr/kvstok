// internal/database/store.go
package database

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

	if err := s.sec.GetMasterKey().SetMasterPassword(password); err != nil { // Correção: usar GetMasterKey()
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
		Value:     value, // será sobrescrito com criptografia
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

	return entry.Value, nil // já vem descriptografado do DecryptJSON? Não! Ajuste abaixo.
}

// GetRaw retorna o valor descriptografado + os metadados completos da entrada
// func (s *Store) GetRaw(key string) (string, *SecretEntry, error) {
// 	if err := s.sec.RequireMasterPassword(); err != nil {
// 		return "", nil, err
// 	}

// 	var encryptedData []byte

// 	s.mu.RLock()
// 	err := s.db.View(func(tx *nutsdb.Tx) error {
// 		val, err := tx.Get(Bucket, []byte(key))
// 		if err != nil {
// 			return err
// 		}
// 		// Fazemos uma cópia para evitar problemas com slice reference
// 		encryptedData = make([]byte, len(val))
// 		copy(encryptedData, val)
// 		return nil
// 	})
// 	s.mu.RUnlock()

// 	if err != nil {
// 		if err == nutsdb.ErrKeyNotFound || err == nutsdb.ErrBucketNotFound {
// 			return "", nil, ErrKeyNotFound
// 		}
// 		return "", nil, fmt.Errorf("erro ao buscar chave: %w", err)
// 	}

// 	// Descriptografa o JSON completo
// 	var entry SecretEntry
// 	if err := s.sec.DecryptJSON(encryptedData, &entry); err != nil {
// 		return "", nil, fmt.Errorf("falha na descriptografia: %w", err)
// 	}

// 	// Verifica expiração (TTL)
// 	if entry.TTL > 0 {
// 		expirationTime := entry.UpdatedAt.Add(time.Duration(entry.TTL) * time.Second)
// 		if time.Now().After(expirationTime) {
// 			// Opcional: auto-delete da chave expirada
// 			_ = s.Delete(key)
// 			return "", nil, ErrKeyExpired
// 		}
// 	}

// 	// entry.Value já vem descriptografado do fluxo anterior? Não!
// 	// Vamos garantir que o valor original esteja correto:
// 	originalValue := entry.Value

// 	return originalValue, &entry, nil
// }

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