#!/bin/bash

# Nuage Identity - Database Migration Helper Script

set -e

MIGRATIONS_PATH="./migrations"

# Default database connection details (can be overridden by DATABASE_URL env var)
DATABASE_HOST="${DATABASE_HOST:-127.0.0.1}"
DATABASE_PORT="${DATABASE_PORT:-5433}"
DATABASE_USER="${DATABASE_USER:-dcim_user}"
DATABASE_PASSWORD="${DATABASE_PASSWORD:-dcim_password}"
DATABASE_NAME="${DATABASE_NAME:-iam}"
DATABASE_SSL_MODE="${DATABASE_SSL_MODE:-disable}"

# Build DATABASE_URL if not provided
if [ -z "$DATABASE_URL" ]; then
    DATABASE_URL="postgres://${DATABASE_USER}:${DATABASE_PASSWORD}@${DATABASE_HOST}:${DATABASE_PORT}/${DATABASE_NAME}?sslmode=${DATABASE_SSL_MODE}"
fi

# Ensure migrate tool is available
export PATH=$PATH:/home/eshwar/go-install/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

if ! command -v migrate &> /dev/null; then
    echo "ERROR: migrate tool is not installed"
    echo "Please run: make install-tools"
    exit 1
fi

# Parse command
case "${1:-help}" in
    up)
        echo "Running migrations up..."
        migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" up
        ;;
    down)
        echo "Rolling back migrations..."
        migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" down
        ;;
    version)
        echo "Checking migration version..."
        migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" version
        ;;
    force)
        if [ -z "$2" ]; then
            echo "ERROR: Force requires a version number"
            echo "Usage: $0 force <version>"
            exit 1
        fi
        echo "Forcing migration to version $2..."
        migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" force "$2"
        ;;
    create)
        if [ -z "$2" ]; then
            echo "ERROR: Create requires a migration name"
            echo "Usage: $0 create <name>"
            exit 1
        fi
        echo "Creating migration: $2"
        migrate create -ext sql -dir "$MIGRATIONS_PATH" -seq "$2"
        ;;
    help|*)
        echo "Nuage Identity - Database Migration Helper"
        echo ""
        echo "Usage: $0 <command> [options]"
        echo ""
        echo "Commands:"
        echo "  up              Run all pending migrations"
        echo "  down            Rollback the last migration"
        echo "  version         Show current migration version"
        echo "  force <version> Force migration to specific version"
        echo "  create <name>   Create a new migration"
        echo "  help            Show this help message"
        echo ""
        echo "Environment Variables:"
        echo "  DATABASE_URL         Database connection string (overrides individual settings)"
        echo "  DATABASE_HOST        Database host (default: 127.0.0.1)"
        echo "  DATABASE_PORT        Database port (default: 5433)"
        echo "  DATABASE_USER        Database user (default: dcim_user)"
        echo "  DATABASE_PASSWORD   Database password (default: dcim_password)"
        echo "  DATABASE_NAME        Database name (default: iam)"
        echo "  DATABASE_SSL_MODE    SSL mode (default: disable)"
        ;;
esac

