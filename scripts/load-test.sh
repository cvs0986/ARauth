#!/bin/bash

# Load testing script for Nuage Identity IAM API
# Requires: Apache Bench (ab) or hey

API_URL="${API_URL:-http://localhost:8080}"
TENANT_ID="${TENANT_ID:-00000000-0000-0000-0000-000000000000}"

echo "Starting load test for Nuage Identity IAM API"
echo "API URL: $API_URL"
echo "Tenant ID: $TENANT_ID"
echo ""

# Check if hey is installed
if command -v hey &> /dev/null; then
    echo "Using 'hey' for load testing"
    
    # Health check endpoint
    echo "Testing /health endpoint..."
    hey -n 1000 -c 10 -m GET "$API_URL/health"
    
    # User list endpoint (requires tenant header)
    echo ""
    echo "Testing /api/v1/users endpoint..."
    hey -n 500 -c 5 -H "X-Tenant-ID: $TENANT_ID" -m GET "$API_URL/api/v1/users"
    
elif command -v ab &> /dev/null; then
    echo "Using 'ab' (Apache Bench) for load testing"
    
    # Health check endpoint
    echo "Testing /health endpoint..."
    ab -n 1000 -c 10 "$API_URL/health"
    
    # User list endpoint
    echo ""
    echo "Testing /api/v1/users endpoint..."
    ab -n 500 -c 5 -H "X-Tenant-ID: $TENANT_ID" "$API_URL/api/v1/users"
    
else
    echo "Error: Neither 'hey' nor 'ab' is installed"
    echo "Install hey: go install github.com/rakyll/hey@latest"
    echo "Or install Apache Bench: sudo apt-get install apache2-utils"
    exit 1
fi

echo ""
echo "Load test completed"

