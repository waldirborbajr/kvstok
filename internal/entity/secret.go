package entity

import "time"

// SecretEntry representa uma entrada criptografada no banco
type SecretEntry struct {
	Value     string    `json:"value"`
	TTL       uint32    `json:"ttl,omitempty"`
	Tags      []string  `json:"tags,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
