package middleware

import (
	"net/http"

	"github.com/arauth-identity/iam/auth/token"
	"github.com/gin-gonic/gin"
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

		// Check token blacklist (Redis)
		if claims.ID != "" {
			revoked, err := tokenService.IsAccessTokenRevoked(c.Request.Context(), claims.ID)
			if err != nil {
				// FAIL CLOSED: If we can't check revocation status, we reject the request
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "unauthorized",
					"message": "Failed to verify token status",
				})
				c.Abort()
				return
			}
			if revoked {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "unauthorized",
					"message": "Token revoked",
				})
				c.Abort()
				return
			}
		}

		// Set user context
		c.Set("user_id", claims.Subject)
		if claims.TenantID != "" {
			c.Set("tenant_id", claims.TenantID)
		}
		c.Set("principal_type", claims.PrincipalType)
		c.Set("user_claims", claims)
		c.Set("system_roles", claims.SystemRoles)
		c.Set("system_permissions", claims.SystemPermissions)
		c.Set("roles", claims.Roles)
		c.Set("permissions", claims.Permissions)

		c.Next()
	}
}
