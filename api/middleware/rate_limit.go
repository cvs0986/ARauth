package middleware

import (
	"github.com/gin-gonic/gin"
)

// RateLimit returns a rate limiting middleware (placeholder)
// TODO: Implement actual rate limiting with Redis
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Placeholder - actual implementation will use Redis
		c.Next()
	}
}

