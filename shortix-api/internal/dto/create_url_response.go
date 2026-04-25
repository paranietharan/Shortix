package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateURLResponse struct {
	ID          uuid.UUID  `json:"id"`
	LongURL     string     `json:"long_url"`
	ShortCode   string     `json:"short_code"`
	CustomAlias *string    `json:"custom_alias,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	ShortURL    string     `json:"short_url"`
}
