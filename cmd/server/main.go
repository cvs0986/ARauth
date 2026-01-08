package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/arauth-identity/iam/api/handlers"
	"github.com/arauth-identity/iam/api/routes"
	"github.com/arauth-identity/iam/auth/hydra"
	"github.com/arauth-identity/iam/auth/claims"
	"github.com/arauth-identity/iam/auth/login"
	"github.com/arauth-identity/iam/auth/mfa"
	"github.com/arauth-identity/iam/auth/token"
	"github.com/arauth-identity/iam/config/loader"
	"github.com/arauth-identity/iam/config/validator"
	"github.com/arauth-identity/iam/identity/permission"
	"github.com/arauth-identity/iam/identity/role"
	"github.com/arauth-identity/iam/identity/tenant"
	"github.com/arauth-identity/iam/identity/user"
	"github.com/arauth-identity/iam/internal/audit"
	"github.com/arauth-identity/iam/internal/cache"
	"github.com/arauth-identity/iam/internal/logger"
	"github.com/arauth-identity/iam/security/encryption"
	"github.com/arauth-identity/iam/security/totp"
	"github.com/arauth-identity/iam/storage/postgres"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	cfg, err := loader.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := validator.Validate(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.Init(
		cfg.Logging.Level,
		cfg.Logging.Format,
		cfg.Logging.Output,
		cfg.Logging.FilePath,
		cfg.Logging.MaxSize,
		cfg.Logging.MaxBackups,
		cfg.Logging.MaxAge,
	); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		_ = logger.Sync() // Ignore sync errors on shutdown
	}()

	logger.Logger.Info("Starting ARauth Identity IAM API",
		zap.String("version", "0.1.0"),
		zap.String("port", fmt.Sprintf("%d", cfg.Server.Port)),
	)

	// Connect to database
	db, err := postgres.NewConnection(&cfg.Database)
	if err != nil {
		logger.Logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	logger.Logger.Info("Database connection established")

	// Connect to Redis
	redisClient, err := postgres.NewRedisConnection(&cfg.Redis)
	if err != nil {
		logger.Logger.Warn("Failed to connect to Redis", zap.Error(err))
		logger.Logger.Info("Continuing without Redis cache")
		redisClient = nil
	} else {
		defer redisClient.Close()
		logger.Logger.Info("Redis connection established")
	}

	// Initialize cache
	var cacheClient *cache.Cache
	if redisClient != nil {
		cacheClient = cache.NewCache(redisClient)
	}

	// Initialize repositories
	tenantRepo := postgres.NewTenantRepository(db)
	userRepo := postgres.NewUserRepository(db)
	credentialRepo := postgres.NewCredentialRepository(db)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(db)
	tenantSettingsRepo := postgres.NewTenantSettingsRepository(db)
	mfaRecoveryCodeRepo := postgres.NewMFARecoveryCodeRepository(db)
	auditRepo := postgres.NewAuditRepository(db)
	roleRepo := postgres.NewRoleRepository(db)
	permissionRepo := postgres.NewPermissionRepository(db)
	systemRoleRepo := postgres.NewSystemRoleRepository(db) // NEW: System role repository

	// Initialize audit logger
	auditLogger := audit.NewLogger(auditRepo)

	// Initialize encryption (for MFA secrets)
	encryptionKey := []byte(cfg.Security.EncryptionKey)
	if len(encryptionKey) != 32 {
		logger.Logger.Fatal("Encryption key must be exactly 32 bytes (AES-256)")
	}
	encryptor, err := encryption.NewEncryptor(encryptionKey)
	if err != nil {
		logger.Logger.Fatal("Failed to initialize encryptor", zap.Error(err))
	}

	// Initialize TOTP generator
	totpIssuer := cfg.Security.TOTPIssuer
	if totpIssuer == "" {
		totpIssuer = "ARauth Identity"
	}
	totpGenerator := totp.NewGenerator(totpIssuer)

	// Initialize Hydra client
	hydraClient := hydra.NewClient(cfg.Hydra.AdminURL)

	// Initialize MFA session manager
	var mfaSessionManager *mfa.SessionManager
	if cacheClient != nil {
		mfaSessionManager = mfa.NewSessionManager(cacheClient)
	} else {
		// Create a no-op session manager if Redis is not available
		// In production, Redis should be required for MFA
		logger.Logger.Warn("Redis not available - MFA sessions will not persist across restarts")
		mfaSessionManager = mfa.NewSessionManager(cache.NewCache(nil)) // Will fail gracefully
	}

	// Initialize claims builder
	claimsBuilder := claims.NewBuilder(roleRepo, permissionRepo, systemRoleRepo)

	// Initialize token lifetime resolver
	lifetimeResolver := token.NewLifetimeResolver(&cfg.Security, tenantSettingsRepo)

	// Initialize token service
	tokenService, err := token.NewService(&cfg.Security, lifetimeResolver)
	if err != nil {
		logger.Logger.Fatal("Failed to initialize token service", zap.Error(err))
	}

	// Initialize services
	tenantService := tenant.NewService(tenantRepo)
	userService := user.NewService(userRepo, credentialRepo) // Pass credentialRepo to create credentials automatically
	loginService := login.NewService(userRepo, credentialRepo, refreshTokenRepo, tenantSettingsRepo, hydraClient, claimsBuilder, tokenService, lifetimeResolver)
	mfaService := mfa.NewService(userRepo, credentialRepo, mfaRecoveryCodeRepo, totpGenerator, encryptor, mfaSessionManager)
	roleService := role.NewService(roleRepo, permissionRepo)
	permissionService := permission.NewService(permissionRepo)

	// Initialize refresh service
	refreshService := token.NewRefreshService(tokenService, refreshTokenRepo, userRepo, claimsBuilder, lifetimeResolver)

	// Initialize handlers
	tenantHandler := handlers.NewTenantHandler(tenantService)
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(loginService, refreshService)
	mfaHandler := handlers.NewMFAHandler(mfaService, auditLogger, tokenService, claimsBuilder, userRepo, lifetimeResolver)
	roleHandler := handlers.NewRoleHandler(roleService)
	permissionHandler := handlers.NewPermissionHandler(permissionService)
	systemHandler := handlers.NewSystemHandler(tenantService, tenantRepo, tenantSettingsRepo) // NEW: System handler with tenant settings

	// Set Gin mode
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.New()

	// Setup routes with dependencies
	routes.SetupRoutes(router, logger.Logger, userHandler, authHandler, mfaHandler, tenantHandler, roleHandler, permissionHandler, systemHandler, tenantRepo, cacheClient, db, redisClient, tokenService)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		logger.Logger.Info("Server starting",
			zap.String("address", srv.Addr),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Logger.Info("Server exited")
}

