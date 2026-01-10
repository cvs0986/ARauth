# ADR-001: RBAC & Permissions Architecture

**Status**: Accepted  
**Date**: 2025-01-10  
**Deciders**: Architecture Team  
**Tags**: security, rbac, permissions, architecture

---

## Context

ARauth Identity IAM requires a robust Role-Based Access Control (RBAC) system that:
- Supports both system-level and tenant-level access control
- Prevents privilege escalation
- Maintains auditability
- Scales to enterprise requirements
- Supports headless IAM architecture

---

## Decision

We have implemented a **two-plane RBAC system** with:

1. **System Plane**: System roles and permissions for platform control
2. **Tenant Plane**: Tenant roles and permissions for tenant management

### Key Design Decisions

#### 1. Explicit Permissions (No Wildcards)

**Decision**: All permissions are explicitly assigned. No `*:*` wildcards.

**Rationale**:
- Enables auditing and forensics
- Meets compliance requirements
- Predictable permission evolution
- Clear security boundaries

**Alternatives Considered**:
- Wildcard permissions (`*:*`) - Rejected: Not auditable
- Permission groups - Considered for future: Too complex for MVP

**Consequences**:
- ✅ More verbose permission assignments
- ✅ Better security and auditability
- ✅ Clearer permission model

---

#### 2. Namespaced Permissions

**Decision**: All tenant permissions use `tenant.*` namespace. System permissions use `system.*`.

**Format**: `{namespace}.{resource}.{action}`

Examples:
- `tenant.users.create`
- `tenant.roles.read`
- `system.tenants.create`
- `system.users.manage`

**Rationale**:
- Prevents collision between system and tenant permissions
- Clear plane separation
- Easier policy evaluation
- Future-proof for additional namespaces

**Alternatives Considered**:
- Flat permissions (`users:create`) - Rejected: Collision risk
- Hierarchical (`tenant/users/create`) - Rejected: Less readable

**Consequences**:
- ✅ Clear namespace boundaries
- ✅ No permission collision
- ✅ Easier to reason about

---

#### 3. Predefined Tenant Roles

**Decision**: Every tenant automatically gets three predefined roles:
- `tenant_owner` - Full control
- `tenant_admin` - Most admin features (no permission management by default)
- `tenant_auditor` - Read-only access

**Rationale**:
- Prevents lockout scenarios
- Simplifies bootstrap
- Industry standard (AWS, Azure, Auth0)
- Clear role hierarchy

**Alternatives Considered**:
- No predefined roles - Rejected: Lockout risk
- More predefined roles - Considered: Deferred to future

**Consequences**:
- ✅ Safe bootstrap process
- ✅ Clear role expectations
- ✅ Prevents accidental lockouts

---

#### 4. Auto-Attach Permissions to `tenant_owner`

**Decision**: When new permissions are created, they are automatically attached to `tenant_owner`.

**Rationale**:
- Maintains "owner has all permissions" invariant
- Prevents accidental lockouts
- No migration surprises
- Consistent behavior

**Alternatives Considered**:
- Manual assignment - Rejected: Error-prone
- Periodic sync - Rejected: Delayed updates

**Consequences**:
- ✅ Owner always has access
- ✅ No lockout scenarios
- ✅ Automatic permission sync

---

#### 5. Namespace Validation

**Decision**: Tenants can only create permissions in allowed namespaces: `tenant.*`, `app.*`, `resource.*`.

**Rationale**:
- Prevents privilege escalation
- Maintains namespace isolation
- Security boundary enforcement

**Alternatives Considered**:
- No validation - Rejected: Security risk
- More restrictive - Considered: Too limiting

**Consequences**:
- ✅ Prevents privilege escalation
- ✅ Clear security boundaries
- ✅ Tenant flexibility within bounds

---

#### 6. Hard Separation: System vs Tenant Roles

**Decision**: System roles and tenant roles are physically separated:
- System roles: `tenant_id = NULL`, `is_system = true`
- Tenant roles: `tenant_id = <uuid>`, `is_system = false` (or `true` for predefined)

**Rationale**:
- Prevents catastrophic privilege bugs
- Clear plane separation
- Easier to reason about
- Prevents accidental mixing

**Alternatives Considered**:
- Single table with flags - Rejected: Mixing risk
- Separate tables - Considered: Over-engineering

**Consequences**:
- ✅ Clear separation
- ✅ Prevents privilege escalation
- ✅ Easier validation

---

#### 7. Permission-Based UI Access

**Decision**: Admin UI access is controlled by `tenant.admin.access` permission, not role names.

**Rationale**:
- Avoids role-name coupling
- Scales for enterprises
- Flexible permission model
- No hard-coded UI logic

**Alternatives Considered**:
- Role-based UI - Rejected: Not scalable
- Hard-coded owner check - Rejected: Inflexible

**Consequences**:
- ✅ Flexible permission model
- ✅ Enterprise-ready
- ✅ No UI-role coupling

---

#### 8. No Role Inheritance (For Now)

**Decision**: Flat RBAC structure. No role inheritance.

**Rationale**:
- Simpler to reason about
- Safer (no circular dependencies)
- Covers 90% of enterprise needs
- Can be added later without breaking changes

**Alternatives Considered**:
- Role inheritance - Deferred: Too complex for MVP
- Permission groups - Deferred: Future enhancement

**Consequences**:
- ✅ Simpler model
- ✅ Easier to understand
- ✅ Can evolve later

---

## Permission Evolution Strategy

### When New Permissions Are Introduced

1. **Create permission** with `tenant.*` namespace
2. **Auto-attach to `tenant_owner`** (automatic)
3. **Other roles unchanged** (explicit assignment required)

### Migration Strategy

For existing tenants when new permissions are added:

1. **Option A**: Run initialization again (idempotent)
2. **Option B**: Manual assignment to specific roles
3. **Option C**: Migration script to assign to all roles

**Current Approach**: Option A (idempotent initialization)

---

## Permission → OAuth Scope Mapping

### Current State

Permissions are used for:
- Admin dashboard access
- API authorization
- Control plane operations

OAuth scopes are used for:
- Application access tokens
- Resource access
- User consent

### Mapping Rules (To Be Defined)

```
permission: tenant.users.read
  → scope: users:read (tenant-defined)
  → scope: tenant.users.read (system-defined)
```

**Future Work**: Define explicit mapping rules and tenant customization.

---

## Security Considerations

### Invariants

See [INVARIANTS.md](../INVARIANTS.md) for complete list.

Key invariants:
1. No wildcard permissions
2. System roles never belong to tenants
3. `tenant_owner` always has all permissions
4. Namespace validation enforced server-side
5. Backend is authoritative (UI is advisory)

### Threat Model

Protected against:
- ✅ Privilege escalation via permission creation
- ✅ Tenant lockout scenarios
- ✅ System role hijacking
- ✅ Permission namespace collision
- ✅ Wildcard permission abuse

---

## Future Enhancements

### Not Implemented (Deferred)

1. **Role Inheritance** - Can be added without breaking changes
2. **Permission Groups** - Can be added as abstraction layer
3. **Dynamic Permission Evaluation** - Can be added via capability model
4. **Permission Templates** - Can be added for common use cases

### Rationale for Deferral

- Current model covers 90% of enterprise needs
- Simpler is safer
- Can evolve without breaking changes
- Focus on core security first

---

## Consequences

### Positive

- ✅ Enterprise-grade security
- ✅ Audit-friendly design
- ✅ Scalable architecture
- ✅ Future-proof permission model
- ✅ Clear separation of concerns

### Negative

- ⚠️ More verbose permission assignments
- ⚠️ Requires explicit permission management
- ⚠️ No role inheritance (yet)

### Neutral

- Permission evolution requires explicit assignment
- Migration needed when adding new permissions

---

## References

- [INVARIANTS.md](../INVARIANTS.md) - Security invariants
- [TENANT_ROLES_PERMISSIONS_IMPLEMENTATION.md](../../implementation/TENANT_ROLES_PERMISSIONS_IMPLEMENTATION.md) - Implementation details
- [CHATGPT_FEEDBACK_APPLIED.md](../../implementation/CHATGPT_FEEDBACK_APPLIED.md) - Expert review and adjustments

---

## Approval

- ✅ Architecture Team
- ✅ Security Team
- ✅ Implementation Team

**Status**: Accepted and Implemented

