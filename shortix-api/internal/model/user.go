package model

import "time"

type User struct {
	ID                 string
	Email              string
	PasswordHash       string
	Role               string
	IsActive           bool
	IsEmailVerified    bool
	EmailVerifiedAt    *time.Time
	LastLoginAt        *time.Time
	LastLoginIP        *string
	LastLoginUserAgent *string
	LastLoginDevice    *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
