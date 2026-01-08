# ARauth Capability Model Implementation Plan

## 📋 Executive Summary

This document provides a comprehensive, line-by-line implementation plan based on `feature_capibility.md`. It defines the three-layer capability model (System → System→Tenant → Tenant → User) and outlines all required changes across backend, frontend, and infrastructure.

**Key Principle**: Capabilities flow strictly downward with no upward overrides.

---

## 🎯 Implementation Phases Overview

| Phase | Name | Duration | Priority |
|-------|------|----------|----------|
| **Phase 1** | Database & Models | 2-3 weeks | P0 |
| **Phase 2** | Backend Core Logic | 3-4 weeks | P0 |
| **Phase 3** | API Endpoints | 2-3 weeks | P0 |
| **Phase 4** | Frontend Admin Dashboard | 3-4 weeks | P0 |
| **Phase 5** | Enforcement & Validation | 2-3 weeks | P0 |
| **Phase 6** | Testing & Documentation | 2-3 weeks | P1 |
| **Phase 7** | Migration & Deployment | 1-2 weeks | P1 |

**Total Estimated Duration**: 15-22 weeks

---

## 📊 Capability Model Architecture

```
┌──────────────────────────────────────────────┐
│ SYSTEM (Platform / Control Plane)             │
│ • Defines WHAT EXISTS                         │
│ • Hard security guardrails                    │
└───────────────────────┬──────────────────────┘
                        │
                        │ Allowed Capabilities
                        ▼
┌──────────────────────────────────────────────┐
│ SYSTEM → TENANT CAPABILITY ASSIGNMENT         │
│ • What THIS tenant is allowed to use          │
│ • Per-tenant feature flags                    │
└───────────────────────┬──────────────────────┘
                        │
                        │ Enabled Features
                        ▼
┌──────────────────────────────────────────────┐
│ TENANT (Organization Plane)                   │
│ • Chooses what to ENABLE                      │
│ • Enforces security policies                  │
└───────────────────────┬──────────────────────┘
                        │
                        │ Enforcement & State
                        ▼
┌──────────────────────────────────────────────┐
│ USER (Identity Plane)                         │
│ • Enrolls & COMPLIES                          │
│ • Has state, not power                        │
└──────────────────────────────────────────────┘
```

---

## 🗄️ Phase 1: Database & Models

### 1.1 Create Tenant Capabilities Table

**Issue**: `#001` - Create tenant_capabilities table  
**Tags**: `database`, `migration`, `p0`

**Migration**: `000018_create_tenant_capabilities.up.sql`

```sql
-- Tenant capabilities table (System → Tenant assignment)
CREATE TABLE tenant_capabilities (
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    capability_key VARCHAR(255) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT false,
    value JSONB, -- For capability-specific configuration (e.g., max_token_ttl: "10m")
    configured_by UUID REFERENCES users(id),
    configured_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, capability_key)
);

CREATE INDEX idx_tenant_capabilities_tenant_id ON tenant_capabilities(tenant_id);
CREATE INDEX idx_tenant_capabilities_key ON tenant_capabilities(capability_key);
```

**Capability Keys**:
- `mfa` - Multi-factor authentication
- `totp` - Time-based OTP
- `saml` - SAML federation
- `oidc` - OAuth2/OIDC
- `oauth2` - OAuth2 protocol
- `passwordless` - Passwordless authentication
- `ldap` - LDAP/AD integration
- `max_token_ttl` - Maximum token TTL (value: duration string)
- `allowed_grant_types` - Allowed OAuth grant types (value: array)
- `allowed_scope_namespaces` - Allowed scope namespaces (value: array)

### 1.2 Create System Capabilities Table

**Issue**: `#002` - Create system_capabilities table  
**Tags**: `database`, `migration`, `p0`

**Migration**: `000019_create_system_capabilities.up.sql`

```sql
-- System capabilities table (Global System Level)
CREATE TABLE system_capabilities (
    capability_key VARCHAR(255) PRIMARY KEY,
    enabled BOOLEAN NOT NULL DEFAULT false,
    default_value JSONB, -- Default configuration
    description TEXT,
    updated_by UUID REFERENCES users(id),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Insert default system capabilities
INSERT INTO system_capabilities (capability_key, enabled, default_value, description) VALUES
    ('mfa', true, '{}', 'Multi-factor authentication support'),
    ('totp', true, '{}', 'Time-based OTP support'),
    ('saml', false, '{}', 'SAML federation support'),
    ('oidc', true, '{}', 'OIDC protocol support'),
    ('oauth2', true, '{}', 'OAuth2 protocol support'),
    ('passwordless', false, '{}', 'Passwordless authentication support'),
    ('ldap', false, '{}', 'LDAP/AD integration support'),
    ('max_token_ttl', true, '{"value": "15m"}', 'Maximum token TTL (15 minutes)'),
    ('allowed_grant_types', true, '{"value": ["authorization_code", "refresh_token", "client_credentials"]}', 'Allowed OAuth grant types'),
    ('allowed_scope_namespaces', true, '{"value": ["openid", "profile", "users", "clients"]}', 'Allowed scope namespaces'),
    ('pkce_mandatory', true, '{"value": true}', 'PKCE mandatory for OAuth flows');
```

### 1.3 Create Tenant Feature Enablement Table

**Issue**: `#003` - Create tenant_feature_enablement table  
**Tags**: `database`, `migration`, `p0`

**Migration**: `000020_create_tenant_feature_enablement.up.sql`

```sql
-- Tenant feature enablement table (Tenant Choice)
CREATE TABLE tenant_feature_enablement (
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    feature_key VARCHAR(255) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT false,
    configuration JSONB, -- Feature-specific configuration
    enabled_by UUID REFERENCES users(id),
    enabled_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, feature_key)
);

CREATE INDEX idx_tenant_feature_enablement_tenant_id ON tenant_feature_enablement(tenant_id);
CREATE INDEX idx_tenant_feature_enablement_key ON tenant_feature_enablement(feature_key);
```

### 1.4 Create User Capability State Table

**Issue**: `#004` - Create user_capability_state table  
**Tags**: `database`, `migration`, `p0`

**Migration**: `000021_create_user_capability_state.up.sql`

```sql
-- User capability state table (User Enrollment/State)
CREATE TABLE user_capability_state (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    capability_key VARCHAR(255) NOT NULL,
    enrolled BOOLEAN NOT NULL DEFAULT false,
    state_data JSONB, -- e.g., TOTP secret, recovery codes
    enrolled_at TIMESTAMP,
    last_used_at TIMESTAMP,
    PRIMARY KEY (user_id, capability_key)
);

CREATE INDEX idx_user_capability_state_user_id ON user_capability_state(user_id);
CREATE INDEX idx_user_capability_state_capability ON user_capability_state(capability_key);
```

### 1.5 Update Go Models

**Issue**: `#005` - Create Go models for capability tables  
**Tags**: `backend`, `models`, `p0`

**Files to Create/Update**:
- `identity/models/system_capability.go`
- `identity/models/tenant_capability.go`
- `identity/models/tenant_feature_enablement.go`
- `identity/models/user_capability_state.go`

---

## 🔧 Phase 2: Backend Core Logic

### 2.1 Create Capability Service

**Issue**: `#006` - Implement capability evaluation service  
**Tags**: `backend`, `service`, `p0`

**File**: `identity/capability/service.go`

**Responsibilities**:
- Evaluate capability inheritance (System → Tenant → User)
- Check if feature is allowed for tenant
- Check if feature is enabled by tenant
- Check if user has enrolled in capability
- Enforce capability checks during authentication

**Key Methods**:
```go
type CapabilityService interface {
    // System level
    IsCapabilitySupported(ctx context.Context, capabilityKey string) (bool, error)
    GetSystemCapability(ctx context.Context, capabilityKey string) (*SystemCapability, error)
    
    // System → Tenant level
    IsCapabilityAllowedForTenant(ctx context.Context, tenantID uuid.UUID, capabilityKey string) (bool, error)
    GetAllowedCapabilitiesForTenant(ctx context.Context, tenantID uuid.UUID) (map[string]bool, error)
    SetTenantCapability(ctx context.Context, tenantID uuid.UUID, capabilityKey string, enabled bool, value *json.RawMessage) error
    
    // Tenant level
    IsFeatureEnabledByTenant(ctx context.Context, tenantID uuid.UUID, featureKey string) (bool, error)
    GetEnabledFeaturesForTenant(ctx context.Context, tenantID uuid.UUID) (map[string]bool, error)
    EnableFeatureForTenant(ctx context.Context, tenantID uuid.UUID, featureKey string, config *json.RawMessage) error
    
    // User level
    IsUserEnrolled(ctx context.Context, userID uuid.UUID, capabilityKey string) (bool, error)
    GetUserCapabilityState(ctx context.Context, userID uuid.UUID, capabilityKey string) (*UserCapabilityState, error)
    EnrollUserInCapability(ctx context.Context, userID uuid.UUID, capabilityKey string, stateData *json.RawMessage) error
    
    // Evaluation (combines all levels)
    EvaluateCapability(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, capabilityKey string) (*CapabilityEvaluation, error)
}
```

### 2.2 Create Capability Repositories

**Issue**: `#007` - Implement capability repositories  
**Tags**: `backend`, `repository`, `p0`

**Files**:
- `storage/interfaces/system_capability_repository.go`
- `storage/interfaces/tenant_capability_repository.go`
- `storage/interfaces/tenant_feature_enablement_repository.go`
- `storage/interfaces/user_capability_state_repository.go`
- `storage/postgres/system_capability_repository.go`
- `storage/postgres/tenant_capability_repository.go`
- `storage/postgres/tenant_feature_enablement_repository.go`
- `storage/postgres/user_capability_state_repository.go`

### 2.3 Update Authentication Flow

**Issue**: `#008` - Integrate capability checks in auth flow  
**Tags**: `backend`, `authentication`, `p0`

**Files to Update**:
- `auth/login/login.go` - Check MFA/TOTP capabilities
- `auth/mfa/mfa.go` - Enforce MFA based on capabilities
- `auth/token/token.go` - Include capability context in tokens

**Changes**:
- Before login: Check if password auth is allowed
- After password: Check if MFA is required and allowed
- Before token issuance: Validate requested scopes against allowed namespaces
- Token claims: Include capability context (informational)

### 2.4 Update OAuth/OIDC Flow

**Issue**: `#009` - Integrate capability checks in OAuth flow  
**Tags**: `backend`, `oauth`, `p0`

**Files to Update**:
- `auth/hydra/hydra.go` - Validate grant types and scopes
- OAuth client creation: Check if OIDC/OAuth2 is allowed
- Scope validation: Check against allowed scope namespaces

---

## 🌐 Phase 3: API Endpoints

### 3.1 System Capability Management APIs

**Issue**: `#010` - System capability management endpoints  
**Tags**: `api`, `system`, `p0`

**Endpoints**:
- `GET /system/capabilities` - List all system capabilities
- `GET /system/capabilities/:key` - Get system capability
- `PUT /system/capabilities/:key` - Update system capability (system_owner only)

### 3.2 Tenant Capability Assignment APIs

**Issue**: `#011` - Tenant capability assignment endpoints  
**Tags**: `api`, `system`, `p0`

**Endpoints**:
- `GET /system/tenants/:id/capabilities` - Get allowed capabilities for tenant
- `PUT /system/tenants/:id/capabilities/:key` - Assign capability to tenant
- `DELETE /system/tenants/:id/capabilities/:key` - Revoke capability from tenant
- `GET /system/tenants/:id/capabilities/evaluation` - Evaluate all capabilities for tenant

### 3.3 Tenant Feature Enablement APIs

**Issue**: `#012` - Tenant feature enablement endpoints  
**Tags**: `api`, `tenant`, `p0`

**Endpoints**:
- `GET /api/v1/tenant/features` - Get enabled features for tenant
- `PUT /api/v1/tenant/features/:key` - Enable feature for tenant
- `DELETE /api/v1/tenant/features/:key` - Disable feature for tenant

**Validation**:
- Tenant can only enable features that are allowed by system
- Tenant cannot exceed system limits

### 3.4 User Capability State APIs

**Issue**: `#013` - User capability state endpoints  
**Tags**: `api`, `user`, `p0`

**Endpoints**:
- `GET /api/v1/users/:id/capabilities` - Get user capability states
- `GET /api/v1/users/:id/capabilities/:key` - Get specific capability state
- `POST /api/v1/users/:id/capabilities/:key/enroll` - Enroll user in capability
- `DELETE /api/v1/users/:id/capabilities/:key` - Unenroll user from capability

---

## 🎨 Phase 4: Frontend Admin Dashboard

### 4.1 System Capability Management UI

**Issue**: `#014` - System capability management page  
**Tags**: `frontend`, `system`, `p0`

**File**: `frontend/admin-dashboard/src/pages/system/Capabilities.tsx`

**Features**:
- List all system capabilities with status
- Enable/disable capabilities globally
- Configure default values
- Show which tenants are using each capability
- Visual indicators for enabled/disabled state

### 4.2 Tenant Capability Assignment UI

**Issue**: `#015` - Tenant capability assignment page  
**Tags**: `frontend`, `system`, `p0`

**File**: `frontend/admin-dashboard/src/pages/system/TenantCapabilities.tsx`

**Features**:
- Select tenant from dropdown
- Show capability matrix (allowed vs not allowed)
- Toggle capabilities for selected tenant
- Configure capability-specific values (e.g., max_token_ttl)
- Visual inheritance diagram
- Bulk assignment for multiple tenants

### 4.3 Tenant Feature Enablement UI

**Issue**: `#016` - Tenant feature enablement page  
**Tags**: `frontend`, `tenant`, `p0`

**File**: `frontend/admin-dashboard/src/pages/tenant/Features.tsx`

**Features**:
- Show available features (based on allowed capabilities)
- Enable/disable features
- Configure feature settings (e.g., MFA enforcement rules)
- Visual indicators showing:
  - System support (green/gray)
  - Tenant allowed (green/gray)
  - Tenant enabled (green/gray)
- Capability inheritance visualization

### 4.4 User Capability State UI

**Issue**: `#017` - User capability enrollment page  
**Tags**: `frontend`, `user`, `p0`

**File**: `frontend/admin-dashboard/src/pages/users/UserCapabilities.tsx`

**Features**:
- Show user's capability enrollment status
- Enroll/unenroll users in capabilities
- View enrollment details (e.g., TOTP secret, recovery codes)
- Show required vs optional capabilities
- Force enrollment for required capabilities

### 4.5 Enhanced Settings Page

**Issue**: `#018` - Enhanced settings page with capability model  
**Tags**: `frontend`, `settings`, `p0`

**File**: `frontend/admin-dashboard/src/pages/Settings.tsx` (update)

**New Tabs**:
- **System Settings** (SYSTEM users only):
  - System Capabilities
  - Global Security Policies
  - Platform Guardrails
- **Tenant Capabilities** (SYSTEM users only):
  - Assign capabilities to tenants
  - Configure per-tenant limits
- **Tenant Features** (TENANT users):
  - Enable/disable features
  - Configure feature settings
- **User Capabilities** (TENANT users):
  - View user enrollment status
  - Manage user enrollments

### 4.6 Interactive Capability Visualization

**Issue**: `#019` - Capability inheritance visualization component  
**Tags**: `frontend`, `ui`, `p1`

**File**: `frontend/admin-dashboard/src/components/CapabilityInheritanceDiagram.tsx`

**Features**:
- Visual diagram showing System → Tenant → User flow
- Color-coded states (enabled, allowed, enrolled)
- Interactive tooltips
- Real-time updates
- Export as image

### 4.7 Dashboard Enhancements

**Issue**: `#020` - Enhanced dashboard with capability metrics  
**Tags**: `frontend`, `dashboard`, `p1`

**File**: `frontend/admin-dashboard/src/pages/Dashboard.tsx` (update)

**New Metrics**:
- System: Total capabilities enabled, tenants using each capability
- Tenant: Enabled features, user enrollment rates
- User: Enrollment status, required vs optional

---

## 🔒 Phase 5: Enforcement & Validation

### 5.1 Capability Enforcement Middleware

**Issue**: `#021` - Capability enforcement middleware  
**Tags**: `backend`, `middleware`, `p0`

**File**: `api/middleware/capability.go`

**Features**:
- Check capability before allowing feature usage
- Validate tenant feature enablement
- Enforce user enrollment requirements
- Return clear error messages

### 5.2 Validation Logic

**Issue**: `#022` - Capability validation logic  
**Tags**: `backend`, `validation`, `p0`

**Validation Rules**:
- Tenant cannot enable feature not allowed by system
- Tenant cannot exceed system limits (e.g., max_token_ttl)
- User cannot skip required enrollments
- System cannot bypass tenant restrictions

### 5.3 Token Capability Context

**Issue**: `#023` - Include capability context in tokens  
**Tags**: `backend`, `token`, `p0`

**File**: `auth/claims/builder.go` (update)

**Token Claims**:
```json
{
  "tenant_id": "uuid",
  "capabilities": {
    "mfa": true,
    "saml": false
  },
  "features": {
    "mfa": {
      "enabled": true,
      "required": true
    }
  }
}
```

**Note**: Informational only, not authoritative for authorization.

---

## 🧪 Phase 6: Testing & Documentation

### 6.1 Unit Tests

**Issue**: `#024` - Unit tests for capability service  
**Tags**: `testing`, `unit`, `p1`

**Coverage**:
- Capability evaluation logic
- Inheritance chain validation
- Edge cases (missing capabilities, disabled features)

### 6.2 Integration Tests

**Issue**: `#025` - Integration tests for capability APIs  
**Tags**: `testing`, `integration`, `p1`

**Test Scenarios**:
- System admin assigns capability to tenant
- Tenant admin enables feature
- User enrolls in capability
- Enforcement during authentication
- OAuth scope validation

### 6.3 E2E Tests

**Issue**: `#026` - E2E tests for capability flow  
**Tags**: `testing`, `e2e`, `p1`

**Test Scenarios**:
- Complete capability assignment → enablement → enrollment flow
- UI interactions for capability management
- Error handling and validation

### 6.4 Documentation

**Issue**: `#027` - Update documentation  
**Tags**: `documentation`, `p1`

**Documents to Create/Update**:
- `docs/architecture/CAPABILITY_MODEL.md` - Architecture overview
- `docs/guides/capability-management.md` - User guide
- `docs/api/capability-endpoints.md` - API documentation
- Update `docs/DOCUMENTATION_INDEX.md`

---

## 🚀 Phase 7: Migration & Deployment

### 7.1 Data Migration

**Issue**: `#028` - Migrate existing data to capability model  
**Tags**: `migration`, `database`, `p0`

**Migration Script**: `migrations/000022_migrate_existing_capabilities.up.sql`

**Tasks**:
- Migrate existing tenant settings to capability model
- Set default capabilities for existing tenants
- Preserve existing feature enablements

### 7.2 Deployment Plan

**Issue**: `#029` - Deployment and rollout plan  
**Tags**: `deployment`, `p1`

**Steps**:
1. Deploy database migrations
2. Run data migration script
3. Deploy backend with capability service
4. Deploy frontend with new UI
5. Monitor and validate
6. Gradual rollout to tenants

### 7.3 Rollback Plan

**Issue**: `#030` - Rollback procedures  
**Tags**: `deployment`, `p1`

**Rollback Steps**:
- Database rollback migrations
- Feature flags to disable capability checks
- Revert to previous version if needed

---

## 📋 Detailed Feature Breakdown

### Feature: MFA/TOTP

**System Level**:
- ✅ MFA supported: `system_capabilities.mfa = true`
- ✅ TOTP supported: `system_capabilities.totp = true`

**System → Tenant**:
- Tenant allowed MFA: `tenant_capabilities.mfa = true`
- Tenant allowed TOTP: `tenant_capabilities.totp = true`

**Tenant Level**:
- MFA enabled: `tenant_feature_enablement.mfa.enabled = true`
- MFA required for admins: `tenant_feature_enablement.mfa.configuration.required_for_admins = true`

**User Level**:
- TOTP enrolled: `user_capability_state.totp.enrolled = true`
- TOTP secret: `user_capability_state.totp.state_data.secret`

### Feature: OAuth2/OIDC

**System Level**:
- ✅ OIDC supported: `system_capabilities.oidc = true`
- ✅ OAuth2 supported: `system_capabilities.oauth2 = true`
- Allowed grant types: `system_capabilities.allowed_grant_types.value = ["authorization_code", ...]`
- PKCE mandatory: `system_capabilities.pkce_mandatory.value = true`

**System → Tenant**:
- Tenant allowed OIDC: `tenant_capabilities.oidc = true`
- Tenant allowed grant types: `tenant_capabilities.allowed_grant_types.value = [...]`

**Tenant Level**:
- OIDC enabled: `tenant_feature_enablement.oidc.enabled = true`
- OAuth clients created by tenant

**User Level**:
- User authenticates via OIDC (no enrollment needed)

### Feature: SAML

**System Level**:
- ❌ SAML supported: `system_capabilities.saml = false` (initially)

**System → Tenant**:
- Tenant not allowed SAML: `tenant_capabilities.saml = false`

**Tenant Level**:
- SAML cannot be enabled (not allowed)

**User Level**:
- N/A

### Feature: Token TTL

**System Level**:
- Max token TTL: `system_capabilities.max_token_ttl.value = "15m"`

**System → Tenant**:
- Tenant max token TTL: `tenant_capabilities.max_token_ttl.value = "10m"` (must be ≤ system max)

**Tenant Level**:
- Token TTL setting: `tenant_settings.access_token_ttl_minutes = 10` (must be ≤ tenant max)

**User Level**:
- User receives token with TTL (no control)

### Feature: Scope Namespaces

**System Level**:
- Allowed namespaces: `system_capabilities.allowed_scope_namespaces.value = ["openid", "profile", "users", "clients"]`

**System → Tenant**:
- Tenant allowed namespaces: `tenant_capabilities.allowed_scope_namespaces.value = ["openid", "users"]`

**Tenant Level**:
- Tenant creates scopes within allowed namespaces: `users:read`, `users:write`

**User Level**:
- User receives scopes via roles (no direct control)

---

## 🏷️ GitHub Tags Structure

### Priority Tags
- `p0` - Critical, must be done first
- `p1` - Important, should be done soon
- `p2` - Nice to have, can be deferred

### Component Tags
- `backend` - Backend changes
- `frontend` - Frontend changes
- `database` - Database/migration changes
- `api` - API endpoint changes
- `testing` - Test-related work
- `documentation` - Documentation updates

### Feature Tags
- `capability-model` - Core capability model
- `system` - System-level features
- `tenant` - Tenant-level features
- `user` - User-level features
- `mfa` - MFA/TOTP features
- `oauth` - OAuth2/OIDC features
- `saml` - SAML features
- `security` - Security-related

### Type Tags
- `migration` - Database migration
- `service` - Service layer
- `repository` - Repository layer
- `middleware` - Middleware
- `ui` - UI component
- `integration` - Integration work

---

## 📊 Progress Tracking

See `docs/status/CAPABILITY_MODEL_STATUS.md` for detailed progress tracking.

---

## 🔗 Related Documents

- `feature_capibility.md` - Source of truth for capability model
- `docs/architecture/DESIGN_VALIDATION.md` - Design validation
- `docs/security/MASTER_TENANT_IMPLEMENTATION_PLAN_V2.md` - System/Tenant separation
- `docs/architecture/frontend/ADMIN_DASHBOARD_V2_ARCHITECTURE.md` - Frontend architecture

---

## ✅ Next Steps

1. Review and approve this implementation plan
2. Create GitHub issues for each phase
3. Set up project board with phases
4. Begin Phase 1 implementation
5. Regular status updates in `docs/status/CAPABILITY_MODEL_STATUS.md`

