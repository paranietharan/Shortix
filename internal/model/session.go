package model

import "time"

type Session struct {
	ID               string
	UserID           string
	AccessTokenHash  string
	RefreshTokenHash string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
	IsRevoked        bool
	IPAddress        *string
	UserAgent        *string
	Device           *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
