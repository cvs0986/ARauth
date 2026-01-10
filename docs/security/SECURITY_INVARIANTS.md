# ARauth Security Invariants & Guardrails

**Version**: 1.0  
**Status**: Security-Critical Contract  
**Purpose**: Non-negotiable security rules that must never be violated

---

## Table of Contents

1. [Core Security Principles](#core-security-principles)
2. [Backend Is Law](#backend-is-law)
3. [One-Time Secret Handling](#one-time-secret-handling)
4. [MFA Invariants](#mfa-invariants)
5. [Token Invariants](#token-invariants)
6. [Impersonation Rules](#impersonation-rules)
7. [Audit Enforcement](#audit-enforcement)
8. [Tenant Isolation](#tenant-isolation)
9. [Permission Enforcement](#permission-enforcement)
10. [No Fake Data](#no-fake-data)
11. [Compliance Requirements](#compliance-requirements)

---

## Core Security Principles

### 1. Defense in Depth

**Rule**: Security is enforced at multiple layers.

**Layers**:
1. **UI**: Permission gates, input validation, user warnings
2. **API**: Permission checks, request validation, rate limiting
3. **Business Logic**: Authority enforcement, audit logging
4. **Database**: Row-level security, tenant isolation

**Violation**: Relying on UI-only security

### 2. Principle of Least Privilege

**Rule**: Users and systems have only the minimum permissions required.

**Enforcement**:
- No "admin" catch-all permissions
- Granular permission model (50+ permissions)
- Permission checks on every API call
- No permission inheritance without explicit grants

**Violation**: Granting broad permissions by default

### 3. Explicit Over Implicit

**Rule**: Security decisions must be explicit, never assumed.

**Examples**:
- Explicit attribute mapping (no auto-mapping)
- Explicit scope indicators (SYSTEM vs TENANT)
- Explicit audit reasons (no silent actions)
- Explicit confirmation dialogs (no silent destructive actions)

**Violation**: Auto-linking external identities, auto-granting permissions

---

## Backend Is Law

### Invariant

**The UI never grants authority. All authority comes from the backend.**

### Enforcement Rules

1. **Permissions are read-only in UI**
   - UI reads permissions from `authStore`
   - UI never modifies permissions
   - Permission checks are pure functions

2. **All authorization decisions made by backend**
   - Every API call checks permissions server-side
   - UI permission gates are UX optimization, not security
   - Backend rejects unauthorized requests with `403 PERMISSION_DENIED`

3. **No client-side security**
   - Browser state can be manipulated
   - Client-side checks can be bypassed
   - Security must be server-enforced

### Violations to Prevent

❌ UI granting permissions  
❌ UI creating tokens  
❌ UI authorizing actions  
❌ Optimistic updates without backend confirmation  
❌ Disabled buttons as security controls  

### Correct Pattern

✅ Backend grants permissions  
✅ Backend returns permissions in auth response  
✅ UI stores permissions (read-only)  
✅ UI shows/hides based on permissions  
✅ Backend enforces permissions on every API call  

---

## One-Time Secret Handling

### Invariant

**Secrets (client secrets, signing secrets, SCIM tokens) are shown exactly once.**

### Enforcement Rules

1. **Display on creation only**
   - Secret returned in create response
   - Secret never returned in list/get responses
   - Secret never stored in plaintext on backend

2. **UI pattern**
   - Password-masked until copied
   - Copy button with confirmation
   - Warning: "This secret will not be shown again"
   - Confirmation before closing without copying

3. **Backend enforcement**
   - Hash secrets before storage
   - Never re-display secrets
   - Return `SECRET_ALREADY_RETRIEVED` error if requested again

### Secrets Covered

- OAuth2 client secrets
- SCIM tokens
- Webhook signing secrets
- API keys
- Impersonation tokens

### Violations to Prevent

❌ Re-displaying secrets  
❌ Storing secrets in plaintext  
❌ Sending secrets in logs  
❌ Allowing secret retrieval after creation  

---

## MFA Invariants

### Invariant

**MFA reset is a destructive, audited action.**

### Enforcement Rules

1. **Reset requires audit reason**
   - UI enforces audit reason field (required)
   - Backend validates audit reason (non-empty)
   - Audit log records reason

2. **Reset is immediate**
   - User's MFA is disabled immediately
   - User must re-enroll MFA on next login
   - No grace period

3. **Reset is logged**
   - `user.mfa.reset` event in audit log
   - Actor, target user, reason, timestamp

### Violations to Prevent

❌ Silent MFA reset  
❌ MFA reset without audit reason  
❌ Allowing user to reset own MFA without re-authentication  

---

## Token Invariants

### Invariant

**All tokens have limited lifetime and can be revoked.**

### Enforcement Rules

1. **Token types**
   - Access tokens (short-lived, 15 minutes)
   - Refresh tokens (long-lived, 30 days)
   - SCIM tokens (long-lived, no expiry)
   - Impersonation tokens (session-bound)

2. **Revocation**
   - All tokens can be revoked
   - Revocation is immediate
   - Revoked tokens never work again

3. **Rotation**
   - OAuth2 client secrets can be rotated
   - Rotation invalidates old secret
   - New secret shown once

### Violations to Prevent

❌ Tokens without expiry (except SCIM)  
❌ Tokens that cannot be revoked  
❌ Reusing revoked tokens  

---

## Impersonation Rules

### Invariant

**Impersonation is always visible, audited, and SYSTEM-only.**

### Enforcement Rules

1. **Who can impersonate**
   - `principalType === 'SYSTEM'`
   - `systemPermissions.includes('users:impersonate')`
   - `selectedTenantId` is set

2. **Visibility**
   - Banner always visible when impersonating
   - Banner cannot be dismissed
   - Banner appears before all other UI
   - Banner shows impersonated user and tenant

3. **Audit**
   - `impersonation.started` event (who, target, tenant, timestamp)
   - `impersonation.ended` event (who, target, duration)
   - All actions during impersonation logged with context

4. **Ending impersonation**
   - One-click "End Impersonation" button
   - Triggers page reload
   - Clears impersonation state
   - Logs audit event

### Violations to Prevent

❌ Silent impersonation  
❌ Impersonation without tenant context  
❌ TENANT users impersonating  
❌ Dismissible impersonation banner  
❌ Impersonation without audit trail  

---

## Audit Enforcement

### Invariant

**All destructive actions require audit reasons and are logged.**

### Enforcement Rules

1. **Actions requiring audit reasons**
   - Suspend user
   - Delete user
   - Reset user MFA
   - Revoke SCIM token
   - Delete external IdP
   - Unlink external identity
   - Revoke session
   - Delete webhook

2. **Audit reason validation**
   - UI enforces non-empty audit reason
   - Backend validates audit reason (minimum 1 character)
   - Audit reason stored with event

3. **Audit log immutability**
   - Audit logs are append-only
   - Audit logs cannot be deleted
   - Audit logs cannot be modified

4. **Audit log retention**
   - Minimum 90 days
   - Configurable per tenant
   - Export capability for compliance

### Violations to Prevent

❌ Destructive actions without audit reasons  
❌ Empty audit reasons  
❌ Modifying audit logs  
❌ Deleting audit logs  

---

## Tenant Isolation

### Invariant

**TENANT users cannot access other tenants' data.**

### Enforcement Rules

1. **Tenant boundary enforcement**
   - All API calls scoped to `homeTenantId` (TENANT users)
   - Backend validates tenant ID on every request
   - Cross-tenant queries rejected with `TENANT_BOUNDARY_VIOLATION`

2. **Data isolation**
   - Users belong to exactly one tenant
   - Roles scoped to tenant
   - Permissions scoped to tenant
   - OAuth2 clients scoped to tenant
   - SCIM tokens scoped to tenant
   - Webhooks scoped to tenant

3. **SYSTEM user exceptions**
   - SYSTEM users can access all tenants
   - SYSTEM users must select tenant for tenant-scoped operations
   - SYSTEM users cannot create cross-tenant resources

### Violations to Prevent

❌ TENANT users accessing other tenants  
❌ Cross-tenant role assignments  
❌ Cross-tenant user queries  
❌ Shared resources across tenants (except SYSTEM webhooks)  

---

## Permission Enforcement

### Invariant

**Every action requires explicit permission check.**

### Enforcement Rules

1. **UI enforcement**
   - Every route wrapped in `ProtectedRoute`
   - Every UI element wrapped in `PermissionGate`
   - Permission checks before API calls

2. **Backend enforcement**
   - Every API endpoint checks permissions
   - Permission check before business logic
   - Return `403 PERMISSION_DENIED` if unauthorized

3. **Permission granularity**
   - No catch-all permissions
   - Separate read/create/update/delete permissions
   - Separate permissions for sensitive actions (impersonate, mfa:reset)

### Violations to Prevent

❌ Unchecked API endpoints  
❌ Catch-all "admin" permissions  
❌ UI-only permission checks  
❌ Permission checks after business logic  

---

## No Fake Data

### Invariant

**The UI never displays invented or mocked data.**

### Enforcement Rules

1. **Missing APIs**
   - Throw `APINotConnectedError`
   - Display "Backend integration pending" message
   - Show empty states for empty lists

2. **Missing data**
   - Display "Coming Soon" for unimplemented features
   - Show `null` or empty values honestly
   - No placeholder data

3. **Error states**
   - Display real error messages
   - No silent failures
   - No optimistic success states

### Violations to Prevent

❌ Mocked API responses  
❌ Fake success states  
❌ Placeholder data  
❌ Silent failures  

---

## Compliance Requirements

### GDPR

**Right to be Forgotten**:
- User deletion must cascade to all related data
- Audit logs retain user ID (not PII)
- Export user data on request

**Data Minimization**:
- Collect only necessary user data
- No unnecessary logging of PII
- Audit logs anonymize where possible

### SOC 2

**Access Control**:
- Principle of least privilege
- Permission-based access
- Audit all access

**Audit Logging**:
- All destructive actions logged
- Immutable audit logs
- 90-day minimum retention

**Change Management**:
- Audit reasons for changes
- Approval workflows (future)
- Rollback capability (future)

### HIPAA (if applicable)

**Access Logging**:
- All data access logged
- Impersonation fully audited
- Session tracking

**Data Encryption**:
- Secrets hashed at rest
- TLS in transit
- No plaintext secrets

---

## Summary

**Non-Negotiable Invariants**:
1. Backend Is Law - UI never grants authority
2. One-Time Secrets - Shown exactly once
3. MFA Reset - Audited and immediate
4. Token Revocation - Always possible
5. Impersonation Visibility - Always visible and audited
6. Audit Enforcement - Destructive actions require reasons
7. Tenant Isolation - No cross-tenant access for TENANT users
8. Permission Enforcement - Every action checked
9. No Fake Data - Honest UI always
10. Compliance - GDPR, SOC 2, HIPAA ready

**These invariants must be preserved during backend integration and all future development.**

**Violation of these invariants is a security incident.**
