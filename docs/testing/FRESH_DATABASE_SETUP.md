# Fresh Database Setup Guide

## Quick Reset

To drop and recreate the database from scratch:

```bash
./scripts/reset-database.sh
```

This script will:
1. Drop the existing `iam` database
2. Create a fresh `iam` database
3. Run all migrations from scratch
4. Verify the setup

---

## Manual Steps (Alternative)

If you prefer to do it manually:

### 1. Drop Database

```bash
export DATABASE_HOST=127.0.0.1
export DATABASE_PORT=5433
export DATABASE_USER=dcim_user
export DATABASE_PASSWORD=dcim_password
export DATABASE_NAME=iam

PGPASSWORD=${DATABASE_PASSWORD} psql -h ${DATABASE_HOST} -p ${DATABASE_PORT} -U ${DATABASE_USER} -d postgres -c "DROP DATABASE IF EXISTS ${DATABASE_NAME};"
```

### 2. Create Fresh Database

```bash
PGPASSWORD=${DATABASE_PASSWORD} psql -h ${DATABASE_HOST} -p ${DATABASE_PORT} -U ${DATABASE_USER} -d postgres -c "CREATE DATABASE ${DATABASE_NAME};"
```

### 3. Run All Migrations

```bash
./scripts/migrate.sh up
```

Or directly:

```bash
DATABASE_URL="postgres://${DATABASE_USER}:${DATABASE_PASSWORD}@${DATABASE_HOST}:${DATABASE_PORT}/${DATABASE_NAME}?sslmode=disable"
migrate -path ./migrations -database "$DATABASE_URL" up
```

### 4. Verify Migration

```bash
./scripts/migrate.sh version
```

Expected output: `23` (or higher)

---

## Verification Checklist

After reset, verify:

- [ ] Migration version is 23 (or higher)
- [ ] Permissions table has `tenant_id`, `updated_at`, `deleted_at` columns
- [ ] All key tables exist: `tenants`, `users`, `roles`, `permissions`, etc.
- [ ] Can start backend server without errors
- [ ] Can create a tenant via API
- [ ] Tenant initialization creates predefined roles and permissions

---

## Testing New Tenant Creation

After database reset, test tenant creation:

```bash
# 1. Start backend
./scripts/start-backend-local.sh

# 2. Create a tenant (in another terminal)
curl -X POST http://localhost:8080/system/tenants \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <system_admin_token>" \
  -d '{
    "name": "Test Tenant",
    "domain": "test.local"
  }'

# 3. Verify roles were created
# Check that tenant_owner, tenant_admin, tenant_auditor roles exist

# 4. Verify permissions were created
# Check that all permissions use tenant.* namespace
```

---

## Expected Permissions After Reset

When a new tenant is created, you should see these permissions (all with `tenant.*` namespace):

- `tenant.users.create`
- `tenant.users.read`
- `tenant.users.update`
- `tenant.users.delete`
- `tenant.users.manage`
- `tenant.roles.create`
- `tenant.roles.read`
- `tenant.roles.update`
- `tenant.roles.delete`
- `tenant.roles.manage`
- `tenant.permissions.create`
- `tenant.permissions.read`
- `tenant.permissions.update`
- `tenant.permissions.delete`
- `tenant.permissions.manage`
- `tenant.settings.read`
- `tenant.settings.update`
- `tenant.audit.read`
- `tenant.admin.access`

---

## Troubleshooting

### Migration Fails

```bash
# Check migration version
./scripts/migrate.sh version

# Force to specific version if needed
./scripts/migrate.sh force 23
```

### Database Connection Issues

```bash
# Test connection
PGPASSWORD=${DATABASE_PASSWORD} psql -h ${DATABASE_HOST} -p ${DATABASE_PORT} -U ${DATABASE_USER} -d ${DATABASE_NAME} -c "SELECT version();"
```

### Permissions Not Created

- Check tenant initialization logs
- Verify `tenantInitializer` is properly injected
- Check that tenant creation succeeded

---

## Notes

- **All data will be lost** when resetting the database
- Make sure to backup any important data first
- The reset script includes a 5-second warning before proceeding
- All migrations will run in order (000001 through 000023)

