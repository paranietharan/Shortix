package model

import "github.com/golang-jwt/jwt/v4"

type AuthClaims struct {
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
	SessionID string `json:"sid"`
	jwt.RegisteredClaims
}
