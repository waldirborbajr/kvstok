package database

import "time"

type SecretEntry struct {
	Value     string
	TTL       uint32
	Tags      []string // ✅ NOVO
	CreatedAt time.Time
	UpdatedAt time.Time
}
