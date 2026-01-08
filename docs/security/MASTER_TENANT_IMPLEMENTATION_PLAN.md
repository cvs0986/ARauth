# Master Tenant & Multi-Tenant Security Implementation Plan

## 📋 Executive Summary

This document outlines a comprehensive plan to implement a **Master Tenant** architecture similar to Keycloak's master realm concept. This addresses critical security vulnerabilities where tenant members can currently access and modify data from other tenants.

---

## 🚨 Current Security Issues

### Issue 1: Cross-Tenant Data Access
**Problem**: Any authenticated user can access tenant management endpoints and see/edit all tenants.

**Evidence**:
- `GET /api/v1/tenants` - Lists all tenants (no tenant isolation)
- `PUT /api/v1/tenants/:id` - Can edit any tenant
- `DELETE /api/v1/tenants/:id` - Can delete any tenant
- Users from Tenant A can see/modify Tenant B's data

**Impact**: **CRITICAL** - Complete breach of tenant isolation, data leakage, unauthorized modifications.

### Issue 2: No Super-Admin Role
**Problem**: No distinction between regular tenant users and system administrators.

**Impact**: **HIGH** - Cannot properly manage multi-tenant system, no way to bootstrap or manage system-wide settings.

### Issue 3: Tenant Settings Access
**Problem**: All users can potentially access system settings that should be tenant-specific or system-wide.

**Impact**: **MEDIUM** - Configuration tampering, security policy violations.

### Issue 4: No Bootstrap Process
**Problem**: No defined process to create the first master tenant and admin user.

**Impact**: **HIGH** - Manual setup required, potential for misconfiguration.

---

## ✅ Proposed Solution: Master Tenant Architecture

### Core Concept

1. **Master Tenant**: A special system tenant (ID: `00000000-0000-0000-0000-000000000000` or flagged as `is_master = true`)
2. **Super-Admin Role**: A special role (`system:admin` or `master:admin`) that grants cross-tenant access
3. **Tenant Isolation**: Regular users can only access their own tenant's data
4. **Bootstrap Process**: Automated creation of master tenant and admin user on first run

### Why This Solution is Best

1. **Security**: Complete tenant isolation by default, explicit permissions for cross-tenant access
2. **Scalability**: Supports unlimited tenants with proper isolation
3. **Flexibility**: Can have multiple super-admins, role-based access control
4. **Industry Standard**: Similar to Keycloak (master realm), Auth0 (management API), AWS Organizations
5. **Auditability**: Clear distinction between tenant-scoped and system-scoped operations
6. **Maintainability**: Clear separation of concerns, easier to reason about security

---

## 🏗️ Architecture Design

### 1. Database Schema Changes

#### 1.1 Update `tenants` Table

```sql
-- Migration: 000013_add_master_tenant_support.up.sql
ALTER TABLE tenants 
ADD COLUMN is_master BOOLEAN DEFAULT FALSE NOT NULL,
ADD COLUMN parent_tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL,
ADD COLUMN metadata JSONB DEFAULT '{}'::jsonb;

-- Create index for master tenant lookup
CREATE INDEX idx_tenants_is_master ON tenants(is_master) WHERE is_master = TRUE;
CREATE INDEX idx_tenants_parent_tenant_id ON tenants(parent_tenant_id);

-- Add constraint: only one master tenant
CREATE UNIQUE INDEX idx_tenants_single_master ON tenants(is_master) WHERE is_master = TRUE;
```

#### 1.2 Update `roles` Table

```sql
-- Migration: 000014_add_system_roles.up.sql
ALTER TABLE roles
ADD COLUMN is_system_role BOOLEAN DEFAULT FALSE NOT NULL,
ADD COLUMN scope VARCHAR(50) DEFAULT 'tenant' NOT NULL CHECK (scope IN ('tenant', 'system'));

-- Create index for system roles
CREATE INDEX idx_roles_is_system_role ON roles(is_system_role) WHERE is_system_role = TRUE;
CREATE INDEX idx_roles_scope ON roles(scope);
```

#### 1.3 Create `system_permissions` Table

```sql
-- Migration: 000015_create_system_permissions.up.sql
CREATE TABLE system_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource VARCHAR(255) NOT NULL,
    action VARCHAR(255) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT TRUE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(resource, action)
);

CREATE INDEX idx_system_permissions_resource_action ON system_permissions(resource, action);
```

#### 1.4 Update `user_roles` Table

```sql
-- Migration: 000016_update_user_roles_for_system.up.sql
-- No changes needed, but ensure tenant_id is properly set
-- System roles should have tenant_id = master_tenant_id
```

### 2. Authorization Middleware

#### 2.1 Enhanced Tenant Middleware

**File**: `api/middleware/tenant.go`

```go
// RequireTenantOrMaster checks if user has access to the requested tenant
func RequireTenantOrMaster(c *gin.Context) (uuid.UUID, bool) {
    // Get tenant ID from context (set by TenantMiddleware)
    tenantID, exists := c.Get("tenantID")
    if !exists {
        RespondWithError(c, http.StatusBadRequest, "tenant_required", 
            "Tenant ID must be provided")
        return uuid.Nil, false
    }

    requestedTenantID := tenantID.(uuid.UUID)
    
    // Get user from context (set by JWTAuth middleware)
    userID, _ := c.Get("userID")
    userTenantID, _ := c.Get("userTenantID") // From JWT claims
    roles, _ := c.Get("roles").([]string)
    permissions, _ := c.Get("permissions").([]string)

    // Check if user is super-admin (has system:admin role or system:* permissions)
    isSuperAdmin := hasSystemAdminAccess(roles, permissions)

    // If super-admin, allow access to any tenant
    if isSuperAdmin {
        return requestedTenantID, true
    }

    // Regular users can only access their own tenant
    if userTenantID != requestedTenantID {
        RespondWithError(c, http.StatusForbidden, "access_denied",
            "You do not have permission to access this tenant")
        return uuid.Nil, false
    }

    return requestedTenantID, true
}

func hasSystemAdminAccess(roles []string, permissions []string) bool {
    // Check for system:admin role
    for _, role := range roles {
        if role == "system:admin" || strings.HasPrefix(role, "system:") {
            return true
        }
    }

    // Check for system:* permissions
    for _, perm := range permissions {
        if perm == "system:*" || strings.HasPrefix(perm, "system:") {
            return true
        }
    }

    return false
}
```

#### 2.2 Resource Authorization Middleware

**File**: `api/middleware/authorization.go` (NEW)

```go
package middleware

import (
    "net/http"
    "strings"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

// RequirePermission checks if user has required permission
func RequirePermission(resource, action string) gin.HandlerFunc {
    return func(c *gin.Context) {
        permissions, exists := c.Get("permissions")
        if !exists {
            RespondWithError(c, http.StatusForbidden, "access_denied",
                "Permission check failed: no permissions found")
            c.Abort()
            return
        }

        perms := permissions.([]string)
        requiredPerm := resource + ":" + action

        // Check for exact permission or wildcard
        hasPermission := false
        for _, perm := range perms {
            if perm == requiredPerm || perm == resource+":*" || perm == "*:*" {
                hasPermission = true
                break
            }
        }

        if !hasPermission {
            RespondWithError(c, http.StatusForbidden, "access_denied",
                "You do not have permission to perform this action")
            c.Abort()
            return
        }

        c.Next()
    }
}

// RequireSystemAdmin checks if user is a system administrator
func RequireSystemAdmin() gin.HandlerFunc {
    return func(c *gin.Context) {
        roles, exists := c.Get("roles")
        if !exists {
            RespondWithError(c, http.StatusForbidden, "access_denied",
                "System admin check failed")
            c.Abort()
            return
        }

        roleList := roles.([]string)
        isSystemAdmin := false

        for _, role := range roleList {
            if role == "system:admin" || strings.HasPrefix(role, "system:") {
                isSystemAdmin = true
                break
            }
        }

        if !isSystemAdmin {
            RespondWithError(c, http.StatusForbidden, "access_denied",
                "System administrator access required")
            c.Abort()
            return
        }

        c.Next()
    }
}
```

### 3. Route Protection Updates

**File**: `api/routes/routes.go`

```go
// Tenant routes - only accessible by super-admins
tenantRoutes := apiV1.Group("/tenants")
tenantRoutes.Use(middleware.JWTAuth(tokenService))
tenantRoutes.Use(middleware.RequireSystemAdmin()) // NEW: Require system admin
{
    tenantRoutes.POST("", tenantHandler.Create)
    tenantRoutes.GET("", tenantHandler.List)
    tenantRoutes.GET("/:id", tenantHandler.GetByID)
    tenantRoutes.PUT("/:id", tenantHandler.Update)
    tenantRoutes.DELETE("/:id", tenantHandler.Delete)
}

// Tenant-scoped routes - users can only access their own tenant
tenantScoped := apiV1.Group("")
tenantScoped.Use(middleware.TenantMiddleware(tenantRepo))
tenantScoped.Use(middleware.JWTAuth(tokenService))
tenantScoped.Use(middleware.RequireTenantOrMaster()) // NEW: Check tenant access
{
    // Users, Roles, Permissions - tenant-scoped
    users := tenantScoped.Group("/users")
    {
        users.POST("", userHandler.Create)
        users.GET("", userHandler.List)
        users.GET("/:id", userHandler.GetByID)
        users.PUT("/:id", userHandler.Update)
        users.DELETE("/:id", userHandler.Delete)
    }

    // ... other tenant-scoped routes
}
```

### 4. Bootstrap Process

#### 4.1 Bootstrap Service

**File**: `cmd/bootstrap/bootstrap.go` (NEW)

```go
package bootstrap

import (
    "context"
    "fmt"
    "github.com/google/uuid"
    "github.com/arauth-identity/iam/identity/tenant"
    "github.com/arauth-identity/iam/identity/user"
    "github.com/arauth-identity/iam/identity/role"
    "github.com/arauth-identity/iam/identity/permission"
    "github.com/arauth-identity/iam/storage/interfaces"
    "github.com/arauth-identity/iam/security/password"
)

const (
    MasterTenantID = "00000000-0000-0000-0000-000000000000"
    MasterTenantName = "Master Tenant"
    MasterTenantDomain = "master.local"
    AdminUsername = "admin"
    AdminEmail = "admin@master.local"
    AdminPassword = "Admin@123456" // Should be changed on first login
)

type BootstrapService struct {
    tenantRepo     interfaces.TenantRepository
    userRepo       interfaces.UserRepository
    roleRepo       interfaces.RoleRepository
    permissionRepo interfaces.PermissionRepository
    credentialRepo interfaces.CredentialRepository
}

func NewBootstrapService(
    tenantRepo interfaces.TenantRepository,
    userRepo interfaces.UserRepository,
    roleRepo interfaces.RoleRepository,
    permissionRepo interfaces.PermissionRepository,
    credentialRepo interfaces.CredentialRepository,
) *BootstrapService {
    return &BootstrapService{
        tenantRepo:     tenantRepo,
        userRepo:       userRepo,
        roleRepo:       roleRepo,
        permissionRepo: permissionRepo,
        credentialRepo: credentialRepo,
    }
}

func (s *BootstrapService) Bootstrap(ctx context.Context) error {
    // Check if master tenant already exists
    masterTenantID, _ := uuid.Parse(MasterTenantID)
    existing, err := s.tenantRepo.GetByID(ctx, masterTenantID)
    if err == nil && existing != nil {
        return fmt.Errorf("system already bootstrapped")
    }

    // 1. Create Master Tenant
    masterTenant := &tenant.Tenant{
        ID:       masterTenantID,
        Name:     MasterTenantName,
        Domain:   MasterTenantDomain,
        Status:   "active",
        IsMaster: true, // NEW field
    }

    if err := s.tenantRepo.Create(ctx, masterTenant); err != nil {
        return fmt.Errorf("failed to create master tenant: %w", err)
    }

    // 2. Create System Permissions
    systemPerms := []struct {
        resource    string
        action      string
        description string
    }{
        {"system", "admin", "Full system administration access"},
        {"tenants", "create", "Create new tenants"},
        {"tenants", "read", "View all tenants"},
        {"tenants", "update", "Update any tenant"},
        {"tenants", "delete", "Delete any tenant"},
        {"system", "settings", "Manage system-wide settings"},
    }

    for _, perm := range systemPerms {
        p := &permission.Permission{
            TenantID:    masterTenantID,
            Resource:    perm.resource,
            Action:      perm.action,
            Description: perm.description,
            IsSystem:    true, // NEW field
        }
        if err := s.permissionRepo.Create(ctx, p); err != nil {
            return fmt.Errorf("failed to create system permission: %w", err)
        }
    }

    // 3. Create System Admin Role
    systemAdminRole := &role.Role{
        TenantID:    masterTenantID,
        Name:        "system:admin",
        Description: "System Administrator - Full access to all tenants and system settings",
        IsSystemRole: true, // NEW field
        Scope:       "system", // NEW field
    }

    if err := s.roleRepo.Create(ctx, systemAdminRole); err != nil {
        return fmt.Errorf("failed to create system admin role: %w", err)
    }

    // 4. Assign all system permissions to system:admin role
    allSystemPerms, _ := s.permissionRepo.ListByTenantID(ctx, masterTenantID)
    for _, perm := range allSystemPerms {
        if perm.IsSystem {
            if err := s.roleRepo.AssignPermission(ctx, systemAdminRole.ID, perm.ID); err != nil {
                return fmt.Errorf("failed to assign permission to system admin role: %w", err)
            }
        }
    }

    // 5. Create Admin User
    adminUser := &user.User{
        TenantID:  masterTenantID,
        Username:  AdminUsername,
        Email:     AdminEmail,
        FirstName: stringPtr("System"),
        LastName:  stringPtr("Administrator"),
        Status:    "active",
    }

    if err := s.userRepo.Create(ctx, adminUser); err != nil {
        return fmt.Errorf("failed to create admin user: %w", err)
    }

    // 6. Set Admin Password
    hasher := password.NewHasher()
    passwordHash, err := hasher.Hash(AdminPassword)
    if err != nil {
        return fmt.Errorf("failed to hash admin password: %w", err)
    }

    cred := &credential.Credential{
        UserID:       adminUser.ID,
        PasswordHash: passwordHash,
    }

    if err := s.credentialRepo.Create(ctx, cred); err != nil {
        return fmt.Errorf("failed to create admin credentials: %w", err)
    }

    // 7. Assign System Admin Role to Admin User
    if err := s.userRepo.AssignRole(ctx, adminUser.ID, systemAdminRole.ID); err != nil {
        return fmt.Errorf("failed to assign system admin role: %w", err)
    }

    return nil
}

func stringPtr(s string) *string {
    return &s
}
```

#### 4.2 Bootstrap Command

**File**: `cmd/bootstrap/main.go` (NEW)

```go
package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "os"

    "github.com/arauth-identity/iam/cmd/bootstrap"
    "github.com/arauth-identity/iam/config"
    "github.com/arauth-identity/iam/config/loader"
    "github.com/arauth-identity/iam/storage/postgres"
    "go.uber.org/zap"
)

func main() {
    configPath := flag.String("config", "config/config.yaml", "Path to config file")
    force := flag.Bool("force", false, "Force bootstrap even if master tenant exists")
    flag.Parse()

    // Load configuration
    cfg, err := loader.Load(*configPath)
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Initialize logger
    logger, _ := zap.NewDevelopment()
    defer logger.Sync()

    // Connect to database
    db, err := postgres.NewConnection(&cfg.Database)
    if err != nil {
        logger.Fatal("Failed to connect to database", zap.Error(err))
    }
    defer db.Close()

    // Initialize repositories
    tenantRepo := postgres.NewTenantRepository(db)
    userRepo := postgres.NewUserRepository(db)
    roleRepo := postgres.NewRoleRepository(db)
    permissionRepo := postgres.NewPermissionRepository(db)
    credentialRepo := postgres.NewCredentialRepository(db)

    // Initialize bootstrap service
    bootstrapService := bootstrap.NewBootstrapService(
        tenantRepo,
        userRepo,
        roleRepo,
        permissionRepo,
        credentialRepo,
    )

    ctx := context.Background()

    // Check if already bootstrapped
    if !*force {
        masterTenantID, _ := uuid.Parse(bootstrap.MasterTenantID)
        existing, err := tenantRepo.GetByID(ctx, masterTenantID)
        if err == nil && existing != nil {
            fmt.Println("⚠️  System already bootstrapped. Use --force to re-bootstrap.")
            os.Exit(0)
        }
    }

    // Run bootstrap
    if err := bootstrapService.Bootstrap(ctx); err != nil {
        logger.Fatal("Bootstrap failed", zap.Error(err))
    }

    fmt.Println("✅ System bootstrapped successfully!")
    fmt.Printf("   Master Tenant ID: %s\n", bootstrap.MasterTenantID)
    fmt.Printf("   Admin Username: %s\n", bootstrap.AdminUsername)
    fmt.Printf("   Admin Email: %s\n", bootstrap.AdminEmail)
    fmt.Printf("   Default Password: %s\n", bootstrap.AdminPassword)
    fmt.Println("\n⚠️  IMPORTANT: Change the admin password on first login!")
}
```

### 5. Claims Builder Updates

**File**: `auth/claims/builder.go`

```go
// BuildClaims should include system role and permission flags
func (b *Builder) BuildClaims(ctx context.Context, user *models.User) (*Claims, error) {
    // ... existing code ...

    // Check if user has system:admin role
    isSystemAdmin := false
    for _, role := range roles {
        if role.IsSystemRole && role.Scope == "system" {
            isSystemAdmin = true
            break
        }
    }

    return &Claims{
        // ... existing fields ...
        IsSystemAdmin: isSystemAdmin, // NEW field
        SystemRoles:   systemRoles,   // NEW field
        SystemPermissions: systemPermissions, // NEW field
    }, nil
}
```

### 6. Frontend Updates

#### 6.1 Admin Dashboard - Tenant Management

**File**: `frontend/admin-dashboard/src/pages/tenants/TenantList.tsx`

- Only show "Create Tenant" button if user is system admin
- Show warning if user tries to access tenant management without permissions
- Filter out master tenant from regular tenant list (or show with special badge)

#### 6.2 Admin Dashboard - User Management

- Users can only see users from their own tenant
- System admins can see users from all tenants (with tenant filter)

---

## 🔒 Security Considerations

### 1. Default Deny Principle
- **All routes default to tenant-scoped access**
- **Explicit permission required for cross-tenant access**
- **System admin role must be explicitly granted**

### 2. JWT Claims
- Include `is_system_admin` flag in JWT
- Include `system_roles` and `system_permissions` in claims
- Validate claims on every request

### 3. Audit Logging
- Log all cross-tenant access attempts
- Log system admin actions separately
- Track tenant creation/deletion by system admins

### 4. Password Policy
- Enforce strong password for admin user
- Require password change on first login
- Implement password rotation policy

### 5. Rate Limiting
- Stricter rate limits for system admin endpoints
- Separate rate limits for tenant management operations

---

## 📋 Implementation Checklist

### Phase 1: Database Schema (Week 1)
- [ ] Create migration for `is_master` field in tenants table
- [ ] Create migration for system roles and permissions
- [ ] Create system_permissions table
- [ ] Update indexes and constraints
- [ ] Test migrations up and down

### Phase 2: Backend Authorization (Week 1-2)
- [ ] Implement `RequireTenantOrMaster` middleware
- [ ] Implement `RequireSystemAdmin` middleware
- [ ] Implement `RequirePermission` middleware
- [ ] Update route protection
- [ ] Update claims builder
- [ ] Add unit tests for authorization

### Phase 3: Bootstrap Process (Week 2)
- [ ] Create bootstrap service
- [ ] Create bootstrap command
- [ ] Test bootstrap process
- [ ] Document bootstrap procedure
- [ ] Add bootstrap to deployment scripts

### Phase 4: Frontend Updates (Week 2-3)
- [ ] Update tenant list to check permissions
- [ ] Hide tenant management for non-admins
- [ ] Add system admin indicators
- [ ] Update user management for tenant isolation
- [ ] Add permission checks in UI

### Phase 5: Testing & Documentation (Week 3)
- [ ] Integration tests for tenant isolation
- [ ] Security testing (penetration testing)
- [ ] Update API documentation
- [ ] Update user guide
- [ ] Create migration guide for existing deployments

### Phase 6: Deployment (Week 4)
- [ ] Run migrations on staging
- [ ] Bootstrap staging environment
- [ ] Test end-to-end scenarios
- [ ] Deploy to production
- [ ] Monitor for issues

---

## 🚀 Migration Strategy

### For Existing Deployments

1. **Backup Database**: Full backup before migration
2. **Run Migrations**: Apply all schema changes
3. **Bootstrap Master Tenant**: Run bootstrap command
4. **Migrate Existing Data**: 
   - Assign existing users to appropriate tenants
   - Create tenant-specific roles
   - Remove cross-tenant access
5. **Update Applications**: Update API clients to use new authorization
6. **Test**: Comprehensive testing before production

### Rollback Plan

1. **Database Rollback**: Revert migrations (if possible)
2. **Code Rollback**: Deploy previous version
3. **Data Recovery**: Restore from backup if needed

---

## 📊 Why This Solution is Secure

### 1. **Defense in Depth**
- Multiple layers of authorization checks
- Middleware-level and handler-level validation
- Database-level constraints

### 2. **Principle of Least Privilege**
- Users default to tenant-scoped access
- System admin requires explicit grant
- Permissions are granular and specific

### 3. **Auditability**
- All system admin actions are logged
- Clear distinction between tenant and system operations
- Traceable access patterns

### 4. **Industry Best Practices**
- Similar to Keycloak, Auth0, AWS Organizations
- Follows OAuth2/OIDC patterns
- Aligns with NIST security guidelines

### 5. **Scalability**
- Supports unlimited tenants
- Efficient permission checking
- Minimal performance impact

### 6. **Maintainability**
- Clear separation of concerns
- Well-documented architecture
- Easy to extend and modify

---

## 🔍 Testing Scenarios

### 1. Tenant Isolation Tests
- ✅ User from Tenant A cannot access Tenant B's data
- ✅ User from Tenant A cannot list Tenant B's users
- ✅ User from Tenant A cannot modify Tenant B's settings

### 2. System Admin Tests
- ✅ System admin can access all tenants
- ✅ System admin can create/edit/delete tenants
- ✅ System admin can manage system settings
- ✅ System admin actions are logged

### 3. Bootstrap Tests
- ✅ Bootstrap creates master tenant correctly
- ✅ Bootstrap creates admin user with correct permissions
- ✅ Bootstrap fails gracefully if already bootstrapped
- ✅ Bootstrap can be forced with --force flag

### 4. Permission Tests
- ✅ Regular users cannot access tenant management
- ✅ Regular users can only access their tenant's data
- ✅ System permissions are separate from tenant permissions

---

## 📚 References

- [Keycloak Master Realm Documentation](https://www.keycloak.org/docs/latest/server_admin/#_master_realm)
- [OAuth2 Multi-Tenant Best Practices](https://oauth.net/2/multi-tenancy/)
- [NIST Security Guidelines](https://csrc.nist.gov/publications/detail/sp/800-53/rev-5/final)
- [OWASP Multi-Tenant Security](https://owasp.org/www-community/Multi-Tenant_Application_Security)

---

## 📝 Notes

- Master tenant ID is fixed UUID: `00000000-0000-0000-0000-000000000000`
- System admin role name: `system:admin`
- Default admin password should be changed on first login
- Consider implementing MFA for system admin accounts
- Regular tenant users should never have system permissions

---

**Document Version**: 1.0  
**Last Updated**: 2026-01-08  
**Author**: ARauth Identity Team

