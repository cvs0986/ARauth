# Tenant Roles and Permissions Implementation Plan

## Overview
This plan implements predefined tenant roles and permissions that are automatically created when a tenant is created, following industry best practices and ChatGPT's recommendations.

## Goals
1. ✅ Create predefined tenant roles (`tenant_owner`, `tenant_admin`, `tenant_auditor`)
2. ✅ Create predefined tenant permissions (users, roles, permissions, settings, audit)
3. ✅ Automatically initialize roles and permissions when tenant is created
4. ✅ Assign `tenant_owner` to first user automatically
5. ✅ Prevent deletion/modification of predefined roles
6. ✅ Implement permission-based UI access
7. ✅ Add "No Access" page for users without permissions

---

## Phase 1: Backend - Predefined Permissions and Roles

### 1.1 Predefined Tenant Permissions

**Permission Categories:**

#### User Management
- `users:create` - Create new users
- `users:read` - View users
- `users:update` - Update users
- `users:delete` - Delete users
- `users:manage` - Full user management (wildcard)

#### Role Management
- `roles:create` - Create roles
- `roles:read` - View roles
- `roles:update` - Update roles
- `roles:delete` - Delete roles
- `roles:manage` - Full role management

#### Permission Management
- `permissions:create` - Create permissions
- `permissions:read` - View permissions
- `permissions:update` - Update permissions
- `permissions:delete` - Delete permissions
- `permissions:manage` - Full permission management

#### Tenant Settings
- `tenant:settings:read` - View tenant settings
- `tenant:settings:update` - Update tenant settings

#### Audit & Logs
- `audit:read` - View audit logs

#### Admin Access
- `admin:access` - Access admin dashboard (required for any UI access)

### 1.2 Predefined Tenant Roles

#### `tenant_owner`
- **Description**: Full tenant ownership and control
- **Permissions**: All permissions (`*:*` or all listed above)
- **Properties**: 
  - `is_system = true` (non-deletable, non-modifiable)
  - Auto-assigned to first user
  - Can manage everything in tenant

#### `tenant_admin`
- **Description**: Tenant administration with most admin features
- **Permissions**:
  - `users:*`
  - `roles:*`
  - `permissions:*`
  - `tenant:settings:read`
  - `tenant:settings:update`
  - `audit:read`
  - `admin:access`
- **Properties**:
  - `is_system = true` (non-deletable, non-modifiable)
  - Cannot delete tenant or modify tenant_owner

#### `tenant_auditor`
- **Description**: Read-only access for auditing and compliance
- **Permissions**:
  - `users:read`
  - `roles:read`
  - `permissions:read`
  - `audit:read`
  - `admin:access`
- **Properties**:
  - `is_system = true` (non-deletable, non-modifiable)
  - Read-only access

---

## Phase 2: Implementation Steps

### Step 1: Create Tenant Initialization Service
**File**: `identity/tenant/initializer.go`

**Responsibilities**:
- Create predefined roles for a tenant
- Create predefined permissions for a tenant
- Assign permissions to roles
- Return created roles for assignment

**Interface**:
```go
type TenantInitializer interface {
    InitializeTenant(ctx context.Context, tenantID uuid.UUID) (*InitializationResult, error)
}

type InitializationResult struct {
    TenantOwnerRoleID uuid.UUID
    TenantAdminRoleID uuid.UUID
    TenantAuditorRoleID uuid.UUID
    PermissionsCreated int
}
```

### Step 2: Integrate with Tenant Creation
**File**: `identity/tenant/service.go`

**Changes**:
- After tenant creation, call initialization service
- Return initialization result for role assignment

### Step 3: Update Tenant Creation Handler
**File**: `api/handlers/system_handler.go`

**Changes**:
- After tenant creation, assign `tenant_owner` role to first user (if provided)
- Or create first user and assign `tenant_owner` role

### Step 4: Protect System Roles
**File**: `identity/role/service.go` and `api/handlers/role_handler.go`

**Changes**:
- Prevent deletion of roles with `is_system = true`
- Prevent modification of roles with `is_system = true` (name, description)
- Allow permission assignment/removal for system roles

---

## Phase 3: Frontend Updates

### Step 5: Add Permission Checks
**File**: `frontend/admin-dashboard/src/components/layout/Sidebar.tsx`

**Changes**:
- Replace `permission: null` with specific permissions
- Only Dashboard should have `permission: null` (or `admin:access`)

### Step 6: Create No Access Page
**File**: `frontend/admin-dashboard/src/pages/NoAccess.tsx`

**Purpose**:
- Show when user has no `admin:access` permission
- Display helpful message
- Option to redirect to app home

### Step 7: Add Admin Access Check
**File**: `frontend/admin-dashboard/src/App.tsx` or `ProtectedRoute.tsx`

**Changes**:
- Check for `admin:access` permission
- Redirect to `/no-access` if missing

### Step 8: Update Navigation Permissions
**File**: `frontend/admin-dashboard/src/components/layout/Sidebar.tsx`

**New Permission Mapping**:
- Dashboard: `admin:access`
- Users: `users:read`
- Roles: `roles:read`
- Permissions: `permissions:read`
- Settings: `tenant:settings:read`
- Audit Logs: `audit:read`

---

## Phase 4: Testing & Validation

### Step 9: Test Tenant Creation
- Verify roles are created
- Verify permissions are created
- Verify role-permission assignments
- Verify first user gets `tenant_owner` role

### Step 10: Test Role Protection
- Verify system roles cannot be deleted
- Verify system roles cannot be modified
- Verify permissions can be assigned/removed

### Step 11: Test Permission-Based Access
- Verify users with no permissions see "No Access" page
- Verify users with partial permissions see filtered navigation
- Verify `tenant_owner` sees everything
- Verify `tenant_auditor` sees read-only views

---

## Implementation Order

1. ✅ Create tenant initializer service
2. ✅ Integrate with tenant creation
3. ✅ Update role service to protect system roles
4. ✅ Update handlers to prevent system role deletion/modification
5. ✅ Create No Access page
6. ✅ Add admin:access permission check
7. ✅ Update navigation permissions
8. ✅ Test end-to-end flow

---

## Database Schema

### Roles Table (existing)
- `is_system BOOLEAN` - Already exists
- Use this to mark predefined roles as non-deletable

### Permissions Table (existing)
- No changes needed
- Permissions are tenant-scoped

### Role Permissions Table (existing)
- No changes needed
- Used to assign permissions to roles

---

## API Changes

### New Endpoints (if needed)
- None required - initialization happens automatically

### Modified Endpoints
- `POST /system/tenants` - Now initializes roles/permissions
- `DELETE /api/v1/roles/:id` - Now prevents deletion of system roles
- `PUT /api/v1/roles/:id` - Now prevents modification of system roles

---

## Security Considerations

1. **System roles are immutable** - Cannot be deleted or renamed
2. **Permission-based access** - UI respects permissions, backend enforces
3. **First user gets owner** - Prevents lockout scenarios
4. **No blank pages** - Users without access see explicit message

---

## Success Criteria

✅ When tenant is created:
- 3 predefined roles are created
- All predefined permissions are created
- Permissions are assigned to roles correctly
- First user (if provided) gets `tenant_owner` role

✅ System roles:
- Cannot be deleted
- Cannot be modified (name, description)
- Can have permissions assigned/removed

✅ UI Access:
- Users without `admin:access` see "No Access" page
- Navigation filtered by permissions
- No blank pages

---

## Future Enhancements

1. Custom roles can be created by tenant owners/admins
2. Custom permissions can be created
3. Role templates for common use cases
4. Bulk role assignment
5. Role inheritance

