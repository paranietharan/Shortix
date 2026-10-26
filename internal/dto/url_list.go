package dto

import (
	"time"

	"github.com/google/uuid"
)

type URLResponse struct {
	ID          uuid.UUID  `json:"id"`
	LongURL     string     `json:"long_url"`
	ShortCode   string     `json:"short_code"`
	CustomAlias *string    `json:"custom_alias,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type ListURLsResponse struct {
	URLs []URLResponse `json:"urls"`
	Total int64        `json:"total"`
	Page  int          `json:"page"`
	Limit int          `json:"limit"`
}

type PaginationQuery struct {
	Page  int `form:"page,default=1"`
	Limit int `form:"limit,default=10"`
}
