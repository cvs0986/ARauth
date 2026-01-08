package middleware

import (
	"net/http"
	"strings"

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

		// Verify tenant_id matches requested tenant
		requestedTenantID, ok := GetTenantID(c)
		if !ok {
			RespondWithError(c, http.StatusBadRequest, "tenant_required", "Tenant context is required", nil)
			c.Abort()
			return
		}

		if userClaims.TenantID != "" {
			if userClaims.TenantID != requestedTenantID.String() {
				RespondWithError(c, http.StatusForbidden, "forbidden",
					"You do not have access to this tenant", nil)
				c.Abort()
				return
			}
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

