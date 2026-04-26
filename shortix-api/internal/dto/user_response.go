package dto

import "time"

type UserResponse struct {
	ID                string     `json:"id"`
	Email             string     `json:"email"`
	FirstName         *string    `json:"first_name,omitempty"`
	LastName          *string    `json:"last_name,omitempty"`
	ProfilePictureURL *string    `json:"profile_picture_url,omitempty"`
	Bio               *string    `json:"bio,omitempty"`
	PhoneNumber       *string    `json:"phone_number,omitempty"`
	Role              string     `json:"role"`
	IsActive          bool       `json:"is_active"`
	IsEmailVerified   bool       `json:"is_email_verified"`
	CreatedAt         time.Time  `json:"created_at"`
	LastLoginAt       *time.Time `json:"last_login_at,omitempty"`
}

type ListUsersResponse struct {
	Users []UserResponse `json:"users"`
	Total int            `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}
