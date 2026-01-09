# Permission Evolution Strategy

## Overview

This document describes how permissions evolve in ARauth IAM when new permissions are introduced to the system.

---

## Core Principle

> **When new permissions are introduced, they are automatically attached to `tenant_owner`. Other roles remain unchanged and require explicit assignment.**

---

## How It Works

### 1. New Permission Creation

When a new permission is created (either by system initialization or tenant creation):

```go
// In identity/permission/service.go::Create
permission := &models.Permission{
    TenantID: tenantID,
    Resource: "tenant.newfeature",
    Action:   "manage",
    // ...
}

// Create permission
s.repo.Create(ctx, permission)

// Auto-attach to tenant_owner
s.tenantInitializer.AttachAllPermissionsToTenantOwner(ctx, tenantID)
```

### 2. Auto-Attach Process

The `AttachAllPermissionsToTenantOwner()` method:

1. Gets the `tenant_owner` role for the tenant
2. Lists all current permissions for the tenant
3. Assigns all permissions (including the new one) to `tenant_owner`
4. Is idempotent (safe to call multiple times)

### 3. Other Roles

Other roles (`tenant_admin`, `tenant_auditor`, custom roles) are **not** automatically updated. They require:

- Explicit permission assignment by tenant administrators
- Or re-running role initialization (if predefined roles)

---

## Scenarios

### Scenario 1: New Tenant Created

**What Happens**:
1. Tenant is created
2. All predefined permissions are created
3. All permissions are assigned to `tenant_owner`
4. Subset of permissions assigned to `tenant_admin` and `tenant_auditor`

**Result**: ✅ All roles have appropriate permissions

---

### Scenario 2: New Permission Added to System

**Example**: System adds `tenant.reports.generate` permission

**What Happens**:
1. New permission is created for existing tenants (or on next tenant creation)
2. Permission is automatically attached to `tenant_owner`
3. `tenant_admin` and `tenant_auditor` do NOT get it automatically

**Result**: 
- ✅ `tenant_owner` can use new feature immediately
- ⚠️ Other roles need explicit assignment

**Migration Options**:
- Option A: Re-run tenant initialization (idempotent)
- Option B: Manually assign to specific roles
- Option C: Migration script to assign to all roles

---

### Scenario 3: Tenant Creates Custom Permission

**Example**: Tenant creates `app.customfeature.manage`

**What Happens**:
1. Permission is created (namespace validated)
2. Permission is automatically attached to `tenant_owner`
3. Other roles do NOT get it automatically

**Result**:
- ✅ `tenant_owner` can use custom feature immediately
- ⚠️ Other roles need explicit assignment if needed

---

## Migration Strategies

### Strategy 1: Idempotent Initialization (Current)

**How**: Re-run tenant initialization

**Pros**:
- Simple
- Idempotent (safe to run multiple times)
- Updates predefined roles

**Cons**:
- Doesn't update custom roles
- May overwrite custom role assignments

**Use When**: Adding new predefined permissions

---

### Strategy 2: Explicit Assignment

**How**: Manually assign permissions to roles

**Pros**:
- Precise control
- Doesn't affect existing assignments
- Works for custom roles

**Cons**:
- Manual work
- Easy to miss roles

**Use When**: Adding permissions to specific roles only

---

### Strategy 3: Migration Script

**How**: Script that assigns new permissions to all roles

**Pros**:
- Automated
- Consistent
- Can target specific roles

**Cons**:
- Requires script maintenance
- May assign to roles that shouldn't have it

**Use When**: Bulk permission updates needed

---

## Best Practices

### For System Developers

1. **Document new permissions**: Add to permission list in `initializer.go`
2. **Update role assignments**: Update predefined role permission lists
3. **Test auto-attach**: Verify `tenant_owner` gets new permissions
4. **Migration guide**: Document migration steps if needed

### For Tenant Administrators

1. **Review new permissions**: Check what `tenant_owner` gets automatically
2. **Assign to roles**: Explicitly assign to `tenant_admin` if needed
3. **Audit regularly**: Review role permissions periodically
4. **Use custom roles**: Create custom roles for specific permission sets

---

## Examples

### Example 1: Adding Reporting Feature

```go
// New permission added to initializer
{"tenant.reports.generate", "tenant.reports", "generate", "Generate reports"}

// What happens:
// 1. Permission created for all tenants (on next initialization)
// 2. Auto-attached to tenant_owner ✅
// 3. NOT attached to tenant_admin (needs explicit assignment)
// 4. NOT attached to tenant_auditor (read-only, doesn't need it)
```

### Example 2: Tenant Creates Custom Permission

```go
// Tenant creates: app.analytics.view
// Namespace validated: ✅ (app.* is allowed)
// Created successfully
// Auto-attached to tenant_owner ✅
// Other roles: Need explicit assignment
```

---

## Invariants

1. **`tenant_owner` always has all permissions** - Maintained automatically
2. **Other roles require explicit assignment** - No automatic updates
3. **New permissions are discoverable** - Available for assignment
4. **Migration is optional** - System works without migration

---

## Future Enhancements

### Permission Templates

Future: Permission templates that define which roles get which permissions automatically.

Example:
```yaml
permission: tenant.reports.generate
auto_assign_to:
  - tenant_owner
  - tenant_admin
```

### Permission Groups

Future: Group related permissions for easier management.

Example:
```yaml
group: reporting
permissions:
  - tenant.reports.generate
  - tenant.reports.export
  - tenant.reports.view
```

---

## References

- [INVARIANTS.md](./INVARIANTS.md) - Security invariants
- [ADR-001-RBAC-PERMISSIONS.md](./adr/ADR-001-RBAC-PERMISSIONS.md) - Architecture decision
- [identity/tenant/initializer.go](../../identity/tenant/initializer.go) - Implementation

---

**Last Updated**: 2025-01-10  
**Status**: Active  
**Owner**: Architecture Team

