# Tenant Settings Management

## Overview

SYSTEM users (Master Admin) can manage tenant-specific configurations and settings for any tenant. This includes:

- Token lifetimes (Access Token, Refresh Token, ID Token)
- Remember Me settings
- Token rotation
- MFA requirements for extended sessions
- Password policies (future)
- MFA policies (future)

## API Endpoints

### Get Tenant Settings
**Endpoint**: `GET /system/tenants/:id/settings`  
**Access**: SYSTEM users only  
**Permission**: `tenant:configure`

**Response**:
```json
{
  "id": "uuid",
  "tenant_id": "uuid",
  "access_token_ttl_minutes": 15,
  "refresh_token_ttl_days": 30,
  "id_token_ttl_minutes": 60,
  "remember_me_enabled": true,
  "remember_me_refresh_token_ttl_days": 90,
  "remember_me_access_token_ttl_minutes": 60,
  "token_rotation_enabled": true,
  "require_mfa_for_extended_sessions": false
}
```

**Note**: If settings don't exist, returns a message indicating they need to be created.

### Update Tenant Settings
**Endpoint**: `PUT /system/tenants/:id/settings`  
**Access**: SYSTEM users only  
**Permission**: `tenant:configure`

**Request Body** (all fields optional):
```json
{
  "access_token_ttl_minutes": 30,
  "refresh_token_ttl_days": 60,
  "id_token_ttl_minutes": 120,
  "remember_me_enabled": true,
  "remember_me_refresh_token_ttl_days": 180,
  "remember_me_access_token_ttl_minutes": 120,
  "token_rotation_enabled": true,
  "require_mfa_for_extended_sessions": true
}
```

**Behavior**:
- If settings don't exist, creates them with provided values (uses defaults for omitted fields)
- If settings exist, updates only the provided fields
- Validates constraints (e.g., TTL ranges)

## Use Cases

### 1. System Admin Configuring New Tenant
1. SYSTEM admin creates a new tenant
2. SYSTEM admin configures tenant settings (token TTLs, security policies)
3. Tenant is ready for use with custom configuration

### 2. System Admin Updating Tenant Configuration
1. SYSTEM admin views tenant settings
2. SYSTEM admin updates specific settings (e.g., increase token TTLs)
3. Changes apply immediately to that tenant

### 3. Tenant Admin Viewing Their Settings
1. TENANT admin views their tenant settings (read-only via `/api/v1/tenants/:id/settings` - future)
2. TENANT admin can request changes from SYSTEM admin
3. Or TENANT admin can update certain settings if permitted (future enhancement)

## Security Considerations

1. **Permission-Based Access**:
   - Only SYSTEM users with `tenant:configure` permission can manage tenant settings
   - TENANT users cannot modify settings via system API

2. **Tenant Isolation**:
   - SYSTEM admin must explicitly specify tenant ID
   - Settings are scoped to specific tenant
   - No cross-tenant data leakage

3. **Validation**:
   - All TTL values are validated against constraints
   - Invalid values are rejected with clear error messages

## Frontend Integration

### Admin Dashboard - SYSTEM User Flow

1. **View Tenant Settings**:
   - Navigate to Tenants → Select Tenant → Settings
   - Shows current tenant configuration
   - Can edit and save changes

2. **Bulk Configuration**:
   - Select multiple tenants
   - Apply same settings to all selected tenants
   - Useful for standardization

3. **Settings Templates**:
   - Save common configurations as templates
   - Apply templates to new tenants
   - Ensures consistency

## Future Enhancements

1. **Password Policies**:
   - Per-tenant password requirements
   - Password expiration policies
   - Password history

2. **MFA Policies**:
   - Per-tenant MFA enforcement
   - MFA methods allowed
   - Recovery code policies

3. **Rate Limiting**:
   - Per-tenant rate limits
   - API throttling policies

4. **Audit Configuration**:
   - Per-tenant audit log retention
   - Audit event types to log

5. **Branding**:
   - Tenant-specific branding
   - Custom login pages
   - Email templates

