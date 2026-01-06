package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/nuage-identity/iam/api/handlers"
	"github.com/nuage-identity/iam/api/middleware"
	"go.uber.org/zap"
)

// SetupRoutes configures all routes
func SetupRoutes(router *gin.Engine, logger *zap.Logger) {
	// Middleware
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
		// Auth routes (to be implemented)
		auth := v1.Group("/auth")
		{
			_ = auth // Placeholder for auth routes
		}

		// User routes (to be wired up with dependency injection)
		_ = v1.Group("/users")

		// Tenant routes (to be implemented)
		tenants := v1.Group("/tenants")
		{
			_ = tenants // Placeholder for tenant routes
		}
	}
}

