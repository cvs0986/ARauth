#!/bin/bash

# Reset IAM Database - Drop and recreate from scratch
# This will delete all data and run all migrations fresh

set -e

# Database configuration (matching start-backend-local.sh)
export DATABASE_HOST=127.0.0.1
export DATABASE_PORT=5433
export DATABASE_USER=dcim_user
export DATABASE_PASSWORD=dcim_password
export DATABASE_NAME=iam
export DATABASE_SSL_MODE=disable

# Build connection strings
DB_URL="postgres://${DATABASE_USER}:${DATABASE_PASSWORD}@${DATABASE_HOST}:${DATABASE_PORT}/${DATABASE_NAME}?sslmode=${DATABASE_SSL_MODE}"
DB_BASE_URL="postgres://${DATABASE_USER}:${DATABASE_PASSWORD}@${DATABASE_HOST}:${DATABASE_PORT}/postgres?sslmode=${DATABASE_SSL_MODE}"

echo "🗑️  Resetting IAM Database..."
echo ""
echo "⚠️  WARNING: This will DELETE ALL DATA in the '${DATABASE_NAME}' database!"
echo "   Press Ctrl+C within 5 seconds to cancel..."
echo ""
sleep 5

echo "📋 Configuration:"
echo "  Host: ${DATABASE_HOST}:${DATABASE_PORT}"
echo "  User: ${DATABASE_USER}"
echo "  Database: ${DATABASE_NAME}"
echo ""

# Step 1: Drop existing database
echo "1️⃣  Dropping existing database..."
PGPASSWORD=${DATABASE_PASSWORD} psql -h ${DATABASE_HOST} -p ${DATABASE_PORT} -U ${DATABASE_USER} -d postgres -c "DROP DATABASE IF EXISTS ${DATABASE_NAME};" 2>/dev/null || {
    echo "   ⚠️  Could not drop database (might not exist or connection failed)"
}

# Step 2: Create fresh database
echo "2️⃣  Creating fresh database..."
PGPASSWORD=${DATABASE_PASSWORD} psql -h ${DATABASE_HOST} -p ${DATABASE_PORT} -U ${DATABASE_USER} -d postgres -c "CREATE DATABASE ${DATABASE_NAME};" || {
    echo "   ❌ Failed to create database"
    exit 1
}
echo "   ✅ Database created"

# Step 3: Run all migrations
echo "3️⃣  Running all migrations..."
export PATH=$PATH:/home/eshwar/go-install/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

if ! command -v migrate &> /dev/null; then
    echo "   ❌ ERROR: migrate tool is not installed"
    echo "   Please run: make install-tools"
    exit 1
fi

migrate -path ./migrations -database "$DB_URL" up || {
    echo "   ❌ Migration failed"
    exit 1
}

echo "   ✅ All migrations applied"

# Step 4: Verify migration status
echo "4️⃣  Verifying migration status..."
VERSION=$(migrate -path ./migrations -database "$DB_URL" version 2>&1 | grep -oE '[0-9]+' | head -1)
echo "   Current migration version: ${VERSION}"

# Step 5: Check key tables exist
echo "5️⃣  Verifying key tables..."
PGPASSWORD=${DATABASE_PASSWORD} psql "$DB_URL" -c "
SELECT 
    table_name,
    (SELECT COUNT(*) FROM information_schema.columns WHERE table_name = t.table_name) as column_count
FROM information_schema.tables t
WHERE table_schema = 'public' 
  AND table_name IN ('tenants', 'users', 'roles', 'permissions', 'user_roles', 'role_permissions')
ORDER BY table_name;
" 2>/dev/null || echo "   ⚠️  Could not verify tables"

# Step 6: Check permissions table structure
echo "6️⃣  Checking permissions table structure..."
PGPASSWORD=${DATABASE_PASSWORD} psql "$DB_URL" -c "
SELECT 
    column_name, 
    data_type, 
    is_nullable
FROM information_schema.columns 
WHERE table_name = 'permissions' 
  AND column_name IN ('tenant_id', 'updated_at', 'deleted_at')
ORDER BY column_name;
" 2>/dev/null || echo "   ⚠️  Could not check permissions structure"

echo ""
echo "✅ Database reset complete!"
echo ""
echo "📝 Next steps:"
echo "   1. Start the backend server: ./scripts/start-backend-local.sh"
echo "   2. Create a tenant via API or admin dashboard"
echo "   3. Verify predefined roles and permissions are created"
echo ""

