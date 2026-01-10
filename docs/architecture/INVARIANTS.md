# ARauth IAM Security Invariants

## 🔒 Core Security Invariants

These invariants **MUST** remain true for ARauth to maintain its security posture. Violating any of these invariants can lead to privilege escalation, lockouts, or security breaches.

---

## 1. No Wildcard Permissions

**Invariant**: `*:*` or wildcard permissions are **never** used.

**Rationale**:
- Wildcards make auditing impossible
- Forensics become unreliable
- Compliance requirements cannot be met
- Permission evolution becomes unpredictable

**Enforcement**:
- All permissions are explicitly assigned
- `tenant_owner` gets all permissions explicitly (not via wildcard)
- Permission checks never use wildcard matching

**Status**: ✅ Enforced

---

## 2. System Roles Never Belong to Tenants

**Invariant**: System roles (`is_system = true`) **never** have a `tenant_id`.

**Rationale**:
- Prevents catastrophic privilege escalation
- Maintains clear control-plane / tenant-plane separation
- Prevents tenants from accessing system-level privileges

**Enforcement**:
- System roles: `tenant_id = NULL`, `is_system = true`
- Tenant roles: `tenant_id = <uuid>`, `is_system = false` (or `true` for predefined tenant roles)
- Tenant API cannot create system roles
- Validation in `identity/role/service.go::Create`

**Status**: ✅ Enforced

---

## 3. Tenant Roles Never Affect System APIs

**Invariant**: Tenant roles and permissions **never** grant access to system-level APIs.

**Rationale**:
- Maintains plane separation
- Prevents privilege escalation
- Ensures system APIs remain system-only

**Enforcement**:
- System APIs check `principal_type = SYSTEM`
- System APIs check `systemPermissions` (not `permissions`)
- Tenant permissions use `tenant.*` namespace (cannot access `system.*`)

**Status**: ✅ Enforced

---

## 4. `tenant_owner` Always Has All Tenant Permissions

**Invariant**: The `tenant_owner` role **always** has all permissions for its tenant.

**Rationale**:
- Prevents accidental lockouts
- Ensures tenant always has a recovery path
- Maintains "owner can do everything" guarantee

**Enforcement**:
- Auto-attach new permissions to `tenant_owner` when created
- `AttachAllPermissionsToTenantOwner()` called automatically
- Initialization assigns all permissions explicitly

**Status**: ✅ Enforced

**Additional Safeguard**: ⚠️ TODO - Prevent removal of last `tenant_owner` assignment

---

## 5. Permission Namespace is Enforced Server-Side

**Invariant**: Tenants can **only** create permissions in allowed namespaces.

**Rationale**:
- Prevents privilege escalation attempts
- Maintains namespace isolation
- Prevents collision with system permissions

**Enforcement**:
- Allowed namespaces: `tenant.*`, `app.*`, `resource.*`
- Forbidden namespaces: `system.*`, `platform.*`
- Validation in `identity/permission/service.go::Create`

**Status**: ✅ Enforced

---

## 6. UI Permissions are Advisory; Backend is Authoritative

**Invariant**: Frontend permission checks are **advisory only**. Backend **always** enforces permissions.

**Rationale**:
- Frontend can be bypassed
- Backend is the source of truth
- Security must be enforced server-side

**Enforcement**:
- All API endpoints check permissions
- Frontend filters UI based on permissions (UX only)
- Backend returns 403 Forbidden if permission missing

**Status**: ✅ Enforced

---

## 7. `tenant_owner` Must Always Exist

**Invariant**: Every tenant **must** have at least one user with `tenant_owner` role.

**Rationale**:
- Prevents tenant self-lockout
- Ensures recovery path exists
- Maintains tenant operability

**Enforcement**:
- ⚠️ TODO - Add validation to prevent removal of last `tenant_owner`
- Auto-assigned to first user in tenant
- Cannot be deleted (system role)

**Status**: ⚠️ Partially Enforced (needs safeguard)

---

## 8. System Roles are Immutable

**Invariant**: System roles (`is_system = true`) **cannot** be deleted or modified.

**Rationale**:
- Prevents privilege erosion
- Maintains system integrity
- Prevents accidental deletion of critical roles

**Enforcement**:
- `identity/role/service.go::Delete` checks `is_system`
- `identity/role/service.go::Update` checks `is_system`
- Returns error if attempting to modify system role

**Status**: ✅ Enforced

---

## 9. Permission Evolution Strategy

**Invariant**: When new permissions are introduced, they are automatically attached to `tenant_owner`.

**Rationale**:
- Maintains "owner has all permissions" guarantee
- No migration surprises
- No accidental lockouts

**Enforcement**:
- `identity/permission/service.go::Create` calls `AttachAllPermissionsToTenantOwner()`
- Other roles remain unchanged (explicit assignment required)

**Status**: ✅ Enforced

---

## 10. Explicit Permission Assignment

**Invariant**: All permissions are **explicitly** assigned. No implicit or inherited permissions.

**Rationale**:
- Auditability
- Predictability
- Security clarity

**Enforcement**:
- All role-permission assignments are explicit
- No wildcard matching
- No implicit inheritance (for now)

**Status**: ✅ Enforced

---

## 🔍 Verification

To verify these invariants are maintained:

1. **Code Review**: Check that all permission checks are explicit
2. **Tests**: Add negative security tests (attempt privilege escalation)
3. **Audits**: Regular security audits to verify invariants
4. **Monitoring**: Log all permission changes and role modifications

---

## 🚨 Break-Glass Procedures

If an invariant is violated:

1. **Immediate**: Disable affected tenant/system
2. **Investigation**: Audit logs to determine scope
3. **Recovery**: Restore from backup or manual fix
4. **Prevention**: Update validation to prevent recurrence

---

## 📝 Maintenance

These invariants should be:
- Reviewed in every security audit
- Tested in CI/CD pipeline
- Documented in onboarding materials
- Referenced in architecture decisions

---

**Last Updated**: 2025-01-10  
**Status**: Active  
**Owner**: Security Team

