# Admin Dashboard Improvements

## Summary

This document addresses three key improvements to the Admin Dashboard:

1. **"All Tenants" Aggregated View** - SYSTEM admin can view aggregated data across all tenants
2. **TENANT User Security Settings** - TENANT users can now configure security settings for their tenant
3. **OAuth2/OIDC Settings** - Clarification on what can be tenant-specific vs system-wide

---

## 1. "All Tenants" Aggregated View

### Problem
When SYSTEM admin selects "All Tenants" in the dropdown, the Dashboard showed no data because tenant-scoped APIs require a tenant context.

### Solution
**Status: Backend Ready, Frontend TODO**

The backend now supports:
- SYSTEM users can access `/system/*` endpoints without tenant context
- When `selectedTenantId` is `null`, the frontend should aggregate data across all tenants

### Implementation Plan

#### Backend (Already Implemented)
- ✅ SYSTEM API endpoints (`/system/tenants`) don't require tenant context
- ✅ Tenant-scoped endpoints can be called with `X-Tenant-ID` header for each tenant

#### Frontend (TODO)
1. **Dashboard.tsx**:
   - When `selectedTenantId === null` and `isSystemUser() === true`:
     - Fetch all tenants
     - For each tenant, fetch users, roles, permissions
     - Aggregate statistics (sum totals, count actives)
     - Display aggregated view

2. **UserList.tsx, RoleList.tsx, PermissionList.tsx**:
   - When "All Tenants" is selected:
     - Fetch data from all tenants
     - Display with tenant column/indicator
     - Aggregate totals

### Example Aggregated Query
```typescript
// Pseudo-code for Dashboard aggregation
if (isSystemUser() && !selectedTenantId) {
  const allTenants = await systemApi.tenants.list();
  const allUsers = await Promise.all(
    allTenants.map(t => userApi.list(t.id))
  );
  const aggregatedUsers = allUsers.flat();
  // Calculate stats from aggregatedUsers
}
```

---

## 2. TENANT User Security Settings

### Problem
TENANT users could not see or configure Security settings (Password policies, MFA, Rate limiting) for their tenant. Only Token settings were available.

### Solution
**Status: ✅ COMPLETED**

### Backend Changes

#### Database Migration
- ✅ Created migration `000017_add_security_settings_to_tenant_settings.up.sql`
- ✅ Added columns to `tenant_settings` table:
  - `min_password_length` (INT, default: 12)
  - `require_uppercase` (BOOLEAN, default: true)
  - `require_lowercase` (BOOLEAN, default: true)
  - `require_numbers` (BOOLEAN, default: true)
  - `require_special_chars` (BOOLEAN, default: true)
  - `password_expiry_days` (INT, nullable - NULL = never expires)
  - `mfa_required` (BOOLEAN, default: false)
  - `rate_limit_requests` (INT, default: 100)
  - `rate_limit_window_seconds` (INT, default: 60)

#### API Endpoints
- ✅ **SYSTEM users**: `GET/PUT /system/tenants/:id/settings` (can manage any tenant)
- ✅ **TENANT users**: `GET/PUT /api/v1/tenant/settings` (can only manage their own tenant)

#### Repository Updates
- ✅ Updated `TenantSettings` struct to include security fields
- ✅ Updated `GetByTenantID`, `Create`, and `Update` methods to handle security settings

### Frontend Changes (TODO)

#### Settings.tsx
1. **Show Security Tab for TENANT Users**:
   ```typescript
   // Currently: Security tab only shown for SYSTEM users
   // Change: Show Security tab for TENANT users too
   {isSystemUser() && (
     <TabsTrigger value="security">Security</TabsTrigger>
   )}
   // Should be:
   <TabsTrigger value="security">Security</TabsTrigger>
   ```

2. **Fetch Tenant Settings for TENANT Users**:
   ```typescript
   // For TENANT users, fetch their own tenant settings
   const { data: tenantSettings } = useQuery({
     queryKey: ['tenant-settings', tenantId],
     queryFn: () => tenantApi.getSettings(tenantId), // NEW endpoint
     enabled: !isSystemUser() && !!tenantId,
   });
   ```

3. **Update Security Form to Use Tenant Settings**:
   - Populate form with `tenantSettings` values
   - Save to `/api/v1/tenant/settings` endpoint

#### API Service Updates
```typescript
// frontend/admin-dashboard/src/services/api.ts
export const tenantApi = {
  // ... existing methods ...
  
  getSettings: async (tenantId: string): Promise<TenantSettings> => {
    const response = await apiClient.get(`/api/v1/tenant/settings`);
    return response.data;
  },
  
  updateSettings: async (data: Partial<TenantSettings>): Promise<TenantSettings> => {
    const response = await apiClient.put(`/api/v1/tenant/settings`, data);
    return response.data;
  },
};
```

---

## 3. OAuth2/OIDC Settings

### Question
Can OAuth2/OIDC settings be different for each tenant? If yes, why aren't we providing that? If no, why not?

### Answer

**OAuth2/OIDC has two types of settings:**

#### A. System-Wide Settings (Cannot be tenant-specific)
These **MUST** be system-wide because they define the OAuth2/OIDC provider infrastructure:

1. **OAuth2 Provider URLs**:
   - `hydra_admin_url` (e.g., `http://localhost:4445`)
   - `hydra_public_url` (e.g., `http://localhost:4444`)
   - These are infrastructure endpoints, not tenant-specific

2. **OIDC Discovery Endpoints**:
   - `.well-known/openid-configuration`
   - `.well-known/oauth-authorization-server`
   - These are standardized endpoints that must be accessible at fixed URLs

3. **JWT Issuer**:
   - `iss` claim in JWT tokens
   - Typically a single domain (e.g., `https://auth.example.com`)
   - Cannot vary per tenant without breaking OIDC compliance

**Why System-Wide?**
- OAuth2/OIDC clients need to know where to connect
- Discovery endpoints must be at predictable URLs
- Multiple issuer domains would break OIDC compliance
- Infrastructure is shared (single Hydra instance)

#### B. Tenant-Specific Settings (CAN be tenant-specific)
These **CAN** be tenant-specific:

1. **OAuth2 Client Registrations**:
   - Each tenant can have their own OAuth2 clients
   - Client IDs, secrets, redirect URIs are tenant-scoped
   - **Status**: Not yet implemented, but should be added

2. **Token Lifetimes** (Already Implemented):
   - Access Token TTL
   - Refresh Token TTL
   - ID Token TTL
   - These are already tenant-specific via `tenant_settings`

3. **Allowed Scopes**:
   - Each tenant can define which OAuth2 scopes they support
   - **Status**: Not yet implemented, but should be added

### Recommendation

**Current State:**
- ✅ Token lifetimes are tenant-specific (already implemented)
- ❌ OAuth2 client management is not tenant-specific (should be added)
- ❌ OAuth2 scopes are not tenant-specific (should be added)

**Future Enhancement:**
1. **Add OAuth2 Client Management**:
   - Table: `oauth2_clients` with `tenant_id`
   - API: `/api/v1/oauth2/clients` (tenant-scoped)
   - SYSTEM users can manage clients for any tenant
   - TENANT users can manage their own clients

2. **Add Tenant-Specific Scopes**:
   - Table: `tenant_oauth2_scopes` with `tenant_id`
   - Each tenant defines which scopes they support
   - Scopes are included in token claims

3. **Keep Provider URLs System-Wide**:
   - These remain in system configuration
   - All tenants use the same Hydra instance
   - Discovery endpoints remain at fixed URLs

### Implementation Priority

**Phase 1 (Current)**:
- ✅ Token lifetimes per tenant
- ✅ Security settings per tenant

**Phase 2 (Recommended)**:
- ⚠️ OAuth2 client management per tenant
- ⚠️ Tenant-specific OAuth2 scopes

**Phase 3 (Future)**:
- ⚠️ Tenant-specific branding (login pages, email templates)
- ⚠️ Tenant-specific consent screens

---

## Summary of Changes

### Completed ✅
1. ✅ Database migration for security settings
2. ✅ Backend API endpoints for TENANT users to access their settings
3. ✅ Updated `TenantSettings` struct and repository methods
4. ✅ Security settings can be configured per tenant

### TODO ⚠️
1. ⚠️ Frontend: Show Security tab for TENANT users
2. ⚠️ Frontend: Fetch and display tenant security settings
3. ⚠️ Frontend: Implement "All Tenants" aggregated view
4. ⚠️ Backend: OAuth2 client management per tenant (future)
5. ⚠️ Backend: Tenant-specific OAuth2 scopes (future)

---

## Testing Checklist

### TENANT User Security Settings
- [ ] TENANT user can see Security tab in Settings
- [ ] TENANT user can view their tenant's security settings
- [ ] TENANT user can update password policies
- [ ] TENANT user can update MFA requirements
- [ ] TENANT user can update rate limiting
- [ ] Changes are persisted and applied

### SYSTEM User "All Tenants" View
- [ ] SYSTEM user can select "All Tenants" in dropdown
- [ ] Dashboard shows aggregated statistics
- [ ] User list shows users from all tenants
- [ ] Role list shows roles from all tenants
- [ ] Permission list shows permissions from all tenants

### OAuth2/OIDC
- [ ] Verify provider URLs are system-wide (config file)
- [ ] Verify token lifetimes are tenant-specific (tenant_settings)
- [ ] Document OAuth2 client management as future enhancement

