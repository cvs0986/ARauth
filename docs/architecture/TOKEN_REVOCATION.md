# Token Revocation & Blacklisting Design
**Status:** IMPLEMENTED
**Owner:** @cvs0986

## Problem
Currently, access tokens are stateless and cannot be revoked before expiry. If a user is blocked or a token is compromised, they retain access until the JWT expires (default 15m). Security requirements mandate immediate revocation capability.

## Solution
Implement a **Redis-based Token Blacklist**.

## Architecture

### 1. Storage (Redis)
- **Key Pattern**: `blacklist:token:<jti>`
- **Value**: `revoked_at` (timestamp) or metadata JSON.
- **TTL**: Set to `exp` (expiration time) of the JWT. This ensures the blacklist doesn't grow indefinitely; once the token naturally expires, it's removed from Redis.

### 2. Revocation Flow (`/api/v1/auth/revoke`)
1.  Receive `token` (access token).
2.  Validate signature (don't trust claims if invalid).
3.  Extract `jti` (JWT ID) and `exp` (Expiration).
4.  Calculate `TTL = exp - now`.
5.  Set `blacklist:token:<jti>` in Redis with TTL.
6.  Emit audit event `token.revoked`.

### 3. Verification Flow (Middleware)
1.  **JWTAuthMiddleware**:
    - existing: Verify signature & expiry.
    - **NEW**: Check Redis `EXISTS blacklist:token:<jti>`.
    - If exists -> Reject (401 Unauthorized).

### 4. Components Involved
- **Redis Client**: `internal/database/redis/client.go` (Existing).
- **Token Service**: `auth/token/service.go`.
    - New method: `RevokeAccessToken(ctx, tokenString)`.
    - New method: `IsAccessTokenRevoked(ctx, jti)`.
- **Middleware**: `api/middleware/jwt_auth.go`.
    - Inject `TokenService` (or new `BlacklistService`).

## Implementation Plan

### Phase 1: Service Layer
- Define `BlacklistRepository` interface (Redis impl).
- Update `TokenService` to include revocation logic.

### Phase 2: Middleware Integration
- Update `JWTAuthMiddleware` to check blacklist.
- Update `NewJWTAuthMiddleware` signature.

### Phase 3: API Endpoint
- Update `RevokeToken` handler to call `TokenService.RevokeAccessToken`.
- Currently it lists a TODO.

## Performance
- Redis lookup is O(1) and very fast (<1ms).
- Adds one network hop per API request.
- **Mitigation**: Use connection pooling and localized Redis.

## Security
- `jti` claim MUST be present in all JWTs.
- Redis availability is critical. If Redis is down, we can:
    - Fail open (security risk).
    - **Fail closed (availability risk) - PREFERRED for high security.**
