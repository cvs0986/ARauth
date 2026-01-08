# Master Tenant & Multi-Tenant Security Implementation Plan V2

## 📋 Executive Summary

This document outlines a comprehensive plan to implement a **Master User / System Admin** architecture that separates platform-level control from tenant-level identity. This addresses critical security vulnerabilities and aligns with industry best practices (AWS, Azure, GCP, Keycloak).

**Key Design Principle**: Master users live in the **Platform Control Plane** (no tenant), while regular users live in the **Tenant Plane** (isolated per tenant).

---

## 🚨 Current Security Issues

### Issue 1: Cross-Tenant Data Access
**Problem**: Any authenticated user can access tenant management endpoints and see/edit all tenants.

**Impact**: **CRITICAL** - Complete breach of tenant isolation.

### Issue 2: No System/Platform Control Plane
**Problem**: No separation between system-level and tenant-level operations.

**Impact**: **CRITICAL** - Cannot properly manage multi-tenant system securely.

### Issue 3: Master User Configuration
**Problem**: No flexible way to create master user (config, CLI, bootstrap).

**Impact**: **HIGH** - Deployment flexibility and automation issues.

### Issue 4: System-Wide Settings Management
**Problem**: No way for master user to configure settings for all tenants.

**Impact**: **HIGH** - Cannot manage system-wide policies and configurations.

---

## ✅ Proposed Solution: Two-Plane Architecture

### Core Concept

```
┌──────────────────────────┐
│ PLATFORM CONTROL PLANE   │  ← Master users (tenant_id = NULL)
│ (System / Global)        │     Principal Type: SYSTEM
└─────────────┬────────────┘
              │
┌─────────────▼────────────┐
│ TENANT PLANE              │  ← Tenant users (tenant_id = <uuid>)
│ (Isolated per tenant)     │     Principal Type: TENANT
└──────────────────────────┘
```

### Key Design Decisions

1. **Master Users Have `tenant_id = NULL`** (not part of any tenant)
2. **Principal Type System** (`SYSTEM`, `TENANT`, `SERVICE`)
3. **Separate API Boundaries** (`/system/*` vs `/tenants/{id}/*`)
4. **Separate Authorization Logic** (no overlap)
5. **Config-Based Bootstrap** (YAML, env vars, CLI flags)

### Why This Solution is Best

1. **Industry Standard**: Matches AWS Organizations, Azure AD, GCP IAM, Keycloak
2. **Security**: Complete isolation, no privilege escalation possible
3. **Scalability**: Works for SaaS, on-prem, MSP, regulated environments
4. **Flexibility**: Multiple ways to create master user (config, CLI, API)
5. **Maintainability**: Clear separation, easier to reason about

---

## 🏗️ Architecture Design

### 1. Database Schema Changes

#### 1.1 Update `users` Table

```sql
-- Migration: 000013_add_principal_type.up.sql
ALTER TABLE users
ADD COLUMN principal_type VARCHAR(50) DEFAULT 'TENANT' NOT NULL 
  CHECK (principal_type IN ('SYSTEM', 'TENANT', 'SERVICE')),
ADD COLUMN tenant_id UUID NULL REFERENCES tenants(id) ON DELETE CASCADE;

-- Add constraint: SYSTEM users must have tenant_id = NULL
ALTER TABLE users
ADD CONSTRAINT chk_system_user_no_tenant 
  CHECK (
    (principal_type = 'SYSTEM' AND tenant_id IS NULL) OR
    (principal_type != 'SYSTEM' AND tenant_id IS NOT NULL)
  );

-- Update existing constraint to allow NULL for SYSTEM users
ALTER TABLE users
DROP CONSTRAINT IF EXISTS users_tenant_id_fkey;
ALTER TABLE users
ADD CONSTRAINT users_tenant_id_fkey 
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

-- Create indexes
CREATE INDEX idx_users_principal_type ON users(principal_type);
CREATE INDEX idx_users_system_users ON users(principal_type) WHERE principal_type = 'SYSTEM';
```

#### 1.2 Create `system_roles` Table

```sql
-- Migration: 000014_create_system_roles.up.sql
CREATE TABLE system_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE system_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource VARCHAR(255) NOT NULL,
    action VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(resource, action)
);

CREATE TABLE system_role_permissions (
    role_id UUID NOT NULL REFERENCES system_roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES system_permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE user_system_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES system_roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    assigned_by UUID REFERENCES users(id),
    PRIMARY KEY (user_id, role_id)
);

-- Insert default system roles
INSERT INTO system_roles (id, name, description) VALUES
    ('00000000-0000-0000-0000-000000000001', 'system_owner', 'Full system ownership and control'),
    ('00000000-0000-0000-0000-000000000002', 'system_admin', 'System administration with tenant management'),
    ('00000000-0000-0000-0000-000000000003', 'system_auditor', 'Read-only system access for auditing');

-- Insert default system permissions
INSERT INTO system_permissions (resource, action, description) VALUES
    ('tenant', 'create', 'Create new tenants'),
    ('tenant', 'read', 'View all tenants'),
    ('tenant', 'update', 'Update any tenant'),
    ('tenant', 'delete', 'Delete any tenant'),
    ('tenant', 'suspend', 'Suspend tenant access'),
    ('tenant', 'configure', 'Configure tenant settings'),
    ('system', 'settings', 'Manage system-wide settings'),
    ('system', 'policy', 'Manage global policies'),
    ('system', 'audit', 'View system audit logs'),
    ('billing', 'manage', 'Manage billing and subscriptions');

-- Assign permissions to system_owner (all permissions)
INSERT INTO system_role_permissions (role_id, permission_id)
SELECT 
    '00000000-0000-0000-0000-000000000001'::uuid,
    id
FROM system_permissions;

-- Assign permissions to system_admin (tenant management + system settings)
INSERT INTO system_role_permissions (role_id, permission_id)
SELECT 
    '00000000-0000-0000-0000-000000000002'::uuid,
    id
FROM system_permissions
WHERE resource IN ('tenant', 'system') AND action != 'delete';

-- Assign permissions to system_auditor (read-only)
INSERT INTO system_role_permissions (role_id, permission_id)
SELECT 
    '00000000-0000-0000-0000-000000000003'::uuid,
    id
FROM system_permissions
WHERE action = 'read' OR action = 'audit';
```

#### 1.3 Create `system_settings` Table

```sql
-- Migration: 000015_create_system_settings.up.sql
CREATE TABLE system_settings (
    key VARCHAR(255) PRIMARY KEY,
    value JSONB NOT NULL,
    description TEXT,
    updated_by UUID REFERENCES users(id),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Insert default system settings
INSERT INTO system_settings (key, value, description) VALUES
    ('password_policy', '{"min_length": 12, "require_uppercase": true, "require_lowercase": true, "require_numbers": true, "require_special": true}', 'Global password policy'),
    ('mfa_policy', '{"enforced_for_system_users": true, "enforced_for_tenant_admins": false}', 'MFA enforcement policy'),
    ('session_policy', '{"max_session_duration": 3600, "idle_timeout": 900}', 'Session management policy'),
    ('rate_limit_policy', '{"system_api_rpm": 1000, "tenant_api_rpm": 100}', 'Rate limiting policy');
```

#### 1.4 Create `tenant_configurations` Table

```sql
-- Migration: 000016_create_tenant_configurations.up.sql
CREATE TABLE tenant_configurations (
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL,
    value JSONB NOT NULL,
    configured_by UUID REFERENCES users(id),
    configured_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, key)
);

CREATE INDEX idx_tenant_configurations_tenant_id ON tenant_configurations(tenant_id);
```

### 2. Configuration-Based Bootstrap

#### 2.1 Update Config Structure

**File**: `config/config.go`

```go
type BootstrapConfig struct {
    Enabled    bool   `yaml:"enabled" env:"BOOTSTRAP_ENABLED" envDefault:"false"`
    MasterUser struct {
        Username string `yaml:"username" env:"BOOTSTRAP_USERNAME" envDefault:"admin"`
        Email    string `yaml:"email" env:"BOOTSTRAP_EMAIL" envDefault:"admin@arauth.io"`
        Password string `yaml:"password" env:"BOOTSTRAP_PASSWORD"` // Required if enabled
        FirstName string `yaml:"first_name" env:"BOOTSTRAP_FIRST_NAME" envDefault:"System"`
        LastName  string `yaml:"last_name" env:"BOOTSTRAP_LAST_NAME" envDefault:"Administrator"`
    } `yaml:"master_user"`
    Force bool `yaml:"force" env:"BOOTSTRAP_FORCE" envDefault:"false"`
}
```

**File**: `config/config.yaml`

```yaml
bootstrap:
  enabled: false  # Set to true to enable bootstrap on startup
  force: false     # Set to true to re-bootstrap even if master user exists
  master_user:
    username: "admin"
    email: "admin@arauth.io"
    password: "${BOOTSTRAP_PASSWORD}"  # Must be set via env var for security
    first_name: "System"
    last_name: "Administrator"
```

#### 2.2 Bootstrap Service

**File**: `cmd/bootstrap/bootstrap.go`

```go
package bootstrap

import (
    "context"
    "fmt"
    "github.com/google/uuid"
    "github.com/arauth-identity/iam/config"
    "github.com/arauth-identity/iam/identity/user"
    "github.com/arauth-identity/iam/storage/interfaces"
    "github.com/arauth-identity/iam/security/password"
)

type BootstrapService struct {
    cfg            *config.BootstrapConfig
    userRepo       interfaces.UserRepository
    credentialRepo interfaces.CredentialRepository
    systemRoleRepo interfaces.SystemRoleRepository
}

func NewBootstrapService(
    cfg *config.BootstrapConfig,
    userRepo interfaces.UserRepository,
    credentialRepo interfaces.CredentialRepository,
    systemRoleRepo interfaces.SystemRoleRepository,
) *BootstrapService {
    return &BootstrapService{
        cfg:            cfg,
        userRepo:       userRepo,
        credentialRepo: credentialRepo,
        systemRoleRepo: systemRoleRepo,
    }
}

func (s *BootstrapService) Bootstrap(ctx context.Context) error {
    // Check if master user already exists
    existing, err := s.userRepo.GetByEmail(ctx, s.cfg.MasterUser.Email)
    if err == nil && existing != nil && existing.PrincipalType == "SYSTEM" {
        if !s.cfg.Force {
            return fmt.Errorf("master user already exists (use --force to re-bootstrap)")
        }
        // Force re-bootstrap: delete existing master user
        if err := s.userRepo.Delete(ctx, existing.ID); err != nil {
            return fmt.Errorf("failed to delete existing master user: %w", err)
        }
    }

    // 1. Create Master User (tenant_id = NULL, principal_type = SYSTEM)
    masterUser := &user.User{
        ID:            uuid.New(),
        Username:      s.cfg.MasterUser.Username,
        Email:         s.cfg.MasterUser.Email,
        FirstName:     &s.cfg.MasterUser.FirstName,
        LastName:      &s.cfg.MasterUser.LastName,
        Status:        "active",
        PrincipalType: "SYSTEM", // NEW: System user
        TenantID:      nil,      // NEW: No tenant
    }

    if err := s.userRepo.Create(ctx, masterUser); err != nil {
        return fmt.Errorf("failed to create master user: %w", err)
    }

    // 2. Set Master User Password
    hasher := password.NewHasher()
    passwordHash, err := hasher.Hash(s.cfg.MasterUser.Password)
    if err != nil {
        return fmt.Errorf("failed to hash password: %w", err)
    }

    cred := &credential.Credential{
        UserID:       masterUser.ID,
        PasswordHash: passwordHash,
    }

    if err := s.credentialRepo.Create(ctx, cred); err != nil {
        return fmt.Errorf("failed to create credentials: %w", err)
    }

    // 3. Assign system_owner role
    systemOwnerRoleID, _ := uuid.Parse("00000000-0000-0000-0000-000000000001")
    if err := s.systemRoleRepo.AssignRoleToUser(ctx, masterUser.ID, systemOwnerRoleID); err != nil {
        return fmt.Errorf("failed to assign system_owner role: %w", err)
    }

    return nil
}
```

#### 2.3 Bootstrap on Server Start

**File**: `cmd/server/main.go`

```go
// After loading config
if cfg.Bootstrap.Enabled {
    logger.Info("Bootstrap enabled, checking for master user...")
    
    bootstrapService := bootstrap.NewBootstrapService(
        &cfg.Bootstrap,
        userRepo,
        credentialRepo,
        systemRoleRepo,
    )
    
    if err := bootstrapService.Bootstrap(context.Background()); err != nil {
        if cfg.Bootstrap.Force {
            logger.Fatal("Bootstrap failed", zap.Error(err))
        } else {
            logger.Info("Bootstrap skipped", zap.Error(err))
        }
    } else {
        logger.Info("✅ Master user bootstrapped successfully",
            zap.String("username", cfg.Bootstrap.MasterUser.Username),
            zap.String("email", cfg.Bootstrap.MasterUser.Email))
    }
}
```

#### 2.4 Standalone Bootstrap Command

**File**: `cmd/bootstrap/main.go`

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
    // ... other imports
)

func main() {
    configPath := flag.String("config", "config/config.yaml", "Path to config file")
    username := flag.String("username", "", "Master user username (overrides config)")
    email := flag.String("email", "", "Master user email (overrides config)")
    password := flag.String("password", "", "Master user password (required)")
    force := flag.Bool("force", false, "Force bootstrap even if master user exists")
    flag.Parse()

    if *password == "" {
        log.Fatal("Password is required. Use --password flag or BOOTSTRAP_PASSWORD env var")
    }

    // Load configuration
    cfg, err := loader.Load(*configPath)
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Override with CLI flags
    if *username != "" {
        cfg.Bootstrap.MasterUser.Username = *username
    }
    if *email != "" {
        cfg.Bootstrap.MasterUser.Email = *email
    }
    cfg.Bootstrap.MasterUser.Password = *password
    cfg.Bootstrap.Force = *force

    // Initialize services and run bootstrap
    // ... (similar to server bootstrap)
}
```

### 3. Claims Builder Updates

**File**: `auth/claims/builder.go`

```go
func (b *Builder) BuildClaims(ctx context.Context, user *models.User) (*Claims, error) {
    claims := &Claims{
        Subject:      user.ID.String(),
        Username:     user.Username,
        Email:        user.Email,
        PrincipalType: user.PrincipalType, // NEW
    }

    // For SYSTEM users: no tenant_id, get system roles
    if user.PrincipalType == "SYSTEM" {
        claims.TenantID = uuid.Nil // No tenant
        systemRoles, err := b.getSystemRoles(ctx, user.ID)
        if err != nil {
            return nil, err
        }
        claims.SystemRoles = systemRoles
        claims.SystemPermissions = b.getSystemPermissions(ctx, systemRoles)
        claims.Scope = "system:*"
    } else {
        // For TENANT users: tenant_id required, get tenant roles
        claims.TenantID = *user.TenantID
        tenantRoles, err := b.getTenantRoles(ctx, user.ID, *user.TenantID)
        if err != nil {
            return nil, err
        }
        claims.Roles = tenantRoles
        claims.Permissions = b.getTenantPermissions(ctx, tenantRoles, *user.TenantID)
        claims.Scope = fmt.Sprintf("tenant:%s", user.TenantID.String())
    }

    return claims, nil
}
```

### 4. API Route Separation

**File**: `api/routes/routes.go`

```go
// System API routes (master users only)
systemAPI := v1.Group("/system")
systemAPI.Use(middleware.JWTAuth(tokenService))
systemAPI.Use(middleware.RequireSystemUser()) // NEW: Check principal_type == SYSTEM
{
    // Tenant management
    systemAPI.POST("/tenants", tenantHandler.Create)
    systemAPI.GET("/tenants", tenantHandler.List)
    systemAPI.GET("/tenants/:id", tenantHandler.GetByID)
    systemAPI.PUT("/tenants/:id", tenantHandler.Update)
    systemAPI.DELETE("/tenants/:id", tenantHandler.Delete)
    systemAPI.POST("/tenants/:id/suspend", tenantHandler.Suspend)
    systemAPI.POST("/tenants/:id/resume", tenantHandler.Resume)

    // System settings
    systemAPI.GET("/settings", systemHandler.GetSettings)
    systemAPI.PUT("/settings", systemHandler.UpdateSettings)

    // Tenant configurations (configure settings for any tenant)
    systemAPI.GET("/tenants/:id/config", tenantConfigHandler.Get)
    systemAPI.PUT("/tenants/:id/config", tenantConfigHandler.Update)

    // System audit logs
    systemAPI.GET("/audit", auditHandler.List)
}

// Tenant API routes (tenant users only)
tenantAPI := v1.Group("/tenants/:tenant_id")
tenantAPI.Use(middleware.TenantMiddleware(tenantRepo))
tenantAPI.Use(middleware.JWTAuth(tokenService))
tenantAPI.Use(middleware.RequireTenantUser()) // NEW: Check principal_type == TENANT
{
    tenantAPI.POST("/users", userHandler.Create)
    tenantAPI.GET("/users", userHandler.List)
    // ... other tenant-scoped routes
}
```

### 5. System Settings Management

**File**: `api/handlers/system_handler.go` (NEW)

```go
package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/arauth-identity/iam/api/middleware"
    "github.com/arauth-identity/iam/identity/system"
)

type SystemHandler struct {
    systemService system.ServiceInterface
}

func (h *SystemHandler) GetSettings(c *gin.Context) {
    settings, err := h.systemService.GetSettings(c.Request.Context())
    if err != nil {
        middleware.RespondWithError(c, http.StatusInternalServerError, "get_failed", err.Error(), nil)
        return
    }
    c.JSON(http.StatusOK, settings)
}

func (h *SystemHandler) UpdateSettings(c *gin.Context) {
    var req system.UpdateSettingsRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request", "Request validation failed", middleware.FormatValidationErrors(err))
        return
    }

    userID, _ := c.Get("userID")
    settings, err := h.systemService.UpdateSettings(c.Request.Context(), userID.(uuid.UUID), &req)
    if err != nil {
        middleware.RespondWithError(c, http.StatusInternalServerError, "update_failed", err.Error(), nil)
        return
    }
    c.JSON(http.StatusOK, settings)
}
```

**File**: `api/handlers/tenant_config_handler.go` (NEW)

```go
package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/arauth-identity/iam/api/middleware"
    "github.com/arauth-identity/iam/identity/tenant"
)

type TenantConfigHandler struct {
    tenantService tenant.ServiceInterface
}

func (h *TenantConfigHandler) Get(c *gin.Context) {
    tenantID, err := uuid.Parse(c.Param("id"))
    if err != nil {
        middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id", "Invalid tenant ID", nil)
        return
    }

    config, err := h.tenantService.GetConfiguration(c.Request.Context(), tenantID)
    if err != nil {
        middleware.RespondWithError(c, http.StatusNotFound, "not_found", err.Error(), nil)
        return
    }
    c.JSON(http.StatusOK, config)
}

func (h *TenantConfigHandler) Update(c *gin.Context) {
    tenantID, err := uuid.Parse(c.Param("id"))
    if err != nil {
        middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id", "Invalid tenant ID", nil)
        return
    }

    var req tenant.UpdateConfigurationRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request", "Request validation failed", middleware.FormatValidationErrors(err))
        return
    }

    userID, _ := c.Get("userID")
    config, err := h.tenantService.UpdateConfiguration(c.Request.Context(), tenantID, userID.(uuid.UUID), &req)
    if err != nil {
        middleware.RespondWithError(c, http.StatusInternalServerError, "update_failed", err.Error(), nil)
        return
    }
    c.JSON(http.StatusOK, config)
}
```

### 6. Middleware Updates

**File**: `api/middleware/authorization.go` (NEW functions)

```go
// RequireSystemUser ensures user is a SYSTEM principal
func RequireSystemUser() gin.HandlerFunc {
    return func(c *gin.Context) {
        claims, exists := c.Get("user_claims")
        if !exists {
            RespondWithError(c, http.StatusUnauthorized, "unauthorized", "User claims not found", nil)
            c.Abort()
            return
        }

        userClaims := claims.(*claims.Claims)
        if userClaims.PrincipalType != "SYSTEM" {
            RespondWithError(c, http.StatusForbidden, "forbidden", 
                "System user access required. This endpoint is only available to system administrators.", nil)
            c.Abort()
            return
        }

        c.Next()
    }
}

// RequireTenantUser ensures user is a TENANT principal
func RequireTenantUser() gin.HandlerFunc {
    return func(c *gin.Context) {
        claims, exists := c.Get("user_claims")
        if !exists {
            RespondWithError(c, http.StatusUnauthorized, "unauthorized", "User claims not found", nil)
            c.Abort()
            return
        }

        userClaims := claims.(*claims.Claims)
        if userClaims.PrincipalType != "TENANT" {
            RespondWithError(c, http.StatusForbidden, "forbidden",
                "Tenant user access required. System users cannot access tenant-scoped endpoints.", nil)
            c.Abort()
            return
        }

        // Verify tenant_id matches
        requestedTenantID, _ := GetTenantID(c)
        if userClaims.TenantID != requestedTenantID {
            RespondWithError(c, http.StatusForbidden, "forbidden",
                "You do not have access to this tenant", nil)
            c.Abort()
            return
        }

        c.Next()
    }
}

// RequireSystemPermission checks if system user has required permission
func RequireSystemPermission(resource, action string) gin.HandlerFunc {
    return func(c *gin.Context) {
        claims, exists := c.Get("user_claims")
        if !exists {
            RespondWithError(c, http.StatusUnauthorized, "unauthorized", "User claims not found", nil)
            c.Abort()
            return
        }

        userClaims := claims.(*claims.Claims)
        if userClaims.PrincipalType != "SYSTEM" {
            RespondWithError(c, http.StatusForbidden, "forbidden", "System user required", nil)
            c.Abort()
            return
        }

        requiredPerm := resource + ":" + action
        hasPermission := false
        for _, perm := range userClaims.SystemPermissions {
            if perm == requiredPerm || perm == resource+":*" || perm == "*:*" {
                hasPermission = true
                break
            }
        }

        if !hasPermission {
            RespondWithError(c, http.StatusForbidden, "forbidden",
                fmt.Sprintf("Required permission: %s", requiredPerm), nil)
            c.Abort()
            return
        }

        c.Next()
    }
}
```

---

## 🔒 Security Considerations

### 1. Master User Security

- **MFA Mandatory**: System users must have MFA enabled
- **Strong Password Policy**: Enforced for system users
- **Short Token Lifetime**: System tokens expire faster (5 minutes)
- **Audit Everything**: All system operations logged
- **Rate Limiting**: Stricter limits for system APIs

### 2. Token Design

**System User Token**:
```json
{
  "sub": "user-001",
  "principal_type": "SYSTEM",
  "system_roles": ["system_owner"],
  "system_permissions": ["tenant:*", "system:*"],
  "iss": "arauth",
  "scope": "system:*",
  "exp": 1234567890
}
```

**Tenant User Token**:
```json
{
  "sub": "user-234",
  "principal_type": "TENANT",
  "tenant_id": "tenant-abc",
  "roles": ["tenant_admin"],
  "permissions": ["users:read", "users:write"],
  "iss": "arauth",
  "scope": "tenant:tenant-abc",
  "exp": 1234567890
}
```

### 3. Absolute Rules

1. **No Overlap**: System roles never evaluated in tenant authorization
2. **No Inheritance**: Tenant users cannot escalate to system roles
3. **No Shortcuts**: Separate middleware for system vs tenant
4. **No Tenant ID for System Users**: `tenant_id = NULL` enforced at DB level

---

## 📋 Implementation Checklist

### Phase 1: Database Schema (Week 1)
- [ ] Add `principal_type` to users table
- [ ] Make `tenant_id` nullable for SYSTEM users
- [ ] Create `system_roles`, `system_permissions` tables
- [ ] Create `system_settings` table
- [ ] Create `tenant_configurations` table
- [ ] Insert default system roles and permissions
- [ ] Test migrations

### Phase 2: Configuration & Bootstrap (Week 1)
- [ ] Add bootstrap config to `config.yaml`
- [ ] Implement bootstrap service
- [ ] Add bootstrap on server start
- [ ] Create standalone bootstrap command
- [ ] Support config, env vars, CLI flags
- [ ] Test bootstrap process

### Phase 3: Backend Authorization (Week 2)
- [ ] Update claims builder for principal types
- [ ] Implement `RequireSystemUser` middleware
- [ ] Implement `RequireTenantUser` middleware
- [ ] Implement `RequireSystemPermission` middleware
- [ ] Separate system and tenant API routes
- [ ] Update JWT token generation
- [ ] Add unit tests

### Phase 4: System Settings Management (Week 2)
- [ ] Implement system settings service
- [ ] Implement tenant configuration service
- [ ] Create system handler
- [ ] Create tenant config handler
- [ ] Add API endpoints
- [ ] Test settings management

### Phase 5: Frontend Updates (Week 3)
- [ ] Add system admin login flow
- [ ] Create system admin dashboard
- [ ] Add tenant management UI
- [ ] Add system settings UI
- [ ] Add tenant configuration UI
- [ ] Hide tenant management for non-system users

### Phase 6: Testing & Documentation (Week 3-4)
- [ ] Integration tests for system vs tenant separation
- [ ] Security testing (penetration testing)
- [ ] Update API documentation
- [ ] Create migration guide
- [ ] Document bootstrap process

---

## 🚀 Bootstrap Methods

### Method 1: Config File

```yaml
# config/config.yaml
bootstrap:
  enabled: true
  master_user:
    username: "admin"
    email: "admin@arauth.io"
    password: "${BOOTSTRAP_PASSWORD}"
```

```bash
export BOOTSTRAP_PASSWORD="SecurePassword123!"
go run cmd/server/main.go
```

### Method 2: Environment Variables

```bash
export BOOTSTRAP_ENABLED=true
export BOOTSTRAP_USERNAME="admin"
export BOOTSTRAP_EMAIL="admin@arauth.io"
export BOOTSTRAP_PASSWORD="SecurePassword123!"
go run cmd/server/main.go
```

### Method 3: CLI Command

```bash
go run cmd/bootstrap/main.go \
  --username="admin" \
  --email="admin@arauth.io" \
  --password="SecurePassword123!" \
  --force
```

### Method 4: Docker/Kubernetes

```yaml
# docker-compose.yml
services:
  iam:
    environment:
      - BOOTSTRAP_ENABLED=true
      - BOOTSTRAP_USERNAME=admin
      - BOOTSTRAP_EMAIL=admin@arauth.io
      - BOOTSTRAP_PASSWORD=${BOOTSTRAP_PASSWORD}
```

---

## ✅ Master User Capabilities

### System-Wide Operations

1. **Tenant Management**
   - Create, read, update, delete tenants
   - Suspend/resume tenants
   - View all tenants

2. **System Settings**
   - Global password policy
   - MFA enforcement policy
   - Session management policy
   - Rate limiting policy
   - System-wide configurations

3. **Tenant Configuration**
   - Configure settings for any tenant
   - Override tenant-specific policies
   - Manage tenant-level configurations

4. **Audit & Monitoring**
   - View system audit logs
   - Monitor system metrics
   - Access all tenant audit logs

5. **Billing & Subscriptions** (if applicable)
   - Manage billing
   - Handle subscriptions
   - View usage metrics

---

## 🎯 Comparison: V1 vs V2

| Aspect | V1 (Master Tenant) | V2 (Principal Type) |
|--------|---------------------|---------------------|
| Master User Location | In master tenant | Outside all tenants (`tenant_id = NULL`) |
| Principal Type | Implicit | Explicit (`SYSTEM`, `TENANT`) |
| API Separation | Same routes, different auth | Separate routes (`/system/*` vs `/tenants/*`) |
| Token Design | Includes tenant_id | No tenant_id for system users |
| Authorization Logic | Mixed (check tenant + system) | Completely separate |
| Industry Alignment | Good | Excellent (AWS/Azure/GCP pattern) |
| Security | Good | Better (no privilege escalation possible) |

**Recommendation**: **V2 (Principal Type)** is the better approach for production.

---

## 📚 References

- [AWS Organizations Architecture](https://docs.aws.amazon.com/organizations/)
- [Azure AD Administrative Units](https://docs.microsoft.com/en-us/azure/active-directory/roles/admin-units-overview)
- [GCP IAM Hierarchy](https://cloud.google.com/iam/docs/resource-hierarchy-access-control)
- [Keycloak Master Realm](https://www.keycloak.org/docs/latest/server_admin/#_master_realm)
- [NIST Multi-Tenant Security](https://csrc.nist.gov/publications/detail/sp/800-53/rev-5/final)

---

**Document Version**: 2.0  
**Last Updated**: 2026-01-08  
**Author**: ARauth Identity Team

