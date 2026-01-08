# Admin Dashboard V2 Implementation Plan

## Overview

The Admin Dashboard needs to support two distinct user types with different capabilities:

1. **SYSTEM Users (Master Admin)**
   - Can manage ALL tenants
   - Can create new tenants
   - Can manage system-wide configurations
   - Can create tenant admins
   - Access to `/system/*` API endpoints

2. **TENANT Users (Tenant Admin)**
   - Can only manage their own tenant
   - Can create users within their tenant
   - Can manage tenant-specific configurations
   - Access to `/api/v1/*` API endpoints (tenant-scoped)

## ✅ Phase 1: Auth Store & API Client (COMPLETE)

### Completed:
- ✅ Updated `authStore.ts` with:
  - `principalType`: `'SYSTEM' | 'TENANT'`
  - `systemPermissions`: Array of system-level permissions
  - `permissions`: Array of tenant-level permissions
  - `selectedTenantId`: For SYSTEM users to switch tenant context
  - Helper methods: `isSystemUser()`, `isTenantUser()`, `hasSystemPermission()`, `hasPermission()`, `getCurrentTenantId()`

- ✅ Created JWT decoder utility (`jwt-decoder.ts`)
  - Extracts `principal_type`, `system_permissions`, `permissions` from JWT token
  - Used during login to populate auth store

- ✅ Updated Login page
  - Decodes JWT token after login
  - Stores user context (principal_type, permissions) in auth store

- ✅ Added System API client
  - `systemApi.tenants.*` for managing tenants (SYSTEM users only)

## 📋 Phase 2: Layout & Navigation (TODO)

### 2.1 Update Sidebar Component

**File**: `frontend/admin-dashboard/src/components/layout/Sidebar.tsx`

**Changes**:
```typescript
// Conditional navigation based on user type
const navigation = isSystemUser ? [
  { name: 'Dashboard', href: '/', icon: '📊' },
  { name: 'Tenants', href: '/tenants', icon: '🏢' }, // SYSTEM only
  { name: 'Users', href: '/users', icon: '👤' }, // With tenant selector
  { name: 'Roles', href: '/roles', icon: '🔑' }, // With tenant selector
  { name: 'Permissions', href: '/permissions', icon: '🛡️' }, // With tenant selector
  { name: 'System Settings', href: '/settings/system', icon: '⚙️' }, // SYSTEM only
  { name: 'Audit Logs', href: '/audit', icon: '📋' }, // All tenants
] : [
  { name: 'Dashboard', href: '/', icon: '📊' },
  { name: 'Users', href: '/users', icon: '👤' }, // Tenant only
  { name: 'Roles', href: '/roles', icon: '🔑' }, // Tenant only
  { name: 'Permissions', href: '/permissions', icon: '🛡️' }, // Tenant only
  { name: 'Tenant Settings', href: '/settings/tenant', icon: '⚙️' }, // Tenant only
  { name: 'Audit Logs', href: '/audit', icon: '📋' }, // Tenant only
];
```

### 2.2 Add Tenant Selector Component

**New File**: `frontend/admin-dashboard/src/components/TenantSelector.tsx`

**Purpose**: Allow SYSTEM users to switch between tenants

**Features**:
- Dropdown showing all available tenants
- "All Tenants" option for system-wide view
- Updates `selectedTenantId` in auth store
- Triggers data refresh when tenant changes

### 2.3 Update Header Component

**File**: `frontend/admin-dashboard/src/components/layout/Header.tsx`

**Changes**:
- Show Tenant Selector (SYSTEM users only)
- Show current tenant name (TENANT users)
- Show user type badge (SYSTEM/TENANT)

## 📋 Phase 3: Dashboard Updates (TODO)

### 3.1 Update Dashboard Page

**File**: `frontend/admin-dashboard/src/pages/Dashboard.tsx`

**SYSTEM User Dashboard**:
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
  - Create New Tenant (if has `tenant:create` permission)
  - Create Tenant Admin
  - View System Settings

**TENANT User Dashboard**:
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
  - Create User (if has `users:create` permission)
  - Create Role
  - View Tenant Settings

## 📋 Phase 4: Settings Page Split (TODO)

### 4.1 Create System Settings Page

**New File**: `frontend/admin-dashboard/src/pages/settings/SystemSettings.tsx`

**Features** (SYSTEM users only):
- Global JWT configuration
- System-wide security policies
- OAuth2/OIDC configuration
- Master user management
- System audit settings

### 4.2 Create Tenant Settings Page

**New File**: `frontend/admin-dashboard/src/pages/settings/TenantSettings.tsx`

**Features** (TENANT users):
- Tenant-specific JWT TTLs
- Tenant password policies
- Tenant MFA requirements
- Tenant branding (future)
- Tenant audit settings

### 4.3 Update Settings Router

**File**: `frontend/admin-dashboard/src/App.tsx`

**Routes**:
- `/settings/system` → SystemSettings (SYSTEM users only)
- `/settings/tenant` → TenantSettings (TENANT users)
- `/settings` → Redirect based on user type

## 📋 Phase 5: User Management Updates (TODO)

### 5.1 Update User List Page

**File**: `frontend/admin-dashboard/src/pages/users/UserList.tsx`

**SYSTEM Users**:
- Show tenant filter dropdown
- Show users from all tenants (when no tenant selected)
- Show users from selected tenant (when tenant selected)
- "Create Tenant Admin" button

**TENANT Users**:
- Show only users from their tenant
- "Create User" button

### 5.2 Update Create User Form

**File**: `frontend/admin-dashboard/src/pages/users/CreateUser.tsx`

**SYSTEM Users**:
- Tenant selector (required)
- Role selector (system roles + tenant roles)
- Can create tenant admins

**TENANT Users**:
- Tenant ID pre-filled (read-only)
- Role selector (tenant roles only)
- Can create regular users

## 📋 Phase 6: Permission-Based UI (TODO)

### 6.1 Create Permission Hooks

**New File**: `frontend/admin-dashboard/src/hooks/usePermissions.ts`

```typescript
export function usePermissions() {
  const { hasSystemPermission, hasPermission, isSystemUser } = useAuthStore();
  
  return {
    canCreateTenant: () => isSystemUser() && hasSystemPermission('tenant:create'),
    canManageTenants: () => isSystemUser() && hasSystemPermission('tenant:*'),
    canCreateUser: () => hasPermission('users:create'),
    canManageRoles: () => hasPermission('roles:*'),
    // ... more permission checks
  };
}
```

### 6.2 Update Components to Use Permissions

- Hide/show buttons based on permissions
- Disable actions user can't perform
- Show appropriate error messages

## 🔐 Security Considerations

1. **Frontend Protection:**
   - Never call `/system/*` endpoints for TENANT users
   - Hide UI elements user can't access
   - Check permissions before showing actions

2. **Backend Enforcement:**
   - Backend always verifies principal_type
   - Backend always verifies permissions
   - Frontend checks are for UX only, not security

3. **Tenant Isolation:**
   - TENANT users should never see other tenants' data
   - SYSTEM users must explicitly select tenant context
   - All tenant-scoped API calls include tenant_id

## 📊 User Experience Flow

### SYSTEM User Flow:
1. Login (no tenant_id required)
2. See "All Tenants" dashboard
3. Can select a tenant from dropdown
4. When tenant selected, see tenant-scoped data
5. Can switch back to "All Tenants" view
6. Can create new tenants
7. Can manage system settings

### TENANT User Flow:
1. Login (tenant_id from JWT)
2. See tenant dashboard
3. Can only see/manage their tenant's data
4. Can manage tenant settings
5. Cannot access system settings
6. Cannot see other tenants

## 🚀 Implementation Priority

1. **High Priority:**
   - Phase 2: Sidebar & Tenant Selector (enables basic navigation)
   - Phase 3: Dashboard Updates (main landing page)

2. **Medium Priority:**
   - Phase 4: Settings Split (important for configuration)
   - Phase 5: User Management Updates (core functionality)

3. **Low Priority:**
   - Phase 6: Permission-Based UI (polish, but backend enforces)

## 📝 Testing Checklist

- [ ] SYSTEM user can login without tenant_id
- [ ] TENANT user can login with tenant_id
- [ ] SYSTEM user sees all tenants in dashboard
- [ ] TENANT user sees only their tenant
- [ ] SYSTEM user can switch tenants
- [ ] TENANT user cannot access system settings
- [ ] SYSTEM user can create tenants
- [ ] TENANT user cannot create tenants
- [ ] Permission checks work correctly
- [ ] UI elements hide/show based on permissions

