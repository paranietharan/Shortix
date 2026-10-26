package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	apperrors "shortix-api/internal/errors"
	"shortix-api/internal/repository"
	"shortix-api/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	tokens   *service.TokenManager
	sessions repository.SessionRepository
}

func NewAuthMiddleware(tokens *service.TokenManager, sessions repository.SessionRepository) *AuthMiddleware {
	return &AuthMiddleware{tokens: tokens, sessions: sessions}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		rawToken, ok := extractBearerToken(c.GetHeader("Authorization"))
		if !ok {
			writeError(c, apperrors.ErrUnauthorized)
			return
		}

		claims, err := m.tokens.ParseAccessToken(rawToken)
		if err != nil {
			writeError(c, apperrors.ErrUnauthorized)
			return
		}

		session, err := m.sessions.GetByAccessHash(c.Request.Context(), hashToken(rawToken))
		if err != nil {
			writeError(c, apperrors.ErrUnauthorized)
			return
		}
		if session.IsRevoked || time.Now().UTC().After(session.AccessExpiresAt) {
			writeError(c, apperrors.ErrSessionRevoked)
			return
		}

		if claims.SessionID != session.ID {
			writeError(c, apperrors.ErrUnauthorized)
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextRoleKey, claims.Role)
		c.Set(ContextSessionIDKey, claims.SessionID)
		c.Next()
	}
}

func (m *AuthMiddleware) RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, ok := RoleFromContext(c)
		if !ok {
			writeError(c, apperrors.ErrForbidden)
			c.Abort()
			return
		}

		hasRole := false
		for _, role := range roles {
			if userRole == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			writeError(c, apperrors.ErrForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}

func extractBearerToken(header string) (string, bool) {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 {
		return "", false
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", false
	}
	return token, true
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func writeError(c *gin.Context, appErr *apperrors.AppError) {
	c.AbortWithStatusJSON(appErr.HTTPStatus, gin.H{
		"error": gin.H{
			"code":    appErr.Code,
			"message": appErr.Message,
		},
	})
}

func UserIDFromContext(c *gin.Context) (string, bool) {
	v, ok := c.Get(ContextUserIDKey)
	if !ok {
		return "", false
	}
	userID, ok := v.(string)
	return userID, ok
}

func SessionIDFromContext(c *gin.Context) (string, bool) {
	v, ok := c.Get(ContextSessionIDKey)
	if !ok {
		return "", false
	}
	sessionID, ok := v.(string)
	return sessionID, ok
}

func RoleFromContext(c *gin.Context) (string, bool) {
	v, ok := c.Get(ContextRoleKey)
	if !ok {
		return "", false
	}
	role, ok := v.(string)
	return role, ok
}

func AbortUnauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error": gin.H{
			"code":    apperrors.ErrUnauthorized.Code,
			"message": apperrors.ErrUnauthorized.Message,
		},
	})
}
