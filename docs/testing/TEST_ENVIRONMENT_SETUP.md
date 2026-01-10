# Test Environment Setup Guide

## Overview

This guide provides step-by-step instructions for setting up a clean test environment for ARauth IAM system testing. The environment should start from a completely clean state (empty database, no tenants, no users).

## Prerequisites

### Required Software
- **Go**: Version 1.21 or later
- **PostgreSQL**: Version 14 or later
- **Redis**: Version 7 or later
- **Docker & Docker Compose**: For containerized services (optional)
- **curl**: For API testing
- **jq**: For JSON processing (optional but recommended)
- **psql**: PostgreSQL client

### System Requirements
- **OS**: Linux, macOS, or Windows (WSL2)
- **RAM**: Minimum 4GB (8GB recommended)
- **Disk**: 10GB free space
- **Network**: Ports 8080, 5432, 6379 available

## Setup Methods

### Method 1: Docker Compose (Recommended)

This is the easiest way to set up all dependencies.

#### Step 1: Clone and Navigate
```bash
cd /path/to/nuage-indentity
```

#### Step 2: Create Environment File
```bash
cat > .env <<EOF
POSTGRES_PASSWORD=test_password_123
HYDRA_DB_PASSWORD=hydra_password_123
REDIS_PASSWORD=redis_password_123
EOF
```

#### Step 3: Start Services
```bash
docker-compose up -d
```

#### Step 4: Wait for Services
```bash
# Wait for PostgreSQL
until docker exec postgres-iam pg_isready -U iam_user; do
  echo "Waiting for PostgreSQL..."
  sleep 2
done

# Wait for Redis
until docker exec redis-iam redis-cli --raw incr ping > /dev/null 2>&1; do
  echo "Waiting for Redis..."
  sleep 2
done

echo "✅ All services are ready"
```

#### Step 5: Run Migrations
```bash
export DATABASE_URL="postgres://iam_user:test_password_123@localhost:5432/iam?sslmode=disable"
migrate -path migrations -database "$DATABASE_URL" up
```

#### Step 6: Verify Database
```bash
psql -h localhost -U iam_user -d iam -c "SELECT COUNT(*) FROM tenants;"
# Should return: 0
```

### Method 2: Manual Setup

#### Step 1: Install PostgreSQL
```bash
# Ubuntu/Debian
sudo apt-get install postgresql-14

# macOS
brew install postgresql@14

# Start PostgreSQL
sudo systemctl start postgresql  # Linux
brew services start postgresql@14  # macOS
```

#### Step 2: Create Database
```bash
# Create user and database
sudo -u postgres psql <<EOF
CREATE USER iam_user WITH PASSWORD 'test_password_123';
CREATE DATABASE iam OWNER iam_user;
GRANT ALL PRIVILEGES ON DATABASE iam TO iam_user;
EOF
```

#### Step 3: Install Redis
```bash
# Ubuntu/Debian
sudo apt-get install redis-server

# macOS
brew install redis

# Start Redis
sudo systemctl start redis  # Linux
brew services start redis  # macOS
```

#### Step 4: Run Migrations
```bash
export DATABASE_URL="postgres://iam_user:test_password_123@localhost:5432/iam?sslmode=disable"
migrate -path migrations -database "$DATABASE_URL" up
```

#### Step 5: Verify Setup
```bash
psql -h localhost -U iam_user -d iam -c "SELECT COUNT(*) FROM tenants;"
# Should return: 0
```

## Configuration

### Environment Variables

Create a `.env` file or export these variables:

```bash
# Database
export DATABASE_HOST=localhost
export DATABASE_PORT=5432
export DATABASE_NAME=iam
export DATABASE_USER=iam_user
export DATABASE_PASSWORD=test_password_123
export DATABASE_SSL_MODE=disable

# Redis
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=redis_password_123  # If Redis has password
export REDIS_DB=0

# JWT
export JWT_SECRET=test-jwt-secret-key-min-32-characters-long-for-local-dev
export JWT_SIGNING_KEY_PATH=  # Optional: path to RSA key

# Encryption
export ENCRYPTION_KEY=01234567890123456789012345678901  # 32 bytes

# Hydra (OAuth2 server - optional for basic testing)
export HYDRA_ADMIN_URL=http://localhost:4445
export HYDRA_PUBLIC_URL=http://localhost:4444

# Server
export SERVER_PORT=8080
export SERVER_HOST=0.0.0.0

# Logging
export LOG_LEVEL=info
export LOG_FORMAT=json

# Bootstrap (for initial system owner)
export BOOTSTRAP_ENABLED=true
export BOOTSTRAP_FORCE=false
export BOOTSTRAP_USERNAME=admin
export BOOTSTRAP_EMAIL=admin@arauth.io
export BOOTSTRAP_PASSWORD=AdminPassword123!
export BOOTSTRAP_FIRST_NAME=System
export BOOTSTRAP_LAST_NAME=Administrator
```

### Configuration File

Alternatively, create `config/config.test.yaml`:

```yaml
server:
  port: 8080
  host: "0.0.0.0"

database:
  host: "localhost"
  port: 5432
  name: "iam"
  user: "iam_user"
  password: "test_password_123"
  ssl_mode: "disable"

redis:
  host: "localhost"
  port: 6379
  password: "redis_password_123"
  db: 0

security:
  encryption_key: "01234567890123456789012345678901"
  jwt:
    secret: "test-jwt-secret-key-min-32-characters-long-for-local-dev"
    access_token_ttl: 15m
    refresh_token_ttl: 30d

bootstrap:
  enabled: true
  force: false
  master_user:
    username: "admin"
    email: "admin@arauth.io"
    password: "AdminPassword123!"
    first_name: "System"
    last_name: "Administrator"
```

## Starting the Server

### Option 1: Using Go Run
```bash
# Load environment variables
source .env  # If using .env file

# Start server
go run cmd/server/main.go
```

### Option 2: Using Make
```bash
make run
```

### Option 3: Using Build
```bash
make build
./bin/iam-api
```

## Verification Steps

### 1. Health Check
```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "database": "connected",
  "redis": "connected"
}
```

### 2. Liveness Check
```bash
curl http://localhost:8080/health/live
```

Expected: `200 OK`

### 3. Readiness Check
```bash
curl http://localhost:8080/health/ready
```

Expected: `200 OK`

### 4. Database Verification
```bash
psql -h localhost -U iam_user -d iam <<EOF
-- Check migrations applied
SELECT COUNT(*) FROM schema_migrations;

-- Check no tenants exist
SELECT COUNT(*) FROM tenants;

-- Check no users exist (except bootstrap user if enabled)
SELECT COUNT(*) FROM users;

-- Check system roles exist
SELECT COUNT(*) FROM system_roles;
-- Should return: 3 (system_owner, system_admin, system_auditor)

-- Check system capabilities exist
SELECT COUNT(*) FROM system_capabilities;
-- Should return: 11 (mfa, totp, saml, oidc, oauth2, etc.)
EOF
```

### 5. Redis Verification
```bash
redis-cli -a redis_password_123 ping
```

Expected: `PONG`

## Clean Database Reset

If you need to reset the database to a clean state:

### Option 1: Drop and Recreate
```bash
# Drop database
psql -h localhost -U postgres <<EOF
DROP DATABASE IF EXISTS iam;
CREATE DATABASE iam OWNER iam_user;
EOF

# Run migrations
export DATABASE_URL="postgres://iam_user:test_password_123@localhost:5432/iam?sslmode=disable"
migrate -path migrations -database "$DATABASE_URL" up
```

### Option 2: Rollback and Reapply
```bash
# Rollback all migrations
export DATABASE_URL="postgres://iam_user:test_password_123@localhost:5432/iam?sslmode=disable"
migrate -path migrations -database "$DATABASE_URL" down -all

# Reapply migrations
migrate -path migrations -database "$DATABASE_URL" up
```

### Option 3: Truncate Tables (Faster)
```bash
psql -h localhost -U iam_user -d iam <<EOF
-- Truncate all data tables (preserve structure)
TRUNCATE TABLE 
  users, tenants, roles, permissions, user_roles, role_permissions,
  refresh_tokens, mfa_recovery_codes, audit_events,
  webhooks, webhook_deliveries, user_invitations,
  identity_providers, federated_identities,
  impersonation_sessions, oauth_scopes, scim_tokens,
  tenant_settings, tenant_capabilities, tenant_feature_enablement,
  user_capability_state
CASCADE;

-- Reset sequences if any
-- (PostgreSQL auto-increment handles this, but verify)
EOF
```

## Frontend Setup (Optional)

If testing the admin console:

### Prerequisites
- Node.js 18+
- npm or yarn

### Setup
```bash
cd frontend/admin-dashboard
npm install
npm run dev
```

Frontend will be available at `http://localhost:5173` (or configured port).

## Troubleshooting

### Database Connection Issues
```bash
# Check PostgreSQL is running
sudo systemctl status postgresql  # Linux
brew services list | grep postgresql  # macOS

# Check connection
psql -h localhost -U iam_user -d iam -c "SELECT 1;"

# Check permissions
psql -h localhost -U postgres -c "\du iam_user"
```

### Redis Connection Issues
```bash
# Check Redis is running
sudo systemctl status redis  # Linux
brew services list | grep redis  # macOS

# Test connection
redis-cli -a redis_password_123 ping

# Check Redis data
redis-cli -a redis_password_123 KEYS "*"
```

### Migration Issues
```bash
# Check migration status
migrate -path migrations -database "$DATABASE_URL" version

# Check for errors
migrate -path migrations -database "$DATABASE_URL" up -verbose

# Verify schema
psql -h localhost -U iam_user -d iam -c "\dt"
```

### Port Conflicts
```bash
# Check if port 8080 is in use
lsof -i :8080  # macOS/Linux
netstat -ano | findstr :8080  # Windows

# Kill process if needed
kill -9 $(lsof -t -i:8080)  # macOS/Linux
```

## Environment Validation Checklist

Before starting tests, verify:

- [ ] PostgreSQL is running and accessible
- [ ] Redis is running and accessible
- [ ] Database migrations are applied
- [ ] Database is empty (no tenants, no users except bootstrap)
- [ ] Server starts without errors
- [ ] Health check returns `200 OK`
- [ ] All environment variables are set
- [ ] System roles exist (3 roles)
- [ ] System capabilities exist (11 capabilities)
- [ ] System permissions exist (13 permissions)

## Next Steps

Once the environment is set up:

1. **Read** `TEST_EXECUTION_GUIDE.md` for test execution instructions
2. **Start** with `TEST_CASES/SYSTEM_LEVEL.md` for system bootstrap tests
3. **Follow** test cases in order for complete coverage

---

**Important**: Always start testing from a clean database state to ensure reproducible results.

