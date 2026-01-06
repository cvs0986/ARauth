# Security Documentation

This document describes the security architecture, implementation, and best practices for Nuage Identity.

## 🎯 Security Principles

1. **Defense in Depth**: Multiple layers of security
2. **Least Privilege**: Minimum required permissions
3. **Secure by Default**: Security built-in, not bolted on
4. **Zero Trust**: Verify everything
5. **Fail Secure**: Fail closed, not open

## 🔐 Authentication Security

### Password Security

#### Argon2id Hashing

**Algorithm**: Argon2id (winner of Password Hashing Competition)

**Parameters**:
```go
const (
    memory      = 64 * 1024  // 64 MB
    iterations  = 3
    parallelism = 4
    saltLength  = 16
    keyLength   = 32
)
```

**Rationale**:
- Memory-hard function (resistant to GPU attacks)
- Time-memory trade-off protection
- Industry standard

**Implementation**:
```go
import "golang.org/x/crypto/argon2"

func HashPassword(password string) (string, error) {
    salt := make([]byte, saltLength)
    if _, err := rand.Read(salt); err != nil {
        return "", err
    }
    
    hash := argon2.IDKey(
        []byte(password),
        salt,
        iterations,
        memory,
        parallelism,
        keyLength,
    )
    
    // Encode: $argon2id$v=19$m=65536,t=3,p=4$salt$hash
    return encodeHash(salt, hash), nil
}
```

#### Password Policies

**Requirements**:
- Minimum length: 12 characters
- Complexity: At least one uppercase, lowercase, number, special character
- No common passwords (check against common password list)
- No username in password
- Password history: Prevent reuse of last 5 passwords

**Validation**:
```go
func ValidatePassword(password string, username string) error {
    if len(password) < 12 {
        return errors.New("password too short")
    }
    
    if strings.Contains(strings.ToLower(password), strings.ToLower(username)) {
        return errors.New("password cannot contain username")
    }
    
    // Check complexity
    // Check against common passwords
    
    return nil
}
```

### Multi-Factor Authentication (MFA)

#### TOTP Implementation

**Algorithm**: TOTP (RFC 6238)

**Parameters**:
- Time step: 30 seconds
- Hash algorithm: SHA1
- Digits: 6

**Secret Generation**:
```go
import "github.com/pquerna/otp"

key, err := totp.Generate(totp.GenerateOpts{
    Issuer:      "Nuage Identity",
    AccountName: user.Email,
    Period:      30,
    Digits:      otp.DigitsSix,
    Algorithm:   otp.AlgorithmSHA1,
})
```

**Validation**:
```go
func ValidateTOTP(secret string, code string) bool {
    return totp.Validate(code, secret)
}
```

**Security**:
- Secret stored encrypted in database
- Rate limiting: 5 attempts per 5 minutes
- Lockout after 10 failed attempts

#### Recovery Codes

**Generation**:
- 10 single-use codes
- 16 characters each
- Stored hashed in database

**Usage**:
- One-time use
- Invalidated after use
- Can regenerate (invalidates old codes)

### Rate Limiting

#### Implementation

**Strategy**: Sliding window with Redis

**Limits**:
- Login: 5 attempts per minute per IP
- MFA: 5 attempts per 5 minutes per user
- Token refresh: 10 requests per minute per token
- API calls: 100 requests per minute per client

**Implementation**:
```go
func (rl *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
    now := time.Now()
    windowStart := now.Add(-window)
    
    // Remove old entries
    rl.redis.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart.Unix(), 10))
    
    // Count current requests
    count, err := rl.redis.ZCard(ctx, key).Result()
    if err != nil {
        return false, err
    }
    
    if count >= int64(limit) {
        return false, nil
    }
    
    // Add current request
    rl.redis.ZAdd(ctx, key, &redis.Z{
        Score:  float64(now.Unix()),
        Member: strconv.FormatInt(now.UnixNano(), 10),
    })
    rl.redis.Expire(ctx, key, window)
    
    return true, nil
}
```

## 🎫 Token Security

### JWT Access Tokens

#### Token Structure

```json
{
  "sub": "user-123",
  "tenant": "tenant-123",
  "roles": ["admin", "user"],
  "permissions": ["user.read", "user.write"],
  "acr": "mfa",
  "iss": "https://iam.example.com",
  "aud": "client-123",
  "exp": 1234567890,
  "iat": 1234567890,
  "jti": "token-id-123"
}
```

#### Security Features

**Signing Algorithm**: RS256 (RSA with SHA-256)

**Key Rotation**:
- JWKS endpoint for key discovery
- Key rotation every 90 days
- Support for multiple keys during rotation

**Token Lifetime**:
- Access token: 15 minutes
- ID token: 1 hour
- Refresh token: 30 days

**Token Validation**:
1. Signature verification (JWKS)
2. Expiration check
3. Issuer validation
4. Audience validation
5. JTI blacklist check (optional)

### Refresh Tokens

#### Security

**Storage**: Opaque tokens in Hydra database

**Rotation**: Enabled
- New refresh token on each refresh
- Old token invalidated

**Blacklist**: Redis
- Invalidated tokens added to blacklist
- TTL: Token expiry time

**Validation**:
```go
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
    // Check blacklist
    if s.blacklist.Exists(ctx, refreshToken) {
        return nil, errors.New("token revoked")
    }
    
    // Validate with Hydra
    // Rotate token
    // Add old token to blacklist
    
    return tokenResponse, nil
}
```

## 🔒 Data Security

### Encryption at Rest

**Database**:
- PostgreSQL encryption (TDE or filesystem encryption)
- Sensitive fields encrypted:
  - Password hashes (already hashed)
  - TOTP secrets (encrypted)
  - Recovery codes (hashed)

**Encryption Algorithm**: AES-256-GCM

**Key Management**:
- Keys stored in secure key management system
- Key rotation every 90 days
- Separate keys per tenant (optional)

### Encryption in Transit

**TLS**:
- TLS 1.3 minimum
- Strong cipher suites only
- Certificate pinning (optional)

**Configuration**:
```go
tlsConfig := &tls.Config{
    MinVersion: tls.VersionTLS13,
    CipherSuites: []uint16{
        tls.TLS_AES_256_GCM_SHA384,
        tls.TLS_CHACHA20_POLY1305_SHA256,
        tls.TLS_AES_128_GCM_SHA256,
    },
}
```

### Data Protection

**PII Handling**:
- Minimize PII in logs
- Encrypt sensitive data
- Access controls

**Data Retention**:
- Audit logs: 1 year
- User data: Per retention policy
- Token data: Per token expiry

## 🛡️ API Security

### Input Validation

**All Inputs Validated**:
- Request body validation
- Query parameter validation
- Path parameter validation
- Header validation

**Tools**:
```go
import "github.com/go-playground/validator/v10"

type LoginRequest struct {
    Username string `json:"username" validate:"required,email"`
    Password string `json:"password" validate:"required,min=12"`
    TenantID string `json:"tenant_id" validate:"required,uuid"`
}
```

### SQL Injection Prevention

**Parameterized Queries**:
```go
// Good
query := "SELECT * FROM users WHERE id = $1"
db.Query(query, userID)

// Bad
query := fmt.Sprintf("SELECT * FROM users WHERE id = '%s'", userID)
```

**ORM/Query Builder**: Use parameterized queries only

### XSS Prevention

**No HTML Output**: API returns JSON only, no HTML rendering

**Content-Type**: Always `application/json`

### CSRF Protection

**Not Required**: Stateless API, no sessions

**Alternative**: Use SameSite cookies if cookies are used

## 🔍 Security Monitoring

### Logging

**Security Events Logged**:
- Failed login attempts
- MFA failures
- Token validation failures
- Permission denials
- Rate limit violations
- Suspicious activity

**Log Format**:
```json
{
  "timestamp": "2024-01-01T00:00:00Z",
  "level": "warn",
  "event": "failed_login",
  "username": "user@example.com",
  "ip": "192.168.1.1",
  "user_agent": "Mozilla/5.0...",
  "request_id": "req-123"
}
```

### Alerting

**Alerts Triggered For**:
- Multiple failed logins from same IP
- Brute force attempts
- Unusual access patterns
- Token validation failures
- System errors

### Audit Logging

**Audited Actions**:
- User creation/deletion
- Role assignments
- Permission changes
- Tenant creation
- MFA enrollment/disabling
- Password changes

**Audit Log Format**:
```json
{
  "timestamp": "2024-01-01T00:00:00Z",
  "actor": "user-123",
  "action": "user.created",
  "target": "user-456",
  "tenant": "tenant-123",
  "ip": "192.168.1.1",
  "metadata": {}
}
```

## 🔐 Secrets Management

### Configuration

**Secrets Stored In**:
- Environment variables (development)
- Secret management system (production)
  - Kubernetes Secrets
  - HashiCorp Vault
  - AWS Secrets Manager

**Never Commit**:
- Database passwords
- API keys
- JWT signing keys
- Encryption keys

### Key Rotation

**Rotation Schedule**:
- JWT signing keys: Every 90 days
- Database passwords: Every 180 days
- Encryption keys: Every 90 days

**Rotation Process**:
1. Generate new key
2. Update configuration
3. Support both keys during transition
4. Remove old key after transition period

## 🚨 Incident Response

### Security Incidents

**Response Process**:
1. Detect and assess
2. Contain incident
3. Eradicate threat
4. Recover systems
5. Post-incident review

### Breach Procedures

**If Breach Detected**:
1. Immediately revoke affected tokens
2. Force password reset for affected users
3. Notify affected users
4. Investigate root cause
5. Implement fixes
6. Document incident

## 📋 Security Checklist

### Development

- [ ] All inputs validated
- [ ] SQL injection prevented
- [ ] XSS prevented
- [ ] CSRF protection (if needed)
- [ ] Secrets not in code
- [ ] Dependencies scanned
- [ ] Security tests written

### Deployment

- [ ] TLS enabled
- [ ] Strong cipher suites
- [ ] Secrets in secure storage
- [ ] Firewall rules configured
- [ ] Rate limiting enabled
- [ ] Monitoring configured
- [ ] Backup strategy

### Operations

- [ ] Regular security audits
- [ ] Dependency updates
- [ ] Key rotation schedule
- [ ] Incident response plan
- [ ] Security training

## 📚 Related Documentation

- [Architecture Overview](../architecture/overview.md)
- [API Design](./api-design.md)
- [Deployment Guide](../deployment/kubernetes.md)

