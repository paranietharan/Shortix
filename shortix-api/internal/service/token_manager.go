package service

import (
	"fmt"
	"time"

	"shortix-api/internal/model"

	"github.com/golang-jwt/jwt/v4"
)

type TokenManager struct {
	secret []byte
}

func NewTokenManager(secret string) *TokenManager {
	return &TokenManager{secret: []byte(secret)}
}

func (m *TokenManager) GenerateAccessToken(userID, role, sessionID string, expiresAt time.Time) (string, error) {
	claims := model.AuthClaims{
		UserID:    userID,
		Role:      role,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			NotBefore: jwt.NewNumericDate(time.Now().UTC()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *TokenManager) ParseAccessToken(tokenString string) (*model.AuthClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &model.AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*model.AuthClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid access token")
	}
	return claims, nil
}
