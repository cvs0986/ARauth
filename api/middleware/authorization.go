package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// AuthorizationMiddleware creates middleware for role-based authorization
func AuthorizationMiddleware(roleRepo interfaces.RoleRepository, permissionRepo interfaces.PermissionRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get tenant ID from context
		tenantID, exists := GetTenantID(c)
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "tenant_required",
				"message": "Tenant context is required",
			})
			c.Abort()
			return
		}

		// Get user ID from JWT token (for now, we'll get it from a header or context)
		// In production, this would come from a validated JWT token
		userIDStr := c.GetHeader("X-User-ID")
		if userIDStr == "" {
			// Try to get from context (set by auth middleware in future)
			if userID, ok := c.Get("user_id"); ok {
				userIDStr = userID.(string)
			}
		}

		if userIDStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "User authentication required",
			})
			c.Abort()
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_user_id",
				"message": "Invalid user ID format",
			})
			c.Abort()
			return
		}

		// Get user roles
		roles, err := roleRepo.GetUserRoles(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "authorization_error",
				"message": "Failed to get user roles",
			})
			c.Abort()
			return
		}

		// Store roles and permissions in context
		c.Set("user_id", userID)
		c.Set("user_roles", roles)

		// Get permissions for all roles
		permissionMap := make(map[string]bool)
		for _, role := range roles {
			// Verify role belongs to tenant
			if role.TenantID != tenantID {
				continue
			}

			permissions, err := permissionRepo.GetRolePermissions(c.Request.Context(), role.ID)
			if err != nil {
				continue
			}

			for _, perm := range permissions {
				permissionKey := perm.Resource + ":" + perm.Action
				permissionMap[permissionKey] = true
			}
		}

		permissions := make([]string, 0, len(permissionMap))
		for perm := range permissionMap {
			permissions = append(permissions, perm)
		}

		c.Set("user_permissions", permissions)

		c.Next()
	}
}

// RequireRole creates middleware that requires a specific role
func RequireRole(roleName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("user_roles")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Role information not available",
			})
			c.Abort()
			return
		}

		userRoles := roles.([]interface{}) // This will need type assertion based on actual type
		hasRole := false

		// Check if user has required role
		// Note: This is simplified - actual implementation would check role names
		_ = userRoles
		_ = roleName

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Required role not found: " + roleName,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission creates middleware that requires a specific permission
func RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := c.Get("user_permissions")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Permission information not available",
			})
			c.Abort()
			return
		}

		userPermissions := permissions.([]string)
		requiredPermission := resource + ":" + action

		hasPermission := false
		for _, perm := range userPermissions {
			if perm == requiredPermission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Required permission not found: " + requiredPermission,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission creates middleware that requires any of the specified permissions
func RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userPermissions, exists := c.Get("user_permissions")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Permission information not available",
			})
			c.Abort()
			return
		}

		userPerms := userPermissions.([]string)
		permissionSet := make(map[string]bool)
		for _, perm := range userPerms {
			permissionSet[perm] = true
		}

		hasAny := false
		for _, requiredPerm := range permissions {
			if permissionSet[requiredPerm] {
				hasAny = true
				break
			}
		}

		if !hasAny {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "None of the required permissions found",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserPermissions retrieves user permissions from context
func GetUserPermissions(c *gin.Context) ([]string, bool) {
	permissions, exists := c.Get("user_permissions")
	if !exists {
		return nil, false
	}

	perms, ok := permissions.([]string)
	return perms, ok
}

// HasPermission checks if user has a specific permission
func HasPermission(c *gin.Context, resource, action string) bool {
	permissions, ok := GetUserPermissions(c)
	if !ok {
		return false
	}

	requiredPermission := resource + ":" + action
	for _, perm := range permissions {
		if perm == requiredPermission {
			return true
		}
	}

	return false
}

