package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/nuage-identity/iam/api/handlers"
	"github.com/nuage-identity/iam/api/middleware"
	"go.uber.org/zap"
)

// SetupRoutes configures all routes
func SetupRoutes(router *gin.Engine, logger *zap.Logger, userHandler *handlers.UserHandler, authHandler *handlers.AuthHandler, mfaHandler *handlers.MFAHandler) {
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
		// Auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
		}

		// MFA routes
		mfa := v1.Group("/mfa")
		{
			mfa.POST("/enroll", mfaHandler.Enroll)
			mfa.POST("/verify", mfaHandler.Verify)
			mfa.POST("/challenge", mfaHandler.Challenge)
			mfa.POST("/challenge/verify", mfaHandler.VerifyChallenge)
		}

		// User routes
		users := v1.Group("/users")
		{
			users.POST("", userHandler.Create)
			users.GET("", userHandler.List)
			users.GET("/:id", userHandler.GetByID)
			users.PUT("/:id", userHandler.Update)
			users.DELETE("/:id", userHandler.Delete)
		}

		// Tenant routes (to be implemented)
		tenants := v1.Group("/tenants")
		{
			_ = tenants // Placeholder for tenant routes
		}
	}
}

