# Future Enhancements Status

## ✅ Implemented Features

### 1. Custom Roles Can Be Created by Tenant Owners/Admins
**Status**: ✅ **FULLY IMPLEMENTED**

**Backend**:
- Endpoint: `POST /api/v1/roles`
- Handler: `api/handlers/role_handler.go::Create`
- Service: `identity/role/service.go::Create`
- Permission required: `roles:create`
- Custom roles have `is_system = false` (unlike predefined roles)
- Custom roles can be deleted and modified (unlike system roles)

**Frontend**:
- Component: `frontend/admin-dashboard/src/pages/roles/CreateRoleDialog.tsx`
- UI: Available in Role List page
- Users with `roles:create` permission can create custom roles

**How it works**:
1. Tenant owners/admins with `roles:create` permission can create custom roles
2. Custom roles are tenant-scoped
3. Permissions can be assigned to custom roles
4. Custom roles can be assigned to users
5. Custom roles can be deleted (unlike system roles)

---

### 2. Custom Permissions Can Be Created
**Status**: ✅ **FULLY IMPLEMENTED**

**Backend**:
- Endpoint: `POST /api/v1/permissions`
- Handler: `api/handlers/permission_handler.go::Create`
- Service: `identity/permission/service.go::Create`
- Permission required: `permissions:create`
- Custom permissions are tenant-scoped (have `tenant_id`)

**Frontend**:
- Component: `frontend/admin-dashboard/src/pages/permissions/CreatePermissionDialog.tsx`
- UI: Available in Permission List page
- Users with `permissions:create` permission can create custom permissions

**How it works**:
1. Tenant owners/admins with `permissions:create` permission can create custom permissions
2. Custom permissions are tenant-scoped
3. Custom permissions can be assigned to roles
4. Custom permissions can be deleted

---

## ❌ Not Implemented Features

### 3. Role Templates for Common Use Cases
**Status**: ❌ **NOT IMPLEMENTED**

**What's missing**:
- No template system in the codebase
- No predefined role templates (e.g., "Security Admin", "Helpdesk", "Developer")
- No UI for selecting/creating roles from templates

**What would be needed**:
- Database table for role templates
- API endpoints for managing templates
- UI for browsing and applying templates
- Template definition structure (name, description, permissions list)

**Potential Implementation**:
```go
// Example structure
type RoleTemplate struct {
    ID          uuid.UUID
    Name        string
    Description string
    Permissions []string // List of permission keys
    IsSystem    bool     // System templates vs custom templates
}
```

---

### 4. Bulk Role Assignment
**Status**: ❌ **NOT IMPLEMENTED**

**What's missing**:
- Only single role assignment exists: `POST /api/v1/users/:id/roles/:role_id`
- No bulk assignment endpoint
- No UI for bulk assignment

**What exists**:
- Single assignment: `api/handlers/role_handler.go::AssignRoleToUser`
- Single removal: `api/handlers/role_handler.go::RemoveRoleFromUser`

**What would be needed**:
- New endpoint: `POST /api/v1/users/bulk-assign-roles`
- Request body: `{ user_ids: [], role_id: uuid }` or `{ user_id: uuid, role_ids: [] }`
- UI: Checkbox selection in user list + bulk action dropdown
- Batch processing logic

**Potential Implementation**:
```go
// Example endpoint
POST /api/v1/users/bulk-assign-roles
{
  "user_ids": ["uuid1", "uuid2", "uuid3"],
  "role_id": "role-uuid"
}
```

---

### 5. Role Inheritance
**Status**: ❌ **NOT IMPLEMENTED**

**What's missing**:
- No inheritance fields in Role model
- No inheritance logic in role service
- No parent/child role relationships

**What exists**:
- Flat role structure (no hierarchy)
- Roles have direct permission assignments only

**What would be needed**:
- Add `parent_role_id` field to `roles` table
- Add `inherited_permissions` logic to permission calculation
- API endpoints for managing role hierarchy
- UI for visualizing role inheritance tree
- Validation to prevent circular inheritance

**Potential Implementation**:
```go
// Role model update
type Role struct {
    // ... existing fields
    ParentRoleID *uuid.UUID `json:"parent_role_id" db:"parent_role_id"`
}

// Permission calculation
func (s *Service) GetEffectivePermissions(ctx context.Context, roleID uuid.UUID) ([]*models.Permission, error) {
    // Get direct permissions
    // Get inherited permissions from parent role
    // Merge and deduplicate
}
```

---

## 📊 Summary

| Feature | Status | Backend | Frontend | Notes |
|---------|--------|---------|----------|-------|
| **Custom Roles** | ✅ Implemented | ✅ | ✅ | Fully functional |
| **Custom Permissions** | ✅ Implemented | ✅ | ✅ | Fully functional |
| **Role Templates** | ❌ Not Implemented | ❌ | ❌ | Would be useful for common use cases |
| **Bulk Role Assignment** | ❌ Not Implemented | ❌ | ❌ | Currently only single assignment |
| **Role Inheritance** | ❌ Not Implemented | ❌ | ❌ | Would enable hierarchical RBAC |

---

## 🎯 Recommendations

### High Priority
1. **Bulk Role Assignment** - Very useful for onboarding multiple users
   - Common use case: Assign "Developer" role to 50 new employees
   - Implementation effort: Medium
   - Business value: High

### Medium Priority
2. **Role Templates** - Speeds up role creation
   - Common use case: Quick setup of standard roles (Security Admin, Helpdesk, etc.)
   - Implementation effort: Medium
   - Business value: Medium

### Low Priority
3. **Role Inheritance** - Advanced feature
   - Common use case: Complex organizational hierarchies
   - Implementation effort: High (requires careful design to avoid circular dependencies)
   - Business value: Medium (most use cases can be handled with flat roles + custom permissions)

---

## 🔍 Code References

### Custom Roles
- Backend: `api/handlers/role_handler.go:34` (Create handler)
- Backend: `identity/role/service.go:42` (Create service)
- Frontend: `frontend/admin-dashboard/src/pages/roles/CreateRoleDialog.tsx`
- Route: `POST /api/v1/roles`

### Custom Permissions
- Backend: `api/handlers/permission_handler.go:25` (Create handler)
- Backend: `identity/permission/service.go` (Create service)
- Frontend: `frontend/admin-dashboard/src/pages/permissions/CreatePermissionDialog.tsx`
- Route: `POST /api/v1/permissions`

### Single Role Assignment (Current)
- Backend: `api/handlers/role_handler.go:348` (AssignRoleToUser)
- Route: `POST /api/v1/users/:id/roles/:role_id`

