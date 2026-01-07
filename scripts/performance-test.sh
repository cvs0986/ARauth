#!/bin/bash

# Performance Testing Script for Nuage Identity IAM
# This script runs performance benchmarks and load tests

set -e

echo "🚀 Nuage Identity - Performance Testing"
echo "========================================"
echo ""

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
API_URL="${API_URL:-http://localhost:8080}"
TENANT_ID="${TENANT_ID:-}"
CONCURRENT_USERS="${CONCURRENT_USERS:-100}"
TOTAL_REQUESTS="${TOTAL_REQUESTS:-10000}"

echo -e "${BLUE}Configuration:${NC}"
echo "  API URL: $API_URL"
echo "  Concurrent Users: $CONCURRENT_USERS"
echo "  Total Requests: $TOTAL_REQUESTS"
echo ""

# Check if hey is installed
if ! command -v hey &> /dev/null; then
    echo -e "${YELLOW}⚠️  'hey' not found. Installing...${NC}"
    go install github.com/rakyll/hey@latest
    export PATH=$PATH:$(go env GOPATH)/bin
fi

# Run benchmarks
echo -e "${GREEN}📊 Running Go Benchmarks...${NC}"
echo ""
make benchmark
echo ""

# Health check load test
echo -e "${GREEN}🔥 Load Testing Health Endpoint...${NC}"
echo ""
hey -n $TOTAL_REQUESTS -c $CONCURRENT_USERS -m GET "$API_URL/health"
echo ""

# If tenant ID is provided, test authenticated endpoints
if [ -n "$TENANT_ID" ]; then
    echo -e "${GREEN}🔥 Load Testing Authenticated Endpoints...${NC}"
    echo ""
    echo "Testing with tenant: $TENANT_ID"
    hey -n $((TOTAL_REQUESTS / 2)) -c $((CONCURRENT_USERS / 2)) -m GET \
        -H "X-Tenant-ID: $TENANT_ID" \
        "$API_URL/api/v1/users"
    echo ""
fi

echo -e "${GREEN}✅ Performance testing complete!${NC}"

