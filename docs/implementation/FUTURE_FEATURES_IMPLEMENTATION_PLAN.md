# Future Features Implementation Plan

**Last Updated**: 2025-01-10  
**Status**: Planning Phase  
**Priority**: Based on Expert Review Feedback

---

## 📋 Table of Contents

1. [Missing Features (Should be Planned)](#missing-features-should-be-planned)
2. [High Value Next Features](#high-value-next-features)
3. [Nice to Have Later Features](#nice-to-have-later-features)
4. [Documentation Improvements](#documentation-improvements)
5. [Implementation Priorities](#implementation-priorities)

---

## Missing Features (Should be Planned)

### 1. Audit Events (Structured)

**Status**: ⚠️ **MISSING - HIGH PRIORITY**

**Current State**: 
- Basic logging exists
- No structured audit event system
- No event storage/querying

**What's Needed**:

**Structured Audit Events**:
- `user.created`
- `user.updated`
- `user.deleted`
- `user.locked`
- `user.unlocked`
- `role.assigned`
- `role.removed`
- `permission.assigned`
- `permission.removed`
- `mfa.enrolled`
- `mfa.verified`
- `mfa.disabled`
- `tenant.created`
- `tenant.suspended`
- `tenant.resumed`
- `tenant.deleted`
- `login.success`
- `login.failure`
- `token.issued`
- `token.revoked`

**Event Structure**:
```go
type AuditEvent struct {
    ID          uuid.UUID              `json:"id"`
    EventType   string                 `json:"event_type"`   // e.g., "user.created"
    Actor       AuditActor             `json:"actor"`        // Who performed the action
    Target      AuditTarget            `json:"target"`       // What was affected
    Timestamp   time.Time              `json:"timestamp"`
    SourceIP    string                 `json:"source_ip"`
    UserAgent   string                 `json:"user_agent"`
    TenantID    *uuid.UUID             `json:"tenant_id,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    Result      string                 `json:"result"`        // "success", "failure", "denied"
    Error       string                 `json:"error,omitempty"`
}

type AuditActor struct {
    UserID      uuid.UUID `json:"user_id"`
    Username    string    `json:"username"`
    PrincipalType string  `json:"principal_type"`
}

type AuditTarget struct {
    Type        string    `json:"type"`        // "user", "role", "tenant", etc.
    ID          uuid.UUID `json:"id"`
    Identifier  string    `json:"identifier"`   // username, role name, etc.
}
```

**Implementation Plan**:

1. **Database Schema**:
   ```sql
   CREATE TABLE audit_events (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       event_type VARCHAR(100) NOT NULL,
       actor_user_id UUID NOT NULL,
       actor_username VARCHAR(255) NOT NULL,
       actor_principal_type VARCHAR(20) NOT NULL,
       target_type VARCHAR(50),
       target_id UUID,
       target_identifier VARCHAR(255),
       timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
       source_ip INET,
       user_agent TEXT,
       tenant_id UUID,
       metadata JSONB,
       result VARCHAR(20) NOT NULL,
       error TEXT,
       created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
   );

   CREATE INDEX idx_audit_events_event_type ON audit_events(event_type);
   CREATE INDEX idx_audit_events_actor_user_id ON audit_events(actor_user_id);
   CREATE INDEX idx_audit_events_target_id ON audit_events(target_id);
   CREATE INDEX idx_audit_events_tenant_id ON audit_events(tenant_id);
   CREATE INDEX idx_audit_events_timestamp ON audit_events(timestamp DESC);
   ```

2. **Service Layer**:
   ```go
   type AuditService interface {
       LogEvent(ctx context.Context, event *AuditEvent) error
       QueryEvents(ctx context.Context, filters *AuditFilters) ([]*AuditEvent, error)
       GetEvent(ctx context.Context, eventID uuid.UUID) (*AuditEvent, error)
   }
   ```

3. **Integration Points**:
   - User service: Log user CRUD operations
   - Role service: Log role assignments/removals
   - Permission service: Log permission changes
   - Auth service: Log login attempts
   - MFA service: Log MFA events
   - Tenant service: Log tenant lifecycle events

4. **API Endpoints**:
   - `GET /api/v1/audit/events` - Query audit events
   - `GET /api/v1/audit/events/:id` - Get specific event
   - `GET /system/audit/events` - System-wide audit (SYSTEM users only)

**Files to Create**:
- `identity/audit/service.go` - Audit service
- `identity/audit/model.go` - Audit event models
- `storage/interfaces/audit_repository.go` - Repository interface
- `storage/postgres/audit_repository.go` - PostgreSQL implementation
- `api/handlers/audit_handler.go` - HTTP handlers
- `migrations/000024_create_audit_events.up.sql` - Migration

**Estimated Effort**: 3-5 days

---

### 2. Event Hooks / Webhooks

**Status**: ⚠️ **MISSING - MEDIUM PRIORITY**

**What's Needed**:

**Webhook System**:
- Configurable webhook endpoints per tenant
- Event subscriptions (which events to send)
- Retry logic with exponential backoff
- Webhook secret signing
- Webhook delivery status tracking

**Implementation Plan**:

1. **Database Schema**:
   ```sql
   CREATE TABLE webhooks (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
       url TEXT NOT NULL,
       secret TEXT NOT NULL,
       events TEXT[] NOT NULL, -- Array of event types
       enabled BOOLEAN NOT NULL DEFAULT true,
       created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
       updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
   );

   CREATE TABLE webhook_deliveries (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       webhook_id UUID REFERENCES webhooks(id) ON DELETE CASCADE,
       event_id UUID REFERENCES audit_events(id),
       status VARCHAR(20) NOT NULL, -- "pending", "success", "failed"
       attempts INT NOT NULL DEFAULT 0,
       last_attempt_at TIMESTAMP WITH TIME ZONE,
       next_retry_at TIMESTAMP WITH TIME ZONE,
       response_code INT,
       response_body TEXT,
       error_message TEXT,
       created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
   );
   ```

2. **Service Layer**:
   ```go
   type WebhookService interface {
       CreateWebhook(ctx context.Context, tenantID uuid.UUID, req *CreateWebhookRequest) (*Webhook, error)
       TriggerWebhook(ctx context.Context, event *AuditEvent) error
       RetryFailedWebhooks(ctx context.Context) error
   }
   ```

3. **Webhook Payload**:
   ```json
   {
     "event": {
       "id": "event-uuid",
       "type": "user.created",
       "timestamp": "2025-01-10T12:00:00Z",
       "actor": {
         "user_id": "actor-uuid",
         "username": "admin"
       },
       "target": {
         "type": "user",
         "id": "user-uuid",
         "identifier": "john.doe"
       }
     },
     "signature": "sha256=..."
   }
   ```

4. **API Endpoints**:
   - `POST /api/v1/webhooks` - Create webhook
   - `GET /api/v1/webhooks` - List webhooks
   - `GET /api/v1/webhooks/:id` - Get webhook
   - `PUT /api/v1/webhooks/:id` - Update webhook
   - `DELETE /api/v1/webhooks/:id` - Delete webhook
   - `GET /api/v1/webhooks/:id/deliveries` - Get delivery history

**Files to Create**:
- `identity/webhook/service.go` - Webhook service
- `identity/webhook/model.go` - Webhook models
- `storage/interfaces/webhook_repository.go` - Repository interface
- `storage/postgres/webhook_repository.go` - PostgreSQL implementation
- `api/handlers/webhook_handler.go` - HTTP handlers
- `internal/webhook/dispatcher.go` - Async webhook dispatcher
- `migrations/000025_create_webhooks.up.sql` - Migration

**Estimated Effort**: 5-7 days

---

### 3. Federation (OIDC/SAML Login)

**Status**: ⚠️ **MISSING - HIGH PRIORITY**

**What's Needed**:

**OIDC Federation**:
- External OIDC provider configuration
- OIDC login flow
- Identity provider discovery
- Token exchange

**SAML Federation**:
- SAML IdP configuration
- SAML SSO flow
- SAML assertion validation
- Attribute mapping

**Implementation Plan**:

1. **Database Schema**:
   ```sql
   CREATE TABLE identity_providers (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
       name VARCHAR(255) NOT NULL,
       type VARCHAR(20) NOT NULL, -- "oidc", "saml"
       enabled BOOLEAN NOT NULL DEFAULT true,
       configuration JSONB NOT NULL,
       attribute_mapping JSONB,
       created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
       updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
   );

   CREATE TABLE federated_identities (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       user_id UUID REFERENCES users(id) ON DELETE CASCADE,
       provider_id UUID REFERENCES identity_providers(id) ON DELETE CASCADE,
       external_id VARCHAR(255) NOT NULL,
       attributes JSONB,
       created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
       UNIQUE(provider_id, external_id)
   );
   ```

2. **OIDC Flow**:
   ```
   1. User clicks "Login with Google" (or other OIDC provider)
   2. Redirect to OIDC provider authorization endpoint
   3. User authenticates with OIDC provider
   4. OIDC provider redirects back with authorization code
   5. Exchange code for ID token
   6. Validate ID token
   7. Extract user attributes
   8. Find or create user in ARauth
   9. Link federated identity
   10. Issue ARauth tokens
   ```

3. **SAML Flow**:
   ```
   1. User clicks "Login with SAML"
   2. Generate SAML AuthnRequest
   3. Redirect to SAML IdP
   4. User authenticates with SAML IdP
   5. SAML IdP redirects back with SAML assertion
   6. Validate SAML assertion
   7. Extract user attributes
   8. Find or create user in ARauth
   9. Link federated identity
   10. Issue ARauth tokens
   ```

4. **Service Layer**:
   ```go
   type FederationService interface {
       CreateIdentityProvider(ctx context.Context, tenantID uuid.UUID, req *CreateIdPRequest) (*IdentityProvider, error)
       InitiateOIDCLogin(ctx context.Context, tenantID uuid.UUID, providerID uuid.UUID) (string, error) // Returns redirect URL
       HandleOIDCCallback(ctx context.Context, code string, state string) (*LoginResponse, error)
       InitiateSAMLLogin(ctx context.Context, tenantID uuid.UUID, providerID uuid.UUID) (string, error) // Returns redirect URL
       HandleSAMLCallback(ctx context.Context, samlResponse string) (*LoginResponse, error)
   }
   ```

5. **API Endpoints**:
   - `POST /api/v1/identity-providers` - Create IdP
   - `GET /api/v1/identity-providers` - List IdPs
   - `GET /api/v1/identity-providers/:id` - Get IdP
   - `PUT /api/v1/identity-providers/:id` - Update IdP
   - `DELETE /api/v1/identity-providers/:id` - Delete IdP
   - `POST /api/v1/auth/federation/oidc/:provider_id/initiate` - Start OIDC login
   - `POST /api/v1/auth/federation/oidc/callback` - Handle OIDC callback
   - `POST /api/v1/auth/federation/saml/:provider_id/initiate` - Start SAML login
   - `POST /api/v1/auth/federation/saml/callback` - Handle SAML callback

**Files to Create**:
- `auth/federation/service.go` - Federation service
- `auth/federation/oidc/client.go` - OIDC client
- `auth/federation/saml/client.go` - SAML client
- `identity/federation/model.go` - Federation models
- `storage/interfaces/federation_repository.go` - Repository interface
- `storage/postgres/federation_repository.go` - PostgreSQL implementation
- `api/handlers/federation_handler.go` - HTTP handlers
- `migrations/000026_create_federation.up.sql` - Migration

**Estimated Effort**: 10-15 days

---

### 4. Identity Linking

**Status**: ⚠️ **MISSING - MEDIUM PRIORITY**

**What's Needed**:

**Identity Linking**:
- One user can have multiple identities (password + SAML + OIDC)
- Link/unlink identities
- Primary identity designation
- Identity verification

**Implementation Plan**:

1. **Database Schema** (extends federated_identities):
   ```sql
   ALTER TABLE federated_identities ADD COLUMN is_primary BOOLEAN NOT NULL DEFAULT false;
   ALTER TABLE federated_identities ADD COLUMN verified BOOLEAN NOT NULL DEFAULT false;
   ALTER TABLE federated_identities ADD COLUMN verified_at TIMESTAMP WITH TIME ZONE;
   ```

2. **Service Layer**:
   ```go
   type IdentityLinkingService interface {
       LinkIdentity(ctx context.Context, userID uuid.UUID, providerID uuid.UUID, externalID string) error
       UnlinkIdentity(ctx context.Context, userID uuid.UUID, providerID uuid.UUID) error
       SetPrimaryIdentity(ctx context.Context, userID uuid.UUID, identityID uuid.UUID) error
       GetUserIdentities(ctx context.Context, userID uuid.UUID) ([]*FederatedIdentity, error)
   }
   ```

3. **Login Flow with Multiple Identities**:
   ```
   1. User attempts login with any identity (password, OIDC, SAML)
   2. System finds user by identity
   3. System checks all linked identities
   4. User can choose which identity to use for login
   5. Issue tokens
   ```

4. **API Endpoints**:
   - `POST /api/v1/users/:id/identities` - Link identity
   - `DELETE /api/v1/users/:id/identities/:identity_id` - Unlink identity
   - `PUT /api/v1/users/:id/identities/:identity_id/primary` - Set primary identity
   - `GET /api/v1/users/:id/identities` - List user identities

**Files to Create**:
- `identity/linking/service.go` - Identity linking service
- `api/handlers/identity_linking_handler.go` - HTTP handlers

**Estimated Effort**: 3-4 days

---

## High Value Next Features

### 1. Permission → OAuth Scope Mapping

**Status**: ⏸️ **DEFERRED - HIGH VALUE**

**What's Needed**:

**Scope Mapping**:
- Map permissions to OAuth scopes
- Tenant-configurable scope definitions
- Scope-based token claims

**Implementation Plan**:

1. **Database Schema**:
   ```sql
   CREATE TABLE oauth_scopes (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
       name VARCHAR(100) NOT NULL,
       description TEXT,
       permissions UUID[] NOT NULL, -- Array of permission IDs
       created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
       UNIQUE(tenant_id, name)
   );
   ```

2. **Service Layer**:
   ```go
   type ScopeMappingService interface {
       CreateScope(ctx context.Context, tenantID uuid.UUID, req *CreateScopeRequest) (*OAuthScope, error)
       MapPermissionsToScope(ctx context.Context, scopeID uuid.UUID, permissionIDs []uuid.UUID) error
       GetScopePermissions(ctx context.Context, scopeID uuid.UUID) ([]*Permission, error)
   }
   ```

3. **Token Claims Enhancement**:
   - Include scopes in token claims
   - Filter permissions by requested scopes
   - Scope-based permission evaluation

**Files to Create**:
- `identity/scope/service.go` - Scope mapping service
- `identity/scope/model.go` - Scope models
- `storage/interfaces/scope_repository.go` - Repository interface
- `storage/postgres/scope_repository.go` - PostgreSQL implementation
- `api/handlers/scope_handler.go` - HTTP handlers
- `migrations/000027_create_oauth_scopes.up.sql` - Migration

**Estimated Effort**: 4-5 days

---

### 2. SCIM Provisioning

**Status**: ⏸️ **DEFERRED - HIGH VALUE**

**What's Needed**:

**SCIM 2.0 API**:
- User provisioning (create, update, delete)
- Group provisioning
- Bulk operations
- SCIM filters

**Implementation Plan**:

1. **SCIM Endpoints**:
   - `GET /scim/v2/Users` - List users
   - `POST /scim/v2/Users` - Create user
   - `GET /scim/v2/Users/:id` - Get user
   - `PUT /scim/v2/Users/:id` - Update user
   - `PATCH /scim/v2/Users/:id` - Partial update
   - `DELETE /scim/v2/Users/:id` - Delete user
   - Similar endpoints for Groups

2. **SCIM Authentication**:
   - Bearer token authentication
   - Tenant-scoped access

3. **Service Layer**:
   ```go
   type SCIMService interface {
       ListUsers(ctx context.Context, tenantID uuid.UUID, filters *SCIMFilters) (*SCIMListResponse, error)
       CreateUser(ctx context.Context, tenantID uuid.UUID, user *SCIMUser) (*SCIMUser, error)
       UpdateUser(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, user *SCIMUser) (*SCIMUser, error)
       DeleteUser(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) error
   }
   ```

**Files to Create**:
- `api/scim/service.go` - SCIM service
- `api/scim/models.go` - SCIM models
- `api/handlers/scim_handler.go` - SCIM handlers
- `api/middleware/scim_auth.go` - SCIM authentication

**Estimated Effort**: 7-10 days

---

### 3. Invite-Based User Onboarding

**Status**: ⏸️ **DEFERRED - HIGH VALUE**

**What's Needed**:

**User Invitations**:
- Generate invitation tokens
- Send invitation emails
- Accept invitation flow
- Invitation expiration

**Implementation Plan**:

1. **Database Schema**:
   ```sql
   CREATE TABLE user_invitations (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
       email VARCHAR(255) NOT NULL,
       token VARCHAR(255) NOT NULL UNIQUE,
       invited_by UUID REFERENCES users(id),
       role_ids UUID[],
       expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
       accepted_at TIMESTAMP WITH TIME ZONE,
       created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
   );
   ```

2. **Service Layer**:
   ```go
   type InvitationService interface {
       CreateInvitation(ctx context.Context, tenantID uuid.UUID, req *CreateInvitationRequest) (*Invitation, error)
       SendInvitation(ctx context.Context, invitationID uuid.UUID) error
       AcceptInvitation(ctx context.Context, token string, password string) (*User, error)
       ResendInvitation(ctx context.Context, invitationID uuid.UUID) error
       RevokeInvitation(ctx context.Context, invitationID uuid.UUID) error
   }
   ```

3. **API Endpoints**:
   - `POST /api/v1/invitations` - Create invitation
   - `GET /api/v1/invitations` - List invitations
   - `GET /api/v1/invitations/:id` - Get invitation
   - `POST /api/v1/invitations/:id/resend` - Resend invitation
   - `DELETE /api/v1/invitations/:id` - Revoke invitation
   - `POST /api/v1/invitations/accept` - Accept invitation

**Files to Create**:
- `identity/invitation/service.go` - Invitation service
- `identity/invitation/model.go` - Invitation models
- `storage/interfaces/invitation_repository.go` - Repository interface
- `storage/postgres/invitation_repository.go` - PostgreSQL implementation
- `api/handlers/invitation_handler.go` - HTTP handlers
- `migrations/000028_create_user_invitations.up.sql` - Migration

**Estimated Effort**: 4-5 days

---

### 4. Session Introspection Endpoint

**Status**: ⏸️ **DEFERRED - MEDIUM VALUE**

**What's Needed**:

**Token Introspection**:
- RFC 7662 compliant endpoint
- Token validation
- Token metadata retrieval

**Implementation Plan**:

1. **Endpoint**:
   - `POST /api/v1/introspect` - Token introspection

2. **Request/Response**:
   ```json
   // Request
   {
     "token": "eyJhbGci...",
     "token_type_hint": "access_token"
   }

   // Response
   {
     "active": true,
     "scope": "openid profile email",
     "client_id": "client-uuid",
     "username": "john.doe",
     "exp": 1234567890,
     "iat": 1234567890,
     "sub": "user-uuid",
     "aud": "client-uuid",
     "iss": "https://iam.example.com"
   }
   ```

3. **Service Layer**:
   ```go
   type IntrospectionService interface {
       IntrospectToken(ctx context.Context, token string) (*TokenIntrospection, error)
   }
   ```

**Files to Create**:
- `auth/introspection/service.go` - Introspection service
- `api/handlers/introspection_handler.go` - HTTP handler

**Estimated Effort**: 2-3 days

---

### 5. Admin Impersonation

**Status**: ⏸️ **DEFERRED - MEDIUM VALUE**

**What's Needed**:

**Impersonation**:
- SYSTEM/TENANT admins can impersonate users
- Explicit impersonation tokens
- Audit logging of impersonation
- Time-limited impersonation

**Implementation Plan**:

1. **Database Schema**:
   ```sql
   CREATE TABLE impersonation_sessions (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       impersonator_id UUID REFERENCES users(id) NOT NULL,
       impersonated_user_id UUID REFERENCES users(id) NOT NULL,
       token TEXT NOT NULL UNIQUE,
       expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
       created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
   );
   ```

2. **Service Layer**:
   ```go
   type ImpersonationService interface {
       StartImpersonation(ctx context.Context, impersonatorID uuid.UUID, targetUserID uuid.UUID) (*ImpersonationSession, error)
       EndImpersonation(ctx context.Context, sessionID uuid.UUID) error
       ValidateImpersonationToken(ctx context.Context, token string) (*ImpersonationSession, error)
   }
   ```

3. **Token Claims Enhancement**:
   - Add `impersonated_by` claim
   - Add `is_impersonation` flag

4. **API Endpoints**:
   - `POST /api/v1/impersonation/start` - Start impersonation
   - `POST /api/v1/impersonation/end` - End impersonation
   - `GET /api/v1/impersonation/sessions` - List active sessions

**Files to Create**:
- `auth/impersonation/service.go` - Impersonation service
- `auth/impersonation/model.go` - Impersonation models
- `storage/interfaces/impersonation_repository.go` - Repository interface
- `storage/postgres/impersonation_repository.go` - PostgreSQL implementation
- `api/handlers/impersonation_handler.go` - HTTP handlers
- `migrations/000029_create_impersonation_sessions.up.sql` - Migration

**Estimated Effort**: 3-4 days

---

## Nice to Have Later Features

### 1. WebAuthn / Passkeys

**Status**: ⏸️ **DEFERRED - FUTURE**

**What's Needed**:
- WebAuthn credential registration
- Passkey authentication
- Multiple passkeys per user
- Backup codes

**Estimated Effort**: 7-10 days

---

### 2. Risk-Based Authentication

**Status**: ⏸️ **DEFERRED - FUTURE**

**What's Needed**:
- IP-based risk scoring
- Geographic risk analysis
- Device fingerprinting
- Behavioral analysis
- Adaptive MFA

**Estimated Effort**: 10-15 days

---

### 3. Conditional Access Policies

**Status**: ⏸️ **DEFERRED - FUTURE**

**What's Needed**:
- Policy engine (OPA-compatible)
- Policy definition language
- Policy evaluation
- Policy-based access control

**Estimated Effort**: 15-20 days

---

### 4. Device Trust

**Status**: ⏸️ **DEFERRED - FUTURE**

**What's Needed**:
- Device registration
- Device fingerprinting
- Trusted device management
- Device-based access policies

**Estimated Effort**: 7-10 days

---

## Documentation Improvements

### 1. Session State Clarification

**Status**: ⚠️ **MISSING**

**What to Add**:
> "ARauth does not store application session state. Applications are responsible for session handling using issued tokens. All authentication state is contained within JWTs or external storage (Redis for temporary data like MFA challenges)."

**Location**: System Overview section

---

### 2. Login Identifiers Documentation

**Status**: ⚠️ **MISSING**

**What to Add**:
- Document supported login identifiers:
  - Username
  - Email
  - Phone (future)
- Case-sensitivity rules
- Identifier uniqueness rules

**Location**: Authentication Features section

---

### 3. MFA Reset/Recovery Flow

**Status**: ⚠️ **MISSING**

**What to Add**:
- Who can reset MFA (tenant admin, tenant owner)
- MFA reset process
- Session invalidation on reset
- Force re-enrollment

**Location**: Multi-Factor Authentication section

---

### 4. Capability vs Feature Key Clarification

**Status**: ⚠️ **MISSING**

**What to Add**:
- Clarify terminology:
  - **Capability** = Platform concept (what ARauth supports)
  - **Feature** = Tenant-enabled usage of a capability
- Document when to use each term

**Location**: Capability Model section

---

### 5. Tenant Deletion Lifecycle

**Status**: ⚠️ **MISSING**

**What to Add**:
- Soft delete vs hard delete
- Retention window
- Token invalidation on delete
- Data retention policy

**Location**: Tenant Management section

---

### 6. User Status Lifecycle

**Status**: ⚠️ **MISSING**

**What to Add**:
- Document user statuses:
  - `active` - User can login
  - `disabled` - User cannot login (admin action)
  - `locked` - User locked due to failed attempts
  - `pending` - User invited but not activated
- Status transition rules

**Location**: User Management section

---

### 7. Allow-Only RBAC Documentation

**Status**: ⚠️ **MISSING**

**What to Add**:
- Document that permissions are additive only
- No deny rules supported
- Permission evaluation is allow-only
- Future consideration: deny rules

**Location**: Permission System section

---

### 8. Token Size Considerations

**Status**: ⚠️ **MISSING**

**What to Add**:
- Document token size implications
- Recommendation: avoid excessive fine-grained permissions
- Future considerations:
  - Permission hashes
  - Permission versioning
  - Token compression

**Location**: Token Management section

---

### 9. Credential Rotation Events

**Status**: ⚠️ **MISSING**

**What to Add**:
- Document revocation strategy:
  - Password change → revoke tokens
  - MFA reset → revoke tokens
  - Role change → optional token refresh
- Token invalidation policies

**Location**: Security Features section

---

### 10. Admin Dashboard as Reference UI

**Status**: ⚠️ **MISSING**

**What to Add**:
- Document that admin dashboard is a reference UI
- Enterprises expected to build custom admin UIs
- Headless positioning maintained

**Location**: Admin Dashboard section

---

## Implementation Priorities

### Phase 1: Critical Missing Features (Next 2-3 Months)

1. **Audit Events** (3-5 days) - HIGH PRIORITY
2. **Federation (OIDC/SAML)** (10-15 days) - HIGH PRIORITY
3. **Event Hooks / Webhooks** (5-7 days) - MEDIUM PRIORITY
4. **Identity Linking** (3-4 days) - MEDIUM PRIORITY

**Total Estimated Effort**: 21-31 days

---

### Phase 2: High Value Features (3-6 Months)

1. **Permission → OAuth Scope Mapping** (4-5 days)
2. **SCIM Provisioning** (7-10 days)
3. **Invite-Based User Onboarding** (4-5 days)
4. **Session Introspection** (2-3 days)
5. **Admin Impersonation** (3-4 days)

**Total Estimated Effort**: 20-27 days

---

### Phase 3: Documentation Improvements (Ongoing)

- All documentation improvements listed above
- Update existing documentation
- Create new documentation sections

**Total Estimated Effort**: 3-5 days

---

### Phase 4: Future Enhancements (6+ Months)

1. WebAuthn / Passkeys
2. Risk-Based Authentication
3. Conditional Access Policies
4. Device Trust

**Total Estimated Effort**: 39-55 days

---

## Summary

**Current Status**: 95% Complete (Core Features)

**Missing Critical Features**: 4 features (21-31 days)
**High Value Features**: 5 features (20-27 days)
**Documentation Improvements**: 10 items (3-5 days)
**Future Enhancements**: 4 features (39-55 days)

**Total Remaining Work**: 83-118 days

**Recommended Next Steps**:
1. Implement Audit Events (foundation for everything)
2. Implement Federation (OIDC/SAML) (biggest enterprise ask)
3. Update documentation with missing clarifications
4. Implement Event Hooks / Webhooks
5. Implement Identity Linking

---

**Last Updated**: 2025-01-10  
**Document Version**: 1.0

