# Break-Glass Procedures

## Overview

Break-glass procedures are emergency administrative actions that bypass normal security controls. These should only be used in critical situations where normal operations cannot proceed.

---

## 🔴 Critical Break-Glass Scenarios

### 1. Last `tenant_owner` Lockout

**Scenario**: All users with `tenant_owner` role are locked out or removed.

**Normal Prevention**: System prevents removal of last `tenant_owner` (enforced in `api/handlers/role_handler.go`)

**Break-Glass Procedure**:

```sql
-- 1. Identify the tenant
SELECT id, name, domain FROM tenants WHERE domain = 'tenant-domain.local';

-- 2. Find tenant_owner role
SELECT id, name FROM roles WHERE tenant_id = '<tenant-id>' AND name = 'tenant_owner';

-- 3. Find a user in the tenant (or create emergency user)
SELECT id, username, email FROM users WHERE tenant_id = '<tenant-id>' LIMIT 1;

-- 4. Directly assign tenant_owner role (bypasses validation)
INSERT INTO user_roles (id, user_id, role_id, assigned_at)
VALUES (gen_random_uuid(), '<user-id>', '<tenant-owner-role-id>', NOW())
ON CONFLICT (user_id, role_id) DO NOTHING;
```

**Verification**:
```sql
-- Verify tenant_owner is assigned
SELECT u.username, r.name 
FROM users u
JOIN user_roles ur ON u.id = ur.user_id
JOIN roles r ON ur.role_id = r.id
WHERE u.tenant_id = '<tenant-id>' AND r.name = 'tenant_owner';
```

**Post-Procedure**:
- Document the incident
- Review audit logs
- Update procedures if needed
- Notify security team

---

### 2. System Admin Lockout

**Scenario**: All system administrators are locked out.

**Break-Glass Procedure**:

```sql
-- 1. Find system_owner role
SELECT id FROM system_roles WHERE name = 'system_owner';

-- 2. Find or create system user
SELECT id, username FROM users WHERE principal_type = 'SYSTEM' LIMIT 1;

-- 3. Directly assign system_owner role
INSERT INTO user_system_roles (user_id, role_id, assigned_by, assigned_at)
VALUES ('<user-id>', '<system-owner-role-id>', '<user-id>', NOW())
ON CONFLICT (user_id, role_id) DO NOTHING;
```

**Verification**:
```sql
-- Verify system_owner is assigned
SELECT u.username, sr.name
FROM users u
JOIN user_system_roles usr ON u.id = usr.user_id
JOIN system_roles sr ON usr.role_id = sr.id
WHERE u.principal_type = 'SYSTEM' AND sr.name = 'system_owner';
```

---

### 3. Tenant Initialization Failure

**Scenario**: Tenant created but roles/permissions not initialized.

**Break-Glass Procedure**:

```go
// Use tenant initializer directly
initializer := tenant.NewInitializer(roleRepo, permissionRepo)
result, err := initializer.InitializeTenant(ctx, tenantID)
```

Or via SQL (manual):

```sql
-- 1. Create tenant_owner role
INSERT INTO roles (id, tenant_id, name, description, is_system, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    '<tenant-id>',
    'tenant_owner',
    'Full tenant ownership and control',
    true,
    NOW(),
    NOW()
) RETURNING id;

-- 2. Create other roles similarly
-- 3. Create permissions
-- 4. Assign permissions to roles
-- (See identity/tenant/initializer.go for full logic)
```

**Recommended**: Use the Go initializer rather than manual SQL.

---

### 4. Permission Namespace Violation Recovery

**Scenario**: Tenant created permissions in forbidden namespace (`system.*`, `platform.*`).

**Break-Glass Procedure**:

```sql
-- 1. Find violating permissions
SELECT id, resource, action, tenant_id 
FROM permissions 
WHERE tenant_id = '<tenant-id>' 
  AND (resource LIKE 'system.%' OR resource LIKE 'platform.%');

-- 2. Delete or rename violating permissions
DELETE FROM permissions 
WHERE tenant_id = '<tenant-id>' 
  AND (resource LIKE 'system.%' OR resource LIKE 'platform.%');

-- Or rename to allowed namespace:
UPDATE permissions 
SET resource = REPLACE(resource, 'system.', 'tenant.')
WHERE tenant_id = '<tenant-id>' AND resource LIKE 'system.%';
```

**Post-Procedure**:
- Review why validation was bypassed
- Fix validation if needed
- Audit all tenant permissions

---

## 🛡️ Security Safeguards

### Before Break-Glass

1. **Verify the emergency**: Confirm normal procedures cannot resolve
2. **Document the situation**: Record why break-glass is needed
3. **Get approval**: If possible, get security team approval
4. **Backup**: Take database backup before making changes

### During Break-Glass

1. **Use direct database access**: Bypass API validation
2. **Minimal changes**: Only make necessary changes
3. **Document actions**: Record all SQL commands executed
4. **Time-stamp**: Note when break-glass was used

### After Break-Glass

1. **Verify fix**: Confirm the issue is resolved
2. **Audit review**: Review audit logs for anomalies
3. **Incident report**: Document the incident
4. **Prevention**: Update procedures to prevent recurrence
5. **Notify**: Inform security team and stakeholders

---

## 📋 Break-Glass Checklist

- [ ] Emergency confirmed (normal procedures exhausted)
- [ ] Database backup taken
- [ ] Incident documented
- [ ] Approval obtained (if possible)
- [ ] Break-glass procedure executed
- [ ] Changes verified
- [ ] Audit logs reviewed
- [ ] Incident report created
- [ ] Prevention measures updated
- [ ] Security team notified

---

## 🚨 When NOT to Use Break-Glass

**Do NOT use break-glass for**:
- Routine operations
- Testing
- Development
- Convenience
- Bypassing normal approval processes

**Only use when**:
- Tenant is completely locked out
- System is non-functional
- No other recovery path exists
- Emergency situation confirmed

---

## 📝 Documentation Requirements

All break-glass procedures must be documented with:

1. **Timestamp**: When break-glass was used
2. **Reason**: Why break-glass was necessary
3. **Actions**: What was done
4. **Verification**: How success was confirmed
5. **Follow-up**: What prevention measures were added

---

## 🔍 Audit Trail

All break-glass actions should be:
- Logged in audit system
- Tagged with `break-glass` flag
- Reviewed within 24 hours
- Included in security reports

---

**Last Updated**: 2025-01-10  
**Status**: Active  
**Owner**: Security Team

