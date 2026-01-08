# GitHub Issues for Capability Model Implementation

This document contains all GitHub issues that need to be created for the capability model implementation. Each issue includes title, description, labels, and acceptance criteria.

---

## Phase 1: Database & Models

### Issue #001: Create tenant_capabilities table
**Title**: `[Phase 1] Create tenant_capabilities table for System→Tenant capability assignment`

**Description**:
Create database migration and table for storing which capabilities are allowed for each tenant. This implements the "System → Tenant" layer of the capability model.

**Labels**: `database`, `migration`, `p0`, `capability-model`, `phase-1`

**Acceptance Criteria**:
- [ ] Migration file `000018_create_tenant_capabilities.up.sql` created
- [ ] Migration file `000018_create_tenant_capabilities.down.sql` created
- [ ] Table includes: tenant_id, capability_key, enabled, value (JSONB), configured_by, configured_at
- [ ] Primary key on (tenant_id, capability_key)
- [ ] Indexes created for tenant_id and capability_key
- [ ] Migration tested and verified

**Related**: #002, #005

---

### Issue #002: Create system_capabilities table
**Title**: `[Phase 1] Create system_capabilities table for global system capabilities`

**Description**:
Create database migration and table for storing global system-level capabilities. This implements the "System" layer of the capability model.

**Labels**: `database`, `migration`, `p0`, `capability-model`, `phase-1`

**Acceptance Criteria**:
- [ ] Migration file `000019_create_system_capabilities.up.sql` created
- [ ] Migration file `000019_create_system_capabilities.down.sql` created
- [ ] Table includes: capability_key, enabled, default_value (JSONB), description, updated_by, updated_at
- [ ] Default capabilities inserted: mfa, totp, saml, oidc, oauth2, passwordless, ldap, max_token_ttl, allowed_grant_types, allowed_scope_namespaces, pkce_mandatory
- [ ] Migration tested and verified

**Related**: #001, #005

---

### Issue #003: Create tenant_feature_enablement table
**Title**: `[Phase 1] Create tenant_feature_enablement table for tenant feature choices`

**Description**:
Create database migration and table for storing which features tenants have actually enabled. This implements the "Tenant" layer of the capability model.

**Labels**: `database`, `migration`, `p0`, `capability-model`, `phase-1`

**Acceptance Criteria**:
- [ ] Migration file `000020_create_tenant_feature_enablement.up.sql` created
- [ ] Migration file `000020_create_tenant_feature_enablement.down.sql` created
- [ ] Table includes: tenant_id, feature_key, enabled, configuration (JSONB), enabled_by, enabled_at
- [ ] Primary key on (tenant_id, feature_key)
- [ ] Indexes created for tenant_id and feature_key
- [ ] Migration tested and verified

**Related**: #001, #002, #005

---

### Issue #004: Create user_capability_state table
**Title**: `[Phase 1] Create user_capability_state table for user enrollment state`

**Description**:
Create database migration and table for storing user-level capability enrollment state (e.g., TOTP secrets, MFA enrollment status).

**Labels**: `database`, `migration`, `p0`, `capability-model`, `phase-1`

**Acceptance Criteria**:
- [ ] Migration file `000021_create_user_capability_state.up.sql` created
- [ ] Migration file `000021_create_user_capability_state.down.sql` created
- [ ] Table includes: user_id, capability_key, enrolled, state_data (JSONB), enrolled_at, last_used_at
- [ ] Primary key on (user_id, capability_key)
- [ ] Indexes created for user_id and capability_key
- [ ] Migration tested and verified

**Related**: #005

---

### Issue #005: Create Go models for capability tables
**Title**: `[Phase 1] Create Go models for all capability-related tables`

**Description**:
Create Go model structs and validation logic for system_capabilities, tenant_capabilities, tenant_feature_enablement, and user_capability_state tables.

**Labels**: `backend`, `models`, `p0`, `capability-model`, `phase-1`

**Acceptance Criteria**:
- [ ] Model file `identity/models/system_capability.go` created
- [ ] Model file `identity/models/tenant_capability.go` created
- [ ] Model file `identity/models/tenant_feature_enablement.go` created
- [ ] Model file `identity/models/user_capability_state.go` created
- [ ] All models include proper JSON tags and validation
- [ ] Models include helper methods (IsEnabled, GetValue, etc.)
- [ ] Unit tests for models

**Related**: #001, #002, #003, #004, #007

---

## Phase 2: Backend Core Logic

### Issue #006: Implement capability evaluation service
**Title**: `[Phase 2] Implement capability service for three-layer evaluation`

**Description**:
Create the core capability service that evaluates capabilities across System → Tenant → User layers. This is the heart of the capability model.

**Labels**: `backend`, `service`, `p0`, `capability-model`, `phase-2`

**Acceptance Criteria**:
- [ ] Service file `identity/capability/service.go` created
- [ ] Interface `CapabilityService` defined with all required methods
- [ ] Implementation handles System level checks
- [ ] Implementation handles System→Tenant level checks
- [ ] Implementation handles Tenant level checks
- [ ] Implementation handles User level checks
- [ ] `EvaluateCapability` method combines all levels correctly
- [ ] Comprehensive unit tests (90%+ coverage)
- [ ] Integration tests with database

**Related**: #007, #008

---

### Issue #007: Implement capability repositories
**Title**: `[Phase 2] Implement repository layer for capability tables`

**Description**:
Create repository interfaces and PostgreSQL implementations for all capability-related tables.

**Labels**: `backend`, `repository`, `p0`, `capability-model`, `phase-2`

**Acceptance Criteria**:
- [ ] Interface files created in `storage/interfaces/`:
  - `system_capability_repository.go`
  - `tenant_capability_repository.go`
  - `tenant_feature_enablement_repository.go`
  - `user_capability_state_repository.go`
- [ ] Implementation files created in `storage/postgres/`:
  - `system_capability_repository.go`
  - `tenant_capability_repository.go`
  - `tenant_feature_enablement_repository.go`
  - `user_capability_state_repository.go`
- [ ] All CRUD operations implemented
- [ ] Proper error handling
- [ ] Unit tests for repositories
- [ ] Integration tests with database

**Related**: #005, #006

---

### Issue #008: Integrate capability checks in auth flow
**Title**: `[Phase 2] Integrate capability checks in authentication flow`

**Description**:
Update authentication flow to check capabilities before allowing login, MFA, and token issuance.

**Labels**: `backend`, `authentication`, `p0`, `capability-model`, `phase-2`, `security`

**Acceptance Criteria**:
- [ ] `auth/login/login.go` checks if password auth is allowed
- [ ] `auth/login/login.go` checks if MFA is required and allowed
- [ ] `auth/mfa/mfa.go` enforces MFA based on capabilities
- [ ] `auth/token/token.go` validates scopes against allowed namespaces
- [ ] Error messages are clear and actionable
- [ ] Integration tests for auth flow with capabilities
- [ ] E2E tests for login with capability restrictions

**Related**: #006, #009, #021

---

### Issue #009: Integrate capability checks in OAuth flow
**Title**: `[Phase 2] Integrate capability checks in OAuth2/OIDC flow`

**Description**:
Update OAuth/OIDC flow to validate grant types, scopes, and PKCE requirements based on capabilities.

**Labels**: `backend`, `oauth`, `p0`, `capability-model`, `phase-2`, `security`

**Acceptance Criteria**:
- [ ] `auth/hydra/hydra.go` validates grant types against allowed list
- [ ] Scope validation checks against allowed scope namespaces
- [ ] PKCE enforcement based on system capability
- [ ] OAuth client creation checks if OIDC/OAuth2 is allowed
- [ ] Integration tests for OAuth flow with capabilities
- [ ] E2E tests for OAuth with capability restrictions

**Related**: #006, #008, #021

---

## Phase 3: API Endpoints

### Issue #010: System capability management endpoints
**Title**: `[Phase 3] Create API endpoints for system capability management`

**Description**:
Create REST API endpoints for system admins to manage global system capabilities.

**Labels**: `api`, `system`, `p0`, `capability-model`, `phase-3`

**Acceptance Criteria**:
- [ ] `GET /system/capabilities` - List all system capabilities
- [ ] `GET /system/capabilities/:key` - Get specific system capability
- [ ] `PUT /system/capabilities/:key` - Update system capability (system_owner only)
- [ ] Proper authorization checks (system_owner permission)
- [ ] Request/response validation
- [ ] API documentation updated
- [ ] Integration tests for all endpoints
- [ ] E2E tests

**Related**: #006, #014

---

### Issue #011: Tenant capability assignment endpoints
**Title**: `[Phase 3] Create API endpoints for tenant capability assignment`

**Description**:
Create REST API endpoints for system admins to assign capabilities to tenants.

**Labels**: `api`, `system`, `p0`, `capability-model`, `phase-3`

**Acceptance Criteria**:
- [ ] `GET /system/tenants/:id/capabilities` - Get allowed capabilities for tenant
- [ ] `PUT /system/tenants/:id/capabilities/:key` - Assign capability to tenant
- [ ] `DELETE /system/tenants/:id/capabilities/:key` - Revoke capability from tenant
- [ ] `GET /system/tenants/:id/capabilities/evaluation` - Evaluate all capabilities
- [ ] Proper authorization checks (system_admin permission)
- [ ] Validation: cannot assign capability not supported by system
- [ ] Request/response validation
- [ ] API documentation updated
- [ ] Integration tests for all endpoints
- [ ] E2E tests

**Related**: #006, #015

---

### Issue #012: Tenant feature enablement endpoints
**Title**: `[Phase 3] Create API endpoints for tenant feature enablement`

**Description**:
Create REST API endpoints for tenant admins to enable/disable features.

**Labels**: `api`, `tenant`, `p0`, `capability-model`, `phase-3`

**Acceptance Criteria**:
- [ ] `GET /api/v1/tenant/features` - Get enabled features for tenant
- [ ] `PUT /api/v1/tenant/features/:key` - Enable feature for tenant
- [ ] `DELETE /api/v1/tenant/features/:key` - Disable feature for tenant
- [ ] Proper authorization checks (tenant_admin permission)
- [ ] Validation: cannot enable feature not allowed by system
- [ ] Validation: cannot exceed system limits
- [ ] Request/response validation
- [ ] API documentation updated
- [ ] Integration tests for all endpoints
- [ ] E2E tests

**Related**: #006, #016

---

### Issue #013: User capability state endpoints
**Title**: `[Phase 3] Create API endpoints for user capability enrollment`

**Description**:
Create REST API endpoints for managing user capability enrollment state.

**Labels**: `api`, `user`, `p0`, `capability-model`, `phase-3`

**Acceptance Criteria**:
- [ ] `GET /api/v1/users/:id/capabilities` - Get user capability states
- [ ] `GET /api/v1/users/:id/capabilities/:key` - Get specific capability state
- [ ] `POST /api/v1/users/:id/capabilities/:key/enroll` - Enroll user in capability
- [ ] `DELETE /api/v1/users/:id/capabilities/:key` - Unenroll user from capability
- [ ] Proper authorization checks (user can view own, admin can manage others)
- [ ] Validation: cannot enroll in capability not enabled by tenant
- [ ] Request/response validation
- [ ] API documentation updated
- [ ] Integration tests for all endpoints
- [ ] E2E tests

**Related**: #006, #017

---

## Phase 4: Frontend Admin Dashboard

### Issue #014: System capability management page
**Title**: `[Phase 4] Create system capability management UI page`

**Description**:
Create React page for system admins to view and manage global system capabilities.

**Labels**: `frontend`, `system`, `p0`, `capability-model`, `phase-4`, `ui`

**Acceptance Criteria**:
- [ ] Page component `frontend/admin-dashboard/src/pages/system/Capabilities.tsx` created
- [ ] List all system capabilities with status (enabled/disabled)
- [ ] Toggle capability enablement
- [ ] Configure default values for capabilities
- [ ] Show which tenants are using each capability
- [ ] Visual indicators (green/gray) for enabled/disabled
- [ ] Responsive design
- [ ] Loading states and error handling
- [ ] Integration with backend API (#010)
- [ ] Unit tests for components
- [ ] E2E tests

**Related**: #010, #015

---

### Issue #015: Tenant capability assignment page
**Title**: `[Phase 4] Create tenant capability assignment UI page`

**Description**:
Create React page for system admins to assign capabilities to tenants.

**Labels**: `frontend`, `system`, `p0`, `capability-model`, `phase-4`, `ui`

**Acceptance Criteria**:
- [ ] Page component `frontend/admin-dashboard/src/pages/system/TenantCapabilities.tsx` created
- [ ] Tenant selector dropdown
- [ ] Capability matrix showing allowed vs not allowed
- [ ] Toggle capabilities for selected tenant
- [ ] Configure capability-specific values (e.g., max_token_ttl)
- [ ] Visual inheritance diagram
- [ ] Bulk assignment for multiple tenants
- [ ] Responsive design
- [ ] Loading states and error handling
- [ ] Integration with backend API (#011)
- [ ] Unit tests for components
- [ ] E2E tests

**Related**: #011, #014, #019

---

### Issue #016: Tenant feature enablement page
**Title**: `[Phase 4] Create tenant feature enablement UI page`

**Description**:
Create React page for tenant admins to enable/disable features.

**Labels**: `frontend`, `tenant`, `p0`, `capability-model`, `phase-4`, `ui`

**Acceptance Criteria**:
- [ ] Page component `frontend/admin-dashboard/src/pages/tenant/Features.tsx` created
- [ ] Show available features (based on allowed capabilities)
- [ ] Enable/disable features toggle
- [ ] Configure feature settings (e.g., MFA enforcement rules)
- [ ] Visual indicators showing:
  - System support (green/gray)
  - Tenant allowed (green/gray)
  - Tenant enabled (green/gray)
- [ ] Capability inheritance visualization
- [ ] Responsive design
- [ ] Loading states and error handling
- [ ] Integration with backend API (#012)
- [ ] Unit tests for components
- [ ] E2E tests

**Related**: #012, #019

---

### Issue #017: User capability enrollment page
**Title**: `[Phase 4] Create user capability enrollment UI page`

**Description**:
Create React page for viewing and managing user capability enrollment.

**Labels**: `frontend`, `user`, `p0`, `capability-model`, `phase-4`, `ui`

**Acceptance Criteria**:
- [ ] Page component `frontend/admin-dashboard/src/pages/users/UserCapabilities.tsx` created
- [ ] Show user's capability enrollment status
- [ ] Enroll/unenroll users in capabilities
- [ ] View enrollment details (e.g., TOTP secret, recovery codes)
- [ ] Show required vs optional capabilities
- [ ] Force enrollment for required capabilities
- [ ] Responsive design
- [ ] Loading states and error handling
- [ ] Integration with backend API (#013)
- [ ] Unit tests for components
- [ ] E2E tests

**Related**: #013

---

### Issue #018: Enhanced settings page with capability model
**Title**: `[Phase 4] Enhance settings page to include capability management`

**Description**:
Update the existing settings page to include new tabs for capability management.

**Labels**: `frontend`, `settings`, `p0`, `capability-model`, `phase-4`, `ui`

**Acceptance Criteria**:
- [ ] Update `frontend/admin-dashboard/src/pages/Settings.tsx`
- [ ] Add "System Capabilities" tab (SYSTEM users only)
- [ ] Add "Tenant Capabilities" tab (SYSTEM users only)
- [ ] Add "Tenant Features" tab (TENANT users)
- [ ] Add "User Capabilities" tab (TENANT users)
- [ ] Integrate with existing settings functionality
- [ ] Maintain backward compatibility
- [ ] Responsive design
- [ ] Unit tests
- [ ] E2E tests

**Related**: #014, #015, #016, #017

---

### Issue #019: Capability inheritance visualization component
**Title**: `[Phase 4] Create interactive capability inheritance diagram component`

**Description**:
Create a reusable React component that visualizes the capability inheritance flow (System → Tenant → User).

**Labels**: `frontend`, `ui`, `p1`, `capability-model`, `phase-4`

**Acceptance Criteria**:
- [ ] Component `frontend/admin-dashboard/src/components/CapabilityInheritanceDiagram.tsx` created
- [ ] Visual diagram showing System → Tenant → User flow
- [ ] Color-coded states (enabled, allowed, enrolled)
- [ ] Interactive tooltips with details
- [ ] Real-time updates when capabilities change
- [ ] Export as image functionality
- [ ] Responsive design
- [ ] Reusable across multiple pages
- [ ] Unit tests
- [ ] Storybook stories (if applicable)

**Related**: #015, #016

---

### Issue #020: Enhanced dashboard with capability metrics
**Title**: `[Phase 4] Enhance dashboard with capability-related metrics`

**Description**:
Update the dashboard to show capability-related statistics and metrics.

**Labels**: `frontend`, `dashboard`, `p1`, `capability-model`, `phase-4`, `ui`

**Acceptance Criteria**:
- [ ] Update `frontend/admin-dashboard/src/pages/Dashboard.tsx`
- [ ] System users see: Total capabilities enabled, tenants using each capability
- [ ] Tenant users see: Enabled features, user enrollment rates
- [ ] User view: Enrollment status, required vs optional
- [ ] Visual charts/graphs for metrics
- [ ] Real-time updates
- [ ] Responsive design
- [ ] Unit tests

**Related**: #014, #015, #016, #017

---

## Phase 5: Enforcement & Validation

### Issue #021: Capability enforcement middleware
**Title**: `[Phase 5] Create middleware for capability enforcement`

**Description**:
Create middleware that enforces capability checks before allowing feature usage.

**Labels**: `backend`, `middleware`, `p0`, `capability-model`, `phase-5`, `security`

**Acceptance Criteria**:
- [ ] Middleware file `api/middleware/capability.go` created
- [ ] Middleware checks capability before allowing feature usage
- [ ] Validates tenant feature enablement
- [ ] Enforces user enrollment requirements
- [ ] Returns clear, actionable error messages
- [ ] Proper HTTP status codes
- [ ] Integration with capability service (#006)
- [ ] Unit tests
- [ ] Integration tests

**Related**: #006, #008, #009, #022

---

### Issue #022: Capability validation logic
**Title**: `[Phase 5] Implement comprehensive capability validation`

**Description**:
Implement validation logic to ensure capability rules are enforced correctly.

**Labels**: `backend`, `validation`, `p0`, `capability-model`, `phase-5`, `security`

**Acceptance Criteria**:
- [ ] Validation: Tenant cannot enable feature not allowed by system
- [ ] Validation: Tenant cannot exceed system limits (e.g., max_token_ttl)
- [ ] Validation: User cannot skip required enrollments
- [ ] Validation: System cannot bypass tenant restrictions
- [ ] Validation functions in `identity/capability/validation.go`
- [ ] Comprehensive unit tests
- [ ] Integration tests
- [ ] Error messages are clear and actionable

**Related**: #006, #021

---

### Issue #023: Include capability context in tokens
**Title**: `[Phase 5] Add capability context to JWT tokens`

**Description**:
Update token claims builder to include capability context (informational only).

**Labels**: `backend`, `token`, `p0`, `capability-model`, `phase-5`

**Acceptance Criteria**:
- [ ] Update `auth/claims/builder.go`
- [ ] Include `capabilities` object in token claims
- [ ] Include `features` object in token claims
- [ ] Mark as informational only (not authoritative)
- [ ] Update token validation if needed
- [ ] Unit tests
- [ ] Integration tests
- [ ] Documentation updated

**Related**: #006, #008

---

## Phase 6: Testing & Documentation

### Issue #024: Unit tests for capability service
**Title**: `[Phase 6] Write comprehensive unit tests for capability service`

**Description**:
Create unit tests for the capability service with high coverage.

**Labels**: `testing`, `unit`, `p1`, `capability-model`, `phase-6`

**Acceptance Criteria**:
- [ ] Test file `identity/capability/service_test.go` created
- [ ] Test coverage ≥ 90%
- [ ] Tests for capability evaluation logic
- [ ] Tests for inheritance chain validation
- [ ] Tests for edge cases (missing capabilities, disabled features)
- [ ] Tests for error conditions
- [ ] Mock repositories used
- [ ] All tests pass

**Related**: #006

---

### Issue #025: Integration tests for capability APIs
**Title**: `[Phase 6] Write integration tests for capability API endpoints`

**Description**:
Create integration tests for all capability-related API endpoints.

**Labels**: `testing`, `integration`, `p1`, `capability-model`, `phase-6`

**Acceptance Criteria**:
- [ ] Test file `api/handlers/capability_handler_test.go` created
- [ ] Tests for system capability management APIs
- [ ] Tests for tenant capability assignment APIs
- [ ] Tests for tenant feature enablement APIs
- [ ] Tests for user capability state APIs
- [ ] Tests for authorization checks
- [ ] Tests for validation logic
- [ ] All tests pass
- [ ] Test coverage documented

**Related**: #010, #011, #012, #013

---

### Issue #026: E2E tests for capability flow
**Title**: `[Phase 6] Write end-to-end tests for complete capability flow`

**Description**:
Create E2E tests that verify the complete capability assignment → enablement → enrollment flow.

**Labels**: `testing`, `e2e`, `p1`, `capability-model`, `phase-6`

**Acceptance Criteria**:
- [ ] E2E test file `api/e2e/capability_flow_test.go` created
- [ ] Test: System admin assigns capability to tenant
- [ ] Test: Tenant admin enables feature
- [ ] Test: User enrolls in capability
- [ ] Test: Enforcement during authentication
- [ ] Test: OAuth scope validation
- [ ] Test: UI interactions for capability management
- [ ] Test: Error handling and validation
- [ ] All tests pass

**Related**: #014, #015, #016, #017

---

### Issue #027: Update documentation
**Title**: `[Phase 6] Create and update documentation for capability model`

**Description**:
Create comprehensive documentation for the capability model architecture, usage, and API.

**Labels**: `documentation`, `p1`, `capability-model`, `phase-6`

**Acceptance Criteria**:
- [ ] Document `docs/architecture/CAPABILITY_MODEL.md` created
- [ ] Document `docs/guides/capability-management.md` created
- [ ] Document `docs/api/capability-endpoints.md` created
- [ ] Update `docs/DOCUMENTATION_INDEX.md`
- [ ] Architecture diagrams included
- [ ] Code examples included
- [ ] API examples included
- [ ] Review and approval

**Related**: All previous issues

---

## Phase 7: Migration & Deployment

### Issue #028: Migrate existing data to capability model
**Title**: `[Phase 7] Create migration script for existing data`

**Description**:
Create migration script to convert existing tenant settings to the new capability model.

**Labels**: `migration`, `database`, `p0`, `capability-model`, `phase-7`

**Acceptance Criteria**:
- [ ] Migration script `migrations/000022_migrate_existing_capabilities.up.sql` created
- [ ] Rollback script `migrations/000022_migrate_existing_capabilities.down.sql` created
- [ ] Migrate existing tenant settings to capability model
- [ ] Set default capabilities for existing tenants
- [ ] Preserve existing feature enablements
- [ ] Data validation after migration
- [ ] Tested on staging environment
- [ ] Documentation for migration process

**Related**: #001, #002, #003

---

### Issue #029: Deployment and rollout plan
**Title**: `[Phase 7] Create deployment and rollout plan`

**Description**:
Create detailed plan for deploying capability model changes to production.

**Labels**: `deployment`, `p1`, `capability-model`, `phase-7`

**Acceptance Criteria**:
- [ ] Deployment plan document created
- [ ] Steps defined: database migrations, backend, frontend
- [ ] Rollout strategy (gradual vs all-at-once)
- [ ] Monitoring and validation steps
- [ ] Rollback procedures defined
- [ ] Communication plan
- [ ] Review and approval

**Related**: #028, #030

---

### Issue #030: Rollback procedures
**Title**: `[Phase 7] Create rollback procedures`

**Description**:
Define procedures for rolling back capability model changes if issues arise.

**Labels**: `deployment`, `p1`, `capability-model`, `phase-7`

**Acceptance Criteria**:
- [ ] Rollback plan document created
- [ ] Database rollback migrations tested
- [ ] Feature flags to disable capability checks
- [ ] Revert procedures for backend
- [ ] Revert procedures for frontend
- [ ] Data preservation strategy
- [ ] Tested on staging environment
- [ ] Review and approval

**Related**: #028, #029

---

## Summary

**Total Issues**: 30  
**Phase 1**: 5 issues  
**Phase 2**: 4 issues  
**Phase 3**: 4 issues  
**Phase 4**: 7 issues  
**Phase 5**: 3 issues  
**Phase 6**: 4 issues  
**Phase 7**: 3 issues

**Priority Distribution**:
- P0 (Critical): 22 issues
- P1 (Important): 8 issues

