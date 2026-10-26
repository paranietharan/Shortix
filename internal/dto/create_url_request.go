package dto

import "time"

type CreateURLRequest struct {
	LongURL     string     `json:"long_url" binding:"required,url"`
	CustomAlias *string    `json:"custom_alias" binding:"omitempty,alphanum,min=4,max=32"`
	ExpiresAt   *time.Time `json:"expires_at" binding:"omitempty,gt"`
}
