package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/arauth-identity/iam/identity/scim"
)

// SCIMAuthMiddleware validates SCIM Bearer tokens
func SCIMAuthMiddleware(tokenService scim.TokenServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract Bearer token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"schemas": []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
				"detail":  "Authorization header required",
				"status":  "401",
			})
			c.Abort()
			return
		}

		// Parse Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"schemas": []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
				"detail":  "Invalid authorization header format",
				"status":  "401",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
		token, err := tokenService.ValidateToken(c.Request.Context(), tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"schemas": []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
				"detail":  "Invalid or expired token",
				"status":  "401",
			})
			c.Abort()
			return
		}

		// Store token and tenant ID in context
		c.Set("scim_token", token)
		c.Set("scim_tenant_id", token.TenantID)
		c.Set("scim_scopes", token.Scopes)

		c.Next()
	}
}

// RequireSCIMScope checks if the request has the required SCIM scope
func RequireSCIMScope(requiredScope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		scopes, exists := c.Get("scim_scopes")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"schemas": []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
				"detail":  "Insufficient permissions",
				"status":  "403",
			})
			c.Abort()
			return
		}

		scopeList := scopes.([]string)
		hasScope := false
		for _, scope := range scopeList {
			if scope == requiredScope || scope == "*" {
				hasScope = true
				break
			}
		}

		if !hasScope {
			c.JSON(http.StatusForbidden, gin.H{
				"schemas": []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
				"detail":  "Insufficient permissions: " + requiredScope + " scope required",
				"status":  "403",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

