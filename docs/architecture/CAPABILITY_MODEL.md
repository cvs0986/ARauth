# Capability Model Architecture

This document describes the ARauth Capability Model architecture, a three-layer model that controls feature availability and enforcement across the system.

## Overview

The Capability Model implements a strict downward inheritance pattern where capabilities flow from System → Tenant → User, with no upward overrides. This ensures that system-level security policies cannot be weakened by tenants or users.

## Three-Layer Model

### Layer 1: System Level (Global)

**Purpose**: Defines what capabilities are supported by ARauth at all.

**Table**: `system_capabilities`

**Responsibilities**:
- Define available capabilities (e.g., MFA, OAuth2, SAML)
- Set global defaults and limits
- Control platform-wide feature availability

**Example**:
```sql
INSERT INTO system_capabilities (capability_key, enabled, default_value) 
VALUES ('mfa', true, '{"max_attempts": 3}');
```

### Layer 2: System → Tenant Assignment

**Purpose**: Defines what capabilities a specific tenant is allowed to use.

**Table**: `tenant_capabilities`

**Responsibilities**:
- Assign capabilities to tenants (SYSTEM admin only)
- Configure per-tenant limits and values
- Control tenant feature availability

**Example**:
```sql
INSERT INTO tenant_capabilities (tenant_id, capability_key, enabled, value)
VALUES ('tenant-uuid', 'mfa', true, '{"max_attempts": 5}');
```

### Layer 3: Tenant Feature Enablement

**Purpose**: Allows tenants to enable features within their allowed capabilities.

**Table**: `tenant_feature_enablement`

**Responsibilities**:
- Enable/disable features for tenant (TENANT admin)
- Configure feature-specific settings
- Control feature activation

**Example**:
```sql
INSERT INTO tenant_feature_enablement (tenant_id, feature_key, enabled, configuration)
VALUES ('tenant-uuid', 'mfa', true, '{"required_for_admins": true}');
```

### Layer 4: User Enrollment

**Purpose**: Tracks user enrollment and compliance with capabilities.

**Table**: `user_capability_state`

**Responsibilities**:
- Track user enrollment status
- Store user-specific state data
- Enforce user-level requirements

**Example**:
```sql
INSERT INTO user_capability_state (user_id, capability_key, enrolled, state_data)
VALUES ('user-uuid', 'mfa', true, '{"totp_secret": "..."}');
```

## Capability Evaluation Flow

The system evaluates capabilities using a strict downward flow:

```
1. System Level: Is capability supported?
   ↓ (if yes)
2. Tenant Assignment: Is capability allowed for tenant?
   ↓ (if yes)
3. Feature Enablement: Is feature enabled by tenant?
   ↓ (if yes)
4. User Enrollment: Is user enrolled? (if required)
   ↓ (if yes)
5. Result: Capability can be used
```

## Key Principles

### 1. Strict Downward Inheritance

Capabilities flow **only** downward:
- System → Tenant → User
- No upward overrides
- System defines limits, tenants enforce, users comply

### 2. Capability vs State

- **Capability**: What is possible/allowed (System/Tenant level)
- **State**: What has happened (User level)
- Users have state, not power

### 3. Enforcement Rules

- **System**: Defines what exists and global guardrails
- **Tenant**: Enables features and enforces policies
- **User**: Enrolls and complies, never enables

## Implementation

### Service Layer

The `CapabilityService` provides methods for:
- System capability management
- Tenant capability assignment
- Tenant feature enablement
- User enrollment
- Capability evaluation (combines all layers)

### Middleware

The `RequireCapability` middleware enforces capability checks:
- Validates all three layers
- Returns clear error messages
- Stores evaluation result in context

### Validation

The `ValidationService` ensures:
- Tenants cannot enable unassigned capabilities
- Tenant values cannot exceed system limits
- Users cannot skip required enrollments

### Token Context

JWT tokens include capability context (informational):
- `capabilities`: Map of available capabilities
- `features`: Map of enabled features with metadata

## API Endpoints

### System Admin Endpoints

- `GET /system/capabilities` - List all system capabilities
- `GET /system/capabilities/:key` - Get system capability
- `PUT /system/capabilities/:key` - Update system capability
- `GET /system/tenants/:tenantId/capabilities` - List tenant capabilities
- `POST /system/tenants/:tenantId/capabilities` - Assign capability to tenant
- `DELETE /system/tenants/:tenantId/capabilities/:key` - Revoke capability

### Tenant Admin Endpoints

- `GET /api/v1/tenant/features` - List enabled features
- `POST /api/v1/tenant/features/:key` - Enable feature
- `DELETE /api/v1/tenant/features/:key` - Disable feature
- `GET /api/v1/users/:userId/capabilities` - List user capabilities
- `POST /api/v1/users/:userId/capabilities/:key/enroll` - Enroll user
- `DELETE /api/v1/users/:userId/capabilities/:key` - Unenroll user

## Frontend Integration

The Admin Dashboard provides:
- System capability management (SYSTEM users)
- Tenant capability assignment (SYSTEM users)
- Tenant feature enablement (TENANT users)
- User capability enrollment (TENANT users)
- Capability inheritance visualization
- Dashboard metrics

## Testing

### Unit Tests

- Service layer tests with mocked repositories
- Validation logic tests
- Middleware tests

### Integration Tests

- API endpoint tests
- Database operation tests
- Service integration tests

### E2E Tests

- Complete capability flow (System → Tenant → User)
- Enforcement validation
- Error handling

## Related Documents

- [Feature Capability Document](../../feature_capibility.md) - Source of truth
- [Implementation Plan](../planning/CAPABILITY_MODEL_IMPLEMENTATION_PLAN.md)
- [Status Tracking](../status/CAPABILITY_MODEL_STATUS.md)

