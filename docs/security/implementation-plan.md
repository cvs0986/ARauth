# 🔐 Secure Authentication Implementation Plan

## Overview

This document outlines the step-by-step implementation of secure JWT token issuance for the Nuage Identity IAM system.

---

## Phase 1: Direct JWT Token Issuance

### Step 1: Create JWT Token Service

**File**: `auth/token/service.go`

**Responsibilities**:
- Generate access tokens (JWT, RS256)
- Generate refresh tokens (opaque UUID)
- Validate tokens
- Extract claims from tokens

**Dependencies**:
- `github.com/golang-jwt/jwt/v5` (already in go.mod)
- RSA private key from config
- Claims builder (already exists)

### Step 2: Create Refresh Token Storage

**Database Migration**: `migrations/XXXXX_create_refresh_tokens.up.sql`

```sql
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
```

**Repository**: `storage/postgres/refresh_token_repository.go`

### Step 3: Update Login Service

**File**: `auth/login/service.go`

**Changes**:
- After successful authentication, call token service
- Generate access token and refresh token
- Store refresh token in database
- Return tokens in response

### Step 4: Create Token Refresh Endpoint

**Handler**: `api/handlers/auth_handler.go` - Add `RefreshToken` method
**Route**: `POST /api/v1/auth/refresh`

**Flow**:
1. Validate refresh token (exists, not expired, not revoked)
2. Get user from token
3. Rotate tokens (invalidate old, create new)
4. Return new tokens

### Step 5: Create Token Revocation Endpoint

**Handler**: `api/handlers/auth_handler.go` - Add `RevokeToken` method
**Route**: `POST /api/v1/auth/revoke`

**Flow**:
1. Accept access token or refresh token
2. Add to blacklist (Redis)
3. If refresh token, mark as revoked in database
4. Return success

### Step 6: Create JWT Validation Middleware

**File**: `api/middleware/auth.go`

**Responsibilities**:
- Extract token from Authorization header
- Validate token signature
- Validate token claims (exp, iss, aud)
- Check token blacklist
- Set user context in Gin context

---

## Phase 2: Enhanced Security Features

### Step 7: JWKS Endpoint

**Route**: `GET /.well-known/jwks.json`

**Response**:
```json
{
  "keys": [
    {
      "kty": "RSA",
      "kid": "key-id-1",
      "use": "sig",
      "alg": "RS256",
      "n": "...",
      "e": "AQAB"
    }
  ]
}
```

### Step 8: Key Rotation

- Support multiple keys during rotation
- JWKS endpoint returns all active keys
- New tokens use new key
- Old tokens validated with old key until expiry

### Step 9: Token Introspection

**Route**: `POST /api/v1/auth/introspect`

**Use Case**: For resource servers to validate tokens

---

## Implementation Order

1. ✅ **JWT Token Service** - Core token generation
2. ✅ **Refresh Token Storage** - Database + repository
3. ✅ **Update Login Service** - Issue tokens on login
4. ✅ **Token Refresh Endpoint** - Allow token renewal
5. ✅ **Token Revocation** - Support logout
6. ✅ **JWT Middleware** - Protect API endpoints
7. ⏳ **JWKS Endpoint** - Key discovery
8. ⏳ **Key Rotation** - Long-term security

---

## Testing Strategy

### Unit Tests
- Token generation
- Token validation
- Claims extraction
- Refresh token rotation

### Integration Tests
- Login flow with tokens
- Token refresh flow
- Token revocation
- Protected endpoint access

### Security Tests
- Token tampering attempts
- Expired token usage
- Revoked token usage
- Invalid signature handling

---

## Security Considerations

1. **Private Key Storage**: Use environment variables or secret manager
2. **Key Rotation**: Plan for 90-day rotation cycle
3. **Token Blacklist**: Use Redis with TTL matching token expiry
4. **Rate Limiting**: Apply to all auth endpoints
5. **Audit Logging**: Log all token operations
6. **HTTPS Only**: Enforce in production

---

## Timeline Estimate

- **Phase 1**: 2-3 days (core functionality)
- **Phase 2**: 1-2 days (enhanced features)
- **Testing**: 1 day
- **Documentation**: 0.5 day

**Total**: ~5-6 days for complete implementation

