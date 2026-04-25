package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	apperrors "shortix-api/internal/errors"

	"github.com/gin-gonic/gin"
)

type rateWindow struct {
	count      int
	windowEnds time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	window  time.Duration
	maxHits int
	store   map[string]*rateWindow
}

func NewRateLimiter(window time.Duration, maxHits int) *RateLimiter {
	return &RateLimiter{
		window:  window,
		maxHits: maxHits,
		store:   make(map[string]*rateWindow),
	}
}

func (r *RateLimiter) Middleware(keyPrefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyPrefix + ":" + c.ClientIP()
		now := time.Now().UTC()

		r.mu.Lock()
		entry, exists := r.store[key]
		if !exists || now.After(entry.windowEnds) {
			r.store[key] = &rateWindow{count: 1, windowEnds: now.Add(r.window)}
			r.mu.Unlock()
			c.Next()
			return
		}

		if entry.count >= r.maxHits {
			retryAfter := int(entry.windowEnds.Sub(now).Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}
			r.mu.Unlock()
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"code":    apperrors.ErrTooManyRequests.Code,
					"message": apperrors.ErrTooManyRequests.Message,
				},
			})
			return
		}

		entry.count++
		r.mu.Unlock()
		c.Next()
	}
}
