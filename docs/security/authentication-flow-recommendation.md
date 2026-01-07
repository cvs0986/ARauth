# 🔐 Secure Authentication Flow Recommendation

## Executive Summary

For a production-ready IAM solution, we recommend a **Hybrid Approach** that supports both:
1. **OAuth2/OIDC Flow** (via ORY Hydra) - For enterprise integrations
2. **Direct JWT Flow** - For simplified API authentication

Both flows share the same security principles and token structure.

---

## 🎯 Recommended Approach: Hybrid Authentication

### Why Hybrid?

1. **Flexibility**: Support different client types (web apps, mobile, APIs)
2. **Standards Compliance**: OAuth2/OIDC for enterprise, direct JWT for simplicity
3. **Security**: Same security standards for both flows
4. **Scalability**: Can evolve from simple to complex as needed

---

## 🔒 Security Principles

### 1. Token Types & Lifetimes

| Token Type | Lifetime | Format | Storage | Purpose |
|------------|----------|--------|---------|---------|
| **Access Token** | 15 minutes | JWT (RS256) | Client | API authorization |
| **Refresh Token** | 30 days | Opaque (UUID) | Database + Redis | Token refresh |
| **ID Token** | 1 hour | JWT (RS256) | Client | User identity (OIDC) |

### 2. Token Signing

**Algorithm**: RS256 (RSA with SHA-256)
- **Why RS256 over HS256?**
  - Public/private key pair allows key rotation without sharing secrets
  - JWKS endpoint enables automatic key discovery
  - Better for distributed systems
  - Industry standard for OAuth2/OIDC

**Key Management**:
- Private key: Stored securely (env var, secret manager, or file with restricted permissions)
- Public key: Exposed via JWKS endpoint (`/.well-known/jwks.json`)
- Key rotation: Every 90 days with overlap period

### 3. Token Structure

```json
{
  "sub": "user-uuid",
  "tenant_id": "tenant-uuid",
  "email": "user@example.com",
  "username": "john.doe",
  "roles": ["admin", "user"],
  "permissions": ["user.read", "user.write"],
  "iss": "https://iam.example.com",
  "aud": "client-id",
  "exp": 1234567890,
  "iat": 1234567890,
  "jti": "token-id-uuid",
  "scope": "role:admin role:user perm:user.read perm:user.write"
}
```

### 4. Refresh Token Security

- **Storage**: Database (for persistence) + Redis (for revocation)
- **Rotation**: New token issued on each refresh, old token invalidated
- **Revocation**: Tokens can be blacklisted in Redis
- **Format**: Opaque UUID (not JWT) to prevent tampering

---

## 🏗️ Implementation Architecture

### Flow 1: Direct JWT Flow (Simplified)

```
┌─────────┐                    ┌──────────┐
│ Client  │                    │ IAM API  │
└────┬────┘                    └────┬─────┘
     │                               │
     │ 1. POST /auth/login           │
     │    {username, password}       │
     ├───────────────────────────────>│
     │                               │
     │ 2. Validate credentials       │
     │    Build claims               │
     │    Sign JWT                   │
     │    Generate refresh token     │
     │                               │
     │ 3. Response                   │
     │    {access_token, refresh,    │
     │     expires_in, token_type}   │
     │<───────────────────────────────┤
     │                               │
     │ 4. Use access_token in        │
     │    Authorization header       │
     │                               │
```

**Use Cases**:
- API-to-API authentication
- Mobile apps
- Single-page applications (SPAs)
- Internal services

### Flow 2: OAuth2/OIDC Flow (via Hydra)

```
┌─────────┐    ┌──────────┐    ┌──────────┐
│ Client  │    │ Hydra    │    │ IAM API  │
└────┬────┘    └────┬─────┘    └────┬─────┘
     │              │               │
     │ 1. GET /oauth2/auth          │
     ├──────────────>│               │
     │              │               │
     │ 2. Login Challenge           │
     │<──────────────┼───────────────┤
     │              │               │
     │ 3. POST /auth/login          │
     │    {login_challenge, ...}    │
     ├──────────────────────────────>│
     │              │               │
     │ 4. Accept Login              │
     │              │<───────────────┤
     │              │               │
     │ 5. Authorization Code        │
     │<──────────────┤               │
     │              │               │
     │ 6. Exchange code for tokens  │
     ├──────────────>│               │
     │              │               │
     │ 7. Tokens                    │
     │<──────────────┤               │
```

**Use Cases**:
- Enterprise integrations
- Third-party applications
- OAuth2/OIDC compliance required
- Multi-tenant SaaS platforms

---

## 🛡️ Security Features

### 1. Token Validation

Every request with an access token must validate:
- ✅ Signature verification (RS256)
- ✅ Expiration check (`exp` claim)
- ✅ Issuer validation (`iss` claim)
- ✅ Audience validation (`aud` claim) - optional but recommended
- ✅ JTI blacklist check (for revoked tokens)
- ✅ Token not before (`nbf` claim) - if used

### 2. Refresh Token Rotation

```go
// On refresh:
1. Validate refresh token exists and not revoked
2. Invalidate old refresh token (add to blacklist)
3. Generate new refresh token
4. Issue new access token
5. Return both new tokens
```

### 3. Token Revocation

- **Access Tokens**: Short-lived, revocation via blacklist (Redis)
- **Refresh Tokens**: Long-lived, revocation via database + Redis
- **Logout**: Revoke both tokens immediately

### 4. Rate Limiting

- Login attempts: 5 per minute per IP
- Token refresh: 10 per minute per user
- API requests: 100 per minute per access token

### 5. MFA Integration

- MFA required users: Return `mfa_required: true`
- After MFA verification: Issue tokens with `acr: "mfa"` claim
- Higher security level indicated in token

---

## 📋 Implementation Plan

### Phase 1: Direct JWT Issuance (Current Priority)

1. ✅ Create JWT token service
   - RS256 signing with configurable key
   - Token generation with claims
   - Token validation middleware

2. ✅ Implement refresh token storage
   - Database table for refresh tokens
   - Redis for revocation blacklist

3. ✅ Update login service
   - Issue access + refresh tokens
   - Store refresh token
   - Return tokens in response

4. ✅ Create token refresh endpoint
   - Validate refresh token
   - Rotate tokens
   - Return new tokens

5. ✅ Create token revocation endpoint
   - Blacklist access token
   - Revoke refresh token
   - Support logout

### Phase 2: Enhanced Security

1. JWKS endpoint for key discovery
2. Key rotation mechanism
3. Token introspection endpoint
4. Audit logging for token operations

### Phase 3: OAuth2/OIDC (Already Supported)

1. ✅ Hydra integration exists
2. ✅ Login challenge flow
3. ✅ Claims injection
4. ⚠️ Needs testing and documentation

---

## 🔧 Configuration

### JWT Configuration

```yaml
security:
  jwt:
    issuer: "https://iam.example.com"
    access_token_ttl: 15m
    refresh_token_ttl: 30d
    id_token_ttl: 1h
    signing_key_path: "/etc/iam/jwt.key"  # RSA private key
    secret: "${JWT_SECRET}"  # Fallback for HS256 (not recommended)
```

### Key Generation

```bash
# Generate RSA key pair
openssl genrsa -out jwt_private.pem 2048
openssl rsa -in jwt_private.pem -pubout -out jwt_public.pem

# Set permissions
chmod 600 jwt_private.pem
chmod 644 jwt_public.pem
```

---

## ✅ Security Checklist

- [x] RS256 signing algorithm
- [x] Short-lived access tokens (15 min)
- [x] Long-lived refresh tokens (30 days)
- [x] Refresh token rotation
- [x] Token revocation support
- [x] Rate limiting
- [x] MFA integration
- [ ] JWKS endpoint
- [ ] Key rotation
- [ ] Token introspection
- [ ] Audit logging

---

## 🚀 Next Steps

1. **Immediate**: Implement direct JWT issuance (Phase 1)
2. **Short-term**: Add JWKS endpoint and key rotation
3. **Long-term**: Enhance OAuth2/OIDC flow documentation and testing

---

## 📚 References

- [OAuth 2.0 Security Best Practices](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-security-topics)
- [JWT Best Practices](https://datatracker.ietf.org/doc/html/rfc8725)
- [ORY Hydra Documentation](https://www.ory.sh/docs/hydra/)

