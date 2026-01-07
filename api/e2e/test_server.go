// +build e2e

package e2e

import (
	"database/sql"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/arauth-identity/iam/api/handlers"
	"github.com/arauth-identity/iam/api/middleware"
	"github.com/arauth-identity/iam/auth/claims"
	"github.com/arauth-identity/iam/auth/login"
	"github.com/arauth-identity/iam/auth/mfa"
	"github.com/arauth-identity/iam/identity/permission"
	"github.com/arauth-identity/iam/identity/role"
	"github.com/arauth-identity/iam/identity/tenant"
	"github.com/arauth-identity/iam/identity/user"
	"github.com/arauth-identity/iam/internal/audit"
	"github.com/arauth-identity/iam/internal/cache"
	"github.com/arauth-identity/iam/storage/postgres"
	"go.uber.org/zap/zaptest"
)

// setupTestServerWithAuth creates a test HTTP server with authentication routes
func setupTestServerWithAuth(db *sql.DB, cacheClient *cache.Cache, loginService login.ServiceInterface, mfaService mfa.ServiceInterface) (*httptest.Server, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup logger
	logger := zaptest.NewLogger(nil)

	// Setup repositories
	tenantRepo := postgres.NewTenantRepository(db)
	roleRepo := postgres.NewRoleRepository(db)
	permissionRepo := postgres.NewPermissionRepository(db)

	// Setup services
	userService := user.NewService(postgres.NewUserRepository(db))
	tenantService := tenant.NewService(tenantRepo)
	roleService := role.NewService(roleRepo, permissionRepo)
	permissionService := permission.NewService(permissionRepo)

	// Setup handlers
	auditLogger := audit.NewLogger(postgres.NewAuditRepository(db))
	authHandler := handlers.NewAuthHandler(loginService)
	mfaHandler := handlers.NewMFAHandler(mfaService, auditLogger)
	userHandler := handlers.NewUserHandler(userService)
	tenantHandler := handlers.NewTenantHandler(tenantService)
	roleHandler := handlers.NewRoleHandler(roleService)
	permissionHandler := handlers.NewPermissionHandler(permissionService)

	// Setup middleware
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.Logging(logger))
	router.Use(middleware.CORS())
	router.Use(middleware.TenantMiddleware(tenantRepo))

	// API routes
	api := router.Group("/api/v1")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
		}

		// MFA routes
		mfaRoutes := api.Group("/mfa")
		{
			mfaRoutes.POST("/enroll", mfaHandler.Enroll)
			mfaRoutes.POST("/challenge", mfaHandler.Challenge)
		}

		// Tenant routes
		tenants := api.Group("/tenants")
		{
			tenants.POST("", tenantHandler.Create)
			tenants.GET("/:id", tenantHandler.GetByID)
		}

		// User routes
		users := api.Group("/users")
		{
			users.POST("", userHandler.Create)
			users.GET("/:id", userHandler.GetByID)
		}

		// Role routes
		roles := api.Group("/roles")
		{
			roles.POST("", roleHandler.Create)
			roles.GET("/:id", roleHandler.GetByID)
		}

		// Permission routes
		permissions := api.Group("/permissions")
		{
			permissions.POST("", permissionHandler.Create)
			permissions.GET("/:id", permissionHandler.GetByID)
		}
	}

	// Create test server
	server := httptest.NewServer(router)
	return server, router
}

// setupTestServer creates a test HTTP server with all routes configured
func setupTestServer(db *sql.DB, cacheClient *cache.Cache) (*httptest.Server, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup logger
	logger := zaptest.NewLogger(nil)

	// Setup repositories
	userRepo := postgres.NewUserRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)
	roleRepo := postgres.NewRoleRepository(db)
	permissionRepo := postgres.NewPermissionRepository(db)

	// Setup services
	userService := user.NewService(userRepo)
	tenantService := tenant.NewService(tenantRepo)
	roleService := role.NewService(roleRepo, permissionRepo)
	permissionService := permission.NewService(permissionRepo)

	// Setup handlers
	userHandler := handlers.NewUserHandler(userService)
	tenantHandler := handlers.NewTenantHandler(tenantService)
	roleHandler := handlers.NewRoleHandler(roleService)
	permissionHandler := handlers.NewPermissionHandler(permissionService)

	// Setup middleware
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.Logging(logger))
	router.Use(middleware.CORS())
	router.Use(middleware.TenantMiddleware(tenantRepo))

	// API routes
	api := router.Group("/api/v1")
	{
		// Tenant routes
		tenants := api.Group("/tenants")
		{
			tenants.POST("", tenantHandler.Create)
			tenants.GET("/:id", tenantHandler.GetByID)
			tenants.GET("", tenantHandler.List)
		}

		// User routes
		users := api.Group("/users")
		{
			users.POST("", userHandler.Create)
			users.GET("/:id", userHandler.GetByID)
			users.GET("", userHandler.List)
		}

		// Role routes
		roles := api.Group("/roles")
		{
			roles.POST("", roleHandler.Create)
			roles.GET("/:id", roleHandler.GetByID)
			roles.GET("", roleHandler.List)
		}

		// Permission routes
		permissions := api.Group("/permissions")
		{
			permissions.POST("", permissionHandler.Create)
			permissions.GET("/:id", permissionHandler.GetByID)
			permissions.GET("", permissionHandler.List)
		}
	}

	// Create test server
	server := httptest.NewServer(router)
	return server, router
}

