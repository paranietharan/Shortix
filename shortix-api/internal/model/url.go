package model

import (
	"time"

	"github.com/google/uuid"
)

type URL struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID      uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	LongURL     string     `json:"long_url" gorm:"type:text;not null"`
	ShortCode   string     `json:"short_code" gorm:"type:varchar(20);uniqueIndex;not null"`
	CustomAlias *string    `json:"custom_alias" gorm:"type:varchar(50);uniqueIndex"`
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
}
