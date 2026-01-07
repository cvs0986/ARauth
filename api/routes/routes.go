package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/nuage-identity/iam/api/handlers"
	"github.com/nuage-identity/iam/api/middleware"
	"github.com/nuage-identity/iam/storage/interfaces"
	"go.uber.org/zap"
)

// SetupRoutes configures all routes
func SetupRoutes(router *gin.Engine, logger *zap.Logger, userHandler *handlers.UserHandler, authHandler *handlers.AuthHandler, mfaHandler *handlers.MFAHandler, tenantHandler *handlers.TenantHandler, roleHandler *handlers.RoleHandler, permissionHandler *handlers.PermissionHandler, tenantRepo interfaces.TenantRepository) {
	// Global middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logging(logger))
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.RateLimit())

	// Health check
	healthHandler := handlers.NewHealthHandler()
	router.GET("/health", healthHandler.Check)

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

		// Tenant-scoped routes (require tenant context)
		tenantScoped := v1.Group("")
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

			// Auth routes (tenant-scoped)
			auth := tenantScoped.Group("/auth")
			{
				auth.POST("/login", authHandler.Login)
			}

			// MFA routes (tenant-scoped)
			mfa := tenantScoped.Group("/mfa")
			{
				mfa.POST("/enroll", mfaHandler.Enroll)
				mfa.POST("/verify", mfaHandler.Verify)
				mfa.POST("/challenge", mfaHandler.Challenge)
				mfa.POST("/challenge/verify", mfaHandler.VerifyChallenge)
			}

			// Role routes (tenant-scoped)
			roles := tenantScoped.Group("/roles")
			{
				roles.POST("", roleHandler.Create)
				roles.GET("", roleHandler.List)
				roles.GET("/:id", roleHandler.GetByID)
				roles.PUT("/:id", roleHandler.Update)
				roles.DELETE("/:id", roleHandler.Delete)
				roles.GET("/:role_id/permissions", roleHandler.GetRolePermissions)
				roles.POST("/:role_id/permissions/:permission_id", roleHandler.AssignPermissionToRole)
				roles.DELETE("/:role_id/permissions/:permission_id", roleHandler.RemovePermissionFromRole)
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

		}
	}
}

