package dto

import "time"

type SessionResponse struct {
	ID               string    `json:"id"`
	Device           string    `json:"device,omitempty"`
	IP               string    `json:"ip,omitempty"`
	UserAgent        string    `json:"user_agent,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}
