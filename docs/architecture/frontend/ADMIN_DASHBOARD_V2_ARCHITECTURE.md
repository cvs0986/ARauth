# Admin Dashboard V2 Architecture - SYSTEM vs TENANT Users

## Overview

The Admin Dashboard now supports two types of users with different capabilities:

1. **SYSTEM Users (Master Admin)**
   - Can see and manage ALL tenants
   - Can create new tenants
   - Can manage system-wide configurations
   - Can create tenant admins
   - Access to `/system/*` API endpoints
   - System-level permissions (e.g., `tenant:create`, `tenant:read`, `system:configure`)

2. **TENANT Users (Tenant Admin)**
   - Can only see and manage their own tenant
   - Can create users within their tenant
   - Can manage tenant-specific configurations
   - Access to `/api/v1/*` API endpoints (tenant-scoped)
   - Tenant-level permissions (e.g., `users:create`, `roles:manage`)

## Architecture Changes

### 1. Authentication & User Context

#### Auth Store Updates
- Store `principal_type`: `"SYSTEM"` | `"TENANT"`
- Store `system_permissions`: Array of system-level permissions (SYSTEM users only)
- Store `tenant_permissions`: Array of tenant-level permissions
- Store `tenant_id`: For TENANT users (required), null for SYSTEM users
- Store `available_tenants`: For SYSTEM users (list of all tenants they can manage)

#### Login Flow
```typescript
// Login response includes:
{
  access_token: string,
  principal_type: "SYSTEM" | "TENANT",
  tenant_id: string | null,
  system_permissions: string[],
  permissions: string[],
  // ... other fields
}
```

### 2. API Client Structure

#### Dual API Endpoints
```typescript
// System API (SYSTEM users only)
systemApi = {
  tenants: {
    list: () => GET /system/tenants
    create: (data) => POST /system/tenants
    update: (id, data) => PUT /system/tenants/:id
    delete: (id) => DELETE /system/tenants/:id
    suspend: (id) => POST /system/tenants/:id/suspend
    resume: (id) => POST /system/tenants/:id/resume
  },
  settings: {
    get: () => GET /system/settings
    update: (data) => PUT /system/settings
  }
}

// Tenant API (TENANT users, or SYSTEM users with tenant context)
tenantApi = {
  users: { ... },
  roles: { ... },
  permissions: { ... },
  settings: { ... }
}
```

### 3. UI Layout & Navigation

#### Conditional Sidebar
```typescript
// SYSTEM User Sidebar
- Dashboard (All Tenants Overview)
- Tenants (Manage All Tenants)
- Tenant Users (with tenant selector)
- System Settings
- Audit Logs (All Tenants)

// TENANT User Sidebar
- Dashboard (Tenant Overview)
- Users (Tenant Users Only)
- Roles (Tenant Roles)
- Permissions (Tenant Permissions)
- Tenant Settings
- Audit Logs (Tenant Only)
```

#### Tenant Selector (SYSTEM Users Only)
- Dropdown in header to switch between tenants
- When tenant selected, show tenant-scoped data
- Can switch back to "All Tenants" view

### 4. Dashboard Page

#### SYSTEM User Dashboard
- **Statistics Cards:**
  - Total Tenants
  - Total Users (across all tenants)
  - Active Tenants
  - Suspended Tenants
- **Recent Activity:**
  - Recent tenant creations
  - Recent tenant admin assignments
  - System-wide audit logs
- **Quick Actions:**
  - Create New Tenant
  - Create Tenant Admin
  - View System Settings

#### TENANT User Dashboard
- **Statistics Cards:**
  - Total Users (in tenant)
  - Active Users
  - Total Roles
  - Total Permissions
- **Recent Activity:**
  - Recent user creations
  - Recent role assignments
  - Tenant audit logs
- **Quick Actions:**
  - Create User
  - Create Role
  - View Tenant Settings

### 5. Settings Page

#### System Settings (SYSTEM Users Only)
- Global JWT configuration
- System-wide security policies
- OAuth2/OIDC configuration
- Master user management
- System audit settings

#### Tenant Settings (TENANT Users)
- Tenant-specific JWT TTLs
- Tenant password policies
- Tenant MFA requirements
- Tenant branding (future)
- Tenant audit settings

### 6. User Management

#### SYSTEM User - User Management
- **View:** All users across all tenants (with tenant filter)
- **Create:** Tenant Admin users (assign to specific tenant)
- **Actions:**
  - Create tenant admin
  - Assign system roles
  - Suspend/activate users across tenants
  - View cross-tenant audit logs

#### TENANT User - User Management
- **View:** Only users in their tenant
- **Create:** Regular tenant users
- **Actions:**
  - Create user
  - Assign tenant roles
  - Manage user status (within tenant)
  - View tenant audit logs

### 7. Permission Checks

#### Frontend Permission Middleware
```typescript
// Check if user has system permission
const hasSystemPermission = (permission: string) => {
  return systemPermissions.includes(permission) || 
         systemPermissions.includes('*:*')
}

// Check if user has tenant permission
const hasTenantPermission = (permission: string) => {
  return permissions.includes(permission) ||
         permissions.includes('*:*')
}

// Conditional rendering
{isSystemUser && hasSystemPermission('tenant:create') && (
  <Button>Create Tenant</Button>
)}
```

## Implementation Plan

### Phase 1: Auth Store & API Client
1. Update `authStore.ts` to store `principal_type`, `system_permissions`
2. Update API client to support both `/system/*` and `/api/v1/*`
3. Add helper functions to check user type and permissions

### Phase 2: Layout & Navigation
1. Update `Sidebar.tsx` to conditionally render menu items
2. Add tenant selector component (SYSTEM users only)
3. Update `Header.tsx` to show tenant selector
4. Update routing to handle SYSTEM vs TENANT routes

### Phase 3: Dashboard Updates
1. Update `Dashboard.tsx` to show different stats based on user type
2. Add tenant selector integration
3. Update quick actions based on permissions

### Phase 4: Settings Page
1. Split Settings into System Settings and Tenant Settings
2. Add permission checks for System Settings
3. Update API calls based on user type

### Phase 5: User Management
1. Update User Management to handle tenant context
2. Add tenant filter for SYSTEM users
3. Update create user form based on user type

### Phase 6: Testing
1. Test SYSTEM user flow
2. Test TENANT user flow
3. Test permission-based UI rendering
4. Test tenant switching (SYSTEM users)

## Security Considerations

1. **API Endpoint Protection:**
   - Frontend should never call `/system/*` endpoints for TENANT users
   - Backend enforces this, but frontend should also check

2. **Permission Checks:**
   - Always check permissions before showing actions
   - Hide UI elements user can't access

3. **Tenant Isolation:**
   - TENANT users should never see other tenants' data
   - SYSTEM users should explicitly select tenant context

4. **Token Validation:**
   - Verify `principal_type` in token
   - Verify `tenant_id` matches for TENANT users

## Migration Notes

- Existing tenant admin users will continue to work as TENANT users
- New SYSTEM users need to be created via bootstrap
- Frontend will automatically detect user type from login response
- No breaking changes to existing tenant-scoped functionality

