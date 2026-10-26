package middleware

import (
	apperrors "shortix-api/internal/errors"

	"github.com/gin-gonic/gin"
)

func RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		role, ok := RoleFromContext(c)
		if !ok {
			writeError(c, apperrors.ErrUnauthorized)
			return
		}
		if _, exists := allowed[role]; !exists {
			writeError(c, apperrors.ErrForbidden)
			return
		}
		c.Next()
	}
}
