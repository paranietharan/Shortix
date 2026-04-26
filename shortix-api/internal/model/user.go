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
	FirstName          *string
	LastName           *string
	ProfilePictureURL  *string
	Bio                *string
	PhoneNumber        *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type EmailChangeData struct {
	NewEmail string `json:"new_email"`
	OTP      string `json:"otp"`
	Attempts int    `json:"attempts"`
}

type PasswordChangeData struct {
	HashedNewPassword string `json:"hashed_new_password"`
	OTP               string `json:"otp"`
	Attempts          int    `json:"attempts"`
}
