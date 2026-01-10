# ARauth Authority & Permission Model

**Version**: 1.0  
**Status**: Security-Critical Contract  
**Purpose**: Define authority boundaries and permission enforcement

---

## Table of Contents

1. [Core Principles](#core-principles)
2. [Authority Split](#authority-split)
3. [Console Mode Computation](#console-mode-computation)
4. [Permission Model](#permission-model)
5. [UI Permission Mapping](#ui-permission-mapping)
6. [Impersonation Rules](#impersonation-rules)
7. [Audit Requirements](#audit-requirements)
8. [Security Invariants](#security-invariants)

---

## Core Principles

### 1. Backend Is Law

**Rule**: The UI **never** grants authority. It only reflects authority granted by the backend.

**Implications**:
- All permissions come from `authStore` (populated by backend)
- UI cannot add, remove, or modify permissions
- Permission checks are read-only operations
- Disabled/hidden UI elements do not constitute security

### 2. Authority First

**Rule**: Every feature must answer: WHO can do this, at WHAT scope, and WHY?

**Implications**:
- Every page wrapped in `PermissionGate`
- Every route protected by `ProtectedRoute`
- Every action requires explicit permission check
- No "admin" catch-all permissions

### 3. Strict Plane Separation

**Rule**: SYSTEM and TENANT authority planes are isolated.

**Implications**:
- SYSTEM users cannot accidentally operate in TENANT mode
- TENANT users cannot access SYSTEM features
- Settings are split by authority plane
- No shared forms or ambiguous contexts

### 4. No Fake UX

**Rule**: The UI never pretends to have authority it lacks.

**Implications**:
- Missing permissions → hidden UI elements
- Missing backend APIs → "Coming Soon" messages
- Failed permission checks → clear error messages
- No optimistic updates without backend confirmation

---

## Authority Split

### SYSTEM Authority

**Principal Type**: `SYSTEM`  
**Scope**: Platform-wide  
**Granted To**: Platform administrators

**Capabilities**:
- Manage all tenants
- Configure system-level settings (OAuth2, JWT, capabilities)
- View cross-tenant metrics and audit logs
- Impersonate users (when tenant selected)
- Create system users or tenant users

**Permission Prefix**: `system:*`

**Examples**:
- `system:configure` - Modify system settings
- `system:tenants:read` - View all tenants
- `system:tenants:create` - Create new tenants

### TENANT Authority

**Principal Type**: `TENANT`  
**Scope**: Single tenant  
**Granted To**: Tenant administrators

**Capabilities**:
- Manage users within tenant
- Configure tenant settings (tokens, password policy, MFA, rate limits)
- View tenant-specific metrics and audit logs
- Create tenant users only
- Cannot access other tenants

**Permission Prefix**: `users:*`, `roles:*`, `settings:*`, etc.

**Examples**:
- `users:read` - View users in tenant
- `users:create` - Create users in tenant
- `settings:update` - Modify tenant settings

---

## Console Mode Computation

Console mode is **computed**, not stored.

### Computation Logic

```typescript
const consoleMode: ConsoleMode = useMemo(() => {
  if (principalType === 'SYSTEM') {
    return selectedTenantId ? 'TENANT' : 'SYSTEM';
  }
  return 'TENANT';
}, [principalType, selectedTenantId]);
```

### Rules

1. **SYSTEM users without tenant selected** → `SYSTEM` mode
2. **SYSTEM users with tenant selected** → `TENANT` mode
3. **TENANT users** → Always `TENANT` mode (no choice)

### Implications

- SYSTEM users can switch modes via tenant selector
- TENANT users cannot access SYSTEM mode
- Mode determines navigation, settings, and data scope
- Mode is always derived, never persisted

---

## Permission Model

### Permission Naming Convention

Format: `<resource>:<action>[:<sub-resource>]`

**Examples**:
- `users:read` - View users
- `users:create` - Create users
- `users:update` - Update users
- `users:delete` - Delete users
- `users:mfa:reset` - Reset user MFA
- `users:impersonate` - Impersonate users

### Permission Scope

Permissions are **scoped** by principal type:

- **SYSTEM permissions**: `systemPermissions[]` array
- **TENANT permissions**: `permissions[]` array

### Permission Checks

**SYSTEM Permission Check**:
```typescript
principalType === 'SYSTEM' && systemPermissions.includes(permission)
```

**TENANT Permission Check**:
```typescript
permissions.includes(permission)
```

**Dual Check** (for features available in both modes):
```typescript
(principalType === 'SYSTEM' && systemPermissions.includes(permission)) ||
permissions.includes(permission)
```

---

## UI Permission Mapping

### Complete Permission List

#### User Management
- `users:read` - View users
- `users:create` - Create users
- `users:update` - Update/suspend/activate users
- `users:delete` - Delete users
- `users:mfa:reset` - Reset user MFA
- `users:impersonate` - Impersonate users (SYSTEM only)

#### Role Management
- `roles:read` - View roles
- `roles:create` - Create roles
- `roles:update` - Update roles
- `roles:delete` - Delete roles
- `roles:assign` - Assign roles to users

#### Permission Management
- `permissions:read` - View permissions
- `permissions:create` - Create permissions
- `permissions:update` - Update permissions
- `permissions:delete` - Delete permissions

#### OAuth2 Clients
- `oauth:clients:read` - View OAuth2 clients
- `oauth:clients:create` - Create OAuth2 clients
- `oauth:clients:update` - Update OAuth2 clients
- `oauth:clients:delete` - Delete OAuth2 clients

#### SCIM
- `scim:read` - View SCIM configuration
- `scim:tokens:create` - Create SCIM tokens
- `scim:tokens:revoke` - Revoke SCIM tokens

#### Federation
- `federation:idp:read` - View external IdPs
- `federation:idp:create` - Create external IdPs
- `federation:idp:update` - Update external IdPs
- `federation:idp:delete` - Delete external IdPs
- `federation:link` - Link external identities
- `federation:unlink` - Unlink external identities

#### Webhooks
- `webhooks:read` - View webhooks
- `webhooks:create` - Create webhooks
- `webhooks:update` - Update webhooks
- `webhooks:delete` - Delete webhooks
- `webhooks:logs:read` - View webhook delivery logs

#### Audit & Observability
- `audit:read` - View audit logs
- `audit:export` - Export audit logs
- `sessions:read` - View active sessions
- `sessions:revoke` - Revoke sessions

#### Settings
- `system:configure` - Modify system settings (SYSTEM only)
- `settings:read` - View tenant settings
- `settings:update` - Update tenant settings

#### Dashboard
- `dashboard:read` - View dashboard metrics

---

## Impersonation Rules

### Who Can Impersonate

**Required**:
1. `principalType === 'SYSTEM'`
2. `systemPermissions.includes('users:impersonate')`
3. `selectedTenantId` is set (tenant context required)

**Forbidden**:
- TENANT users cannot impersonate
- SYSTEM users cannot impersonate without tenant selected
- Cannot impersonate users in other tenants

### Impersonation State

**Stored In**: `authStore`

**Properties**:
- `impersonatedUser: User | null`
- `impersonatedTenant: Tenant | null`

**Actions**:
- `startImpersonation(userId: string): Promise<void>`
- `endImpersonation(): Promise<void>`
- `clearImpersonation(): void` (client-side cleanup)

### Impersonation Banner

**Behavior**:
- Always visible when `impersonatedUser` is set
- Cannot be dismissed
- Appears before all other UI
- Displays impersonated user email and tenant name
- "End Impersonation" button triggers `endImpersonation()` + page reload

### Audit Requirements

**Events Logged**:
- `impersonation.started` - Who, target user, tenant, timestamp
- `impersonation.ended` - Who, target user, duration
- All actions taken during impersonation (with impersonation context)

---

## Audit Requirements

### When Audit Reasons Are Required

**Destructive Actions**:
- Suspend user
- Delete user
- Reset user MFA
- Revoke SCIM token
- Delete external IdP
- Unlink external identity
- Revoke session
- Delete webhook

**Audit Reason Format**:
- Free-text field
- Minimum 1 character (non-empty)
- Recorded in audit log with action
- Visible to auditors and compliance teams

### Audit Reason UI Pattern

1. Action triggered (e.g., "Revoke Session")
2. Confirmation dialog opens
3. Audit reason textarea (required)
4. Warning message about action impact
5. User submits with reason
6. Backend receives action + reason
7. Audit log records both

**No Bypass**: UI enforces audit reason before API call. Backend must also enforce.

---

## Security Invariants

### 1. UI Never Grants Authority

**Invariant**: The UI cannot grant permissions, create tokens, or authorize actions.

**Enforcement**:
- All permissions from backend
- All tokens from backend
- All authorization decisions from backend
- UI only reflects backend state

### 2. Permission Checks Are Read-Only

**Invariant**: Permission checks in UI do not modify state.

**Enforcement**:
- `hasPermission()` is pure function
- `PermissionGate` only shows/hides
- No permission mutations in frontend

### 3. Tenant Isolation

**Invariant**: TENANT users cannot access other tenants' data.

**Enforcement**:
- `homeTenantId` is immutable
- All API calls scoped to `homeTenantId`
- No tenant selector for TENANT users
- Backend enforces tenant boundaries

### 4. Impersonation Visibility

**Invariant**: Impersonation is always visible and audited.

**Enforcement**:
- Banner cannot be dismissed
- Banner appears before all UI
- All actions logged with impersonation context
- End impersonation triggers page reload

### 5. No Fake Data

**Invariant**: UI never displays invented or mocked data.

**Enforcement**:
- Missing APIs throw `APINotConnectedError`
- "Coming Soon" messages for missing features
- Empty states for empty lists
- No placeholder data

### 6. One-Time Secrets

**Invariant**: Secrets (client secrets, signing secrets, SCIM tokens) shown once.

**Enforcement**:
- Secret displayed immediately after creation
- Password-masked until copied
- Confirmation before closing without copying
- Backend never re-displays secrets

### 7. Explicit Scope Indicators

**Invariant**: Users always know their current scope (SYSTEM vs TENANT).

**Enforcement**:
- Mode badges on settings pages
- Scope indicators on lists ("Applies to all tenants" vs "Applies only to this tenant")
- Authority badges on user creation ("System User" vs "Tenant User")
- Clear navigation separation

---

## Why UI Never Grants Authority

### The Problem

If the UI could grant authority, it would create security vulnerabilities:
- Client-side permission checks can be bypassed
- Browser state can be manipulated
- No audit trail for permission grants
- No centralized enforcement

### The Solution

**Backend Is Law**:
1. Backend grants all permissions
2. Backend returns permissions in auth response
3. UI stores permissions in `authStore`
4. UI reads permissions (never writes)
5. UI shows/hides based on permissions
6. Backend enforces permissions on every API call

**Result**: UI is a **view** of authority, not a **source** of authority.

---

## Summary

**Authority Model**:
- SYSTEM vs TENANT split
- Console mode computed from principal type + tenant selection
- Permissions are backend-derived, never UI-granted

**Permission Enforcement**:
- Every feature permission-gated
- Every route protected
- Every action requires explicit permission
- No catch-all permissions

**Impersonation**:
- SYSTEM only, tenant selected
- Always visible (banner)
- Always audited

**Audit**:
- Destructive actions require audit reasons
- All actions logged
- Impersonation context preserved

**Security**:
- UI never grants authority
- No fake data
- One-time secrets
- Explicit scope indicators

**This model is non-negotiable and must be preserved during backend integration.**
