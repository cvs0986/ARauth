#!/bin/bash
# Test migration locally
# Usage: ./scripts/test-migration.sh [database_url]

set -e

DB_URL="${1:-postgres://postgres:postgres@localhost:5433/iam_test?sslmode=disable}"

echo "Testing migration with database: $DB_URL"
echo ""

# Check if migrate is installed
if ! command -v migrate &> /dev/null; then
    echo "Installing migrate..."
    export PATH=$PATH:~/go/bin
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
fi

# Create test database
echo "Creating test database..."
DB_NAME=$(echo "$DB_URL" | sed -n 's/.*\/\([^?]*\).*/\1/p')
DB_BASE=$(echo "$DB_URL" | sed -n 's|postgres://\([^@]*\)@\([^/]*\)/.*|\1@\2/postgres|p')

psql "$DB_BASE?sslmode=disable" -c "DROP DATABASE IF EXISTS ${DB_NAME};" 2>/dev/null || true
psql "$DB_BASE?sslmode=disable" -c "CREATE DATABASE ${DB_NAME};" 2>/dev/null || echo "Database might already exist"

# Run migrations
echo "Running migrations..."
migrate -path migrations -database "$DB_URL" up

# Check migration version
echo ""
echo "Migration version:"
migrate -path migrations -database "$DB_URL" version

# List indexes
echo ""
echo "Created indexes:"
psql "$DB_URL" -c "SELECT schemaname, tablename, indexname FROM pg_indexes WHERE tablename IN ('users', 'tenants', 'roles', 'permissions', 'credentials', 'user_roles', 'role_permissions', 'mfa_recovery_codes', 'audit_logs') ORDER BY tablename, indexname;" 2>/dev/null || echo "Could not list indexes"

echo ""
echo "✅ Migration test completed!"

