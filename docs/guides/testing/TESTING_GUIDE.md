# Complete Testing Guide - ARauth Identity IAM

## 📋 Prerequisites

- ✅ PostgreSQL running on 127.0.0.1:5433
- ✅ Database `iam` created
- ✅ All migrations applied
- ✅ Go installed
- ✅ Node.js and npm installed
- ✅ Frontend dependencies installed

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

## 🎨 Step 2: Start Frontend Applications

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

## 🧪 Step 3: Complete Testing Workflow

### Phase 1: Initial Setup via Admin Dashboard

#### 3.1 Access Admin Dashboard

1. Open browser: `http://localhost:5173`
2. You should see the **Login** page
3. **Note**: You'll need to create a tenant and admin user first via API

#### 3.2 Create First Tenant (via API)

```bash
# Create a tenant
curl -X POST http://localhost:8080/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Company",
    "domain": "test.local",
    "status": "active"
  }'

# Save the tenant ID from response (you'll need it)
```

**Expected Response**:
```json
{
  "id": "uuid-here",
  "name": "Test Company",
  "domain": "test.local",
  "status": "active",
  "created_at": "...",
  "updated_at": "..."
}
```

#### 3.3 Create Admin User (via API)

```bash
# Replace TENANT_ID with the ID from step 3.2
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: TENANT_ID" \
  -d '{
    "username": "admin",
    "email": "admin@test.local",
    "password": "Admin@123456",
    "first_name": "Admin",
    "last_name": "User"
  }'
```

**Expected Response**:
```json
{
  "id": "uuid-here",
  "username": "admin",
  "email": "admin@test.local",
  "first_name": "Admin",
  "last_name": "User",
  "status": "active",
  "tenant_id": "TENANT_ID",
  "created_at": "...",
  "updated_at": "..."
}
```

#### 3.4 Login to Admin Dashboard

1. Go to `http://localhost:5173`
2. Enter credentials:
   - **Username**: `admin`
   - **Password**: `Admin@123456`
   - **Tenant ID**: (the tenant ID from step 3.2)
3. Click **Login**
4. **✅ Expected**: Redirected to Dashboard with statistics

---

### Phase 2: Admin Dashboard Testing

#### 4.1 Dashboard Home Page

**Test**:
- [ ] View statistics cards (Tenants, Users, Roles, Permissions)
- [ ] Check counts are displayed correctly
- [ ] Click "View all" links navigate correctly
- [ ] Quick actions section is visible
- [ ] System overview shows status

**✅ Expected**: Dashboard displays with all statistics

#### 4.2 Tenant Management

**Navigate**: Click "Tenants" in sidebar or Dashboard → View all Tenants

**Test Create Tenant**:
1. Click **"Create Tenant"** button
2. Fill form:
   - Name: `Acme Corp`
   - Domain: `acme.local`
   - Status: `Active`
3. Click **"Create"**
4. **✅ Expected**: Tenant appears in list

**Test Search & Filter**:
- [ ] Search by name: Type "Acme" → Should filter results
- [ ] Search by domain: Type "acme.local" → Should filter results
- [ ] Filter by status: Select "Active" → Should show only active tenants

**Test Pagination**:
- [ ] Create multiple tenants (if needed)
- [ ] Change page size: Select 10, 20, 50, 100
- [ ] Navigate pages: First, Previous, Next, Last
- [ ] Verify item count display

**Test Edit Tenant**:
1. Click **"Edit"** on a tenant
2. Change name or status
3. Click **"Save"**
4. **✅ Expected**: Changes reflected in list

**Test Delete Tenant**:
1. Click **"Delete"** on a tenant
2. Confirm deletion
3. **✅ Expected**: Tenant removed from list

#### 4.3 User Management

**Navigate**: Click "Users" in sidebar

**Test Create User**:
1. Click **"Create User"** button
2. Fill form:
   - Username: `john.doe`
   - Email: `john@test.local`
   - Password: `Secure@123456`
   - First Name: `John`
   - Last Name: `Doe`
   - Status: `Active`
3. Click **"Create"**
4. **✅ Expected**: User appears in list

**Test User Operations**:
- [ ] Search by username, email, or name
- [ ] Filter by status (Active, Inactive, Locked)
- [ ] Edit user details
- [ ] Delete user
- [ ] Pagination works

#### 4.4 Role Management

**Navigate**: Click "Roles" in sidebar

**Test Create Role**:
1. Click **"Create Role"** button
2. Fill form:
   - Name: `Developer`
   - Description: `Developer role with read/write permissions`
3. Click **"Create"**
4. **✅ Expected**: Role appears in list

**Test Assign Permissions to Role**:
1. Click **"Manage Permissions"** on a role
2. Select permissions from the list
3. Click **"Save"**
4. **✅ Expected**: Permissions assigned to role

**Test Role Operations**:
- [ ] Search roles by name or description
- [ ] Edit role details
- [ ] Delete role
- [ ] Pagination works

#### 4.5 Permission Management

**Navigate**: Click "Permissions" in sidebar

**Test Create Permission**:
1. Click **"Create Permission"** button
2. Fill form:
   - Resource: `users`
   - Action: `read`
   - Description: `Read user information`
3. Click **"Create"**
4. **✅ Expected**: Permission appears in list

**Test Permission Operations**:
- [ ] Search by resource, action, or description
- [ ] Edit permission details
- [ ] Delete permission
- [ ] Pagination works

**Create Common Permissions** (for testing):
- `users:read`
- `users:write`
- `users:delete`
- `roles:read`
- `roles:write`
- `tenants:read`
- `tenants:write`

#### 4.6 System Settings

**Navigate**: Click "Settings" in sidebar

**Test Security Settings**:
- [ ] View password policy settings
- [ ] Modify minimum password length
- [ ] Toggle password requirements (uppercase, lowercase, numbers, special)
- [ ] Configure MFA requirements
- [ ] Set rate limiting values
- [ ] Click **"Save Security Settings"**
- [ ] **✅ Expected**: Success message displayed

**Test OAuth2/OIDC Settings**:
- [ ] View Hydra configuration
- [ ] Modify token TTLs
- [ ] Click **"Save OAuth Settings"**
- [ ] **✅ Expected**: Success message displayed

**Test Token Settings**:
- [ ] View token lifetime settings (Access Token TTL, Refresh Token TTL, ID Token TTL)
- [ ] Modify token lifetimes
- [ ] Configure "Remember Me" settings
- [ ] Toggle token rotation
- [ ] Set MFA requirement for extended sessions
- [ ] Click **"Save Token Settings"**
- [ ] **✅ Expected**: Success message displayed

**Test System Configuration**:
- [ ] View JWT settings
- [ ] Modify session timeout
- [ ] Configure account lockout settings
- [ ] Click **"Save System Settings"**
- [ ] **✅ Expected**: Success message displayed

#### 4.7 Audit Logs

**Navigate**: Click "Audit Logs" in sidebar

**Test Audit Log Viewer**:
- [ ] View log entries (if any exist)
- [ ] Search by user, action, resource, or IP
- [ ] Filter by action type
- [ ] Filter by status (success/failure)
- [ ] Test pagination
- [ ] Clear filters

**Note**: Audit logs will populate as you perform actions

---

### Phase 3: E2E Testing App

#### 5.1 User Registration

**Navigate**: `http://localhost:5174`

**Test Registration**:
1. Click **"Register"** or navigate to `/register`
2. Fill registration form:
   - Username: `testuser`
   - Email: `testuser@test.local`
   - Password: `Secure@123456`
   - Confirm Password: `Secure@123456`
   - First Name: `Test`
   - Last Name: `User`
   - Tenant ID: (use the tenant ID from step 3.2)
3. Click **"Register"**
4. **✅ Expected**: Success message and redirect to login

#### 5.2 Login Flow

**Test Login**:
1. Navigate to `/login`
2. Enter credentials:
   - Username: `testuser` (or `admin`)
   - Password: `Secure@123456` (or `Admin@123456`)
   - Tenant ID: (your tenant ID)
   - Remember Me: (optional checkbox)
3. Click **"Login"**
4. **✅ Expected**: Redirected to Dashboard with access token and refresh token

**Test Invalid Credentials**:
- [ ] Enter wrong password → Should show error
- [ ] Enter wrong username → Should show error
- [ ] Enter wrong tenant ID → Should show error

#### 5.3 Dashboard

**Test Dashboard**:
- [ ] View welcome message
- [ ] See navigation cards (Profile, MFA, Roles & Permissions)
- [ ] Click "Logout" → Should redirect to login

#### 5.4 MFA Flow

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

#### 5.5 Profile Management

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

#### 5.6 Roles and Permissions View

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

**Note**: If no roles assigned, assign roles via Admin Dashboard first

---

### Phase 4: Integration Testing

#### 6.1 Cross-App Testing

**Test Flow**:
1. Create tenant in Admin Dashboard
2. Create user in Admin Dashboard
3. Assign role to user in Admin Dashboard
4. Login to E2E Test App with that user
5. View roles and permissions in E2E Test App
6. **✅ Expected**: User sees assigned roles and permissions

#### 6.2 Role Assignment Flow

**Test Complete Flow**:
1. **Admin Dashboard**: Create role "Manager"
2. **Admin Dashboard**: Create permissions (e.g., `users:read`, `users:write`)
3. **Admin Dashboard**: Assign permissions to "Manager" role
4. **Admin Dashboard**: Assign "Manager" role to a user
5. **E2E Test App**: Login as that user
6. **E2E Test App**: View roles and permissions
7. **✅ Expected**: User sees "Manager" role with assigned permissions

#### 6.3 MFA End-to-End

**Test Complete MFA Flow**:
1. **E2E Test App**: Register new user
2. **E2E Test App**: Login
3. **E2E Test App**: Enroll in MFA
4. **E2E Test App**: Logout
5. **E2E Test App**: Login again
6. **E2E Test App**: Enter MFA code
7. **✅ Expected**: Complete MFA flow works

---

## 🔍 Step 4: API Testing (Optional)

### Test API Endpoints Directly

```bash
# Get tenant ID first (from step 3.2)
TENANT_ID="your-tenant-id-here"

# List users
curl -H "X-Tenant-ID: $TENANT_ID" \
  http://localhost:8080/api/v1/users

# List roles
curl -H "X-Tenant-ID: $TENANT_ID" \
  http://localhost:8080/api/v1/roles

# List permissions
curl -H "X-Tenant-ID: $TENANT_ID" \
  http://localhost:8080/api/v1/permissions

# Login and get token
LOGIN_RESPONSE=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "username": "admin",
    "password": "Admin@123456",
    "remember_me": false
  }')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.access_token')
REFRESH_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.refresh_token')

# Use token for authenticated requests
curl -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID" \
  http://localhost:8080/api/v1/users

# Refresh token (when access token expires)
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
- [x] API endpoints accessible

### Admin Dashboard
- [ ] Login works
- [ ] Dashboard displays statistics
- [ ] Tenant CRUD operations
- [ ] User CRUD operations
- [ ] Role CRUD operations
- [ ] Permission CRUD operations
- [ ] Search and filtering
- [ ] Pagination
- [ ] Settings page
- [ ] Audit logs viewer

### E2E Testing App
- [ ] User registration
- [ ] Login/logout
- [ ] MFA enrollment
- [ ] MFA verification
- [ ] Profile management
- [ ] Password change
- [ ] Roles and permissions view

### Integration
- [ ] Cross-app workflows
- [ ] Role assignment flow
- [ ] Permission inheritance
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

---

## 📝 Notes

- **Redis**: Optional - server works without it (caching disabled)
- **Hydra**: Optional - OAuth2 features may not work without it
- **Tenant Context**: Most API calls require `X-Tenant-ID` header
- **Authentication**: Login endpoint returns JWT access token and refresh token
- **Token Refresh**: Use `/api/v1/auth/refresh` to get new tokens when access token expires
- **Token Revocation**: Use `/api/v1/auth/revoke` to logout and invalidate refresh tokens
- **Remember Me**: Extends token lifetimes for longer sessions

---

## 🎯 Next Steps After Testing

1. **Document Issues**: Note any bugs or issues found
2. **Create GitHub Issues**: For any problems discovered
3. **Test Edge Cases**: Invalid inputs, boundary conditions
4. **Performance Testing**: Load testing, stress testing
5. **Security Testing**: Test authentication, authorization, input validation

---

**Happy Testing!** 🚀

**Last Updated**: 2026-01-08

