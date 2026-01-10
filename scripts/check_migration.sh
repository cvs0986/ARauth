#!/bin/bash

# Check if migration 000023 has been applied
# Usage: ./scripts/check_migration.sh [database_url]

set -e

# Default database connection details (can be overridden by DATABASE_URL env var or argument)
DATABASE_HOST="${DATABASE_HOST:-127.0.0.1}"
DATABASE_PORT="${DATABASE_PORT:-5433}"
DATABASE_USER="${DATABASE_USER:-dcim_user}"
DATABASE_PASSWORD="${DATABASE_PASSWORD:-dcim_password}"
DATABASE_NAME="${DATABASE_NAME:-iam}"
DATABASE_SSL_MODE="${DATABASE_SSL_MODE:-disable}"

# Use provided argument, then DATABASE_URL env var, or build from components
if [ -n "$1" ]; then
    DATABASE_URL="$1"
elif [ -n "$DATABASE_URL" ]; then
    # Use existing DATABASE_URL
    :
else
    DATABASE_URL="postgres://${DATABASE_USER}:${DATABASE_PASSWORD}@${DATABASE_HOST}:${DATABASE_PORT}/${DATABASE_NAME}?sslmode=${DATABASE_SSL_MODE}"
fi

echo "Checking migration status for: $DATABASE_URL"
echo ""

# Check migration version
echo "1. Current migration version:"
psql "$DATABASE_URL" -c "SELECT version, dirty FROM schema_migrations;" 2>/dev/null || echo "  ❌ Could not check migration version (database connection failed)"

echo ""
echo "2. Permissions table structure (checking for tenant_id, updated_at, deleted_at):"
psql "$DATABASE_URL" -c "
SELECT 
    column_name, 
    data_type, 
    is_nullable
FROM information_schema.columns 
WHERE table_name = 'permissions' 
  AND column_name IN ('tenant_id', 'updated_at', 'deleted_at')
ORDER BY column_name;
" 2>/dev/null || echo "  ❌ Could not check permissions table structure"

echo ""
echo "3. Checking for required indexes:"
psql "$DATABASE_URL" -c "
SELECT 
    indexname
FROM pg_indexes 
WHERE tablename = 'permissions' 
  AND indexname IN (
    'idx_permissions_tenant_id',
    'idx_permissions_tenant_resource_action',
    'idx_permissions_global_resource_action',
    'idx_permissions_deleted_at'
  )
ORDER BY indexname;
" 2>/dev/null || echo "  ❌ Could not check indexes"

echo ""
echo "✅ Migration check complete!"
echo ""
echo "Expected results if migration 000023 is applied:"
echo "  - schema_migrations.version should be 23 or higher"
echo "  - permissions table should have: tenant_id, updated_at, deleted_at columns"
echo "  - All 4 indexes should exist"

