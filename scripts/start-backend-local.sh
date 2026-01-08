#!/bin/bash

# Start IAM Backend with Local PostgreSQL Configuration
# Database: 127.0.0.1:5433
# User: dcim_user
# Password: dcim_password

set -e

echo "🚀 Starting ARauth Identity IAM Backend..."
echo ""

# Set database configuration
export DATABASE_HOST=127.0.0.1
export DATABASE_PORT=5433
export DATABASE_USER=dcim_user
export DATABASE_PASSWORD=dcim_password
export DATABASE_NAME=iam
export DATABASE_SSL_MODE=disable

# Set other required environment variables
export JWT_SECRET=test-jwt-secret-key-min-32-characters-long-for-local-dev
export ENCRYPTION_KEY=01234567890123456789012345678901

# Optional: Redis (skip if not running)
export REDIS_HOST=localhost
export REDIS_PORT=6379
# export REDIS_PASSWORD=your_redis_password

# Optional: Hydra (skip if not running)
export HYDRA_ADMIN_URL=http://localhost:4445
export HYDRA_PUBLIC_URL=http://localhost:4444

# Logging
export LOG_LEVEL=info
export LOG_FORMAT=json

echo "📋 Configuration:"
echo "  Database: ${DATABASE_HOST}:${DATABASE_PORT}/${DATABASE_NAME}"
echo "  User: ${DATABASE_USER}"
echo "  Server: http://localhost:8080"
echo ""

# Check database connection
echo "🔍 Checking database connection..."
PGPASSWORD=${DATABASE_PASSWORD} psql -h ${DATABASE_HOST} -p ${DATABASE_PORT} -U ${DATABASE_USER} -d ${DATABASE_NAME} -c "SELECT version();" > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✅ Database connection successful"
else
    echo "❌ Database connection failed. Please check:"
    echo "   - PostgreSQL is running on ${DATABASE_HOST}:${DATABASE_PORT}"
    echo "   - Database '${DATABASE_NAME}' exists"
    echo "   - User '${DATABASE_USER}' has access"
    echo "   - Migrations have been run"
    exit 1
fi

echo ""
echo "🚀 Starting server..."
echo ""

# Start the server
go run cmd/server/main.go

