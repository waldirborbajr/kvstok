// internal/audit/logger.go - NOVO

package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type AuditEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Operation string    `json:"operation"` // add, get, del, exp, imp
	Key       string    `json:"key"`
	User      string    `json:"user"` // $USER env var
	Success   bool      `json:"success"`
	ErrorMsg  string    `json:"error_msg,omitempty"`
}

var (
	auditLogMu sync.Mutex
	auditPath  string
)

// Init initializes the audit logger path
func Init() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("falha ao obter diretório home: %w", err)
	}

	auditDir := filepath.Join(home, ".config", "kvstok")

	// Cria diretório se não existir
	if err := os.MkdirAll(auditDir, 0700); err != nil {
		return fmt.Errorf("falha ao criar diretório de audit: %w", err)
	}

	auditPath = filepath.Join(auditDir, "audit.log")
	return nil
}

// LogOperation registra uma operação no log de auditoria
func LogOperation(op string, key string, success bool, err string) error {
	if auditPath == "" {
		// Se não foi inicializado, tenta inicializar silenciosamente
		if initErr := Init(); initErr != nil {
			return initErr
		}
	}

	entry := AuditEntry{
		Timestamp: time.Now(),
		Operation: op,
		Key:       key,
		User:      os.Getenv("USER"),
		Success:   success,
		ErrorMsg:  err,
	}

	return writeAuditEntry(entry)
}

// writeAuditEntry escreve uma entrada no arquivo de auditoria em formato JSON lines
func writeAuditEntry(entry AuditEntry) error {
	auditLogMu.Lock()
	defer auditLogMu.Unlock()

	// Converte para JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("falha ao serializar entrada de auditoria: %w", err)
	}

	// Abre arquivo em modo append com permissões restritas
	file, err := os.OpenFile(auditPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("falha ao abrir arquivo de auditoria: %w", err)
	}
	defer file.Close()

	// Escreve JSON + newline (JSON lines format)
	if _, err := file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("falha ao escrever no log de auditoria: %w", err)
	}

	return nil
}

// ReadAuditLog lê e retorna todas as entradas do log de auditoria
func ReadAuditLog() ([]AuditEntry, error) {
	if auditPath == "" {
		return nil, fmt.Errorf("audit logger não foi inicializado")
	}

	auditLogMu.Lock()
	defer auditLogMu.Unlock()

	data, err := os.ReadFile(auditPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []AuditEntry{}, nil
		}
		return nil, fmt.Errorf("falha ao ler log de auditoria: %w", err)
	}

	var entries []AuditEntry
	scanner := NewLineScanner(data)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry AuditEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			// Log silencioso para linhas malformadas
			continue
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// ReadAuditLogFiltered retorna entradas filtradas por critérios
func ReadAuditLogFiltered(operation string, user string, startTime time.Time) ([]AuditEntry, error) {
	entries, err := ReadAuditLog()
	if err != nil {
		return nil, err
	}

	var filtered []AuditEntry
	for _, entry := range entries {
		if operation != "" && entry.Operation != operation {
			continue
		}
		if user != "" && entry.User != user {
			continue
		}
		if !startTime.IsZero() && entry.Timestamp.Before(startTime) {
			continue
		}

		filtered = append(filtered, entry)
	}

	return filtered, nil
}

// ClearAuditLog limpa o arquivo de auditoria
func ClearAuditLog() error {
	if auditPath == "" {
		return fmt.Errorf("audit logger não foi inicializado")
	}

	auditLogMu.Lock()
	defer auditLogMu.Unlock()

	return os.Remove(auditPath)
}

// GetAuditPath retorna o caminho do arquivo de auditoria
func GetAuditPath() string {
	return auditPath
}

// LineScanner é um scanner simples para processar linhas de um slice de bytes
type LineScanner struct {
	data   []byte
	offset int
	line   []byte
}

// NewLineScanner cria um novo scanner de linhas
func NewLineScanner(data []byte) *LineScanner {
	return &LineScanner{
		data: data,
	}
}

// Scan avança para a próxima linha
func (s *LineScanner) Scan() bool {
	if s.offset >= len(s.data) {
		return false
	}

	start := s.offset
	for s.offset < len(s.data) && s.data[s.offset] != '\n' {
		s.offset++
	}

	s.line = s.data[start:s.offset]

	// Pula o '\n'
	if s.offset < len(s.data) {
		s.offset++
	}

	return true
}

// Bytes retorna a linha atual
func (s *LineScanner) Bytes() []byte {
	return s.line
}
