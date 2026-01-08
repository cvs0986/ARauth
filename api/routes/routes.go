package routes

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/arauth-identity/iam/api/handlers"
	"github.com/arauth-identity/iam/api/middleware"
	"github.com/arauth-identity/iam/auth/token"
	"github.com/arauth-identity/iam/internal/cache"
	"github.com/arauth-identity/iam/storage/interfaces"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Helper functions to extract database and Redis clients
func getDB(db interface{}) *sql.DB {
	if db == nil {
		return nil
	}
	if sqlDB, ok := db.(*sql.DB); ok {
		return sqlDB
	}
	return nil
}

func getRedis(redisClient interface{}) *redis.Client {
	if redisClient == nil {
		return nil
	}
	if client, ok := redisClient.(*redis.Client); ok {
		return client
	}
	return nil
}

// SetupRoutes configures all routes
func SetupRoutes(router *gin.Engine, logger *zap.Logger, userHandler *handlers.UserHandler, authHandler *handlers.AuthHandler, mfaHandler *handlers.MFAHandler, tenantHandler *handlers.TenantHandler, roleHandler *handlers.RoleHandler, permissionHandler *handlers.PermissionHandler, systemHandler *handlers.SystemHandler, tenantRepo interfaces.TenantRepository, cacheClient *cache.Cache, db interface{}, redisClient interface{}, tokenService interface{}) {
	// Global middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logging(logger))
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.RateLimit(cacheClient))

	// Health check
	healthHandler := handlers.NewHealthHandlerWithDeps(getDB(db), cacheClient, getRedis(redisClient))
	router.GET("/health", healthHandler.Check)
	router.GET("/health/live", healthHandler.Liveness)
	router.GET("/health/ready", healthHandler.Readiness)

	// Metrics endpoint
	SetupMetricsRoutes(router)

	// System API routes (for SYSTEM users only)
	systemAPI := router.Group("/system")
	{
		// System routes require JWT authentication and SYSTEM principal type
		if ts, ok := tokenService.(token.ServiceInterface); ok {
			systemAPI.Use(middleware.JWTAuthMiddleware(ts))
			systemAPI.Use(middleware.RequireSystemUser(ts))
		}

		// Tenant management (system admin only)
		systemTenants := systemAPI.Group("/tenants")
		{
			systemTenants.GET("", systemHandler.ListTenants)
			systemTenants.POST("", middleware.RequireSystemPermission("tenant", "create"), systemHandler.CreateTenant)
			systemTenants.GET("/:id", middleware.RequireSystemPermission("tenant", "read"), systemHandler.GetTenant)
			systemTenants.PUT("/:id", middleware.RequireSystemPermission("tenant", "update"), systemHandler.UpdateTenant)
			systemTenants.DELETE("/:id", middleware.RequireSystemPermission("tenant", "delete"), systemHandler.DeleteTenant)
			systemTenants.POST("/:id/suspend", middleware.RequireSystemPermission("tenant", "suspend"), systemHandler.SuspendTenant)
			systemTenants.POST("/:id/resume", middleware.RequireSystemPermission("tenant", "resume"), systemHandler.ResumeTenant)
			
			// Tenant settings management (system admin only)
			systemTenants.GET("/:id/settings", middleware.RequireSystemPermission("tenant", "configure"), systemHandler.GetTenantSettings)
			systemTenants.PUT("/:id/settings", middleware.RequireSystemPermission("tenant", "configure"), systemHandler.UpdateTenantSettings)
		}

		// System settings management (future)
		// systemAPI.GET("/settings", systemHandler.GetSystemSettings)
		// systemAPI.PUT("/settings", systemHandler.UpdateSystemSettings)
	}

	// API v1 routes
	v1 := router.Group("/api/v1")
	{

		// Tenant routes (public - tenant creation doesn't require tenant context)
		tenants := v1.Group("/tenants")
		{
			tenants.POST("", tenantHandler.Create)
			tenants.GET("/domain/:domain", tenantHandler.GetByDomain)
			tenants.GET("/:id", tenantHandler.GetByID)
			tenants.PUT("/:id", tenantHandler.Update)
			tenants.DELETE("/:id", tenantHandler.Delete)
			tenants.GET("", tenantHandler.List)
		}

		// Auth routes (public - no tenant middleware required)
		// SYSTEM users can login without tenant, TENANT users can provide tenant_id in request
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/revoke", authHandler.RevokeToken)
		}

		// MFA challenge endpoints (public - called during login flow before token is issued)
		mfaPublic := v1.Group("/mfa")
		{
			mfaPublic.POST("/challenge", mfaHandler.Challenge)
			mfaPublic.POST("/challenge/verify", mfaHandler.VerifyChallenge)
		}

		// Tenant-scoped routes (require tenant context)
		// These routes can be accessed by:
		// 1. TENANT users (automatically use their tenant from JWT token)
		// 2. SYSTEM users (must provide X-Tenant-ID header to select tenant context)
		tenantScoped := v1.Group("")
		// Apply JWT authentication middleware first
		if ts, ok := tokenService.(token.ServiceInterface); ok {
			tenantScoped.Use(middleware.JWTAuthMiddleware(ts))
			// Allow both SYSTEM and TENANT users to access tenant-scoped routes
			// RequireTenantUser is removed - TenantMiddleware will handle tenant context extraction
		}
		tenantScoped.Use(middleware.TenantMiddleware(tenantRepo))
		{
			// User routes (tenant-scoped)
			users := tenantScoped.Group("/users")
			{
				users.POST("", userHandler.Create)
				users.GET("", userHandler.List)
				users.GET("/:id", userHandler.GetByID)
				users.PUT("/:id", userHandler.Update)
				users.DELETE("/:id", userHandler.Delete)
			}

			// MFA routes (tenant-scoped - require authentication)
			mfa := tenantScoped.Group("/mfa")
			{
				mfa.POST("/enroll", mfaHandler.Enroll)
				mfa.POST("/verify", mfaHandler.Verify)
			}

			// Role routes (tenant-scoped)
			// Note: More specific routes (with /permissions) must come before generic :id routes
			roles := tenantScoped.Group("/roles")
			{
				roles.POST("", roleHandler.Create)
				roles.GET("", roleHandler.List)
				// Permission routes (must come before :id routes to avoid conflict)
				roles.GET("/:id/permissions", roleHandler.GetRolePermissions)
				roles.POST("/:id/permissions/:permission_id", roleHandler.AssignPermissionToRole)
				roles.DELETE("/:id/permissions/:permission_id", roleHandler.RemovePermissionFromRole)
				// Generic role routes
				roles.GET("/:id", roleHandler.GetByID)
				roles.PUT("/:id", roleHandler.Update)
				roles.DELETE("/:id", roleHandler.Delete)
			}

			// Permission routes (tenant-scoped)
			permissions := tenantScoped.Group("/permissions")
			{
				permissions.POST("", permissionHandler.Create)
				permissions.GET("", permissionHandler.List)
				permissions.GET("/:id", permissionHandler.GetByID)
				permissions.PUT("/:id", permissionHandler.Update)
				permissions.DELETE("/:id", permissionHandler.Delete)
			}

			// Tenant settings routes (tenant-scoped - TENANT users can access their own settings)
			// Route: GET/PUT /api/v1/tenant/settings (uses tenant from context)
			tenantScoped.GET("/tenant/settings", systemHandler.GetTenantSettingsFromContext)
			tenantScoped.PUT("/tenant/settings", systemHandler.UpdateTenantSettingsFromContext)

		}
	}
}

