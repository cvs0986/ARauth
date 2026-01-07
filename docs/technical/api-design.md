# API Design

This document describes the API design, endpoints, request/response formats, and authentication for ARauth Identity.

## 🎯 API Design Principles

1. **RESTful**: Follow REST conventions
2. **Stateless**: No server-side sessions
3. **JSON**: JSON request/response format
4. **Versioning**: API versioning in URL path
5. **Error Handling**: Consistent error responses
6. **Documentation**: OpenAPI specification

## 📋 API Base

```
Base URL: https://iam.example.com/api/v1
Content-Type: application/json
```

## 🔐 Authentication Endpoints

### POST /auth/login

**Description**: Authenticate user and obtain tokens.

**Request**:
```json
{
  "username": "user@example.com",
  "password": "password123",
  "tenant_id": "tenant-123",
  "client_id": "client-123",
  "login_challenge": "challenge-123"  // Optional, for OAuth2 flow
}
```

**Response** (200 OK):
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "refresh_token": "refresh-token-123",
  "id_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 900,
  "mfa_required": false
}
```

**Response** (200 OK, MFA Required):
```json
{
  "mfa_required": true,
  "mfa_session_id": "mfa-session-123"
}
```

**Errors**:
- `400 Bad Request`: Invalid request
- `401 Unauthorized`: Invalid credentials
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error

### POST /auth/mfa/verify

**Description**: Verify MFA code.

**Request**:
```json
{
  "mfa_session_id": "mfa-session-123",
  "totp_code": "123456",
  "recovery_code": "recovery-code-123"  // Alternative to TOTP
}
```

**Response** (200 OK):
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "refresh_token": "refresh-token-123",
  "id_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 900
}
```

**Errors**:
- `400 Bad Request`: Invalid request
- `401 Unauthorized`: Invalid MFA code
- `404 Not Found`: MFA session not found
- `429 Too Many Requests`: Rate limit exceeded

### POST /auth/refresh

**Description**: Refresh access token.

**Request**:
```json
{
  "refresh_token": "refresh-token-123",
  "client_id": "client-123"
}
```

**Response** (200 OK):
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "refresh_token": "new-refresh-token-123",
  "id_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 900
}
```

**Errors**:
- `400 Bad Request`: Invalid request
- `401 Unauthorized`: Invalid refresh token
- `429 Too Many Requests`: Rate limit exceeded

### POST /auth/logout

**Description**: Logout and invalidate refresh token.

**Request**:
```json
{
  "refresh_token": "refresh-token-123"
}
```

**Response** (200 OK):
```json
{
  "success": true
}
```

**Errors**:
- `400 Bad Request`: Invalid request
- `401 Unauthorized`: Invalid token

## 👤 User Management Endpoints

### POST /users

**Description**: Create a new user.

**Authentication**: Required (Admin role)

**Request**:
```json
{
  "username": "user@example.com",
  "email": "user@example.com",
  "password": "password123",
  "tenant_id": "tenant-123",
  "first_name": "John",
  "last_name": "Doe",
  "roles": ["user"]
}
```

**Response** (201 Created):
```json
{
  "id": "user-123",
  "username": "user@example.com",
  "email": "user@example.com",
  "tenant_id": "tenant-123",
  "first_name": "John",
  "last_name": "Doe",
  "roles": ["user"],
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### GET /users/:id

**Description**: Get user by ID.

**Authentication**: Required

**Response** (200 OK):
```json
{
  "id": "user-123",
  "username": "user@example.com",
  "email": "user@example.com",
  "tenant_id": "tenant-123",
  "first_name": "John",
  "last_name": "Doe",
  "roles": ["user"],
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### GET /users

**Description**: List users (with pagination).

**Authentication**: Required

**Query Parameters**:
- `tenant_id` (optional): Filter by tenant
- `page` (default: 1): Page number
- `limit` (default: 20, max: 100): Items per page

**Response** (200 OK):
```json
{
  "data": [
    {
      "id": "user-123",
      "username": "user@example.com",
      "email": "user@example.com",
      "tenant_id": "tenant-123",
      "first_name": "John",
      "last_name": "Doe",
      "roles": ["user"],
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

### PUT /users/:id

**Description**: Update user.

**Authentication**: Required

**Request**:
```json
{
  "email": "newemail@example.com",
  "first_name": "Jane",
  "last_name": "Smith"
}
```

**Response** (200 OK):
```json
{
  "id": "user-123",
  "username": "user@example.com",
  "email": "newemail@example.com",
  "tenant_id": "tenant-123",
  "first_name": "Jane",
  "last_name": "Smith",
  "roles": ["user"],
  "updated_at": "2024-01-01T01:00:00Z"
}
```

### DELETE /users/:id

**Description**: Delete user.

**Authentication**: Required (Admin role)

**Response** (204 No Content)

## 🏢 Tenant Management Endpoints

### POST /tenants

**Description**: Create a new tenant.

**Authentication**: Required (Super Admin)

**Request**:
```json
{
  "name": "Acme Corp",
  "domain": "acme.com",
  "status": "active"
}
```

**Response** (201 Created):
```json
{
  "id": "tenant-123",
  "name": "Acme Corp",
  "domain": "acme.com",
  "status": "active",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### GET /tenants/:id

**Description**: Get tenant by ID.

**Authentication**: Required

**Response** (200 OK): Same as POST response

### GET /tenants

**Description**: List tenants.

**Authentication**: Required (Super Admin)

**Response** (200 OK):
```json
{
  "data": [
    {
      "id": "tenant-123",
      "name": "Acme Corp",
      "domain": "acme.com",
      "status": "active",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 10,
    "total_pages": 1
  }
}
```

## 🔑 Role & Permission Endpoints

### POST /roles

**Description**: Create a new role.

**Authentication**: Required (Admin role)

**Request**:
```json
{
  "name": "admin",
  "description": "Administrator role",
  "permissions": ["user.read", "user.write", "tenant.read"]
}
```

**Response** (201 Created):
```json
{
  "id": "role-123",
  "name": "admin",
  "description": "Administrator role",
  "permissions": ["user.read", "user.write", "tenant.read"],
  "created_at": "2024-01-01T00:00:00Z"
}
```

### GET /roles

**Description**: List roles.

**Authentication**: Required

**Response** (200 OK):
```json
{
  "data": [
    {
      "id": "role-123",
      "name": "admin",
      "description": "Administrator role",
      "permissions": ["user.read", "user.write", "tenant.read"]
    }
  ]
}
```

### POST /users/:id/roles

**Description**: Assign roles to user.

**Authentication**: Required (Admin role)

**Request**:
```json
{
  "role_ids": ["role-123", "role-456"]
}
```

**Response** (200 OK):
```json
{
  "user_id": "user-123",
  "roles": [
    {
      "id": "role-123",
      "name": "admin"
    },
    {
      "id": "role-456",
      "name": "user"
    }
  ]
}
```

## 🔒 MFA Endpoints

### POST /users/:id/mfa/enroll

**Description**: Enroll user in MFA.

**Authentication**: Required

**Response** (200 OK):
```json
{
  "secret": "JBSWY3DPEHPK3PXP",
  "qr_code": "data:image/png;base64,iVBORw0KGgo...",
  "recovery_codes": [
    "recovery-code-1",
    "recovery-code-2",
    "..."
  ]
}
```

### DELETE /users/:id/mfa

**Description**: Disable MFA for user.

**Authentication**: Required

**Response** (204 No Content)

## 🏥 Health & Status Endpoints

### GET /health

**Description**: Health check endpoint.

**Response** (200 OK):
```json
{
  "status": "healthy",
  "checks": {
    "database": "up",
    "redis": "up",
    "hydra": "up"
  }
}
```

### GET /metrics

**Description**: Prometheus metrics endpoint.

**Response**: Prometheus format

## 📝 Error Response Format

All errors follow this format:

```json
{
  "error": "error_code",
  "message": "Human readable error message",
  "details": {
    "field": "additional error details"
  },
  "request_id": "request-id-123"
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `invalid_request` | 400 | Invalid request format |
| `unauthorized` | 401 | Authentication required |
| `forbidden` | 403 | Insufficient permissions |
| `not_found` | 404 | Resource not found |
| `conflict` | 409 | Resource conflict |
| `rate_limit_exceeded` | 429 | Rate limit exceeded |
| `internal_error` | 500 | Internal server error |

## 🔐 Authentication

### Bearer Token

Most endpoints require authentication via Bearer token:

```
Authorization: Bearer <access_token>
```

### Token Validation

Tokens are validated:
1. Signature verification (JWKS)
2. Expiration check
3. Issuer validation
4. Audience validation

## 📊 Rate Limiting

Rate limits are applied per IP and per user:

- **Login**: 5 attempts per minute
- **Token Refresh**: 10 requests per minute
- **API Calls**: 100 requests per minute

Rate limit headers:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640995200
```

## 📚 OpenAPI Specification

Full OpenAPI 3.0 specification available at:
```
GET /api/v1/openapi.json
```

## 🔄 API Versioning

API versioning in URL path:
- Current: `/api/v1`
- Future: `/api/v2`

Breaking changes require new version.

## 📚 Related Documentation

- [Architecture Overview](../architecture/overview.md)
- [Data Flow](../architecture/data-flow.md)
- [Security](./security.md)

