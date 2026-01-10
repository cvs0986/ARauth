# ARauth Admin Console User Guide

**Version**: 1.0  
**Status**: UI Complete - Backend Integration Pending  
**Audience**: Platform Admins, Tenant Admins, Security Operators

---

## Table of Contents

1. [Introduction](#introduction)
2. [Console Modes](#console-modes)
3. [Navigation](#navigation)
4. [Dashboards](#dashboards)
5. [Identity Management](#identity-management)
6. [OAuth2 Client Management](#oauth2-client-management)
7. [SCIM Provisioning](#scim-provisioning)
8. [Federation](#federation)
9. [Webhooks](#webhooks)
10. [Audit Logs](#audit-logs)
11. [Active Sessions](#active-sessions)
12. [Impersonation](#impersonation)
13. [Backend Integration Status](#backend-integration-status)

---

## Introduction

The ARauth Admin Console is an enterprise-grade IAM control plane for managing users, roles, permissions, OAuth2 clients, SCIM provisioning, federation, and operational observability.

**Key Principles**:
- **Authority First**: Every action is permission-gated
- **Backend Is Law**: UI never grants authority, only reflects it
- **Strict Plane Separation**: SYSTEM and TENANT modes are isolated
- **No Fake UX**: Missing APIs are explicitly marked "Coming Soon"

---

## Console Modes

The admin console operates in two distinct modes based on your principal type and context.

### SYSTEM Mode

**Who**: Users with `principalType: SYSTEM`  
**Scope**: Platform-wide operations  
**Capabilities**:
- Manage all tenants
- Configure system-level settings (OAuth2, JWT, capabilities)
- View cross-tenant metrics
- Impersonate users (when tenant selected)

**Tenant Selector**: SYSTEM users can select a tenant to operate in tenant context.

### TENANT Mode

**Who**: Users with `principalType: TENANT`  
**Scope**: Single tenant operations  
**Capabilities**:
- Manage users within tenant
- Configure tenant settings (tokens, password policy, MFA, rate limits)
- View tenant-specific metrics
- Cannot access other tenants

**No Tenant Selector**: TENANT users always operate within their home tenant.

---

## Navigation

Navigation is mode-based and permission-filtered. Inaccessible items are hidden (not disabled).

### SYSTEM Mode Navigation

**Platform**
- Dashboard - Platform overview
- Tenants - Tenant management

**Security**
- Audit Logs - All events
- Active Sessions - Cross-tenant sessions

**Configuration**
- System Settings - OAuth2, JWT, capabilities

### TENANT Mode Navigation

**Identity**
- Users - User management
- Roles - Role management
- Permissions - Permission management

**Access**
- OAuth2 Clients - Client applications
- SCIM - Provisioning

**Federation**
- OIDC Providers - External OIDC IdPs
- Identity Linking - External identity links

**Security**
- Audit Logs - Tenant events
- Active Sessions - Tenant sessions
- Webhooks - Event subscriptions

**Advanced**
- Tenant Settings - Configuration

---

## Dashboards

### System Dashboard

**Permission**: `dashboard:read` (SYSTEM)  
**Displays**:
- Total Tenants
- Active Users (cross-tenant)
- System Health

**Data Gaps**: Cross-tenant aggregation, MFA adoption, security posture (marked "Coming Soon")

### Tenant Dashboard

**Permission**: `dashboard:read` (TENANT)  
**Displays**:
- Total Users
- Active Roles
- Assigned Permissions

**Data Gaps**: MFA enrollment, user activity timeline (marked "Coming Soon")

---

## Identity Management

### User List

**Permission**: `users:read`  
**Columns**:
- Email
- Status (active/suspended)
- MFA Enabled
- Roles
- Last Login
- Tenant (SYSTEM mode only)

**Actions** (permission-gated):
- Create User (`users:create`)
- Edit User (`users:update`)
- Delete User (`users:delete`)
- Suspend User (`users:update` + audit reason)
- Activate User (`users:update`)
- Reset MFA (`users:mfa:reset` + audit reason)
- Impersonate (`users:impersonate`, SYSTEM only, tenant selected)

### Create User

**SYSTEM Mode**:
- Tenant selector (create system user or tenant user)
- Role list filtered by selected tenant scope

**TENANT Mode**:
- Always creates tenant user
- Role list shows only tenant roles

**Authority Badges**: Clear visual indicators for "System User" vs "Tenant User"

---

## OAuth2 Client Management

**Permission**: `oauth:clients:read`, `oauth:clients:create`  
**Scope**: Tenant-scoped

### Features

**List View**:
- Client Name
- Client ID
- Grant Types
- Redirect URIs
- Created Date

**Create Client**:
- Client Name
- Issuer URL
- Client ID
- Client Secret (one-time display)
- Grant Types (authorization_code, client_credentials, refresh_token)
- Redirect URIs (dynamic list, HTTPS validation)
- Scopes (openid, profile, email, offline_access)

**Security**:
- Client secret shown once
- Copy confirmation before closing
- No default scopes or grant types

---

## SCIM Provisioning

**Permission**: `scim:read`, `scim:tokens:create`, `scim:tokens:revoke`  
**Scope**: Tenant-scoped

### Features

**Configuration**:
- SCIM Base URL (read-only, with copy button)
- Tenant ID
- Status (enabled/disabled)

**Token Management**:
- List tokens (name, status, created, last used)
- Create token (one-time secret display)
- Revoke token (audit reason required)

**Security**:
- SCIM tokens are root credentials
- Token shown once (password-masked until copied)
- Revoke requires audit reason
- Confirmation before closing without copying

---

## Federation

### External OIDC Providers

**Permission**: `federation:idp:read`, `federation:idp:create`  
**Scope**: Tenant-scoped

**Features**:
- Provider Name
- Issuer URL
- Client ID/Secret (secret is one-time display)
- Scopes
- Explicit Attribute Mapping (email required)
- Test Connection (UI contract)
- Enable/Disable
- Delete (audit reason required)

**Security**:
- No automatic linking
- Explicit attribute mapping (no auto-mapping)
- Test connection before enabling
- Clear warnings for login impact

### Identity Linking

**Permission**: `federation:link`, `federation:unlink`  
**Scope**: Per-user

**Features**:
- View linked external identities (provider, type, external ID, linked date)
- Unlink identity (audit reason required, confirmation dialog)

**Security**:
- No automatic linking
- Unlink requires audit reason
- Clear warnings for authentication impact

---

## Webhooks

**Permission**: `webhooks:read`, `webhooks:create`, `webhooks:delete`  
**Scope**: System-wide OR Tenant-scoped

### Features

**List View**:
- Name
- URL
- Events
- Status (active/disabled)
- Last Delivery (timestamp, status)

**Create Webhook**:
- Name
- URL (HTTPS required)
- Event Selection (user.created, user.updated, role.assigned, etc.)
- Signing Secret (one-time display)
- Retry Policy

**Security**:
- Signing secret shown once
- Copy confirmation before closing
- HTTPS URL required
- Clear warnings about sensitive data

**Scope Indicators**:
- "System webhooks apply to all tenants"
- "Webhooks apply only to this tenant"

---

## Audit Logs

**Permission**: `audit:read`, `audit:export`  
**Scope**: System-wide OR Tenant-scoped

### Features

**Columns**:
- Timestamp
- Actor
- Action
- Target
- Result (success/failure)
- IP Address

**Filters** (UI contract):
- Actor
- Action type
- Result
- Time range

**Export** (UI contract):
- CSV
- JSON

**Scope Indicators**:
- "Viewing all system and tenant audit events"
- "Viewing audit events for this tenant only"

---

## Active Sessions

**Permission**: `sessions:read`, `sessions:revoke`  
**Scope**: System-wide OR Tenant-scoped

### Features

**Columns**:
- User
- IP Address
- User Agent
- Started At
- Last Activity
- Status (active/expired)

**Actions**:
- Revoke Session (audit reason required)
- Revoke All Sessions for User (UI contract)

**Security**:
- Revoke requires audit reason
- Clear warnings: "Revoking sessions logs user out immediately"

**Scope Indicators**:
- "Viewing all active sessions across all tenants"
- "Viewing active sessions for this tenant only"

---

## Impersonation

**Permission**: `users:impersonate` (SYSTEM only)  
**Requirements**:
- SYSTEM user
- Tenant must be selected
- Target user must exist in selected tenant

### Impersonation Banner

**Behavior**:
- Always visible at top of screen
- Cannot be dismissed
- Displays impersonated user email and tenant name
- "End Impersonation" button (one-click exit)

**Security**:
- Ending impersonation triggers audit event
- Page reload after ending impersonation
- Banner appears before any tenant UI

**Audit**: All impersonation events are logged (start, end, actions taken)

---

## Backend Integration Status

### What "Coming Soon" Means

Throughout the admin console, you will see "Coming Soon" or "Backend Integration Pending" messages. These indicate:

1. **The UI is complete** and serves as the contract for implementation
2. **The backend API does not yet exist** for this feature
3. **No fake data is shown** - the UI honestly reflects missing capabilities

### Current Status

**Fully Implemented UI** (Backend Integration Pending):
- User suspend/activate/reset MFA
- OAuth2 client management
- SCIM token management
- Federation (OIDC IdPs, Identity Linking)
- Webhooks
- Audit log filtering and export
- Active session management
- Impersonation state management

**UI Contract Mode**:
All unimplemented backend calls throw `APINotConnectedError` with clear user-facing messages. This prevents silent failures and ensures operators understand system capabilities.

### When Backend Integration Happens

Backend integration will occur via **vertical slices**:
1. Pick one feature (e.g., User Suspend)
2. Implement backend API
3. Connect UI to API
4. Test end-to-end
5. Repeat for next feature

The UI will **not change** during integration - only the backend connections will be added.

---

## Support

For questions or issues:
- Review this guide
- Check authority model documentation
- Consult API contract documentation
- Contact platform team

**Remember**: The UI never grants authority. All permissions come from the backend. If you cannot access a feature, you lack the required permission.
