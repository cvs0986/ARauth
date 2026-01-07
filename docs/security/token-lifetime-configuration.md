# 🔧 Token Lifetime Configuration System

## Overview

Token lifetimes are configurable through multiple sources with a priority hierarchy. This allows flexibility for different deployment scenarios.

---

## 📊 Configuration Priority (Highest to Lowest)

1. **Per-Request** (Remember Me checkbox)
2. **Per-Tenant Settings** (Database)
3. **Environment Variables**
4. **Config File** (config.yaml)
5. **System Defaults**

---

## 🏗️ Architecture

### Configuration Sources

```
┌─────────────────────────────────────────┐
│  Per-Request (Remember Me)             │  ← Highest Priority
│  - Extends refresh token lifetime      │
│  - Optional access token extension     │
└─────────────────────────────────────────┘
           ↓
┌─────────────────────────────────────────┐
│  Per-Tenant Settings (Database)         │
│  - Tenant-specific token lifetimes     │
│  - Managed via Admin Dashboard          │
└─────────────────────────────────────────┘
           ↓
┌─────────────────────────────────────────┐
│  Environment Variables                  │
│  - JWT_ACCESS_TOKEN_TTL                │
│  - JWT_REFRESH_TOKEN_TTL               │
│  - JWT_ID_TOKEN_TTL                    │
└─────────────────────────────────────────┘
           ↓
┌─────────────────────────────────────────┐
│  Config File (config.yaml)              │
│  - Default system-wide settings         │
└─────────────────────────────────────────┘
           ↓
┌─────────────────────────────────────────┐
│  System Defaults                        │  ← Lowest Priority
│  - Access: 15m                          │
│  - Refresh: 30d                        │
│  - ID: 1h                              │
└─────────────────────────────────────────┘
```

---

## 💾 Database Schema

### Tenant Settings Table

```sql
CREATE TABLE tenant_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Token Lifetime Settings
    access_token_ttl_minutes INT NOT NULL DEFAULT 15,
    refresh_token_ttl_days INT NOT NULL DEFAULT 30,
    id_token_ttl_minutes INT NOT NULL DEFAULT 60,
    
    -- Remember Me Settings
    remember_me_enabled BOOLEAN NOT NULL DEFAULT true,
    remember_me_refresh_token_ttl_days INT NOT NULL DEFAULT 90,
    remember_me_access_token_ttl_minutes INT NOT NULL DEFAULT 60,
    
    -- Security Settings
    token_rotation_enabled BOOLEAN NOT NULL DEFAULT true,
    require_mfa_for_extended_sessions BOOLEAN NOT NULL DEFAULT false,
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    UNIQUE(tenant_id)
);

CREATE INDEX idx_tenant_settings_tenant_id ON tenant_settings(tenant_id);
```

---

## 🔧 Configuration Structure

### Config File (config.yaml)

```yaml
security:
  jwt:
    issuer: "https://iam.example.com"
    
    # Default token lifetimes
    access_token_ttl: 15m
    refresh_token_ttl: 30d
    id_token_ttl: 1h
    
    # Remember Me settings
    remember_me:
      enabled: true
      refresh_token_ttl: 90d
      access_token_ttl: 60m
    
    # Token settings
    token_rotation: true
    require_mfa_for_extended_sessions: false
```

### Environment Variables

```bash
# Token Lifetimes
JWT_ACCESS_TOKEN_TTL=15m
JWT_REFRESH_TOKEN_TTL=30d
JWT_ID_TOKEN_TTL=1h

# Remember Me
JWT_REMEMBER_ME_REFRESH_TTL=90d
JWT_REMEMBER_ME_ACCESS_TTL=60m
JWT_REMEMBER_ME_ENABLED=true
```

### Per-Tenant Settings (Database)

Stored in `tenant_settings` table, managed via Admin Dashboard.

---

## 🎯 Remember Me Functionality

### Behavior

When "Remember Me" is checked:
- **Refresh Token**: Extended lifetime (e.g., 90 days instead of 30)
- **Access Token**: Optionally extended (e.g., 60 minutes instead of 15)
- **Session**: Persists across browser restarts

When "Remember Me" is NOT checked:
- **Refresh Token**: Standard lifetime (e.g., 30 days)
- **Access Token**: Standard lifetime (e.g., 15 minutes)
- **Session**: May expire on browser close (depends on storage)

### Security Considerations

1. **MFA Requirement**: Option to require MFA for extended sessions
2. **Audit Logging**: Log all "Remember Me" logins
3. **Revocation**: Allow admins to revoke extended sessions
4. **Rate Limiting**: Stricter limits for extended sessions

---

## 🖥️ Admin Dashboard UI

### Settings Page: Token Configuration

**Location**: `/settings/security/tokens`

**Features**:
- Token lifetime configuration (per tenant)
- Remember Me settings
- Token rotation toggle
- MFA requirements
- Save/Reset buttons

**UI Components**:
```
┌─────────────────────────────────────────┐
│  Token Lifetime Settings               │
├─────────────────────────────────────────┤
│                                         │
│  Access Token Lifetime:                │
│  [15] minutes                          │
│                                         │
│  Refresh Token Lifetime:               │
│  [30] days                             │
│                                         │
│  ID Token Lifetime:                    │
│  [60] minutes                          │
│                                         │
│  ┌─────────────────────────────────┐  │
│  │ Remember Me Settings            │  │
│  ├─────────────────────────────────┤  │
│  │ ☑ Enable Remember Me            │  │
│  │                                 │  │
│  │ Extended Refresh Token:         │  │
│  │ [90] days                       │  │
│  │                                 │  │
│  │ Extended Access Token:         │  │
│  │ [60] minutes                    │  │
│  └─────────────────────────────────┘  │
│                                         │
│  [Save Changes] [Reset to Defaults]    │
└─────────────────────────────────────────┘
```

---

## 🔄 Login Flow with Remember Me

### Request

```json
POST /api/v1/auth/login
{
  "username": "user@example.com",
  "password": "password123",
  "remember_me": true  // ← New field
}
```

### Response

```json
{
  "access_token": "eyJ...",
  "refresh_token": "uuid-v4",
  "token_type": "Bearer",
  "expires_in": 3600,  // Extended if remember_me=true
  "refresh_expires_in": 7776000,  // Extended if remember_me=true
  "remember_me": true
}
```

---

## 📝 Implementation Details

### Token Lifetime Resolver

```go
type TokenLifetimeResolver struct {
    config      *config.SecurityConfig
    tenantRepo  interfaces.TenantRepository
}

func (r *TokenLifetimeResolver) GetAccessTokenTTL(
    ctx context.Context,
    tenantID uuid.UUID,
    rememberMe bool,
) time.Duration {
    // 1. Check per-tenant settings
    if settings := r.getTenantSettings(ctx, tenantID); settings != nil {
        if rememberMe && settings.RememberMeEnabled {
            return time.Duration(settings.RememberMeAccessTokenTTLMinutes) * time.Minute
        }
        return time.Duration(settings.AccessTokenTTLMinutes) * time.Minute
    }
    
    // 2. Check environment variables
    if envTTL := os.Getenv("JWT_ACCESS_TOKEN_TTL"); envTTL != "" {
        if ttl, err := time.ParseDuration(envTTL); err == nil {
            return ttl
        }
    }
    
    // 3. Check config file
    if r.config.JWT.AccessTokenTTL > 0 {
        baseTTL := r.config.JWT.AccessTokenTTL
        if rememberMe {
            return r.config.JWT.RememberMe.AccessTokenTTL
        }
        return baseTTL
    }
    
    // 4. System default
    if rememberMe {
        return 60 * time.Minute
    }
    return 15 * time.Minute
}
```

---

## ✅ Benefits

1. **Flexibility**: Different lifetimes for different tenants
2. **Security**: Shorter lifetimes for sensitive tenants
3. **User Experience**: Remember Me for convenience
4. **Compliance**: Meet different regulatory requirements
5. **Operational**: Easy to adjust without code changes

---

## 🔒 Security Best Practices

1. **Minimum Lifetimes**: Enforce minimums (e.g., access token ≥ 5 min)
2. **Maximum Lifetimes**: Enforce maximums (e.g., refresh token ≤ 90 days)
3. **Validation**: Validate all lifetime values on save
4. **Audit**: Log all lifetime configuration changes
5. **MFA**: Require MFA for extended sessions (optional)

---

## 📋 Migration Path

1. **Phase 1**: Add database schema for tenant settings
2. **Phase 2**: Implement configuration resolver
3. **Phase 3**: Update token service to use resolver
4. **Phase 4**: Add Remember Me to login
5. **Phase 5**: Create Admin Dashboard UI
6. **Phase 6**: Add validation and security checks

---

## 🚀 Next Steps

1. Create database migration for `tenant_settings`
2. Implement configuration resolver
3. Update token service
4. Add Remember Me to login flow
5. Create Admin Dashboard UI

