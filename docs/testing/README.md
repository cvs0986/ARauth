# Testing Guide

This document describes the testing strategy and how to run tests for the Nuage Identity IAM platform.

## Test Structure

```
.
├── storage/postgres/
│   └── *_test.go          # Repository tests
├── identity/*/service_test.go  # Service layer tests
├── api/handlers/*_test.go      # Handler tests
└── internal/testutil/          # Test utilities and mocks
```

## Running Tests

### All Tests

```bash
make test
# or
go test ./... -v
```

### Unit Tests Only

```bash
make test-unit
# or
go test ./... -v -short
```

### Integration Tests

```bash
make test-integration
# or
go test ./... -v -tags=integration
```

### Test Coverage

```bash
make test-coverage
# Generates coverage.html
```

## Test Categories

### Unit Tests

Unit tests test individual components in isolation using mocks.

**Location**: `*_test.go` files alongside source files

**Examples**:
- Repository tests (with test database)
- Service tests (with mocked repositories)
- Handler tests (with mocked services)

### Integration Tests

Integration tests test multiple components working together.

**Requirements**:
- Test database (PostgreSQL)
- Test Redis instance (optional)
- Environment variable: `TEST_DATABASE_URL`

**Setup**:

```bash
export TEST_DATABASE_URL="postgres://user:password@localhost:5432/iam_test?sslmode=disable"
```

### Performance Benchmarks

Benchmark tests measure performance of critical operations.

```bash
make benchmark
# or
go test ./... -bench=. -benchmem
```

**Key Benchmarks**:
- Password hashing
- Password verification
- TOTP generation
- Encryption/decryption

## Test Database Setup

1. **Create test database**:

```sql
CREATE DATABASE iam_test;
CREATE USER iam_test_user WITH PASSWORD 'test_password';
GRANT ALL PRIVILEGES ON DATABASE iam_test TO iam_test_user;
```

2. **Run migrations**:

```bash
export DATABASE_URL="postgres://iam_test_user:test_password@localhost:5432/iam_test?sslmode=disable"
make migrate-up
```

3. **Run tests**:

```bash
export TEST_DATABASE_URL="postgres://iam_test_user:test_password@localhost:5432/iam_test?sslmode=disable"
make test-integration
```

## Load Testing

### Using the Load Test Script

```bash
# Set API URL and tenant ID
export API_URL="http://localhost:8080"
export TENANT_ID="your-tenant-id"

# Run load test
./scripts/load-test.sh
```

### Using hey

```bash
# Install hey
go install github.com/rakyll/hey@latest

# Run load test
hey -n 10000 -c 100 -m GET http://localhost:8080/health
```

### Using Apache Bench

```bash
# Install Apache Bench
sudo apt-get install apache2-utils  # Ubuntu/Debian
sudo yum install httpd-tools         # CentOS/RHEL

# Run load test
ab -n 10000 -c 100 http://localhost:8080/health
```

## Test Coverage Goals

- **Unit Tests**: 80%+ coverage
- **Integration Tests**: Critical paths covered
- **E2E Tests**: Main user flows covered

## Writing Tests

### Example: Service Test

```go
func TestService_Create(t *testing.T) {
    mockRepo := new(MockUserRepository)
    service := NewService(mockRepo)
    
    req := &CreateUserRequest{
        Username: "testuser",
        Email:    "test@example.com",
        Password: "SecurePass123!",
    }
    
    mockRepo.On("GetByUsername", mock.Anything, req.Username, tenantID).
        Return(nil, errors.New("not found"))
    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).
        Return(nil)
    
    user, err := service.Create(context.Background(), req)
    require.NoError(t, err)
    assert.NotNil(t, user)
    
    mockRepo.AssertExpectations(t)
}
```

## Continuous Integration

Tests should be run in CI/CD pipeline:

```yaml
# Example GitHub Actions
- name: Run tests
  run: make test

- name: Generate coverage
  run: make test-coverage

- name: Upload coverage
  uses: codecov/codecov-action@v3
```

## Troubleshooting

### Tests Skipped

If tests are skipped with "database not available":
1. Check `TEST_DATABASE_URL` is set
2. Verify database is running
3. Check database credentials

### Slow Tests

- Use `-short` flag for unit tests only
- Use test database on SSD
- Run tests in parallel: `go test ./... -parallel 4`

## Best Practices

1. **Isolation**: Each test should be independent
2. **Cleanup**: Clean up test data after tests
3. **Mocks**: Use mocks for external dependencies
4. **Fixtures**: Use test fixtures for complex data
5. **Naming**: Use descriptive test names
6. **Assertions**: Use clear assertions with messages

