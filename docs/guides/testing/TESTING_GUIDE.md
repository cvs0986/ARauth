# Complete Testing Guide - ARauth Identity IAM (V2 - Master Tenant Architecture)

## 📋 Prerequisites

- ✅ PostgreSQL running on `127.0.0.1:5433`
- ✅ Database `iam` created
- ✅ User: `dcim_user`, Password: `dcim_password`
- ✅ **All migrations applied (version 16/16)** ✅
- ✅ Go 1.21+ installed
- ✅ Node.js 18+ and npm installed
- ✅ Frontend dependencies installed

### Verify Database Setup

```bash
# Check PostgreSQL connection
PGPASSWORD=dcim_password psql -h 127.0.0.1 -p 5433 -U dcim_user -d iam -c "SELECT version();"

# Verify migrations are applied (should show version 16)
cd /home/eshwar/Documents/Veer/nuage-indentity
export DATABASE_URL="postgres://dcim_user:dcim_password@127.0.0.1:5433/iam?sslmode=disable"
migrate -path migrations -database "$DATABASE_URL" version
```

**Expected**: Should show `16` (all migrations applied)

---

## 🏗️ Architecture Overview (V2 - Master Tenant)

ARauth Identity now supports a **two-plane architecture**:

- **SYSTEM Users** (Master/Platform Admins):
  - `principal_type = 'SYSTEM'`
  - `tenant_id = NULL`
  - Can manage all tenants
  - Can create tenant admins
  - Access to `/system/*` API endpoints
  - System-level roles and permissions

- **TENANT Users** (Tenant Admins):
  - `principal_type = 'TENANT'`
  - `tenant_id = <tenant-uuid>`
  - Can only manage their own tenant
  - Access to `/api/v1/*` API endpoints
  - Tenant-level roles and permissions

---

## 🚀 Step 1: Start Backend Server

### Option A: Using the Start Script (Recommended)

```bash
cd /home/eshwar/Documents/Veer/nuage-indentity
./scripts/start-backend-local.sh
```

### Option B: Manual Start

```bash
cd /home/eshwar/Documents/Veer/nuage-indentity

# Set environment variables
export DATABASE_HOST=127.0.0.1
export DATABASE_PORT=5433
export DATABASE_USER=dcim_user
export DATABASE_PASSWORD=dcim_password
export DATABASE_NAME=iam
export DATABASE_SSL_MODE=disable
export JWT_SECRET=test-jwt-secret-key-min-32-characters-long
export ENCRYPTION_KEY=01234567890123456789012345678901

# Start server
go run cmd/server/main.go
```

### Verify Backend is Running

```bash
# Health check
curl http://localhost:8080/health

# Expected response:
# {"status":"healthy","timestamp":"...","version":"0.1.0","checks":{"database":"healthy","redis":"not_configured"}}
```

**✅ Backend Status**: Server should be running on `http://localhost:8080`

---

## 👑 Step 2: Bootstrap Master User (SYSTEM Admin)

The master user is a SYSTEM-level admin who can manage all tenants. You can create it via:

### Option A: Using Config File

1. Create or edit `config/bootstrap.yaml`:

```yaml
bootstrap:
  enabled: true
  master_user:
    username: "system_admin"
    email: "admin@arauth.local"
    password: "SystemAdmin@123456"
    first_name: "System"
    last_name: "Administrator"
  master_role:
    name: "system_owner"
    assign_all_permissions: true
```

2. Start the server (it will auto-bootstrap on first run)

### Option B: Using Environment Variables

```bash
export BOOTSTRAP_ENABLED=true
export BOOTSTRAP_MASTER_USERNAME=system_admin
export BOOTSTRAP_MASTER_EMAIL=admin@arauth.local
export BOOTSTRAP_MASTER_PASSWORD=SystemAdmin@123456
export BOOTSTRAP_MASTER_FIRST_NAME=System
export BOOTSTRAP_MASTER_LAST_NAME=Administrator
export BOOTSTRAP_MASTER_ROLE=system_owner

# Start server
go run cmd/server/main.go
```

### Option C: Using Bootstrap CLI (if available)

```bash
go run cmd/bootstrap/main.go
```

### Verify Master User Created

```bash
# Check if SYSTEM user exists
PGPASSWORD=dcim_password psql -h 127.0.0.1 -p 5433 -U dcim_user -d iam -c \
  "SELECT id, username, email, principal_type, tenant_id FROM users WHERE principal_type = 'SYSTEM';"
```

**✅ Expected**: Should show the master user with `principal_type = 'SYSTEM'` and `tenant_id = NULL`

---

## 🎨 Step 3: Start Frontend Applications

### Terminal 1: Admin Dashboard

```bash
cd /home/eshwar/Documents/Veer/nuage-indentity/frontend/admin-dashboard
npm run dev
```

**✅ Admin Dashboard**: Should be available at `http://localhost:5173`

### Terminal 2: E2E Testing App

```bash
cd /home/eshwar/Documents/Veer/nuage-indentity/frontend/e2e-test-app
npm run dev
```

**✅ E2E Test App**: Should be available at `http://localhost:5174`

---

## 🧪 Step 4: Complete Testing Workflow

### Phase 1: SYSTEM User Testing (Master Admin)

#### 4.1 Login as SYSTEM Admin

1. Open browser: `http://localhost:5173`
2. Enter credentials:
   - **Username**: `system_admin` (or your bootstrap username)
   - **Password**: `SystemAdmin@123456` (or your bootstrap password)
   - **Tenant ID**: (leave empty - SYSTEM users don't need tenant ID)
3. Click **Login**
4. **✅ Expected**: 
   - Redirected to Dashboard
   - Header shows "System Admin" badge
   - Tenant selector dropdown visible in header
   - Sidebar shows: Dashboard, Tenants, Users, Roles, Permissions, Audit Logs, Settings

#### 4.2 SYSTEM User - Dashboard

**Test**:
- [ ] View statistics cards (Tenants, Users, Roles, Permissions)
- [ ] Check "Tenants" card shows total tenant count
- [ ] Check "Users" card shows total user count (all tenants)
- [ ] Click "View all" links navigate correctly
- [ ] Quick actions section shows "Manage Tenants" and "System Settings"
- [ ] System overview shows "System Status: Operational"

**✅ Expected**: Dashboard displays with system-wide statistics

#### 4.3 SYSTEM User - Tenant Management

**Navigate**: Click "Tenants" in sidebar

**Test Create Tenant**:
1. Click **"Create Tenant"** button
2. Fill form:
   - Name: `Acme Corp`
   - Domain: `acme.local`
   - Status: `Active`
3. Click **"Create"**
4. **✅ Expected**: Tenant appears in list

**Test Tenant Operations**:
- [ ] Search by name: Type "Acme" → Should filter results
- [ ] Search by domain: Type "acme.local" → Should filter results
- [ ] Filter by status: Select "Active" → Should show only active tenants
- [ ] Edit tenant: Change name or status → Changes saved
- [ ] Suspend tenant: Click "Suspend" → Tenant status changes to "suspended"
- [ ] Resume tenant: Click "Resume" → Tenant status changes to "active"
- [ ] Delete tenant: Click "Delete" → Tenant removed

**Test Tenant Selector**:
- [ ] Click tenant selector in header
- [ ] Select a tenant from dropdown
- [ ] Verify tenant context is selected
- [ ] Select "All Tenants (System View)" → Returns to system view

#### 4.4 SYSTEM User - User Management

**Navigate**: Click "Users" in sidebar

**Test Without Tenant Selected**:
- [ ] Should show message: "Please select a tenant from the header to view and manage users."

**Test With Tenant Selected**:
1. Select a tenant from header dropdown
2. Click "Users" in sidebar
3. **✅ Expected**: Users list shows users for selected tenant

**Test Create User for Tenant**:
1. Select a tenant from header
2. Click **"Create User"** button
3. Fill form:
   - Username: `tenant_admin`
   - Email: `admin@acme.local`
   - Password: `TenantAdmin@123456`
   - First Name: `Tenant`
   - Last Name: `Admin`
4. Click **"Create"**
5. **✅ Expected**: User created for selected tenant

**Test User Operations**:
- [ ] Search by username, email, or name
- [ ] Filter by status (Active, Inactive, Locked)
- [ ] Edit user details
- [ ] Delete user
- [ ] Pagination works
- [ ] Table shows "Tenant" column with tenant ID

#### 4.5 SYSTEM User - Settings

**Navigate**: Click "Settings" in sidebar

**Test System Settings Tab**:
- [ ] View "System Settings" tab
- [ ] Configure JWT settings (Issuer, Audience)
- [ ] Modify session timeout
- [ ] Configure account lockout (max attempts, lockout duration)
- [ ] Click **"Save System Settings"**
- [ ] **✅ Expected**: Success message displayed

**Test Security Tab**:
- [ ] View "Security" tab
- [ ] Modify password policy (min length, requirements)
- [ ] Configure MFA requirements
- [ ] Set rate limiting values
- [ ] Click **"Save Security Settings"**
- [ ] **✅ Expected**: Success message displayed

**Test OAuth2/OIDC Tab**:
- [ ] View "OAuth2/OIDC" tab
- [ ] Configure Hydra endpoints
- [ ] Modify token TTLs
- [ ] Click **"Save OAuth Settings"**
- [ ] **✅ Expected**: Success message displayed

**Test Tenant Settings Tab**:
1. Select a tenant from header dropdown
2. Click "Settings" → "Tenant Settings" tab
3. **✅ Expected**: Shows message if no tenant selected
4. With tenant selected:
   - [ ] View token lifetime settings
   - [ ] Modify Access Token TTL, Refresh Token TTL, ID Token TTL
   - [ ] Configure "Remember Me" settings
   - [ ] Toggle token rotation
   - [ ] Set MFA requirement for extended sessions
   - [ ] Click **"Save Tenant Settings"**
   - [ ] **✅ Expected**: Success message, settings saved for selected tenant

---

### Phase 2: TENANT User Testing

#### 5.1 Create Tenant Admin User

**As SYSTEM User**:
1. Select a tenant from header
2. Navigate to "Users"
3. Create a user with admin role (or assign admin role later)

**Or via API**:
```bash
# Get tenant ID first
TENANT_ID="your-tenant-id-here"

# Create tenant admin user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "username": "tenant_admin",
    "email": "admin@acme.local",
    "password": "TenantAdmin@123456",
    "first_name": "Tenant",
    "last_name": "Admin"
  }'
```

#### 5.2 Login as TENANT User

1. Open browser: `http://localhost:5173`
2. Enter credentials:
   - **Username**: `tenant_admin`
   - **Password**: `TenantAdmin@123456`
   - **Tenant ID**: (the tenant ID - required for TENANT users)
3. Click **Login**
4. **✅ Expected**: 
   - Redirected to Dashboard
   - Header shows "Tenant Admin" badge
   - **No tenant selector** in header (locked to their tenant)
   - Sidebar shows: Dashboard, Users, Roles, Permissions, Audit Logs, Settings

#### 5.3 TENANT User - Dashboard

**Test**:
- [ ] View statistics cards (Users, Roles, Permissions)
- [ ] **No "Tenants" card** (TENANT users can't see other tenants)
- [ ] Check "Users" card shows only their tenant's users
- [ ] Quick actions section shows tenant-specific actions only
- [ ] System overview shows "Tenant Overview: [Tenant Name]"

**✅ Expected**: Dashboard displays with tenant-scoped statistics

#### 5.4 TENANT User - User Management

**Navigate**: Click "Users" in sidebar

**Test Create User**:
1. Click **"Create User"** button
2. Fill form (same as SYSTEM user)
3. Click **"Create"**
4. **✅ Expected**: User created for their tenant (tenant_id automatically set)

**Test User Operations**:
- [ ] Search by username, email, or name
- [ ] Filter by status
- [ ] Edit user details
- [ ] Delete user
- [ ] **No "Tenant" column** in table (all users are from same tenant)
- [ ] Pagination works

#### 5.5 TENANT User - Settings

**Navigate**: Click "Settings" in sidebar

**Test Token Settings Tab**:
- [ ] Only "Token Settings" tab visible (no System, Security, OAuth tabs)
- [ ] View token lifetime settings
- [ ] Modify Access Token TTL, Refresh Token TTL, ID Token TTL
- [ ] Configure "Remember Me" settings
- [ ] Toggle token rotation
- [ ] Set MFA requirement for extended sessions
- [ ] Click **"Save Token Settings"**
- [ ] **✅ Expected**: Success message, settings saved for their tenant

**Note**: TENANT users can only configure token settings for their own tenant

---

### Phase 3: Role & Permission Testing

#### 6.1 Create Roles and Permissions

**As SYSTEM or TENANT User**:

**Navigate**: Click "Roles" in sidebar

**Test Create Role**:
1. Click **"Create Role"** button
2. Fill form:
   - Name: `Developer`
   - Description: `Developer role with read/write permissions`
3. Click **"Create"**
4. **✅ Expected**: Role appears in list

**Navigate**: Click "Permissions" in sidebar

**Test Create Permissions**:
1. Click **"Create Permission"** button
2. Create common permissions:
   - Resource: `users`, Action: `read`
   - Resource: `users`, Action: `write`
   - Resource: `users`, Action: `delete`
   - Resource: `roles`, Action: `read`
   - Resource: `roles`, Action: `write`
3. **✅ Expected**: Permissions appear in list

#### 6.2 Assign Permissions to Roles

**Navigate**: Click "Roles" → Click "Manage Permissions" on a role

**Test**:
- [ ] View available permissions list
- [ ] Select permissions to assign
- [ ] Click **"Save"**
- [ ] **✅ Expected**: Permissions assigned to role

#### 6.3 Assign Roles to Users

**Navigate**: Click "Users" → Click "Edit" on a user

**Test** (if role assignment UI exists):
- [ ] View user's current roles
- [ ] Select roles to assign
- [ ] Click **"Save"**
- [ ] **✅ Expected**: Roles assigned to user

**Or via API**:
```bash
# Assign role to user (if API endpoint exists)
curl -X POST http://localhost:8080/api/v1/users/{user_id}/roles \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{"role_id": "role-uuid-here"}'
```

---

### Phase 4: E2E Testing App

#### 7.1 User Registration

**Navigate**: `http://localhost:5174`

**Test Registration**:
1. Click **"Register"** or navigate to `/register`
2. Fill registration form:
   - Username: `testuser`
   - Email: `testuser@acme.local`
   - Password: `Secure@123456`
   - Confirm Password: `Secure@123456`
   - First Name: `Test`
   - Last Name: `User`
   - Tenant ID: (use a tenant ID)
3. Click **"Register"**
4. **✅ Expected**: Success message and redirect to login

#### 7.2 Login Flow

**Test Login**:
1. Navigate to `/login`
2. Enter credentials:
   - Username: `testuser` (or `tenant_admin`)
   - Password: `Secure@123456` (or `TenantAdmin@123456`)
   - Tenant ID: (your tenant ID)
   - Remember Me: (optional checkbox)
3. Click **"Login"**
4. **✅ Expected**: 
   - Redirected to Dashboard
   - Access token and refresh token stored
   - User info displayed

**Test Invalid Credentials**:
- [ ] Enter wrong password → Should show error
- [ ] Enter wrong username → Should show error
- [ ] Enter wrong tenant ID → Should show error

#### 7.3 Token Refresh

**Test**:
- [ ] Wait for access token to expire (or manually expire it)
- [ ] Make API request with expired token
- [ ] **✅ Expected**: Token automatically refreshed using refresh token

#### 7.4 MFA Flow

**Navigate**: Click "Manage MFA" on Dashboard

**Test MFA Enrollment**:
1. Click **"Enroll in MFA"**
2. **✅ Expected**: QR code displayed
3. Scan QR code with authenticator app (Google Authenticator, Authy, etc.)
4. Enter the 6-digit code from app
5. Click **"Verify and Enable"**
6. **✅ Expected**: MFA enabled, recovery codes displayed

**Test MFA Verification**:
1. Logout and login again
2. After entering credentials, you should be prompted for MFA code
3. Enter code from authenticator app
4. Click **"Verify"**
5. **✅ Expected**: Successfully logged in

**Test MFA Disable**:
1. Go to MFA page
2. Click **"Disable MFA"**
3. Confirm
4. **✅ Expected**: MFA disabled

#### 7.5 Profile Management

**Navigate**: Click "Go to Profile" on Dashboard

**Test View Profile**:
- [ ] View user information (username, email, name)
- [ ] View status and tenant ID

**Test Edit Profile**:
1. Click **"Edit Profile"**
2. Modify first name or last name
3. Click **"Save"**
4. **✅ Expected**: Changes saved and displayed

**Test Change Password**:
1. Click **"Change Password"**
2. Enter:
   - Current Password: `Secure@123456`
   - New Password: `NewSecure@123456`
   - Confirm Password: `NewSecure@123456`
3. Click **"Change Password"**
4. **✅ Expected**: Password changed successfully

#### 7.6 Roles and Permissions View

**Navigate**: Click "View Roles & Permissions" on Dashboard

**Test View Roles**:
- [ ] View user information card
- [ ] View assigned roles section
- [ ] See role details (name, description, permissions)
- [ ] View all permissions from roles

**Test Permissions Display**:
- [ ] View permissions grid
- [ ] See resource:action format
- [ ] View permission descriptions

---

### Phase 5: Integration Testing

#### 8.1 Cross-App Testing

**Test Flow**:
1. **Admin Dashboard (SYSTEM)**: Create tenant
2. **Admin Dashboard (SYSTEM)**: Create user for tenant
3. **Admin Dashboard (SYSTEM/TENANT)**: Assign role to user
4. **E2E Test App**: Login with that user
5. **E2E Test App**: View roles and permissions
6. **✅ Expected**: User sees assigned roles and permissions

#### 8.2 Tenant Isolation Testing

**Test**:
1. **SYSTEM User**: Create Tenant A and Tenant B
2. **SYSTEM User**: Create users in both tenants
3. **TENANT User (Tenant A)**: Login and view users
4. **✅ Expected**: Only sees users from Tenant A
5. **TENANT User (Tenant B)**: Login and view users
6. **✅ Expected**: Only sees users from Tenant B

#### 8.3 SYSTEM vs TENANT Permission Testing

**Test**:
1. **SYSTEM User**: Try to access `/system/tenants` endpoint
2. **✅ Expected**: Success (has system permissions)
3. **TENANT User**: Try to access `/system/tenants` endpoint
4. **✅ Expected**: 403 Forbidden (no system permissions)

---

## 🔍 Step 5: API Testing

### Test SYSTEM API Endpoints

```bash
# Login as SYSTEM user
LOGIN_RESPONSE=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "system_admin",
    "password": "SystemAdmin@123456"
  }')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.access_token')

# List all tenants (SYSTEM endpoint)
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/system/tenants

# Create tenant (SYSTEM endpoint)
curl -X POST http://localhost:8080/system/tenants \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "New Tenant",
    "domain": "new.local",
    "status": "active"
  }'

# Get tenant settings (SYSTEM endpoint)
TENANT_ID="tenant-uuid-here"
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/system/tenants/$TENANT_ID/settings

# Update tenant settings (SYSTEM endpoint)
curl -X PUT http://localhost:8080/system/tenants/$TENANT_ID/settings \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "'$TENANT_ID'",
    "access_token_ttl_minutes": 30,
    "refresh_token_ttl_days": 60,
    "id_token_ttl_minutes": 120,
    "remember_me_enabled": true,
    "remember_me_refresh_token_ttl_days": 180,
    "remember_me_access_token_ttl_minutes": 120,
    "token_rotation_enabled": true,
    "require_mfa_for_extended_sessions": false
  }'
```

### Test TENANT API Endpoints

```bash
# Login as TENANT user
TENANT_ID="tenant-uuid-here"
LOGIN_RESPONSE=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "username": "tenant_admin",
    "password": "TenantAdmin@123456"
  }')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.access_token')

# List users (TENANT endpoint - automatically scoped to tenant)
curl -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID" \
  http://localhost:8080/api/v1/users

# Create user (TENANT endpoint)
curl -X POST http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "username": "newuser",
    "email": "newuser@acme.local",
    "password": "Secure@123456",
    "first_name": "New",
    "last_name": "User"
  }'

# List roles (TENANT endpoint)
curl -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID" \
  http://localhost:8080/api/v1/roles

# List permissions (TENANT endpoint)
curl -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID" \
  http://localhost:8080/api/v1/permissions
```

### Test Token Operations

```bash
# Refresh token
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\": \"$REFRESH_TOKEN\"}"

# Revoke token (logout)
curl -X POST http://localhost:8080/api/v1/auth/revoke \
  -H "Content-Type: application/json" \
  -d "{\"token\": \"$REFRESH_TOKEN\", \"token_type_hint\": \"refresh_token\"}"
```

---

## ✅ Testing Checklist Summary

### Backend
- [x] Server starts successfully
- [x] Database connection works
- [x] Health endpoint responds
- [x] All migrations applied (version 16)
- [x] Master user bootstrap works
- [x] API endpoints accessible

### SYSTEM User (Master Admin)
- [ ] Login as SYSTEM user
- [ ] Dashboard shows system-wide statistics
- [ ] Tenant selector works
- [ ] Create/Read/Update/Delete tenants
- [ ] Suspend/Resume tenants
- [ ] Create users for tenants
- [ ] View users across all tenants
- [ ] System Settings tab visible and works
- [ ] Security Settings tab visible and works
- [ ] OAuth2/OIDC Settings tab visible and works
- [ ] Tenant Settings tab works (with tenant selected)
- [ ] Access to `/system/*` endpoints

### TENANT User (Tenant Admin)
- [ ] Login as TENANT user
- [ ] Dashboard shows tenant-scoped statistics
- [ ] No tenant selector (locked to own tenant)
- [ ] Create/Read/Update/Delete users (own tenant only)
- [ ] View only own tenant's users
- [ ] Token Settings tab visible and works
- [ ] No access to System/Security/OAuth tabs
- [ ] No access to `/system/*` endpoints
- [ ] Access to `/api/v1/*` endpoints (tenant-scoped)

### Admin Dashboard
- [ ] Login works (both SYSTEM and TENANT)
- [ ] Dashboard displays correct statistics
- [ ] Tenant CRUD operations (SYSTEM only)
- [ ] User CRUD operations
- [ ] Role CRUD operations
- [ ] Permission CRUD operations
- [ ] Search and filtering
- [ ] Pagination
- [ ] Settings page (conditional based on user type)
- [ ] Audit logs viewer

### E2E Testing App
- [ ] User registration
- [ ] Login/logout
- [ ] Token refresh
- [ ] MFA enrollment
- [ ] MFA verification
- [ ] Profile management
- [ ] Password change
- [ ] Roles and permissions view

### Integration
- [ ] Cross-app workflows
- [ ] Role assignment flow
- [ ] Permission inheritance
- [ ] Tenant isolation
- [ ] SYSTEM vs TENANT permission checks
- [ ] Complete user journey

---

## 🐛 Troubleshooting

### Backend Issues

**Server won't start**:
```bash
# Check server logs
tail -f server.log

# Verify database connection
PGPASSWORD=dcim_password psql -h 127.0.0.1 -p 5433 -U dcim_user -d iam -c "SELECT 1;"
```

**Port already in use**:
```bash
# Kill existing process
pkill -f "go run cmd/server/main.go"

# Or change port
export SERVER_PORT=8081
```

**Migrations not applied**:
```bash
# Check current version
export DATABASE_URL="postgres://dcim_user:dcim_password@127.0.0.1:5433/iam?sslmode=disable"
migrate -path migrations -database "$DATABASE_URL" version

# Apply migrations
migrate -path migrations -database "$DATABASE_URL" up
```

### Frontend Issues

**Apps won't start**:
```bash
# Clear Vite cache
rm -rf node_modules/.vite

# Reinstall dependencies
npm install

# Try again
npm run dev
```

**CORS errors**:
- Verify backend CORS middleware is enabled
- Check API base URL in frontend config

**API connection errors**:
- Verify backend is running on port 8080
- Check `VITE_API_BASE_URL` in frontend `.env` files

**Settings page shows only Token Settings**:
- Check `localStorage.getItem('principal_type')` in browser console
- Should be `"SYSTEM"` for system admin, `"TENANT"` for tenant admin
- If incorrect, logout and login again

### Authentication Issues

**SYSTEM user can't login**:
- Verify user has `principal_type = 'SYSTEM'` in database
- Verify `tenant_id = NULL` for SYSTEM users
- Check JWT token contains `principal_type: "SYSTEM"`

**TENANT user can't login**:
- Verify user has `principal_type = 'TENANT'` in database
- Verify `tenant_id` is set and valid
- Provide tenant ID in login form

**403 Forbidden on SYSTEM endpoints**:
- Verify user has system permissions
- Check JWT token contains `system_permissions` array
- Verify user has required system role assigned

---

## 📝 Notes

- **Redis**: Optional - server works without it (caching disabled)
- **Hydra**: Optional - OAuth2 features may not work without it
- **Tenant Context**: 
  - TENANT users: Must provide `X-Tenant-ID` header
  - SYSTEM users: Can access `/system/*` without tenant context, or select tenant for tenant-scoped operations
- **Authentication**: 
  - Login endpoint returns JWT access token and refresh token
  - Token contains `principal_type`, `system_permissions`, and `permissions`
- **Token Refresh**: Use `/api/v1/auth/refresh` to get new tokens when access token expires
- **Token Revocation**: Use `/api/v1/auth/revoke` to logout and invalidate refresh tokens
- **Remember Me**: Extends token lifetimes for longer sessions
- **Master Tenant Architecture**: 
  - SYSTEM users can manage all tenants
  - TENANT users are isolated to their tenant
  - System roles and permissions are separate from tenant roles and permissions

---

## 🎯 Next Steps After Testing

1. **Document Issues**: Note any bugs or issues found
2. **Create GitHub Issues**: For any problems discovered
3. **Test Edge Cases**: Invalid inputs, boundary conditions
4. **Performance Testing**: Load testing, stress testing
5. **Security Testing**: Test authentication, authorization, input validation
6. **Test SYSTEM User Scenarios**: Multi-tenant management, tenant settings configuration
7. **Test TENANT User Scenarios**: Tenant isolation, tenant-scoped operations

---

## 📚 Related Documentation

- [Architecture Overview](../architecture/overview.md)
- [Master Tenant Architecture](../architecture/backend/MASTER_TENANT_ARCHITECTURE.md)
- [Authentication Flows](../guides/authentication/AUTHENTICATION_FLOWS_GUIDE.md)
- [Frontend Quick Start](../guides/frontend-quick-start.md)
- [Database Configuration](../guides/database-configuration.md)

---

**Happy Testing!** 🚀

**Last Updated**: 2026-01-08  
**Version**: 2.0 (Master Tenant Architecture)
