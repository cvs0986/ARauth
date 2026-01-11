package routes

import (
	"database/sql"

	"github.com/arauth-identity/iam/api/handlers"
	"github.com/arauth-identity/iam/api/middleware"
	"github.com/arauth-identity/iam/auth/token"
	"github.com/arauth-identity/iam/identity/ratelimit"
	"github.com/arauth-identity/iam/identity/scim"
	"github.com/arauth-identity/iam/internal/cache"
	"github.com/arauth-identity/iam/observability/security_events"
	"github.com/arauth-identity/iam/storage/interfaces"
	"github.com/gin-gonic/gin"
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
func SetupRoutes(router *gin.Engine, logger *zap.Logger, userHandler *handlers.UserHandler, authHandler *handlers.AuthHandler, mfaHandler *handlers.MFAHandler, tenantHandler *handlers.TenantHandler, roleHandler *handlers.RoleHandler, permissionHandler *handlers.PermissionHandler, systemHandler *handlers.SystemHandler, capabilityHandler *handlers.CapabilityHandler, auditHandler *handlers.AuditHandler, federationHandler *handlers.FederationHandler, webhookHandler *handlers.WebhookHandler, identityLinkingHandler *handlers.IdentityLinkingHandler, introspectionHandler *handlers.IntrospectionHandler, impersonationHandler *handlers.ImpersonationHandler, oauthScopeHandler *handlers.OAuthScopeHandler, scimHandler *handlers.SCIMHandler, scimTokenHandler *handlers.SCIMTokenHandler, scimTokenService scim.TokenServiceInterface, invitationHandler *handlers.InvitationHandler, sessionHandler *handlers.SessionHandler, oauthClientHandler *handlers.OAuthClientHandler, tenantRepo interfaces.TenantRepository, cacheClient *cache.Cache, db interface{}, redisClient interface{}, tokenService interface{}, rateLimiter ratelimit.Limiter, eventLogger security_events.Logger) {
	// Global middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logging(logger))
	router.Use(middleware.Recovery(logger))

	// Apply multi-tier rate limiting if available
	if rateLimiter != nil {
		router.Use(middleware.MultiTierRateLimit(rateLimiter, eventLogger))
	} else {
		// Fallback to legacy rate limiting (development only)
		router.Use(middleware.RateLimit(cacheClient))
	}

	// Health check (no rate limiting)
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
			systemAPI.Use(middleware.JWTAuthMiddleware(ts, eventLogger))
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

			// Tenant capability assignment (system admin only)
			systemTenants.GET("/:id/capabilities", middleware.RequireSystemPermission("tenant", "configure"), capabilityHandler.GetTenantCapabilities)
			systemTenants.PUT("/:id/capabilities/:key", middleware.RequireSystemPermission("tenant", "configure"), capabilityHandler.SetTenantCapability)
			systemTenants.DELETE("/:id/capabilities/:key", middleware.RequireSystemPermission("tenant", "configure"), capabilityHandler.DeleteTenantCapability)
			systemTenants.GET("/:id/capabilities/evaluation", middleware.RequireSystemPermission("tenant", "read"), capabilityHandler.EvaluateTenantCapabilities)
		}

		// System capability management (system owner only)
		systemCapabilities := systemAPI.Group("/capabilities")
		{
			systemCapabilities.GET("", capabilityHandler.ListSystemCapabilities)
			systemCapabilities.GET("/:key", capabilityHandler.GetSystemCapability)
			systemCapabilities.PUT("/:key", middleware.RequireSystemPermission("system", "configure"), capabilityHandler.UpdateSystemCapability)
		}

		// System users management (system admin only)
		systemUsers := systemAPI.Group("/users")
		{
			systemUsers.GET("", userHandler.ListSystem)
			systemUsers.POST("", userHandler.CreateSystem)
			systemUsers.POST("/:id/change-password", userHandler.ChangePassword)
		}

		// System roles management (system admin only) - show predefined system roles
		systemRoles := systemAPI.Group("/roles")
		{
			systemRoles.GET("", roleHandler.ListSystem)
		}

		// System permissions management (system admin only) - show predefined system permissions
		systemPermissions := systemAPI.Group("/permissions")
		{
			systemPermissions.GET("", permissionHandler.ListSystem)
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
			mfaPublic.POST("/enroll/login", mfaHandler.EnrollForLogin)
		}

		// Tenant-scoped routes (require tenant context)
		// These routes can be accessed by:
		// 1. TENANT users (automatically use their tenant from JWT token)
		// 2. SYSTEM users (must provide X-Tenant-ID header to select tenant context)
		tenantScoped := v1.Group("")
		// Apply JWT authentication middleware first
		if ts, ok := tokenService.(token.ServiceInterface); ok {
			tenantScoped.Use(middleware.JWTAuthMiddleware(ts, eventLogger))
			// Allow both SYSTEM and TENANT users to access tenant-scoped routes
			// RequireTenantUser is removed - TenantMiddleware will handle tenant context extraction
		}
		tenantScoped.Use(middleware.TenantMiddleware(tenantRepo))
		{
			// User routes (tenant-scoped)
			// Note: More specific routes (with /roles) must come before generic :id routes
			users := tenantScoped.Group("/users")
			{
				users.POST("", middleware.RequirePermission("users", "create", eventLogger), userHandler.Create)
				users.GET("", middleware.RequirePermission("users", "read", eventLogger), userHandler.List)
				// User roles routes (must come before /:id to avoid route conflict)
				// Use :id instead of :user_id to avoid wildcard name conflict
				users.GET("/:id/roles", middleware.RequirePermission("users", "read", eventLogger), roleHandler.GetUserRoles)
				users.POST("/:id/roles/:role_id", middleware.RequirePermission("users", "roles:assign", eventLogger), roleHandler.AssignRoleToUser)
				users.DELETE("/:id/roles/:role_id", middleware.RequirePermission("users", "roles:remove", eventLogger), roleHandler.RemoveRoleFromUser)
				// User permissions route (must come before /:id)
				users.GET("/:id/permissions", middleware.RequirePermission("users", "read", eventLogger), userHandler.GetUserPermissions)
				// User capabilities routes (must come before /:id)
				users.GET("/:id/capabilities", middleware.RequirePermission("users", "read", eventLogger), capabilityHandler.GetUserCapabilities)
				users.GET("/:id/capabilities/:key", middleware.RequirePermission("users", "read", eventLogger), capabilityHandler.GetUserCapability)
				users.POST("/:id/capabilities/:key/enroll", middleware.RequirePermission("users", "capabilities:manage", eventLogger), capabilityHandler.EnrollUserCapability)
				users.DELETE("/:id/capabilities/:key", middleware.RequirePermission("users", "capabilities:manage", eventLogger), capabilityHandler.UnenrollUserCapability)
				// User identity linking routes (must come before /:id)
				users.GET("/:id/identities", middleware.RequirePermission("users", "read", eventLogger), identityLinkingHandler.GetUserIdentities)
				users.POST("/:id/identities", middleware.RequirePermission("users", "identities:link", eventLogger), identityLinkingHandler.LinkIdentity)
				users.DELETE("/:id/identities/:identity_id", middleware.RequirePermission("users", "identities:unlink", eventLogger), identityLinkingHandler.UnlinkIdentity)
				users.PUT("/:id/identities/:identity_id/primary", middleware.RequirePermission("users", "identities:manage", eventLogger), identityLinkingHandler.SetPrimaryIdentity)
				users.POST("/:id/identities/:identity_id/verify", middleware.RequirePermission("users", "identities:verify", eventLogger), identityLinkingHandler.VerifyIdentity)
				// Generic user routes
				users.POST("/:id/change-password", middleware.RequirePermission("users", "update", eventLogger), userHandler.ChangePassword)
				users.GET("/:id", middleware.RequirePermission("users", "read", eventLogger), userHandler.GetByID)
				users.PUT("/:id", middleware.RequirePermission("users", "update", eventLogger), userHandler.Update)
				users.DELETE("/:id", middleware.RequirePermission("users", "delete", eventLogger), userHandler.Delete)
			}

			// Session routes (tenant-scoped)
			sessions := tenantScoped.Group("/sessions")
			{
				sessions.GET("", middleware.RequirePermission("sessions", "read", eventLogger), sessionHandler.ListSessions)
				sessions.POST("/:id/revoke", middleware.RequirePermission("sessions", "revoke", eventLogger), sessionHandler.RevokeSession)
			}

			// OAuth client routes (tenant-scoped)
			oauthClients := tenantScoped.Group("/oauth/clients")
			{
				oauthClients.POST("", middleware.RequirePermission("oauth", "clients:create", eventLogger), oauthClientHandler.CreateClient)
				oauthClients.GET("", middleware.RequirePermission("oauth", "clients:read", eventLogger), oauthClientHandler.ListClients)
				oauthClients.GET("/:id", middleware.RequirePermission("oauth", "clients:read", eventLogger), oauthClientHandler.GetClient)
				oauthClients.POST("/:id/rotate-secret", middleware.RequirePermission("oauth", "clients:rotate-secret", eventLogger), oauthClientHandler.RotateSecret)
				oauthClients.DELETE("/:id", middleware.RequirePermission("oauth", "clients:delete", eventLogger), oauthClientHandler.DeleteClient)
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
				roles.POST("", middleware.RequirePermission("roles", "create", eventLogger), roleHandler.Create)
				roles.GET("", middleware.RequirePermission("roles", "read", eventLogger), roleHandler.List)
				// Permission routes (must come before :id routes to avoid conflict)
				roles.GET("/:id/permissions", middleware.RequirePermission("roles", "read", eventLogger), roleHandler.GetRolePermissions)
				roles.POST("/:id/permissions/:permission_id", middleware.RequirePermission("roles", "permissions:assign", eventLogger), roleHandler.AssignPermissionToRole)
				roles.DELETE("/:id/permissions/:permission_id", middleware.RequirePermission("roles", "permissions:remove", eventLogger), roleHandler.RemovePermissionFromRole)
				// Generic role routes
				roles.GET("/:id", middleware.RequirePermission("roles", "read", eventLogger), roleHandler.GetByID)
				roles.PUT("/:id", middleware.RequirePermission("roles", "update", eventLogger), roleHandler.Update)
				roles.DELETE("/:id", middleware.RequirePermission("roles", "delete", eventLogger), roleHandler.Delete)
			}

			// Permission routes (tenant-scoped)
			permissions := tenantScoped.Group("/permissions")
			{
				permissions.POST("", middleware.RequirePermission("permissions", "create", eventLogger), permissionHandler.Create)
				permissions.GET("", middleware.RequirePermission("permissions", "read", eventLogger), permissionHandler.List)
				permissions.GET("/:id", middleware.RequirePermission("permissions", "read", eventLogger), permissionHandler.GetByID)
				permissions.PUT("/:id", middleware.RequirePermission("permissions", "update", eventLogger), permissionHandler.Update)
				permissions.DELETE("/:id", middleware.RequirePermission("permissions", "delete", eventLogger), permissionHandler.Delete)
			}

			// Tenant settings routes (tenant-scoped - TENANT users can access their own settings)
			// Route: GET/PUT /api/v1/tenant/settings (uses tenant from context)
			tenantScoped.GET("/tenant/settings", middleware.RequirePermission("tenant", "settings:read", eventLogger), systemHandler.GetTenantSettingsFromContext)
			tenantScoped.PUT("/tenant/settings", middleware.RequirePermission("tenant", "settings:update", eventLogger), systemHandler.UpdateTenantSettingsFromContext)

			// System capabilities viewing (tenant-scoped, read-only for TENANT users)
			tenantScoped.GET("/tenant/system-capabilities", capabilityHandler.ListSystemCapabilitiesFromContext)
			tenantScoped.GET("/tenant/system-capabilities/:key", capabilityHandler.GetSystemCapabilityFromContext)

			// Tenant capability viewing (tenant-scoped)
			tenantScoped.GET("/tenant/capabilities", capabilityHandler.GetTenantCapabilitiesFromContext)

			// Tenant feature enablement (tenant-scoped)
			tenantScoped.GET("/tenant/features", middleware.RequirePermission("tenant", "read", eventLogger), capabilityHandler.GetTenantFeatures)
			tenantScoped.PUT("/tenant/features/:key", middleware.RequirePermission("tenant", "features:manage", eventLogger), capabilityHandler.EnableTenantFeature)
			tenantScoped.DELETE("/tenant/features/:key", middleware.RequirePermission("tenant", "features:manage", eventLogger), capabilityHandler.DisableTenantFeature)

			// Tenant capability evaluation (tenant-scoped)
			tenantScoped.GET("/tenant/capabilities/evaluation", capabilityHandler.EvaluateTenantCapabilitiesFromContext)

			// Audit events routes (tenant-scoped)
			// Audit routes
			auditRoutes := tenantScoped.Group("/audit")
			{
				auditRoutes.GET("/events", middleware.RequirePermission("audit", "read", eventLogger), auditHandler.QueryEvents)
				auditRoutes.GET("/events/:id", middleware.RequirePermission("audit", "read", eventLogger), auditHandler.GetEvent)
				auditRoutes.GET("/export", middleware.RequirePermission("audit", "export", eventLogger), auditHandler.ExportEvents)
			}

			// Federation routes (Identity Providers)
			identityProviders := tenantScoped.Group("/identity-providers")
			{
				identityProviders.POST("", middleware.RequirePermission("federation", "create", eventLogger), federationHandler.CreateIdentityProvider)
				identityProviders.GET("", middleware.RequirePermission("federation", "read", eventLogger), federationHandler.ListIdentityProviders)
				identityProviders.GET("/:id", middleware.RequirePermission("federation", "read", eventLogger), federationHandler.GetIdentityProvider)
				identityProviders.PUT("/:id", middleware.RequirePermission("federation", "update", eventLogger), federationHandler.UpdateIdentityProvider)
				identityProviders.DELETE("/:id", middleware.RequirePermission("federation", "delete", eventLogger), federationHandler.DeleteIdentityProvider)
				identityProviders.POST("/:id/verify", middleware.RequirePermission("federation", "verify", eventLogger), federationHandler.VerifyIdentityProvider)
			}
		}

		// Federation authentication routes (public, no auth required for initiation)
		// These routes handle OIDC/SAML login flows
		federationAuth := v1.Group("/auth")
		{
			federationAuth.GET("/oidc/:provider_id/initiate", federationHandler.InitiateOIDCLogin)
			federationAuth.GET("/oidc/:provider_id/callback", federationHandler.HandleOIDCCallback)
			federationAuth.GET("/saml/:provider_id/initiate", federationHandler.InitiateSAMLLogin)
			federationAuth.POST("/saml/:provider_id/callback", federationHandler.HandleSAMLCallback)
		}

		// Token introspection endpoint (RFC 7662)
		// Requires authentication (client credentials or bearer token)
		v1.POST("/introspect", introspectionHandler.IntrospectToken)

		// Impersonation endpoints (tenant-scoped, requires admin permission)
		impersonation := tenantScoped.Group("/impersonation")
		{
			impersonation.POST("/users/:id/impersonate", middleware.RequirePermission("users", "impersonate", eventLogger), impersonationHandler.StartImpersonation)
			impersonation.GET("", middleware.RequirePermission("users", "impersonate", eventLogger), impersonationHandler.ListImpersonationSessions)
			impersonation.GET("/:session_id", middleware.RequirePermission("users", "impersonate", eventLogger), impersonationHandler.GetImpersonationSession)
			impersonation.DELETE("/:session_id", middleware.RequirePermission("users", "impersonate", eventLogger), impersonationHandler.EndImpersonation)
		}

		// OAuth Scope endpoints (tenant-scoped)
		oauthScopes := tenantScoped.Group("/oauth/scopes")
		{
			oauthScopes.POST("", middleware.RequirePermission("oauth", "scopes:create", eventLogger), oauthScopeHandler.CreateScope)
			oauthScopes.GET("", middleware.RequirePermission("oauth", "scopes:read", eventLogger), oauthScopeHandler.ListScopes)
			oauthScopes.GET("/:id", middleware.RequirePermission("oauth", "scopes:read", eventLogger), oauthScopeHandler.GetScope)
			oauthScopes.PUT("/:id", middleware.RequirePermission("oauth", "scopes:update", eventLogger), oauthScopeHandler.UpdateScope)
			oauthScopes.DELETE("/:id", middleware.RequirePermission("oauth", "scopes:delete", eventLogger), oauthScopeHandler.DeleteScope)
		}

		// SCIM 2.0 API routes (public, authenticated via Bearer token)
		scimV2 := router.Group("/scim/v2")
		{
			// SCIM discovery endpoints (no auth required)
			scimV2.GET("/ServiceProviderConfig", scimHandler.ServiceProviderConfig)
			scimV2.GET("/ResourceTypes", scimHandler.ResourceTypes)
			scimV2.GET("/Schemas", scimHandler.Schemas)

			// SCIM resource endpoints (require authentication)
			scimUsers := scimV2.Group("/Users")
			scimUsers.Use(middleware.SCIMAuthMiddleware(scimTokenService))
			scimUsers.Use(middleware.RequireSCIMScope("users"))
			{
				scimUsers.POST("", scimHandler.CreateUser)
				scimUsers.GET("", scimHandler.ListUsers)
				scimUsers.GET("/:id", scimHandler.GetUser)
				scimUsers.PUT("/:id", scimHandler.UpdateUser)
				scimUsers.DELETE("/:id", scimHandler.DeleteUser)
			}

			scimGroups := scimV2.Group("/Groups")
			scimGroups.Use(middleware.SCIMAuthMiddleware(scimTokenService))
			scimGroups.Use(middleware.RequireSCIMScope("groups"))
			{
				scimGroups.POST("", scimHandler.CreateGroup)
				scimGroups.GET("", scimHandler.ListGroups)
				scimGroups.GET("/:id", scimHandler.GetGroup)
				scimGroups.PUT("/:id", scimHandler.UpdateGroup)
				scimGroups.DELETE("/:id", scimHandler.DeleteGroup)
			}

			// Bulk operations
			scimBulk := scimV2.Group("/Bulk")
			scimBulk.Use(middleware.SCIMAuthMiddleware(scimTokenService))
			{
				scimBulk.POST("", scimHandler.BulkOperations)
			}
		}

		// SCIM Token Management routes (tenant-scoped)
		scimTokens := tenantScoped.Group("/scim/tokens")
		{
			scimTokens.POST("", middleware.RequirePermission("scim_tokens", "create", eventLogger), scimTokenHandler.CreateToken)
			scimTokens.GET("", middleware.RequirePermission("scim_tokens", "read", eventLogger), scimTokenHandler.ListTokens)
			scimTokens.GET("/:id", middleware.RequirePermission("scim_tokens", "read", eventLogger), scimTokenHandler.GetToken)
			scimTokens.POST("/:id/rotate", middleware.RequirePermission("scim_tokens", "write", eventLogger), scimTokenHandler.RotateToken)
			scimTokens.DELETE("/:id", middleware.RequirePermission("scim_tokens", "delete", eventLogger), scimTokenHandler.DeleteToken)
		}

		// System audit events route (SYSTEM users only - system-wide audit)
		if ts, ok := tokenService.(token.ServiceInterface); ok {
			systemAPI := router.Group("/system")
			systemAPI.Use(middleware.JWTAuthMiddleware(ts, eventLogger))
			systemAPI.Use(middleware.RequireSystemUser(ts))
			{
				systemAPI.GET("/audit/events", auditHandler.QueryEvents)
				systemAPI.GET("/audit/events/:id", auditHandler.GetEvent)
			}
		}
	}
}
