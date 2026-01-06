#!/bin/bash

# Nuage Identity - Database Migration Helper Script

set -e

MIGRATIONS_PATH="./migrations"
DATABASE_URL="${DATABASE_URL:-postgres://iam_user:change-me@localhost:5432/iam?sslmode=disable}"

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
        echo "  DATABASE_URL    Database connection string (default: postgres://iam_user:change-me@localhost:5432/iam?sslmode=disable)"
        ;;
esac

