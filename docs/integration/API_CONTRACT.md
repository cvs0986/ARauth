# ARauth Backend API Integration Contract

**Version**: 1.0  
**Status**: Definitive Contract  
**Purpose**: Single source of truth for backend API implementation

---

## Table of Contents

1. [Overview](#overview)
2. [User Management APIs](#user-management-apis)
3. [Role Management APIs](#role-management-apis)
4. [OAuth2 Client APIs](#oauth2-client-apis)
5. [SCIM APIs](#scim-apis)
6. [Federation APIs](#federation-apis)
7. [Webhook APIs](#webhook-apis)
8. [Audit Log APIs](#audit-log-apis)
9. [Session Management APIs](#session-management-apis)
10. [Impersonation APIs](#impersonation-apis)
11. [Error Handling](#error-handling)

---

## Overview

This document defines the exact API contracts required to integrate the ARauth Admin Console UI with the backend. Every endpoint listed here is **required** for the UI to function fully.

**Contract Rules**:
- All endpoints must enforce permissions server-side
- All destructive actions must record audit reasons
- All responses must include proper error codes
- All tenant-scoped endpoints must validate tenant boundaries

---

## User Management APIs

### List Users

**Endpoint**: `GET /api/v1/users` (TENANT) OR `GET /api/v1/system/users` (SYSTEM)  
**Permission**: `users:read`  
**Scope**: Tenant-scoped OR System-wide

**Query Parameters**:
- `tenant_id` (SYSTEM only) - Filter by tenant
- `status` - Filter by status (active, suspended)
- `page` - Page number
- `limit` - Items per page

**Response**:
```json
{
  "users": [
    {
      "id": "user_123",
      "email": "user@example.com",
      "status": "active",
      "mfa_enabled": true,
      "roles": ["role_1", "role_2"],
      "last_login_at": "2024-01-01T00:00:00Z",
      "tenant_id": "tenant_123",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 100,
  "page": 1,
  "limit": 50
}
```

### Suspend User

**Endpoint**: `POST /api/v1/users/{user_id}/suspend`  
**Permission**: `users:update`  
**Scope**: Tenant-scoped

**Request**:
```json
{
  "audit_reason": "User violated terms of service"
}
```

**Response**:
```json
{
  "id": "user_123",
  "status": "suspended",
  "suspended_at": "2024-01-01T00:00:00Z"
}
```

**Audit**: Log `user.suspended` event with reason

### Activate User

**Endpoint**: `POST /api/v1/users/{user_id}/activate`  
**Permission**: `users:update`  
**Scope**: Tenant-scoped

**Response**:
```json
{
  "id": "user_123",
  "status": "active",
  "activated_at": "2024-01-01T00:00:00Z"
}
```

**Audit**: Log `user.activated` event

### Reset User MFA

**Endpoint**: `POST /api/v1/users/{user_id}/mfa/reset`  
**Permission**: `users:mfa:reset`  
**Scope**: Tenant-scoped

**Request**:
```json
{
  "audit_reason": "User lost MFA device"
}
```

**Response**:
```json
{
  "id": "user_123",
  "mfa_enabled": false,
  "mfa_reset_at": "2024-01-01T00:00:00Z"
}
```

**Audit**: Log `user.mfa.reset` event with reason

---

## OAuth2 Client APIs

### List OAuth2 Clients

**Endpoint**: `GET /api/v1/tenants/{tenant_id}/oauth/clients`  
**Permission**: `oauth:clients:read`  
**Scope**: Tenant-scoped

**Response**:
```json
{
  "clients": [
    {
      "id": "client_123",
      "client_id": "oauth_abc123",
      "name": "My Application",
      "grant_types": ["authorization_code", "refresh_token"],
      "redirect_uris": ["https://example.com/callback"],
      "scopes": ["openid", "profile", "email"],
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### Create OAuth2 Client

**Endpoint**: `POST /api/v1/tenants/{tenant_id}/oauth/clients`  
**Permission**: `oauth:clients:create`  
**Scope**: Tenant-scoped

**Request**:
```json
{
  "name": "My Application",
  "grant_types": ["authorization_code", "refresh_token"],
  "redirect_uris": ["https://example.com/callback"],
  "scopes": ["openid", "profile", "email"]
}
```

**Response**:
```json
{
  "id": "client_123",
  "client_id": "oauth_abc123",
  "client_secret": "secret_xyz789",
  "name": "My Application",
  "grant_types": ["authorization_code", "refresh_token"],
  "redirect_uris": ["https://example.com/callback"],
  "scopes": ["openid", "profile", "email"],
  "created_at": "2024-01-01T00:00:00Z"
}
```

**Security**: `client_secret` is returned ONLY on creation, never again

**Audit**: Log `oauth.client.created` event

### Rotate Client Secret

**Endpoint**: `POST /api/v1/tenants/{tenant_id}/oauth/clients/{client_id}/rotate-secret`  
**Permission**: `oauth:clients:update`  
**Scope**: Tenant-scoped

**Response**:
```json
{
  "client_id": "oauth_abc123",
  "client_secret": "secret_new456",
  "rotated_at": "2024-01-01T00:00:00Z"
}
```

**Security**: New `client_secret` returned ONLY once

**Audit**: Log `oauth.client.secret_rotated` event

---

## SCIM APIs

### Get SCIM Configuration

**Endpoint**: `GET /api/v1/tenants/{tenant_id}/scim/config`  
**Permission**: `scim:read`  
**Scope**: Tenant-scoped

**Response**:
```json
{
  "enabled": true,
  "base_url": "https://api.arauth.example.com/scim/v2/tenants/{tenant_id}",
  "tenant_id": "tenant_123"
}
```

### List SCIM Tokens

**Endpoint**: `GET /api/v1/tenants/{tenant_id}/scim/tokens`  
**Permission**: `scim:tokens:read`  
**Scope**: Tenant-scoped

**Response**:
```json
{
  "tokens": [
    {
      "id": "token_123",
      "name": "Production IdP",
      "status": "active",
      "created_at": "2024-01-01T00:00:00Z",
      "last_used_at": "2024-01-02T00:00:00Z"
    }
  ]
}
```

### Create SCIM Token

**Endpoint**: `POST /api/v1/tenants/{tenant_id}/scim/tokens`  
**Permission**: `scim:tokens:create`  
**Scope**: Tenant-scoped

**Request**:
```json
{
  "name": "Production IdP"
}
```

**Response**:
```json
{
  "id": "token_123",
  "name": "Production IdP",
  "token": "scim_abc123xyz789...",
  "created_at": "2024-01-01T00:00:00Z"
}
```

**Security**: `token` is returned ONLY on creation, never again

**Audit**: Log `scim.token.created` event

### Revoke SCIM Token

**Endpoint**: `POST /api/v1/tenants/{tenant_id}/scim/tokens/{token_id}/revoke`  
**Permission**: `scim:tokens:revoke`  
**Scope**: Tenant-scoped

**Request**:
```json
{
  "audit_reason": "Token compromised"
}
```

**Response**:
```json
{
  "id": "token_123",
  "status": "revoked",
  "revoked_at": "2024-01-01T00:00:00Z"
}
```

**Audit**: Log `scim.token.revoked` event with reason

---

## Federation APIs

### List OIDC Identity Providers

**Endpoint**: `GET /api/v1/tenants/{tenant_id}/federation/oidc`  
**Permission**: `federation:idp:read`  
**Scope**: Tenant-scoped

**Response**:
```json
{
  "providers": [
    {
      "id": "idp_123",
      "name": "Google Workspace",
      "issuer_url": "https://accounts.google.com",
      "client_id": "client_abc123",
      "status": "active",
      "users_linked": 42,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### Create OIDC Identity Provider

**Endpoint**: `POST /api/v1/tenants/{tenant_id}/federation/oidc`  
**Permission**: `federation:idp:create`  
**Scope**: Tenant-scoped

**Request**:
```json
{
  "name": "Google Workspace",
  "issuer_url": "https://accounts.google.com",
  "client_id": "client_abc123",
  "client_secret": "secret_xyz789",
  "scopes": "openid profile email",
  "attribute_mapping": {
    "email": "email",
    "name": "name",
    "given_name": "given_name",
    "family_name": "family_name"
  }
}
```

**Response**:
```json
{
  "id": "idp_123",
  "name": "Google Workspace",
  "issuer_url": "https://accounts.google.com",
  "client_id": "client_abc123",
  "status": "disabled",
  "created_at": "2024-01-01T00:00:00Z"
}
```

**Security**: Provider created in `disabled` state by default

**Audit**: Log `federation.oidc.created` event

### Test OIDC Connection

**Endpoint**: `POST /api/v1/tenants/{tenant_id}/federation/oidc/{idp_id}/test`  
**Permission**: `federation:idp:update`  
**Scope**: Tenant-scoped

**Response**:
```json
{
  "success": true,
  "message": "Connection successful",
  "tested_at": "2024-01-01T00:00:00Z"
}
```

**Error Response**:
```json
{
  "success": false,
  "error": "Invalid client credentials",
  "tested_at": "2024-01-01T00:00:00Z"
}
```

### List User External Identities

**Endpoint**: `GET /api/v1/users/{user_id}/identities`  
**Permission**: `federation:link`  
**Scope**: Tenant-scoped

**Response**:
```json
{
  "identities": [
    {
      "id": "identity_123",
      "provider_name": "Google Workspace",
      "provider_type": "oidc",
      "external_id": "google_user_123",
      "linked_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### Unlink External Identity

**Endpoint**: `POST /api/v1/users/{user_id}/identities/{identity_id}/unlink`  
**Permission**: `federation:unlink`  
**Scope**: Tenant-scoped

**Request**:
```json
{
  "audit_reason": "User requested unlinking"
}
```

**Response**:
```json
{
  "id": "identity_123",
  "unlinked_at": "2024-01-01T00:00:00Z"
}
```

**Audit**: Log `federation.identity.unlinked` event with reason

---

## Webhook APIs

### List Webhooks

**Endpoint**: `GET /api/v1/webhooks` (SYSTEM) OR `GET /api/v1/tenants/{tenant_id}/webhooks` (TENANT)  
**Permission**: `webhooks:read`  
**Scope**: System-wide OR Tenant-scoped

**Response**:
```json
{
  "webhooks": [
    {
      "id": "webhook_123",
      "name": "Production Notifications",
      "url": "https://example.com/webhooks/arauth",
      "events": ["user.created", "user.updated"],
      "status": "active",
      "last_delivery_at": "2024-01-01T00:00:00Z",
      "last_delivery_status": "success",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### Create Webhook

**Endpoint**: `POST /api/v1/webhooks` (SYSTEM) OR `POST /api/v1/tenants/{tenant_id}/webhooks` (TENANT)  
**Permission**: `webhooks:create`  
**Scope**: System-wide OR Tenant-scoped

**Request**:
```json
{
  "name": "Production Notifications",
  "url": "https://example.com/webhooks/arauth",
  "events": ["user.created", "user.updated"],
  "retry_policy": {
    "max_attempts": 3,
    "backoff_seconds": 60
  }
}
```

**Response**:
```json
{
  "id": "webhook_123",
  "name": "Production Notifications",
  "url": "https://example.com/webhooks/arauth",
  "events": ["user.created", "user.updated"],
  "signing_secret": "whsec_abc123xyz789...",
  "status": "active",
  "created_at": "2024-01-01T00:00:00Z"
}
```

**Security**: `signing_secret` returned ONLY on creation, never again

**Audit**: Log `webhook.created` event

---

## Audit Log APIs

### List Audit Logs

**Endpoint**: `GET /api/v1/audit/logs` (SYSTEM) OR `GET /api/v1/tenants/{tenant_id}/audit/logs` (TENANT)  
**Permission**: `audit:read`  
**Scope**: System-wide OR Tenant-scoped

**Query Parameters**:
- `actor` - Filter by actor email
- `action` - Filter by action type
- `result` - Filter by result (success, failure)
- `start_time` - Filter by start timestamp
- `end_time` - Filter by end timestamp
- `page` - Page number
- `limit` - Items per page

**Response**:
```json
{
  "logs": [
    {
      "id": "log_123",
      "timestamp": "2024-01-01T00:00:00Z",
      "actor": "admin@example.com",
      "action": "user.suspended",
      "target": "user_123",
      "result": "success",
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0...",
      "audit_reason": "User violated terms",
      "metadata": {}
    }
  ],
  "total": 1000,
  "page": 1,
  "limit": 50
}
```

### Export Audit Logs

**Endpoint**: `POST /api/v1/audit/logs/export`  
**Permission**: `audit:export`  
**Scope**: System-wide OR Tenant-scoped

**Request**:
```json
{
  "format": "csv",
  "filters": {
    "start_time": "2024-01-01T00:00:00Z",
    "end_time": "2024-01-31T23:59:59Z"
  }
}
```

**Response**:
```json
{
  "export_id": "export_123",
  "status": "processing",
  "download_url": null,
  "created_at": "2024-01-01T00:00:00Z"
}
```

**Audit**: Log `audit.export.requested` event

---

## Session Management APIs

### List Active Sessions

**Endpoint**: `GET /api/v1/sessions` (SYSTEM) OR `GET /api/v1/tenants/{tenant_id}/sessions` (TENANT)  
**Permission**: `sessions:read`  
**Scope**: System-wide OR Tenant-scoped

**Response**:
```json
{
  "sessions": [
    {
      "id": "session_123",
      "user_id": "user_123",
      "user_email": "user@example.com",
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0...",
      "started_at": "2024-01-01T00:00:00Z",
      "last_activity_at": "2024-01-01T01:00:00Z",
      "status": "active"
    }
  ]
}
```

### Revoke Session

**Endpoint**: `POST /api/v1/sessions/{session_id}/revoke`  
**Permission**: `sessions:revoke`  
**Scope**: Tenant-scoped

**Request**:
```json
{
  "audit_reason": "Security incident"
}
```

**Response**:
```json
{
  "id": "session_123",
  "status": "revoked",
  "revoked_at": "2024-01-01T00:00:00Z"
}
```

**Audit**: Log `session.revoked` event with reason

---

## Impersonation APIs

### Start Impersonation

**Endpoint**: `POST /api/v1/system/impersonate`  
**Permission**: `users:impersonate` (SYSTEM only)  
**Scope**: System-wide

**Request**:
```json
{
  "user_id": "user_123",
  "tenant_id": "tenant_123"
}
```

**Response**:
```json
{
  "impersonation_token": "imp_abc123xyz789...",
  "user": {
    "id": "user_123",
    "email": "user@example.com"
  },
  "tenant": {
    "id": "tenant_123",
    "name": "Example Tenant"
  },
  "started_at": "2024-01-01T00:00:00Z"
}
```

**Audit**: Log `impersonation.started` event

### End Impersonation

**Endpoint**: `POST /api/v1/system/impersonate/end`  
**Permission**: `users:impersonate` (SYSTEM only)  
**Scope**: System-wide

**Response**:
```json
{
  "ended_at": "2024-01-01T00:00:00Z",
  "duration_seconds": 3600
}
```

**Audit**: Log `impersonation.ended` event

---

## Error Handling

### Standard Error Response

```json
{
  "error": {
    "code": "PERMISSION_DENIED",
    "message": "You do not have permission to perform this action",
    "details": {
      "required_permission": "users:update"
    }
  }
}
```

### Error Codes

- `PERMISSION_DENIED` - User lacks required permission
- `TENANT_BOUNDARY_VIOLATION` - Attempted cross-tenant access
- `RESOURCE_NOT_FOUND` - Resource does not exist
- `VALIDATION_ERROR` - Request validation failed
- `AUDIT_REASON_REQUIRED` - Destructive action missing audit reason
- `IMPERSONATION_NOT_ALLOWED` - Impersonation requirements not met
- `SECRET_ALREADY_RETRIEVED` - Attempted to retrieve one-time secret again

---

## Summary

**Total Endpoints**: 30+  
**Permission-Gated**: All  
**Audit-Logged**: All destructive actions  
**Tenant-Scoped**: Most endpoints

**Integration Priority** (Vertical Slices):
1. User Management (suspend, activate, reset MFA)
2. OAuth2 Clients (list, create, rotate secret)
3. SCIM Tokens (list, create, revoke)
4. Federation (OIDC IdPs, identity linking)
5. Webhooks (list, create)
6. Audit Logs (list, export)
7. Sessions (list, revoke)
8. Impersonation (start, end)

**This contract is binding and must be implemented exactly as specified.**
