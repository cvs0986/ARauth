# Complete Testing Guide - ARauth Identity IAM (V3 - Capability Model)

## 📋 Table of Contents

1. [Prerequisites](#prerequisites)
2. [Environment Setup](#environment-setup)
3. [Database Setup & Migrations](#database-setup--migrations)
4. [Local Development Testing](#local-development-testing)
5. [Kubernetes Testing](#kubernetes-testing)
6. [Cloud Deployment Testing](#cloud-deployment-testing)
7. [On-Premise Testing](#on-premise-testing)
8. [Capability Model Testing](#capability-model-testing)
9. [Complete Feature Testing](#complete-feature-testing)
10. [API Testing](#api-testing)
11. [Troubleshooting](#troubleshooting)

---

## 📋 Prerequisites

### Required Software

- ✅ **PostgreSQL** 14+ (or compatible database)
- ✅ **Go** 1.22+ installed
- ✅ **Node.js** 18+ and npm installed
- ✅ **Redis** (optional, for caching and MFA sessions)
- ✅ **Docker** & **Docker Compose** (for containerized testing)
- ✅ **kubectl** (for Kubernetes testing)
- ✅ **migrate** tool (for database migrations)

### Verify Prerequisites

```bash
# Check Go version
go version  # Should be 1.22+

# Check Node.js version
node --version  # Should be 18+

# Check PostgreSQL
psql --version  # Should be 14+

# Check Docker
docker --version

# Check kubectl (if testing Kubernetes)
kubectl version --client

# Install migrate tool if not installed
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

---

## 🏗️ Architecture Overview

ARauth Identity supports:

1. **Two-Plane Architecture**:
   - **SYSTEM Users** (Master/Platform Admins)
   - **TENANT Users** (Tenant Admins)

2. **Three-Layer Capability Model**:
   - **System Level**: Global capabilities
   - **System → Tenant**: Capability assignments
   - **Tenant Level**: Feature enablement
   - **User Level**: User enrollment

---

## 🔧 Environment Setup

### Option 1: Local Development (Recommended for Testing)

**Database Configuration**:
```bash
# PostgreSQL connection details
export DATABASE_HOST=127.0.0.1
export DATABASE_PORT=5432
export DATABASE_USER=iam_user
export DATABASE_PASSWORD=change-me
export DATABASE_NAME=iam
export DATABASE_SSL_MODE=disable
```

**Application Configuration**:
```bash
export JWT_SECRET=test-jwt-secret-key-min-32-characters-long
export ENCRYPTION_KEY=01234567890123456789012345678901
export SERVER_PORT=8080
export LOG_LEVEL=info
```

**Redis Configuration** (Optional):
```bash
export REDIS_URL=redis://localhost:6379
```

### Option 2: Docker Compose

```bash
# Start all services
docker-compose up -d

# Check services
docker-compose ps

# View logs
docker-compose logs -f
```

### Option 3: Kubernetes

See [Kubernetes Testing](#kubernetes-testing) section below.

---

## 🗄️ Database Setup & Migrations

### Step 1: Create Database

```bash
# Connect to PostgreSQL
psql -U postgres

# Create database
CREATE DATABASE iam;

# Create user (if needed)
CREATE USER iam_user WITH PASSWORD 'change-me';
GRANT ALL PRIVILEGES ON DATABASE iam TO iam_user;

# Exit psql
\q
```

### Step 2: Run Migrations

```bash
# Set database URL
export DATABASE_URL="postgres://iam_user:change-me@localhost:5432/iam?sslmode=disable"

# Navigate to project root
cd /path/to/nuage-indentity

# Check current migration version
migrate -path migrations -database "$DATABASE_URL" version

# Run all migrations (up to version 22)
migrate -path migrations -database "$DATABASE_URL" up

# Verify migrations applied
migrate -path migrations -database "$DATABASE_URL" version
```

**Expected Output**: Should show version `22` (all migrations including capability model)

### Step 3: Verify Database Schema

```bash
# Connect to database
psql $DATABASE_URL

# Check capability tables exist
\dt system_capabilities
\dt tenant_capabilities
\dt tenant_feature_enablement
\dt user_capability_state

# Check system capabilities are populated
SELECT capability_key, enabled FROM system_capabilities;

# Exit
\q
```

**Expected**: Should see 11 default system capabilities (mfa, totp, saml, oidc, oauth2, etc.)

### Step 4: Run Data Migration (If Upgrading)

If you have existing data, run the migration script:

```bash
# The migration script (000022) will automatically:
# - Assign default capabilities to existing tenants
# - Migrate tenant_settings to capability model
# - Migrate MFA settings
# - Migrate user MFA enrollment

# It's already included in "migrate up" command above
# To run manually if needed:
migrate -path migrations -database "$DATABASE_URL" up
```

---

## 🚀 Local Development Testing

### Step 1: Start Backend Server

```bash
cd /path/to/nuage-indentity

# Set environment variables
export DATABASE_HOST=127.0.0.1
export DATABASE_PORT=5432
export DATABASE_USER=iam_user
export DATABASE_PASSWORD=change-me
export DATABASE_NAME=iam
export DATABASE_SSL_MODE=disable
export JWT_SECRET=test-jwt-secret-key-min-32-characters-long
export ENCRYPTION_KEY=01234567890123456789012345678901

# Start server
go run cmd/server/main.go
```

**Verify Backend**:
```bash
curl http://localhost:8080/health

# Expected:
# {"status":"healthy","timestamp":"...","version":"0.1.0","checks":{"database":"healthy","redis":"not_configured"}}
```

### Step 2: Bootstrap Master User (SYSTEM Admin)

**Option A: Using Config File**

Create `config/bootstrap.yaml`:
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

**Option B: Using Environment Variables**

```bash
export BOOTSTRAP_ENABLED=true
export BOOTSTRAP_MASTER_USERNAME=system_admin
export BOOTSTRAP_MASTER_EMAIL=admin@arauth.local
export BOOTSTRAP_MASTER_PASSWORD=SystemAdmin@123456
export BOOTSTRAP_MASTER_FIRST_NAME=System
export BOOTSTRAP_MASTER_LAST_NAME=Administrator
export BOOTSTRAP_MASTER_ROLE=system_owner

# Start server (will auto-bootstrap)
go run cmd/server/main.go
```

**Verify Master User**:
```bash
psql $DATABASE_URL -c \
  "SELECT id, username, email, principal_type, tenant_id FROM users WHERE principal_type = 'SYSTEM';"
```

### Step 3: Start Frontend Applications

**Terminal 1: Admin Dashboard**
```bash
cd frontend/admin-dashboard
npm install  # First time only
npm run dev
```

**Terminal 2: E2E Testing App** (Optional)
```bash
cd frontend/e2e-test-app
npm install  # First time only
npm run dev
```

**Access Points**:
- Admin Dashboard: `http://localhost:5173`
- E2E Test App: `http://localhost:5174`

---

## ☸️ Kubernetes Testing

### Prerequisites

- Kubernetes cluster (local: minikube, kind, or cloud: EKS, GKE, AKS)
- kubectl configured
- Helm (optional, for easier deployment)

### Step 1: Build Docker Images

```bash
# Build backend image
docker build -t arauth-identity/iam-api:latest -f Dockerfile .

# Build frontend image
cd frontend/admin-dashboard
docker build -t arauth-identity/admin-dashboard:latest -f Dockerfile .
```

### Step 2: Deploy to Kubernetes

**Option A: Using kubectl**

```bash
# Apply database secret
kubectl create secret generic db-credentials \
  --from-literal=username=iam_user \
  --from-literal=password=change-me

# Apply Redis secret (if using)
kubectl create secret generic redis-credentials \
  --from-literal=url=redis://redis-service:6379

# Deploy PostgreSQL (or use managed service)
kubectl apply -f k8s/postgresql.yaml

# Deploy Redis (or use managed service)
kubectl apply -f k8s/redis.yaml

# Deploy backend
kubectl apply -f k8s/backend.yaml

# Deploy frontend
kubectl apply -f k8s/frontend.yaml
```

**Option B: Using Helm**

```bash
# Install with Helm
helm install arauth-identity ./helm/arauth-identity \
  --set database.host=postgres-service \
  --set database.name=iam \
  --set redis.enabled=true
```

### Step 3: Run Migrations in Kubernetes

```bash
# Create migration job
kubectl create job migrate-db --from=cronjob/migrate-db

# Or run manually in a pod
kubectl run migrate --image=arauth-identity/iam-api:latest --rm -it -- \
  migrate -path /migrations -database "$DATABASE_URL" up
```

### Step 4: Access Services

```bash
# Port forward to access services
kubectl port-forward svc/iam-api 8080:8080
kubectl port-forward svc/admin-dashboard 5173:80

# Or use LoadBalancer/Ingress
kubectl get svc
```

### Step 5: Verify Deployment

```bash
# Check pods
kubectl get pods

# Check services
kubectl get svc

# Check logs
kubectl logs -f deployment/iam-api

# Test health endpoint
curl http://localhost:8080/health
```

---

## ☁️ Cloud Deployment Testing

### AWS (EKS)

```bash
# Configure AWS CLI
aws configure

# Create EKS cluster
eksctl create cluster --name arauth-cluster --region us-east-1

# Deploy using kubectl (same as Kubernetes section)
kubectl apply -f k8s/

# Use AWS RDS for PostgreSQL
# Use AWS ElastiCache for Redis
```

### Google Cloud (GKE)

```bash
# Configure gcloud
gcloud init

# Create GKE cluster
gcloud container clusters create arauth-cluster --zone us-central1-a

# Deploy
kubectl apply -f k8s/

# Use Cloud SQL for PostgreSQL
# Use Cloud Memorystore for Redis
```

### Azure (AKS)

```bash
# Configure Azure CLI
az login

# Create AKS cluster
az aks create --resource-group arauth-rg --name arauth-cluster

# Deploy
kubectl apply -f k8s/

# Use Azure Database for PostgreSQL
# Use Azure Cache for Redis
```

---

## 🏢 On-Premise Testing

### Step 1: Prepare Infrastructure

- Physical/Virtual servers
- Network configuration
- Firewall rules
- SSL certificates

### Step 2: Install Dependencies

```bash
# On each server, install:
# - PostgreSQL
# - Redis (optional)
# - Go runtime
# - Node.js (for frontend)
```

### Step 3: Deploy Application

```bash
# Build binaries
go build -o bin/iam-api ./cmd/server

# Copy to server
scp bin/iam-api user@server:/opt/arauth-identity/

# Create systemd service
sudo systemctl enable arauth-identity
sudo systemctl start arauth-identity
```

### Step 4: Configure Reverse Proxy

**Nginx Example**:
```nginx
server {
    listen 80;
    server_name arauth.example.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

---

## 🎯 Capability Model Testing

### Phase 1: System Capability Management (SYSTEM Admin Only)

#### 1.1 Login as SYSTEM Admin

1. Open `http://localhost:5173`
2. Login with:
   - Username: `system_admin`
   - Password: `SystemAdmin@123456`
   - Tenant ID: (leave empty)

#### 1.2 View System Capabilities

**Navigate**: Settings → Capabilities Tab (or `/system/capabilities`)

**Test**:
- [ ] View list of all system capabilities
- [ ] See capability keys: mfa, totp, saml, oidc, oauth2, etc.
- [ ] See enabled/disabled status
- [ ] See descriptions
- [ ] See default values

**Expected**: List shows 11 default capabilities

#### 1.3 Edit System Capability

**Test**:
1. Click "Edit" on a capability (e.g., `mfa`)
2. Modify description
3. Toggle enabled status
4. Update default value (JSON)
5. Click "Save"

**Expected**: Changes saved, capability updated

**API Test**:
```bash
TOKEN="your-system-admin-token"

# Get system capabilities
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/system/capabilities

# Update capability
curl -X PUT http://localhost:8080/system/capabilities/mfa \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "enabled": true,
    "description": "Multi-factor authentication support",
    "default_value": {"max_attempts": 3}
  }'
```

### Phase 2: Tenant Capability Assignment (SYSTEM Admin Only)

#### 2.1 Assign Capability to Tenant

**Navigate**: Tenants → Select Tenant → Capabilities Tab

**Test**:
1. Select a tenant from header dropdown
2. Navigate to "Capabilities" tab
3. Click "Assign Capability"
4. Select capability (e.g., `mfa`)
5. Set enabled: `true`
6. Set value (JSON, optional): `{"max_attempts": 5}`
7. Click "Assign"

**Expected**: Capability assigned to tenant

**API Test**:
```bash
TENANT_ID="tenant-uuid-here"

# Assign capability
curl -X PUT http://localhost:8080/system/tenants/$TENANT_ID/capabilities/mfa \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "enabled": true,
    "value": {"max_attempts": 5}
  }'

# Get tenant capabilities
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/system/tenants/$TENANT_ID/capabilities
```

#### 2.2 Revoke Capability from Tenant

**Test**:
1. Navigate to tenant capabilities
2. Click "Revoke" on a capability
3. Confirm deletion

**Expected**: Capability removed from tenant

#### 2.3 View Capability Evaluation

**Test**:
1. Navigate to tenant capabilities
2. Click "View Evaluation"
3. See complete capability evaluation:
   - System supported: ✅
   - Tenant allowed: ✅
   - Tenant enabled: ⚠️
   - User enrolled: ⚠️

**Expected**: Shows full evaluation chain

### Phase 3: Tenant Feature Enablement (TENANT Admin)

#### 3.1 Login as TENANT Admin

1. Create tenant admin user (as SYSTEM admin)
2. Login with tenant admin credentials
3. Provide tenant ID

#### 3.2 Enable Feature for Tenant

**Navigate**: Settings → Capabilities Tab → Features

**Test**:
1. View available features (capabilities allowed by system)
2. Click "Enable Feature" on `mfa`
3. Set configuration (JSON, optional): `{"required_for_admins": true}`
4. Click "Enable"

**Expected**: Feature enabled for tenant

**API Test**:
```bash
# Login as tenant admin
TENANT_ID="tenant-uuid"
TENANT_TOKEN="tenant-admin-token"

# Enable feature
curl -X PUT http://localhost:8080/api/v1/tenant/features/mfa \
  -H "Authorization: Bearer $TENANT_TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "enabled": true,
    "configuration": {"required_for_admins": true}
  }'

# Get enabled features
curl -H "Authorization: Bearer $TENANT_TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID" \
  http://localhost:8080/api/v1/tenant/features
```

#### 3.3 Disable Feature

**Test**:
1. Navigate to features
2. Click "Disable" on enabled feature
3. Confirm

**Expected**: Feature disabled

### Phase 4: User Capability Enrollment (TENANT Admin)

#### 4.1 Enroll User in Capability

**Navigate**: Users → Select User → Capabilities Tab

**Test**:
1. Select a user
2. Navigate to "Capabilities" tab
3. View available capabilities (enabled by tenant)
4. Click "Enroll" on `mfa`
5. Fill enrollment form (if required)
6. Click "Enroll"

**Expected**: User enrolled in capability

**API Test**:
```bash
USER_ID="user-uuid"

# Enroll user
curl -X POST http://localhost:8080/api/v1/users/$USER_ID/capabilities/mfa/enroll \
  -H "Authorization: Bearer $TENANT_TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "state_data": {}
  }'

# Get user capabilities
curl -H "Authorization: Bearer $TENANT_TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID" \
  http://localhost:8080/api/v1/users/$USER_ID/capabilities
```

#### 4.2 Unenroll User

**Test**:
1. Navigate to user capabilities
2. Click "Unenroll" on enrolled capability
3. Confirm

**Expected**: User unenrolled

### Phase 5: Capability Enforcement Testing

#### 5.1 Test MFA Enforcement

**Scenario**: MFA required but user not enrolled

1. **SYSTEM Admin**: Enable `mfa` capability for tenant
2. **TENANT Admin**: Enable `mfa` feature with `required: true`
3. **User**: Try to login
4. **Expected**: Login blocked, MFA enrollment required

**Test**:
```bash
# Try login without MFA enrollment
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "username": "testuser",
    "password": "password"
  }'

# Expected: 403 Forbidden or MFA required response
```

#### 5.2 Test OAuth2 Scope Validation

**Scenario**: OAuth2 scope not allowed for tenant

1. **SYSTEM Admin**: Set `allowed_scope_namespaces` for tenant
2. **User**: Request OAuth token with unauthorized scope
3. **Expected**: Token request rejected

#### 5.3 Test Capability Inheritance Visualization

**Navigate**: Settings → Capabilities Tab → Inheritance View

**Test**:
- [ ] View three-layer visualization
- [ ] See System → Tenant → User flow
- [ ] See capability status at each level
- [ ] See inheritance path

**Expected**: Visual representation of capability inheritance

---

## 🧪 Complete Feature Testing

### Authentication & Authorization

#### Login Testing

**SYSTEM User Login**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "system_admin",
    "password": "SystemAdmin@123456"
  }'
```

**TENANT User Login**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "username": "tenant_admin",
    "password": "TenantAdmin@123456"
  }'
```

#### Token Refresh

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "$REFRESH_TOKEN"
  }'
```

#### Token Revocation

```bash
curl -X POST http://localhost:8080/api/v1/auth/revoke \
  -H "Content-Type: application/json" \
  -d '{
    "token": "$REFRESH_TOKEN",
    "token_type_hint": "refresh_token"
  }'
```

### MFA Testing

#### Enroll in MFA

```bash
curl -X POST http://localhost:8080/api/v1/mfa/enroll \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "$USER_ID"
  }'
```

#### Verify MFA

```bash
curl -X POST http://localhost:8080/api/v1/mfa/verify \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "$USER_ID",
    "totp_code": "123456"
  }'
```

### User Management

#### Create User

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "username": "newuser",
    "email": "newuser@example.com",
    "password": "Secure@123456",
    "first_name": "New",
    "last_name": "User"
  }'
```

#### List Users

```bash
curl -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID" \
  http://localhost:8080/api/v1/users
```

### Role & Permission Management

#### Create Role

```bash
curl -X POST http://localhost:8080/api/v1/roles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "name": "Developer",
    "description": "Developer role"
  }'
```

#### Assign Permission to Role

```bash
curl -X POST http://localhost:8080/api/v1/roles/$ROLE_ID/permissions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "permission_id": "$PERMISSION_ID"
  }'
```

---

## 🔍 API Testing

### Complete API Test Suite

See [API Testing Guide](./API_TESTING_GUIDE.md) for comprehensive API testing scenarios.

### Quick API Health Check

```bash
#!/bin/bash

BASE_URL="http://localhost:8080"

# Health check
echo "Testing Health Endpoint..."
curl -s $BASE_URL/health | jq .

# Login
echo "Testing Login..."
LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "system_admin", "password": "SystemAdmin@123456"}')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.access_token')
echo "Token: $TOKEN"

# Test protected endpoint
echo "Testing Protected Endpoint..."
curl -s -H "Authorization: Bearer $TOKEN" \
  $BASE_URL/system/capabilities | jq .
```

---

## 🐛 Troubleshooting

### Database Issues

**Connection Failed**:
```bash
# Check PostgreSQL is running
pg_isready -h localhost -p 5432

# Check credentials
psql -h localhost -U iam_user -d iam -c "SELECT 1;"
```

**Migrations Failed**:
```bash
# Check current version
migrate -path migrations -database "$DATABASE_URL" version

# Force version (if needed)
migrate -path migrations -database "$DATABASE_URL" force 22

# Run migrations again
migrate -path migrations -database "$DATABASE_URL" up
```

### Backend Issues

**Server Won't Start**:
```bash
# Check logs
tail -f server.log

# Check port availability
lsof -i :8080

# Check environment variables
env | grep DATABASE
env | grep JWT
```

**Capability Service Errors**:
```bash
# Check capability tables exist
psql $DATABASE_URL -c "\dt *capability*"

# Check system capabilities populated
psql $DATABASE_URL -c "SELECT COUNT(*) FROM system_capabilities;"
```

### Frontend Issues

**CORS Errors**:
- Verify backend CORS middleware enabled
- Check `VITE_API_BASE_URL` in `.env`

**Capability Pages Not Loading**:
- Check user has correct `principal_type` (SYSTEM or TENANT)
- Verify JWT token contains capability claims
- Check browser console for errors

### Kubernetes Issues

**Pods Not Starting**:
```bash
# Check pod status
kubectl get pods

# Check pod logs
kubectl logs <pod-name>

# Check events
kubectl describe pod <pod-name>
```

**Services Not Accessible**:
```bash
# Check services
kubectl get svc

# Check ingress
kubectl get ingress

# Port forward for testing
kubectl port-forward svc/iam-api 8080:8080
```

---

## ✅ Complete Testing Checklist

### Setup
- [ ] Database created and accessible
- [ ] All migrations applied (version 22)
- [ ] System capabilities populated
- [ ] Backend server running
- [ ] Frontend applications running
- [ ] Master user (SYSTEM admin) created

### Capability Model
- [ ] System capabilities visible (SYSTEM admin)
- [ ] System capability editing works
- [ ] Tenant capability assignment works
- [ ] Tenant feature enablement works
- [ ] User capability enrollment works
- [ ] Capability enforcement works
- [ ] Inheritance visualization works

### Authentication
- [ ] SYSTEM user login works
- [ ] TENANT user login works
- [ ] Token refresh works
- [ ] Token revocation works
- [ ] MFA enrollment works
- [ ] MFA verification works

### Authorization
- [ ] SYSTEM user can access `/system/*` endpoints
- [ ] TENANT user cannot access `/system/*` endpoints
- [ ] Tenant isolation works
- [ ] Role-based access control works

### User Management
- [ ] Create user works
- [ ] List users works
- [ ] Update user works
- [ ] Delete user works
- [ ] User search works

### Role & Permission Management
- [ ] Create role works
- [ ] Assign permissions works
- [ ] Assign roles to users works
- [ ] Permission inheritance works

### Deployment Environments
- [ ] Local development works
- [ ] Docker Compose works
- [ ] Kubernetes deployment works
- [ ] Cloud deployment works
- [ ] On-premise deployment works

---

## 📚 Related Documentation

- [Architecture Overview](../architecture/overview.md)
- [Capability Model Architecture](../architecture/CAPABILITY_MODEL.md)
- [Master Tenant Architecture](../architecture/backend/MASTER_TENANT_ARCHITECTURE.md)
- [Deployment Plan](../deployment/CAPABILITY_MODEL_DEPLOYMENT_PLAN.md)
- [API Documentation](../api/capability-endpoints.md)

---

**Happy Testing!** 🚀

**Last Updated**: 2025-01-27  
**Version**: 3.0 (Capability Model)
