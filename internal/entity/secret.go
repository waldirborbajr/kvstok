package entity

import "time"

// SecretEntry represents an encrypted database entry
type SecretEntry struct {
	Value     string    `json:"value"`
	TTL       uint32    `json:"ttl,omitempty"`
	Tags      []string  `json:"tags,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
