package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/arauth-identity/iam/auth/claims"
	"github.com/arauth-identity/iam/auth/token"
	"github.com/arauth-identity/iam/identity/models"
)

// RequireSystemUser ensures user is a SYSTEM principal
func RequireSystemUser(tokenService token.ServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get claims from context (set by JWTAuth middleware)
		claimsObj, exists := c.Get("user_claims")
		if !exists {
			RespondWithError(c, http.StatusUnauthorized, "unauthorized", "User claims not found", nil)
			c.Abort()
			return
		}

		userClaims := claimsObj.(*claims.Claims)
		if userClaims.PrincipalType != string(models.PrincipalTypeSystem) {
			RespondWithError(c, http.StatusForbidden, "forbidden",
				"System user access required. This endpoint is only available to system administrators.", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireTenantUser ensures user is a TENANT principal
func RequireTenantUser(tokenService token.ServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get claims from context (set by JWTAuth middleware)
		claimsObj, exists := c.Get("user_claims")
		if !exists {
			RespondWithError(c, http.StatusUnauthorized, "unauthorized", "User claims not found", nil)
			c.Abort()
			return
		}

		userClaims := claimsObj.(*claims.Claims)
		if userClaims.PrincipalType != string(models.PrincipalTypeTenant) {
			RespondWithError(c, http.StatusForbidden, "forbidden",
				"Tenant user access required. System users cannot access tenant-scoped endpoints.", nil)
			c.Abort()
			return
		}

		// For TENANT users, verify they have a tenant_id in their JWT token
		// The tenant_id will be extracted by TenantMiddleware from the token
		// We don't check GetTenantID here because TenantMiddleware hasn't run yet
		if userClaims.TenantID == "" {
			RespondWithError(c, http.StatusBadRequest, "tenant_required",
				"TENANT users must have tenant_id in their JWT token", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireSystemPermission checks if system user has required permission
func RequireSystemPermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claimsObj, exists := c.Get("user_claims")
		if !exists {
			RespondWithError(c, http.StatusUnauthorized, "unauthorized", "User claims not found", nil)
			c.Abort()
			return
		}

		userClaims := claimsObj.(*claims.Claims)
		if userClaims.PrincipalType != string(models.PrincipalTypeSystem) {
			RespondWithError(c, http.StatusForbidden, "forbidden", "System user required", nil)
			c.Abort()
			return
		}

		requiredPerm := resource + ":" + action
		hasPermission := false
		for _, perm := range userClaims.SystemPermissions {
			if perm == requiredPerm || perm == resource+":*" || perm == "*:*" {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			RespondWithError(c, http.StatusForbidden, "forbidden",
				"Required permission: "+requiredPerm, nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

