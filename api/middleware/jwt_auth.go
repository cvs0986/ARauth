package middleware

import (
	"net/http"

	"github.com/arauth-identity/iam/auth/token"
	"github.com/arauth-identity/iam/observability/security_events"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// JWTAuthMiddleware creates middleware for JWT token validation
func JWTAuthMiddleware(tokenService token.ServiceInterface, eventLogger security_events.Logger) gin.HandlerFunc {
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
			// Log token validation failure
			if eventLogger != nil {
				event := security_events.NewSecurityEvent(
					security_events.EventTokenValidationFailed,
					security_events.SeverityWarning,
				).WithIP(c.ClientIP()).
					WithResource(c.Request.URL.Path).
					WithAction(c.Request.Method).
					WithResult("failure").
					WithDetail("reason", err.Error())
				eventLogger.LogEvent(c.Request.Context(), event)
			}

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
				// Log blacklisted token usage (CRITICAL)
				if eventLogger != nil {
					event := security_events.NewSecurityEvent(
						security_events.EventBlacklistedTokenUsed,
						security_events.SeverityCritical,
					).WithIP(c.ClientIP()).
						WithResource(c.Request.URL.Path).
						WithAction(c.Request.Method).
						WithResult("blocked").
						WithDetail("token_id", claims.ID)

					if claims.Subject != "" {
						if userID, err := uuid.Parse(claims.Subject); err == nil {
							event.WithUser(userID)
						}
					}
					if claims.TenantID != "" {
						if tenantID, err := uuid.Parse(claims.TenantID); err == nil {
							event.WithTenant(tenantID)
						}
					}

					eventLogger.LogEvent(c.Request.Context(), event)
				}

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
