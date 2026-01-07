package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuage-identity/iam/auth/token"
)

// JWTAuthMiddleware creates middleware for JWT token validation
func JWTAuthMiddleware(tokenService token.ServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Extract Bearer token
		tokenString := ""
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid authorization header format. Expected: Bearer <token>",
			})
			c.Abort()
			return
		}

		// Validate token
		claims, err := tokenService.ValidateAccessToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// TODO: Check token blacklist (Redis)
		// This will be implemented when Redis blacklist is set up

		// Set user context
		c.Set("user_id", claims.Subject)
		c.Set("tenant_id", claims.TenantID)
		c.Set("user_claims", claims)

		c.Next()
	}
}

