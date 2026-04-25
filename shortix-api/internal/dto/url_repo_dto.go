package dto

import "time"

type URL struct {
	ID          string
	UserID      *string
	LongURL     string
	ShortCode   string
	CustomAlias *string
	ExpiresAt   *time.Time
	Metadata    map[string]interface{}
	CreatedAt   time.Time
}
