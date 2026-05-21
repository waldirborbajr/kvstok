// internal/audit/logger.go - NOVO

package audit

import (
	"os"
	"time"
)

type AuditEntry struct {
	Timestamp time.Time
	Operation string // add, get, del, exp, imp
	Key       string
	User      string // $USER env var
	Success   bool
	ErrorMsg  string
}

func LogOperation(op string, key string, success bool, err string) {
	entry := AuditEntry{
		Timestamp: time.Now(),
		Operation: op,
		Key:       key,
		User:      os.Getenv("USER"),
		Success:   success,
		ErrorMsg:  err,
	}

	// Salvar em ~/.config/kvstok/audit.log
	// Formato: JSON lines para easy parsing
}
