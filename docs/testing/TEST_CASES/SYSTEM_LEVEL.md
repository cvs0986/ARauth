# System Level Test Cases

## Overview

This document contains test cases for system-level features including bootstrap, system owner creation, system roles, system permissions, and system capabilities.

## Test Case 1: System Bootstrap - Master User Creation

### Feature Name
System Bootstrap - Master User Creation

### Feature Source
- **File**: `cmd/bootstrap/bootstrap_service.go`
- **Module**: `bootstrap`
- **Endpoint**: N/A (automatic on server start)

### Why This Feature Exists
The system needs an initial administrator (system owner) to manage the platform. Bootstrap creates this user automatically on first startup, allowing initial system administration.

### Preconditions
1. Database is empty (no users exist)
2. Migrations have been applied
3. Bootstrap is enabled in configuration
4. Bootstrap password is set via environment variable

### Step-by-Step Test Execution

#### Step 1: Configure Bootstrap
```bash
export BOOTSTRAP_ENABLED=true
export BOOTSTRAP_FORCE=false
export BOOTSTRAP_USERNAME=admin
export BOOTSTRAP_EMAIL=admin@arauth.io
export BOOTSTRAP_PASSWORD=AdminPassword123!
export BOOTSTRAP_FIRST_NAME=System
export BOOTSTRAP_LAST_NAME=Administrator
```

#### Step 2: Verify Database is Empty
```bash
psql -h localhost -U iam_user -d iam -c "SELECT COUNT(*) FROM users;"
# Expected: 0
```

#### Step 3: Start Server
```bash
go run cmd/server/main.go
```

#### Step 4: Verify Bootstrap User Created
```bash
psql -h localhost -U iam_user -d iam <<EOF
SELECT id, username, email, principal_type, tenant_id, status 
FROM users 
WHERE principal_type = 'SYSTEM';
EOF
```

**Expected Output:**
```
                  id                  | username |      email       | principal_type | tenant_id | status
--------------------------------------+----------+------------------+----------------+-----------+--------
 <uuid>                               | admin    | admin@arauth.io  | SYSTEM         | NULL      | active
```

#### Step 5: Verify System Owner Role Assigned
```bash
psql -h localhost -U iam_user -d iam <<EOF
SELECT u.username, sr.name as role_name
FROM users u
JOIN user_system_roles usr ON u.id = usr.user_id
JOIN system_roles sr ON usr.role_id = sr.id
WHERE u.username = 'admin';
EOF
```

**Expected Output:**
```
 username |   role_name
----------+---------------
 admin    | system_owner
```

### Expected Functional Behavior
1. System owner user is created with:
   - Username: `admin`
   - Email: `admin@arauth.io`
   - Principal Type: `SYSTEM`
   - Tenant ID: `NULL`
   - Status: `active`
2. Password is hashed and stored securely
3. `system_owner` role is automatically assigned
4. User can login without tenant context

### Expected Security Behavior
1. Password is hashed (not stored in plaintext)
2. User has `SYSTEM` principal type (not `TENANT`)
3. User has no tenant association
4. User has `system_owner` role with all system permissions

### Negative / Abuse Test Cases

#### Test 1.1: Bootstrap Without Password
```bash
unset BOOTSTRAP_PASSWORD
go run cmd/server/main.go
# Expected: Server fails to start or bootstrap fails with error
```

#### Test 1.2: Bootstrap When User Exists
```bash
# Run bootstrap once (creates user)
go run cmd/server/main.go
# Stop server

# Run bootstrap again (user exists)
go run cmd/server/main.go
# Expected: Bootstrap skipped (unless BOOTSTRAP_FORCE=true)
```

#### Test 1.3: Force Re-bootstrap
```bash
export BOOTSTRAP_FORCE=true
go run cmd/server/main.go
# Expected: Existing user deleted and recreated
```

### Audit Events Expected
- **Event Type**: `user.created`
- **Actor**: System (bootstrap process)
- **Target**: Created user
- **Result**: `success`
- **Metadata**: Should include bootstrap flag

**Verification:**
```bash
# After bootstrap, query audit events
curl -X GET "http://localhost:8080/system/audit/events?event_type=user.created&limit=1" \
  -H "Authorization: Bearer $SYSTEM_TOKEN" | jq '.'
```

### Recovery / Rollback Behavior
- If bootstrap fails, server should not start
- If user creation fails, no partial user should exist
- If role assignment fails, user should be rolled back

### Pass / Fail Criteria
- ✅ User created with correct attributes
- ✅ Password is hashed
- ✅ System owner role assigned
- ✅ User can login
- ✅ Audit event created
- ❌ Fail if user creation fails
- ❌ Fail if password is plaintext
- ❌ Fail if role not assigned

---

## Test Case 2: System Roles Verification

### Feature Name
System Roles - Predefined Roles Existence

### Feature Source
- **File**: `migrations/000014_create_system_roles.up.sql`
- **Module**: `migrations`
- **Endpoint**: `GET /system/roles`

### Why This Feature Exists
System roles define what SYSTEM users can do. Three predefined roles provide different levels of system access: owner (full control), admin (tenant management), and auditor (read-only).

### Preconditions
1. Migrations have been applied
2. System owner exists (from Test Case 1)
3. System owner can authenticate

### Step-by-Step Test Execution

#### Step 1: Login as System Owner
```bash
SYSTEM_TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "AdminPassword123!"
  }' | jq -r '.access_token')

echo "Token: $SYSTEM_TOKEN"
```

#### Step 2: List System Roles
```bash
curl -X GET http://localhost:8080/system/roles \
  -H "Authorization: Bearer $SYSTEM_TOKEN" \
  -H "Content-Type: application/json" | jq '.'
```

**Expected Response:**
```json
[
  {
    "id": "00000000-0000-0000-0000-000000000001",
    "name": "system_owner",
    "description": "Full system ownership and control. Can manage all aspects of the platform."
  },
  {
    "id": "00000000-0000-0000-0000-000000000002",
    "name": "system_admin",
    "description": "System administration with tenant management capabilities. Cannot delete system or modify system owner."
  },
  {
    "id": "00000000-0000-0000-0000-000000000003",
    "name": "system_auditor",
    "description": "Read-only system access for auditing and compliance. Can view all system data but cannot modify."
  }
]
```

#### Step 3: Verify Roles in Database
```bash
psql -h localhost -U iam_user -d iam <<EOF
SELECT id, name, description FROM system_roles ORDER BY name;
EOF
```

**Expected:** 3 roles with correct names and descriptions

### Expected Functional Behavior
1. Three system roles exist:
   - `system_owner`: Full control
   - `system_admin`: Tenant management
   - `system_auditor`: Read-only
2. Roles are non-deletable (enforced by application logic)
3. Roles can be queried via API

### Expected Security Behavior
1. Only SYSTEM users can access `/system/roles`
2. TENANT users cannot access system roles
3. Roles are immutable (cannot be modified)

### Negative / Abuse Test Cases

#### Test 2.1: Access Without Authentication
```bash
curl -X GET http://localhost:8080/system/roles
# Expected: 401 Unauthorized
```

#### Test 2.2: Access as TENANT User
```bash
# Create tenant user and login
TENANT_TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "tenant_user",
    "password": "Password123!",
    "tenant_id": "<tenant_id>"
  }' | jq -r '.access_token')

curl -X GET http://localhost:8080/system/roles \
  -H "Authorization: Bearer $TENANT_TOKEN"
# Expected: 403 Forbidden (not a SYSTEM user)
```

### Audit Events Expected
- No audit events for read-only operations (role listing)

### Recovery / Rollback Behavior
- Roles are created by migrations, cannot be rolled back without migration rollback
- If migration fails, roles don't exist and system is unusable

### Pass / Fail Criteria
- ✅ Three roles exist
- ✅ Roles have correct names and descriptions
- ✅ API returns roles correctly
- ✅ Unauthorized access blocked
- ❌ Fail if roles missing
- ❌ Fail if unauthorized access allowed

---

## Test Case 3: System Permissions Verification

### Feature Name
System Permissions - Predefined Permissions Existence

### Feature Source
- **File**: `migrations/000014_create_system_roles.up.sql`
- **Module**: `migrations`
- **Endpoint**: `GET /system/permissions`

### Why This Feature Exists
System permissions define what actions SYSTEM users can perform. Permissions are assigned to roles, and roles are assigned to users.

### Preconditions
1. Migrations have been applied
2. System owner exists
3. System owner can authenticate

### Step-by-Step Test Execution

#### Step 1: Login as System Owner
```bash
SYSTEM_TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "AdminPassword123!"
  }' | jq -r '.access_token')
```

#### Step 2: List System Permissions
```bash
curl -X GET http://localhost:8080/system/permissions \
  -H "Authorization: Bearer $SYSTEM_TOKEN" \
  -H "Content-Type: application/json" | jq '.'
```

**Expected Response:** Array of permissions including:
- `tenant:create`, `tenant:read`, `tenant:update`, `tenant:delete`
- `tenant:suspend`, `tenant:resume`, `tenant:configure`
- `system:settings`, `system:policy`, `system:audit`, `system:users`
- `billing:manage`, `billing:read`

#### Step 3: Verify Permissions in Database
```bash
psql -h localhost -U iam_user -d iam <<EOF
SELECT resource, action, description 
FROM system_permissions 
ORDER BY resource, action;
EOF
```

**Expected:** 13 permissions with correct resource:action format

#### Step 4: Verify Role-Permission Assignments
```bash
psql -h localhost -U iam_user -d iam <<EOF
SELECT sr.name as role_name, sp.resource, sp.action
FROM system_roles sr
JOIN system_role_permissions srp ON sr.id = srp.role_id
JOIN system_permissions sp ON srp.permission_id = sp.id
WHERE sr.name = 'system_owner'
ORDER BY sp.resource, sp.action;
EOF
```

**Expected:** `system_owner` has all 13 permissions

### Expected Functional Behavior
1. 13 system permissions exist
2. Permissions follow `resource:action` format
3. `system_owner` has all permissions
4. `system_admin` has subset of permissions
5. `system_auditor` has read-only permissions

### Expected Security Behavior
1. Only SYSTEM users can access system permissions
2. Permissions cannot be modified via API (immutable)
3. Permission checks enforce authorization

### Negative / Abuse Test Cases

#### Test 3.1: Unauthorized Access
```bash
curl -X GET http://localhost:8080/system/permissions
# Expected: 401 Unauthorized
```

#### Test 3.2: Modify Permission (Should Fail)
```bash
# Attempt to update permission (if endpoint exists)
# Expected: 405 Method Not Allowed or 403 Forbidden
```

### Audit Events Expected
- No audit events for read-only operations

### Recovery / Rollback Behavior
- Permissions are created by migrations
- Cannot be modified without migration

### Pass / Fail Criteria
- ✅ 13 permissions exist
- ✅ Permissions have correct format
- ✅ Role-permission assignments correct
- ✅ API returns permissions
- ❌ Fail if permissions missing
- ❌ Fail if unauthorized access allowed

---

## Test Case 4: System Capabilities Verification

### Feature Name
System Capabilities - Predefined Capabilities Existence

### Feature Source
- **File**: `migrations/000019_create_system_capabilities.up.sql`
- **Module**: `migrations`, `identity/capability`
- **Endpoint**: `GET /system/capabilities`

### Why This Feature Exists
System capabilities define what features are supported by the platform. Capabilities flow from system → tenant → user, controlling feature availability.

### Preconditions
1. Migrations have been applied
2. System owner exists
3. System owner can authenticate

### Step-by-Step Test Execution

#### Step 1: Login as System Owner
```bash
SYSTEM_TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "AdminPassword123!"
  }' | jq -r '.access_token')
```

#### Step 2: List System Capabilities
```bash
curl -X GET http://localhost:8080/system/capabilities \
  -H "Authorization: Bearer $SYSTEM_TOKEN" \
  -H "Content-Type: application/json" | jq '.'
```

**Expected Response:** Array of capabilities including:
- `mfa`, `totp`, `saml`, `oidc`, `oauth2`
- `passwordless`, `ldap`
- `max_token_ttl`, `allowed_grant_types`, `allowed_scope_namespaces`, `pkce_mandatory`

#### Step 3: Verify Capabilities in Database
```bash
psql -h localhost -U iam_user -d iam <<EOF
SELECT capability_key, enabled, description 
FROM system_capabilities 
ORDER BY capability_key;
EOF
```

**Expected:** 11 capabilities with correct keys

#### Step 4: Get Specific Capability
```bash
curl -X GET http://localhost:8080/system/capabilities/mfa \
  -H "Authorization: Bearer $SYSTEM_TOKEN" | jq '.'
```

**Expected Response:**
```json
{
  "capability_key": "mfa",
  "enabled": true,
  "default_value": {},
  "description": "Multi-factor authentication support"
}
```

### Expected Functional Behavior
1. 11 system capabilities exist
2. Capabilities have `enabled` flag
3. Capabilities have `default_value` (JSON)
4. Capabilities can be queried individually

### Expected Security Behavior
1. Only SYSTEM users can view system capabilities
2. Only SYSTEM users with `system:configure` permission can update capabilities
3. Capabilities control feature availability

### Negative / Abuse Test Cases

#### Test 4.1: Unauthorized Access
```bash
curl -X GET http://localhost:8080/system/capabilities
# Expected: 401 Unauthorized
```

#### Test 4.2: Update Without Permission
```bash
# Attempt to update capability without system:configure permission
# Expected: 403 Forbidden
```

### Audit Events Expected
- No audit events for read operations
- `system.capability.updated` event when capability is updated

### Recovery / Rollback Behavior
- Capabilities are created by migrations
- Updates are logged in audit

### Pass / Fail Criteria
- ✅ 11 capabilities exist
- ✅ Capabilities have correct structure
- ✅ API returns capabilities
- ✅ Individual capability query works
- ❌ Fail if capabilities missing
- ❌ Fail if unauthorized access allowed

---

## Test Case 5: System Owner Login

### Feature Name
System Owner Authentication

### Feature Source
- **File**: `api/handlers/auth_handler.go`
- **Module**: `auth/login`
- **Endpoint**: `POST /api/v1/auth/login`

### Why This Feature Exists
System owner needs to authenticate to access system APIs. SYSTEM users login without tenant context.

### Preconditions
1. System owner exists (from Test Case 1)
2. Server is running

### Step-by-Step Test Execution

#### Step 1: Login as System Owner (No Tenant)
```bash
RESPONSE=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "AdminPassword123!"
  }')

echo "$RESPONSE" | jq '.'

TOKEN=$(echo "$RESPONSE" | jq -r '.access_token')
REFRESH_TOKEN=$(echo "$RESPONSE" | jq -r '.refresh_token')
```

**Expected Response:**
```json
{
  "access_token": "<jwt_token>",
  "refresh_token": "<opaque_token>",
  "token_type": "Bearer",
  "expires_in": 900,
  "principal_type": "SYSTEM"
}
```

#### Step 2: Verify Token Claims
```bash
# Decode JWT (using jwt.io or jq if available)
# Token should contain:
# - sub: user ID
# - principal_type: "SYSTEM"
# - tenant_id: null or absent
# - system_roles: ["system_owner"]
# - system_permissions: [all permissions]
```

#### Step 3: Access System API
```bash
curl -X GET http://localhost:8080/system/tenants \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

**Expected:** List of tenants (empty initially)

### Expected Functional Behavior
1. Login succeeds with correct credentials
2. Access token is JWT format
3. Refresh token is opaque
4. Token contains SYSTEM principal type
5. Token contains system roles and permissions
6. Token can be used to access system APIs

### Expected Security Behavior
1. Password is verified securely
2. Token is signed with JWT secret
3. Token expires after configured TTL
4. Invalid credentials return 401
5. Token required for system API access

### Negative / Abuse Test Cases

#### Test 5.1: Invalid Password
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "WrongPassword123!"
  }'
# Expected: 401 Unauthorized
```

#### Test 5.2: Invalid Username
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "nonexistent",
    "password": "AdminPassword123!"
  }'
# Expected: 401 Unauthorized
```

#### Test 5.3: Expired Token
```bash
# Use token after expiration
# Expected: 401 Unauthorized
```

### Audit Events Expected
- **Event Type**: `login.success` or `login.failure`
- **Actor**: User attempting login
- **Result**: `success` or `failure`
- **Metadata**: IP address, user agent

**Verification:**
```bash
curl -X GET "http://localhost:8080/system/audit/events?event_type=login.success&limit=1" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

### Recovery / Rollback Behavior
- Failed login does not create session
- Rate limiting prevents brute force
- Account lockout after multiple failures (if implemented)

### Pass / Fail Criteria
- ✅ Login succeeds with correct credentials
- ✅ Token is valid JWT
- ✅ Token contains correct claims
- ✅ System API access works
- ✅ Invalid credentials rejected
- ✅ Audit events created
- ❌ Fail if invalid credentials accepted
- ❌ Fail if token missing claims

---

## Summary

### Test Execution Order
1. Test Case 1: System Bootstrap
2. Test Case 2: System Roles
3. Test Case 3: System Permissions
4. Test Case 4: System Capabilities
5. Test Case 5: System Owner Login

### Dependencies
- Test Case 1 must run first (creates system owner)
- Test Cases 2-4 can run in any order
- Test Case 5 requires Test Case 1 (needs system owner)

### Common Issues
1. **Bootstrap fails**: Check password is set, database is empty
2. **Roles missing**: Verify migrations applied
3. **Login fails**: Check credentials, user exists
4. **API access fails**: Check token, permissions

### Next Steps
After completing system-level tests, proceed to:
- `TENANT_LIFECYCLE.md`: Tenant creation and management
- `AUTHENTICATION.md`: Full authentication flows

