package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// TenantContextKey is the key for storing tenant context
const TenantContextKey = "tenant_id"

// TenantContext represents tenant information in the request context
type TenantContext struct {
	TenantID uuid.UUID
	Domain   string
}

// TenantMiddleware extracts tenant information from request
// Supports multiple methods:
// 1. JWT token claims (tenant_id from token - for TENANT users)
// 2. X-Tenant-ID header
// 3. X-Tenant-Domain header
// 4. tenant_id query parameter
// 5. domain query parameter
func TenantMiddleware(tenantRepo interfaces.TenantRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tenantID uuid.UUID
		var err error

		// Check user principal type to determine tenant ID source priority
		principalType, _ := c.Get("principal_type")
		isSystemUser := principalType == "SYSTEM"

		// For TENANT users: tenant ID must come from JWT token (security - they can't switch tenants)
		// For SYSTEM users: tenant ID can come from header (they can select tenant context)
		if !isSystemUser {
			// TENANT users: get tenant ID from JWT token claims (set by JWTAuthMiddleware)
			if tenantIDStr, exists := c.Get("tenant_id"); exists {
				// tenant_id is stored as string in context
				var tenantIDStrValue string
				switch v := tenantIDStr.(type) {
				case string:
					tenantIDStrValue = v
				case uuid.UUID:
					tenantID = v
				default:
					// Try to convert to string
					if str, ok := tenantIDStr.(string); ok {
						tenantIDStrValue = str
					}
				}
				
				if tenantIDStrValue != "" && tenantID == uuid.Nil {
					tenantID, err = uuid.Parse(tenantIDStrValue)
					if err != nil {
						c.JSON(http.StatusBadRequest, gin.H{
							"error":   "invalid_tenant_id",
							"message": "Invalid tenant ID format in JWT token: " + err.Error(),
						})
						c.Abort()
						return
					}
				}
			}
		}

		// Try X-Tenant-ID header (for SYSTEM users, or if TENANT user's JWT didn't have tenant_id)
		if tenantIDStr := c.GetHeader("X-Tenant-ID"); tenantIDStr != "" {
			headerTenantID, err := uuid.Parse(tenantIDStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "invalid_tenant_id",
					"message": "Invalid tenant ID format in X-Tenant-ID header",
				})
				c.Abort()
				return
			}
			// For SYSTEM users: header can set/override tenant ID
			// For TENANT users: header must match JWT tenant_id (security check)
			if isSystemUser {
				tenantID = headerTenantID
			} else {
				// TENANT user: verify header matches JWT tenant_id
				if tenantID != uuid.Nil && tenantID != headerTenantID {
					c.JSON(http.StatusForbidden, gin.H{
						"error":   "tenant_mismatch",
						"message": "X-Tenant-ID header does not match tenant ID in JWT token. TENANT users cannot access other tenants.",
					})
					c.Abort()
					return
				}
				// If JWT didn't have tenant_id, use header (fallback)
				if tenantID == uuid.Nil {
					tenantID = headerTenantID
				}
			}
		} else if domain := c.GetHeader("X-Tenant-Domain"); domain != "" {
			// Try X-Tenant-Domain header
			tenant, err := tenantRepo.GetByDomain(c.Request.Context(), domain)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "tenant_not_found",
					"message": "Tenant not found for domain: " + domain,
				})
				c.Abort()
				return
			}
			tenantID = tenant.ID
		} else if tenantIDStr := c.Query("tenant_id"); tenantIDStr != "" {
			// Try tenant_id query parameter
			tenantID, err = uuid.Parse(tenantIDStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "invalid_tenant_id",
					"message": "Invalid tenant ID format in query parameter",
				})
				c.Abort()
				return
			}
		} else if domain := c.Query("domain"); domain != "" {
			// Try domain query parameter
			tenant, err := tenantRepo.GetByDomain(c.Request.Context(), domain)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "tenant_not_found",
					"message": "Tenant not found for domain: " + domain,
				})
				c.Abort()
				return
			}
			tenantID = tenant.ID
		} else {
			// Try to extract from subdomain (e.g., tenant.example.com)
			host := c.GetHeader("Host")
			if host != "" {
				parts := strings.Split(host, ".")
				if len(parts) >= 2 {
					subdomain := parts[0]
					tenant, err := tenantRepo.GetByDomain(c.Request.Context(), subdomain)
					if err == nil {
						tenantID = tenant.ID
					}
				}
			}
		}

		// If tenant ID is still not found, check if user is SYSTEM user
		// SYSTEM users can access tenant-scoped endpoints if they provide tenant context
		// TENANT users always need tenant context
		if tenantID == uuid.Nil {
			// Check if this is a tenant management endpoint (allowed without tenant context)
			path := c.Request.URL.Path
			if strings.HasPrefix(path, "/api/v1/tenants") && c.Request.Method == "POST" {
				// Creating a tenant doesn't require tenant context
				c.Next()
				return
			}

			// Check if this is a user endpoint - system users can be accessed without tenant context
			// The handler will determine if it's a system user or tenant user
			if strings.HasPrefix(path, "/api/v1/users/") {
				principalType, exists := c.Get("principal_type")
				if exists && principalType == "SYSTEM" {
					// SYSTEM users can access user endpoints without tenant context
					// The handler will check if it's a system user or tenant user
					c.Next()
					return
				}
			}

			// Check if this is a role assignment endpoint - system role assignment doesn't require tenant
			// The handler will check if it's a system role and handle accordingly
			if strings.Contains(path, "/users/") && strings.Contains(path, "/roles/") && c.Request.Method == "POST" {
				// Allow SYSTEM users to assign roles without tenant context
				// The handler will determine if it's a system role or tenant role
				principalType, exists := c.Get("principal_type")
				if exists && principalType == "SYSTEM" {
					// SYSTEM users can assign system roles without tenant context
					// The handler will check if it's a system role
					c.Next()
					return
				}
			}

			// Check if user is SYSTEM user (from JWT claims set by JWTAuthMiddleware)
			principalType, exists := c.Get("principal_type")
			if exists && principalType == "SYSTEM" {
				// SYSTEM users can access tenant-scoped endpoints, but they need to provide tenant context
				// If no tenant context is provided, return error asking for tenant selection
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "tenant_required",
					"message": "Tenant ID must be provided via X-Tenant-ID header for tenant-scoped operations. SYSTEM users must select a tenant context to access tenant-scoped resources.",
				})
				c.Abort()
				return
			}

			// For TENANT users, tenant context is always required
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "tenant_required",
				"message": "Tenant ID or domain must be provided via X-Tenant-ID, X-Tenant-Domain header, query parameter, or subdomain",
			})
			c.Abort()
			return
		}

		// Verify tenant exists and is active
		tenant, err := tenantRepo.GetByID(c.Request.Context(), tenantID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "tenant_not_found",
				"message": "Tenant not found",
			})
			c.Abort()
			return
		}

		if !tenant.IsActive() {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "tenant_inactive",
				"message": "Tenant is not active",
			})
			c.Abort()
			return
		}

		// Store tenant context
		c.Set(TenantContextKey, tenantID)
		c.Set("tenant", tenant)

		c.Next()
	}
}

// GetTenantID retrieves tenant ID from context
func GetTenantID(c *gin.Context) (uuid.UUID, bool) {
	tenantID, exists := c.Get(TenantContextKey)
	if !exists {
		return uuid.Nil, false
	}

	id, ok := tenantID.(uuid.UUID)
	return id, ok
}

// RequireTenant is a helper that ensures tenant context exists
func RequireTenant(c *gin.Context) (uuid.UUID, bool) {
	tenantID, exists := GetTenantID(c)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "tenant_required",
			"message": "Tenant context is required",
		})
		c.Abort()
		return uuid.Nil, false
	}
	return tenantID, true
}

