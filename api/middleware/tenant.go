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
// 1. X-Tenant-ID header
// 2. X-Tenant-Domain header
// 3. tenant_id query parameter
// 4. domain query parameter
func TenantMiddleware(tenantRepo interfaces.TenantRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tenantID uuid.UUID
		var err error

		// Try X-Tenant-ID header first
		if tenantIDStr := c.GetHeader("X-Tenant-ID"); tenantIDStr != "" {
			tenantID, err = uuid.Parse(tenantIDStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "invalid_tenant_id",
					"message": "Invalid tenant ID format in X-Tenant-ID header",
				})
				c.Abort()
				return
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

