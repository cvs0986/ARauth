# Data Flow Documentation

This document describes the authentication and authorization flows in Nuage Identity.

## 🔐 Authentication Flows

### 1. Direct Login Flow (Simplified)

This is the primary flow for applications that want a simple username/password login.

```
┌─────────────┐                    ┌─────────────┐                    ┌─────────────┐
│ Client App  │                    │   IAM API   │                    │    Hydra    │
└──────┬──────┘                    └──────┬──────┘                    └──────┬──────┘
       │                                   │                                   │
       │ 1. POST /auth/login               │                                   │
       │    {username, password, tenant}   │                                   │
       ├──────────────────────────────────>│                                   │
       │                                   │                                   │
       │                                   │ 2. Validate credentials           │
       │                                   │    (check DB)                     │
       │                                   │                                   │
       │                                   │ 3. Check MFA requirement          │
       │                                   │    (if enabled)                    │
       │                                   │                                   │
       │                                   │ 4. Build claims                   │
       │                                   │    (roles, permissions)           │
       │                                   │                                   │
       │                                   │ 5. POST /admin/oauth2/auth/requests/login/accept
       │                                   │    {login_challenge, subject, claims}
       │                                   ├──────────────────────────────────>│
       │                                   │                                   │
       │                                   │ 6. Create OAuth2 client (if needed)
       │                                   │                                   │
       │                                   │ 7. Issue tokens                   │
       │                                   │<──────────────────────────────────┤
       │                                   │                                   │
       │ 8. Response:                     │                                   │
       │    {access_token, refresh_token,  │                                   │
       │     id_token, expires_in}         │                                   │
       │<──────────────────────────────────┤                                   │
       │                                   │                                   │
```

**Steps:**

1. **Client Request**: Client sends login credentials to IAM API
2. **Credential Validation**: IAM validates username/password against database
3. **MFA Check**: If MFA is enabled, return MFA challenge
4. **Claims Building**: Build JWT claims from user roles and permissions
5. **Hydra Integration**: Call Hydra Admin API to accept login and issue tokens
6. **Token Response**: Return tokens to client

### 2. OAuth2 Authorization Code Flow (with PKCE)

This flow is for applications that want full OAuth2 compliance.

```
┌─────────────┐                    ┌─────────────┐                    ┌─────────────┐
│ Client App  │                    │    Hydra    │                    │   IAM API   │
└──────┬──────┘                    └──────┬──────┘                    └──────┬──────┘
       │                                   │                                   │
       │ 1. Generate code_verifier         │                                   │
       │    Generate code_challenge        │                                   │
       │                                   │                                   │
       │ 2. GET /oauth2/auth               │                                   │
       │    ?client_id=...                 │                                   │
       │    &redirect_uri=...              │                                   │
       │    &code_challenge=...            │                                   │
       │    &code_challenge_method=S256    │                                   │
       ├──────────────────────────────────>│                                   │
       │                                   │                                   │
       │                                   │ 3. POST /oauth2/auth/requests/login
       │                                   │    (login_challenge)              │
       │                                   ├──────────────────────────────────>│
       │                                   │                                   │
       │                                   │ 4. Return login_challenge         │
       │                                   │<──────────────────────────────────┤
       │                                   │                                   │
       │ 5. Redirect to login UI           │                                   │
       │    with login_challenge           │                                   │
       │<──────────────────────────────────┤                                   │
       │                                   │                                   │
       │ 6. POST /auth/login               │                                   │
       │    {login_challenge, username,    │                                   │
       │     password}                     │                                   │
       ├──────────────────────────────────────────────────────────────────────>│
       │                                   │                                   │
       │                                   │ 7. Validate credentials           │
       │                                   │                                   │
       │                                   │ 8. Build claims                   │
       │                                   │                                   │
       │                                   │ 9. POST /admin/oauth2/auth/requests/login/accept
       │                                   │    {login_challenge, subject, claims}
       │                                   │                                   │
       │                                   │ 10. Return redirect_uri with code  │
       │                                   │<──────────────────────────────────┤
       │                                   │                                   │
       │ 11. GET redirect_uri?code=...     │                                   │
       │<──────────────────────────────────┤                                   │
       │                                   │                                   │
       │ 12. POST /oauth2/token            │                                   │
       │     {code, code_verifier, ...}    │                                   │
       ├──────────────────────────────────>│                                   │
       │                                   │                                   │
       │ 13. Return tokens                 │                                   │
       │     {access_token, refresh_token, │                                   │
       │      id_token}                    │                                   │
       │<──────────────────────────────────┤                                   │
       │                                   │                                   │
```

**Steps:**

1. **PKCE Setup**: Client generates code_verifier and code_challenge
2. **Authorization Request**: Client redirects to Hydra authorization endpoint
3. **Login Challenge**: Hydra calls IAM API with login_challenge
4. **Login UI**: Client shows login UI with login_challenge
5. **Login Request**: Client sends credentials to IAM API
6. **Validation**: IAM validates credentials
7. **Claims Building**: Build JWT claims
8. **Accept Login**: IAM accepts login in Hydra
9. **Authorization Code**: Hydra returns authorization code
10. **Token Exchange**: Client exchanges code for tokens

### 3. MFA Flow

```
┌─────────────┐                    ┌─────────────┐                    ┌─────────────┐
│ Client App  │                    │   IAM API   │                    │    Redis    │
└──────┬──────┘                    └──────┬──────┘                    └──────┬──────┘
       │                                   │                                   │
       │ 1. POST /auth/login               │                                   │
       │    {username, password}           │                                   │
       ├──────────────────────────────────>│                                   │
       │                                   │                                   │
       │                                   │ 2. Validate credentials           │
       │                                   │                                   │
       │                                   │ 3. Check MFA enabled              │
       │                                   │                                   │
       │                                   │ 4. Generate MFA session           │
       │                                   │    Store in Redis (TTL: 5 min)    │
       │                                   ├──────────────────────────────────>│
       │                                   │                                   │
       │ 5. Response:                      │                                   │
       │    {mfa_required: true,           │                                   │
       │     mfa_session_id: "..."}        │                                   │
       │<──────────────────────────────────┤                                   │
       │                                   │                                   │
       │ 6. POST /auth/mfa/verify          │                                   │
       │    {mfa_session_id, totp_code}    │                                   │
       ├──────────────────────────────────>│                                   │
       │                                   │                                   │
       │                                   │ 7. Get MFA session from Redis     │
       │                                   │<──────────────────────────────────┤
       │                                   │                                   │
       │                                   │ 8. Validate TOTP code             │
       │                                   │                                   │
       │                                   │ 9. Delete MFA session             │
       │                                   ├──────────────────────────────────>│
       │                                   │                                   │
       │                                   │ 10. Continue with token issuance  │
       │                                   │                                   │
       │ 11. Response: tokens              │                                   │
       │<──────────────────────────────────┤                                   │
       │                                   │                                   │
```

### 4. Token Refresh Flow

```
┌─────────────┐                    ┌─────────────┐                    ┌─────────────┐
│ Client App  │                    │   IAM API   │                    │    Hydra    │
└──────┬──────┘                    └──────┬──────┘                    └──────┬──────┘
       │                                   │                                   │
       │ 1. POST /auth/refresh             │                                   │
       │    {refresh_token}                │                                   │
       ├──────────────────────────────────>│                                   │
       │                                   │                                   │
       │                                   │ 2. Validate refresh token          │
       │                                   │    (check Redis blacklist)         │
       │                                   │                                   │
       │                                   │ 3. POST /oauth2/token              │
       │                                   │    {grant_type: refresh_token, ...}│
       │                                   ├──────────────────────────────────>│
       │                                   │                                   │
       │                                   │ 4. Rotate refresh token            │
       │                                   │                                   │
       │                                   │ 5. Return new tokens               │
       │                                   │<──────────────────────────────────┤
       │                                   │                                   │
       │ 6. Response:                      │                                   │
       │    {access_token, refresh_token,  │                                   │
       │     id_token}                     │                                   │
       │<──────────────────────────────────┤                                   │
       │                                   │                                   │
```

### 5. Logout Flow

```
┌─────────────┐                    ┌─────────────┐                    ┌─────────────┐
│ Client App  │                    │   IAM API   │                    │    Redis    │
└──────┬──────┘                    └──────┬──────┘                    └──────┬──────┘
       │                                   │                                   │
       │ 1. POST /auth/logout              │                                   │
       │    {refresh_token}                │                                   │
       ├──────────────────────────────────>│                                   │
       │                                   │                                   │
       │                                   │ 2. Validate refresh token          │
       │                                   │                                   │
       │                                   │ 3. Add to blacklist (Redis)        │
       │                                   │    TTL: refresh_token_expiry        │
       │                                   ├──────────────────────────────────>│
       │                                   │                                   │
       │                                   │ 4. Revoke in Hydra (optional)     │
       │                                   │                                   │
       │ 5. Response: {success: true}      │                                   │
       │<──────────────────────────────────┤                                   │
       │                                   │                                   │
```

## 🔑 Authorization Flow

### Service-to-Service Authorization

Microservices validate JWTs without calling IAM:

```
┌─────────────┐                    ┌─────────────┐                    ┌─────────────┐
│   Service   │                    │   IAM API   │                    │   Service   │
└──────┬──────┘                    └──────┬──────┘                    └──────┬──────┘
       │                                   │                                   │
       │ 1. Request with JWT               │                                   │
       │    Authorization: Bearer <token>  │                                   │
       ├──────────────────────────────────────────────────────────────────────>│
       │                                   │                                   │
       │                                   │ 2. Extract JWT from header        │
       │                                   │                                   │
       │                                   │ 3. Validate JWT signature         │
       │                                   │    (using JWKS endpoint)          │
       │                                   ├──────────────────────────────────>│
       │                                   │                                   │
       │                                   │ 4. Check expiration               │
       │                                   │                                   │
       │                                   │ 5. Extract claims                  │
       │                                   │    {sub, tenant, roles, permissions}
       │                                   │                                   │
       │                                   │ 6. Authorize based on claims       │
       │                                   │                                   │
       │ 7. Response (if authorized)       │                                   │
       │<──────────────────────────────────────────────────────────────────────┤
       │                                   │                                   │
```

### Permission Check Flow

```
Service receives request with JWT
    ↓
Extract claims from JWT
    ↓
Check required permission in claims.permissions
    ↓
If permission exists → Allow
If permission missing → Deny (403)
```

## 📊 Data Flow Summary

### Request Flow

```
Client Request
    ↓
API Middleware (CORS, Rate Limit, Logging)
    ↓
Route Handler
    ↓
Service Layer (Auth/Identity/Policy)
    ↓
Repository Layer
    ↓
Database/Redis
```

### Response Flow

```
Database/Redis
    ↓
Repository Layer
    ↓
Service Layer
    ↓
Route Handler
    ↓
Response Formatter
    ↓
Client
```

## 🔄 State Management

### Stateless Design

- **No server-side sessions**: All state in JWTs or external storage
- **Redis for temporary state**:
  - MFA sessions (TTL: 5 minutes)
  - Rate limiting counters (TTL: 1 minute)
  - Refresh token blacklist (TTL: token expiry)

### Token State

- **Access Token**: Stateless JWT, validated by signature
- **Refresh Token**: Opaque token, stored in Hydra DB
- **ID Token**: Stateless JWT, contains user info

## 🚨 Error Flow

```
Error occurs in service
    ↓
Error wrapped with context
    ↓
Error handler middleware
    ↓
Error response formatted
    ↓
Client receives error
    {
      "error": "error_code",
      "message": "Human readable message",
      "details": {...}
    }
```

## 📚 Related Documentation

- [Architecture Overview](./overview.md)
- [Components](./components.md)
- [Integration Patterns](./integration-patterns.md)

