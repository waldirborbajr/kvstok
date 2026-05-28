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
		return fmt.Errorf("failed to obtain home directory: %w", err)
	}

	auditDir := filepath.Join(home, ".config", "kvstok")

	// Create the directory if it does not exist
	if err := os.MkdirAll(auditDir, 0700); err != nil {
		return fmt.Errorf("failed to create audit directory: %w", err)
	}

	auditPath = filepath.Join(auditDir, "audit.log")
	return nil
}

// LogOperation records an operation in the audit log
func LogOperation(op string, key string, success bool, err string) error {
	if auditPath == "" {
		// If not initialized, attempt to initialize silently
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

// writeAuditEntry writes an entry to the audit log in JSON lines format
func writeAuditEntry(entry AuditEntry) error {
	auditLogMu.Lock()
	defer auditLogMu.Unlock()

	// Converte para JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to serialize audit entry: %w", err)
	}

	// Open file in append mode with restricted permissions
	file, err := os.OpenFile(auditPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open audit file: %w", err)
	}
	defer file.Close()

	// Escreve JSON + newline (JSON lines format)
	if _, err := file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write audit log: %w", err)
	}

	return nil
}

// ReadAuditLog reads and returns all entries from the audit log
func ReadAuditLog() ([]AuditEntry, error) {
	if auditPath == "" {
		return nil, fmt.Errorf("audit logger has not been initialized")
	}

	auditLogMu.Lock()
	defer auditLogMu.Unlock()

	data, err := os.ReadFile(auditPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []AuditEntry{}, nil
		}
		return nil, fmt.Errorf("failed to read audit log: %w", err)
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

// ReadAuditLogFiltered returns entries filtered by the provided criteria
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

// ClearAuditLog clears the audit log file
func ClearAuditLog() error {
	if auditPath == "" {
		return fmt.Errorf("audit logger has not been initialized")
	}

	auditLogMu.Lock()
	defer auditLogMu.Unlock()

	return os.Remove(auditPath)
}

// GetAuditPath returns the path to the audit log file
func GetAuditPath() string {
	return auditPath
}

// LineScanner is a simple scanner for processing lines from a byte slice
type LineScanner struct {
	data   []byte
	offset int
	line   []byte
}

// NewLineScanner creates a new line scanner
func NewLineScanner(data []byte) *LineScanner {
	return &LineScanner{
		data: data,
	}
}

// Scan advances to the next line
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

// Bytes returns the current line
func (s *LineScanner) Bytes() []byte {
	return s.line
}
