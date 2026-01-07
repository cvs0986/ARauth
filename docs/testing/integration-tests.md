# Integration Tests Guide

This document describes how to write and run integration tests for the Nuage Identity IAM platform.

## Overview

Integration tests verify that multiple components work together correctly. Unlike unit tests, integration tests require:
- Real database connection (PostgreSQL)
- Redis connection (optional, for some tests)
- ORY Hydra instance (optional, for authentication tests)

## Setup

### 1. Test Database

Create a separate test database:

```sql
CREATE DATABASE iam_test;
CREATE USER iam_test_user WITH PASSWORD 'test_password';
GRANT ALL PRIVILEGES ON DATABASE iam_test TO iam_test_user;
```

### 2. Environment Variables

Set the test database URL:

```bash
export TEST_DATABASE_URL="postgres://iam_test_user:test_password@localhost:5432/iam_test?sslmode=disable"
```

### 3. Run Migrations

```bash
export DATABASE_URL="$TEST_DATABASE_URL"
make migrate-up
```

## Running Integration Tests

### All Integration Tests

```bash
go test ./... -v -tags=integration
```

### Specific Package

```bash
go test ./storage/postgres/... -v -tags=integration
```

### With Coverage

```bash
go test ./... -tags=integration -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## Writing Integration Tests

### Example: User Repository Test

```go
// +build integration

package postgres

import (
    "context"
    "testing"
    
    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestUserRepository_Create_Integration(t *testing.T) {
    db, cleanup := setupTestDB(t)
    defer cleanup()
    
    repo := NewUserRepository(db)
    
    user := &models.User{
        ID:       uuid.New(),
        TenantID: uuid.New(),
        Username: "testuser",
        Email:    "test@example.com",
    }
    
    err := repo.Create(context.Background(), user)
    require.NoError(t, err)
    assert.NotEqual(t, uuid.Nil, user.ID)
}
```

### Test Structure

1. **Setup**: Create test database connection
2. **Arrange**: Prepare test data
3. **Act**: Execute the operation
4. **Assert**: Verify results
5. **Cleanup**: Remove test data

## Test Categories

### Database Integration Tests

- Repository operations
- Transaction handling
- Query performance
- Data integrity

### Service Integration Tests

- Business logic with real repositories
- Multi-step workflows
- Error handling across layers

### API Integration Tests

- HTTP endpoints
- Request/response handling
- Middleware chain
- Authentication/authorization

## Best Practices

1. **Isolation**: Each test should be independent
2. **Cleanup**: Always clean up test data
3. **Fixtures**: Use test fixtures for complex data
4. **Parallel**: Run tests in parallel when possible
5. **Naming**: Use descriptive test names

## Continuous Integration

Integration tests should run in CI/CD:

```yaml
# .github/workflows/integration-tests.yml
name: Integration Tests

on:
  push:
    branches: [main]
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: test_password
          POSTGRES_DB: iam_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run migrations
        run: |
          export DATABASE_URL="postgres://postgres:test_password@localhost:5432/iam_test?sslmode=disable"
          make migrate-up
      
      - name: Run integration tests
        run: |
          export TEST_DATABASE_URL="postgres://postgres:test_password@localhost:5432/iam_test?sslmode=disable"
          go test ./... -v -tags=integration
```

## Troubleshooting

### Tests Skipped

If tests are skipped:
1. Check `TEST_DATABASE_URL` is set
2. Verify database is running
3. Check database credentials

### Slow Tests

- Use test database on SSD
- Run tests in parallel
- Optimize test data setup

### Database Locks

- Ensure proper cleanup
- Use transactions for test isolation
- Check for hanging connections

