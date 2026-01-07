# 🔐 Authentication Flows - Complete Guide

**Date**: 2026-01-08

This document explains how both authentication flows work in Nuage Identity.

---

## 🎯 Two Authentication Flows

Nuage Identity supports **two authentication flows**:

1. **Direct JWT Flow** ✅ **FULLY IMPLEMENTED**
2. **OAuth2/OIDC Flow** ⚠️ **PARTIALLY IMPLEMENTED** (Hydra integration exists)

---

## 1️⃣ Direct JWT Flow (Primary - Fully Implemented)

### Overview

The **Direct JWT Flow** is a simplified authentication flow where the IAM API directly issues JWT tokens without going through an OAuth2 authorization server. This is ideal for:
- API-to-API authentication
- Mobile applications
- Single Page Applications (SPAs)
- Internal services
- Simple integrations

### How It Works

```
┌─────────┐                    ┌──────────────┐                    ┌──────────┐
│ Client  │                    │  IAM API     │                    │ Database │
└────┬────┘                    └──────┬───────┘                    └────┬─────┘
     │                                 │                                 │
     │ 1. POST /api/v1/auth/login     │                                 │
     │    {username, password,         │                                 │
     │     tenant_id (header),         │                                 │
     │     remember_me}                │                                 │
     ├────────────────────────────────>│                                 │
     │                                 │                                 │
     │                                 │ 2. Validate credentials        │
     │                                 ├───────────────────────────────>│
     │                                 │<───────────────────────────────┤
     │                                 │                                 │
     │                                 │ 3. Build JWT claims             │
     │                                 │    (roles, permissions)         │
     │                                 │                                 │
     │                                 │ 4. Sign JWT (RS256)             │
     │                                 │                                 │
     │                                 │ 5. Generate refresh token        │
     │                                 │                                 │
     │                                 │ 6. Store refresh token           │
     │                                 ├───────────────────────────────>│
     │                                 │<───────────────────────────────┤
     │                                 │                                 │
     │ 7. {access_token, refresh_token,│                                 │
     │     expires_in, token_type}     │                                 │
     │<────────────────────────────────┤                                 │
     │                                 │                                 │
```

### Step-by-Step Flow

#### Step 1: Login Request

**Endpoint**: `POST /api/v1/auth/login`

**Headers**:
```
X-Tenant-ID: <tenant-uuid>  (Required)
Content-Type: application/json
```

**Request Body**:
```json
{
  "username": "john.doe",
  "password": "SecurePassword123!",
  "remember_me": false
}
```

**Response** (Success):
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "550e8400-e29b-41d4-a716-446655440000",
  "id_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 900,
  "refresh_expires_in": 2592000,
  "remember_me": false
}
```

#### Step 2: Use Access Token

**Endpoint**: Any protected API endpoint

**Headers**:
```
Authorization: Bearer <access_token>
X-Tenant-ID: <tenant-uuid>
```

**Example**:
```bash
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "X-Tenant-ID: 6e3a1985-61e6-4446-9b42-d4d0c39dad7a"
```

#### Step 3: Refresh Token (When Access Token Expires)

**Endpoint**: `POST /api/v1/auth/refresh`

**Request Body**:
```json
{
  "refresh_token": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Response**:
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "660e8400-e29b-41d4-a716-446655440001",
  "token_type": "Bearer",
  "expires_in": 900,
  "refresh_expires_in": 2592000
}
```

**Note**: The old refresh token is automatically revoked (token rotation).

#### Step 4: Revoke Token (Logout)

**Endpoint**: `POST /api/v1/auth/revoke`

**Request Body**:
```json
{
  "token": "550e8400-e29b-41d4-a716-446655440000",
  "token_type_hint": "refresh_token"
}
```

**Response**:
```json
{
  "message": "Token revoked successfully"
}
```

### Token Structure

**Access Token (JWT)**:
```json
{
  "sub": "user-uuid",
  "tenant_id": "tenant-uuid",
  "email": "john.doe@example.com",
  "username": "john.doe",
  "roles": ["admin", "user"],
  "permissions": ["user.read", "user.write"],
  "iss": "https://iam.example.com",
  "exp": 1234567890,
  "iat": 1234567890,
  "jti": "token-id-uuid"
}
```

**Refresh Token**: Opaque UUID stored in database

### Security Features

- ✅ **RS256 Signing**: RSA with SHA-256 (industry standard)
- ✅ **Short-lived Access Tokens**: 15 minutes (configurable)
- ✅ **Long-lived Refresh Tokens**: 30 days (configurable)
- ✅ **Token Rotation**: New refresh token on each refresh
- ✅ **Token Revocation**: Immediate logout capability
- ✅ **Remember Me**: Extended lifetimes (90 days refresh, 60 min access)

### Configuration

Token lifetimes can be configured via:
1. **Per-Tenant Settings** (Database) - Highest priority
2. **Environment Variables** - `JWT_ACCESS_TOKEN_TTL`, `JWT_REFRESH_TOKEN_TTL`
3. **Config File** - `config/config.yaml`
4. **System Defaults** - Fallback values

---

## 2️⃣ OAuth2/OIDC Flow (Secondary - Partially Implemented)

### Overview

The **OAuth2/OIDC Flow** uses ORY Hydra as the authorization server. This is ideal for:
- Enterprise integrations
- Third-party applications
- Standard OAuth2/OIDC compliance
- Multi-tenant SaaS platforms

### Current Implementation Status

**✅ Implemented**:
- Hydra client integration
- Login challenge handling
- Claims injection into Hydra
- Basic OAuth2 flow support

**⚠️ Partially Implemented**:
- Login endpoint accepts `login_challenge` parameter
- Hydra integration code exists
- Needs testing and documentation

**❌ Not Implemented**:
- Authorization endpoint (handled by Hydra)
- Token endpoint (handled by Hydra)
- Consent endpoint (handled by Hydra)
- JWKS endpoint (for Hydra)
- Userinfo endpoint (OIDC)

### How It Works (Current Implementation)

```
┌─────────┐    ┌──────────┐    ┌──────────────┐    ┌──────────┐
│ Client  │    │  Hydra   │    │  IAM API     │    │ Database │
└────┬────┘    └────┬─────┘    └──────┬───────┘    └────┬─────┘
     │              │                 │                 │
     │ 1. Initiate OAuth2 flow        │                 │
     ├─────────────>│                 │                 │
     │              │                 │                 │
     │              │ 2. Redirect to login              │
     │              │    with login_challenge           │
     │<─────────────┤                 │                 │
     │              │                 │                 │
     │ 3. POST /api/v1/auth/login     │                 │
     │    {username, password,         │                 │
     │     login_challenge}            │                 │
     ├────────────────────────────────>│                 │
     │                                 │                 │
     │                                 │ 4. Validate     │
     │                                 ├────────────────>│
     │                                 │<────────────────┤
     │                                 │                 │
     │                                 │ 5. Accept login │
     │                                 │    challenge    │
     │                                 ├───────────────>│
     │                                 │                 │
     │              │ 6. Redirect to consent            │
     │<─────────────┤                 │                 │
     │              │                 │                 │
     │              │ 7. Get tokens from Hydra          │
     │<─────────────┤                 │                 │
     │              │                 │                 │
```

### Step-by-Step Flow

#### Step 1: Client Initiates OAuth2 Flow

Client redirects user to Hydra's authorization endpoint:
```
GET https://hydra.example.com/oauth2/auth?
  client_id=client-id
  &redirect_uri=https://client.example.com/callback
  &response_type=code
  &scope=openid profile email
  &state=random-state
```

#### Step 2: Hydra Redirects to IAM Login

Hydra redirects to IAM API with `login_challenge`:
```
GET https://iam.example.com/login?
  login_challenge=abc123...
```

#### Step 3: IAM Login with Challenge

**Endpoint**: `POST /api/v1/auth/login`

**Request Body**:
```json
{
  "username": "john.doe",
  "password": "SecurePassword123!",
  "login_challenge": "abc123..."
}
```

**What Happens**:
1. IAM validates credentials
2. IAM accepts the login challenge with Hydra
3. Hydra redirects to consent screen (or auto-consent)
4. Hydra issues tokens

**Response**:
```json
{
  "redirect_to": "https://hydra.example.com/oauth2/auth?login=xyz789...",
  "mfa_required": false
}
```

#### Step 4: Client Gets Tokens from Hydra

After consent, client exchanges authorization code for tokens:
```
POST https://hydra.example.com/oauth2/token
  grant_type=authorization_code
  &code=authorization-code
  &redirect_uri=https://client.example.com/callback
```

Hydra returns:
```json
{
  "access_token": "...",
  "refresh_token": "...",
  "id_token": "...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

### Current Limitations

1. **Hydra Must Be Running**: OAuth2 flow requires Hydra to be deployed
2. **Consent Handling**: Currently relies on Hydra's consent screen
3. **Testing**: Needs end-to-end testing with Hydra
4. **Documentation**: Needs complete OAuth2 flow documentation

### Configuration

**Hydra Configuration** (in `config/config.yaml`):
```yaml
hydra:
  admin_url: "http://localhost:4445"
  public_url: "http://localhost:4444"
```

**Environment Variables**:
```bash
HYDRA_ADMIN_URL=http://localhost:4445
HYDRA_PUBLIC_URL=http://localhost:4444
```

---

## 🔄 Flow Comparison

| Feature | Direct JWT Flow | OAuth2/OIDC Flow |
|---------|----------------|------------------|
| **Complexity** | Simple | Complex |
| **Use Case** | API, Mobile, SPA | Enterprise, Third-party |
| **Token Issuance** | IAM API directly | Via Hydra |
| **Standards** | Custom JWT | OAuth2/OIDC compliant |
| **Implementation** | ✅ Complete | ⚠️ Partial |
| **Testing** | ✅ Ready | ⚠️ Needs testing |
| **Documentation** | ✅ Complete | ⚠️ Needs docs |

---

## 🚀 Quick Start Examples

### Direct JWT Flow (Recommended)

```bash
# 1. Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: 6e3a1985-61e6-4446-9b42-d4d0c39dad7a" \
  -d '{
    "username": "john.doe",
    "password": "SecurePassword123!",
    "remember_me": false
  }'

# 2. Use Access Token
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer <access_token>" \
  -H "X-Tenant-ID: 6e3a1985-61e6-4446-9b42-d4d0c39dad7a"

# 3. Refresh Token
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "<refresh_token>"
  }'

# 4. Revoke Token
curl -X POST http://localhost:8080/api/v1/auth/revoke \
  -H "Content-Type: application/json" \
  -d '{
    "token": "<refresh_token>",
    "token_type_hint": "refresh_token"
  }'
```

### OAuth2/OIDC Flow (Requires Hydra)

1. **Start Hydra** (if not running)
2. **Configure OAuth2 Client** in Hydra
3. **Initiate Authorization Flow** from client
4. **IAM handles login** via `login_challenge`
5. **Hydra issues tokens**

---

## ✅ Current Status

### Direct JWT Flow: ✅ **PRODUCTION READY**

- Fully implemented
- Fully tested
- Complete documentation
- Ready for use

### OAuth2/OIDC Flow: ⚠️ **PARTIALLY READY**

- Hydra integration exists
- Login challenge handling works
- Needs:
  - End-to-end testing
  - Complete documentation
  - Consent flow customization (optional)
  - JWKS endpoint (optional)

---

## 🎯 Recommendation

**For Most Use Cases**: Use **Direct JWT Flow** ✅

**For Enterprise/OAuth2 Requirements**: Use **OAuth2/OIDC Flow** ⚠️ (with Hydra)

Both flows can coexist - the system automatically detects which flow to use based on the presence of `login_challenge` parameter.

---

## 📚 Additional Resources

- `docs/security/authentication-flow-recommendation.md` - Detailed security architecture
- `docs/security/implementation-plan.md` - Implementation details
- `TESTING_GUIDE.md` - Testing instructions

