# Testing Strategy

This document describes the testing approach, strategies, and best practices for ARauth Identity.

## 🎯 Testing Principles

1. **Test Pyramid**: More unit tests, fewer integration tests, minimal E2E tests
2. **Test Coverage**: > 80% code coverage
3. **Fast Tests**: Unit tests should run in < 1 second
4. **Isolated Tests**: Tests should not depend on each other
5. **Deterministic**: Tests should produce consistent results

## 📊 Test Pyramid

```
        /\
       /  \      E2E Tests (5%)
      /____\
     /      \    Integration Tests (15%)
    /________\
   /          \  Unit Tests (80%)
  /____________\
```

## 🧪 Unit Tests

### Purpose

Test individual functions and methods in isolation.

### Scope

- Business logic
- Utility functions
- Service methods
- Repository methods (with mocks)

### Example

```go
func TestPasswordHasher_Hash(t *testing.T) {
    hasher := NewPasswordHasher()
    
    password := "test-password-123"
    hash, err := hasher.Hash(password)
    
    assert.NoError(t, err)
    assert.NotEmpty(t, hash)
    assert.True(t, hasher.Verify(password, hash))
    assert.False(t, hasher.Verify("wrong-password", hash))
}
```

### Best Practices

- Use table-driven tests for multiple scenarios
- Mock external dependencies
- Test both success and error cases
- Keep tests focused and simple

## 🔗 Integration Tests

### Purpose

Test component interactions and database operations.

### Scope

- API endpoints
- Database operations
- Service integrations
- Repository implementations

### Setup

**Test Containers**:
```go
func setupTestDB(t *testing.T) *sql.DB {
    // Start PostgreSQL test container
    container := testcontainers.PostgreSQLContainer{
        Image: "postgres:14",
    }
    // ...
    return db
}
```

### Example

```go
func TestUserRepository_Create(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    repo := NewUserRepository(db)
    
    user := &User{
        Username: "testuser",
        Email:    "test@example.com",
        TenantID: "tenant-123",
    }
    
    err := repo.Create(context.Background(), user)
    assert.NoError(t, err)
    assert.NotEmpty(t, user.ID)
    
    // Verify in database
    retrieved, err := repo.GetByID(context.Background(), user.ID)
    assert.NoError(t, err)
    assert.Equal(t, user.Username, retrieved.Username)
}
```

## 🌐 API Tests

### Purpose

Test HTTP endpoints end-to-end.

### Scope

- Request/response handling
- Authentication/authorization
- Error handling
- Status codes

### Example

```go
func TestLoginEndpoint(t *testing.T) {
    app := setupTestApp(t)
    
    req := LoginRequest{
        Username: "testuser",
        Password: "password123",
        TenantID: "tenant-123",
    }
    
    body, _ := json.Marshal(req)
    w := httptest.NewRecorder()
    httpReq := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
    httpReq.Header.Set("Content-Type", "application/json")
    
    app.ServeHTTP(w, httpReq)
    
    assert.Equal(t, 200, w.Code)
    
    var response LoginResponse
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.NotEmpty(t, response.AccessToken)
}
```

## 🔐 Security Tests

### Purpose

Test security features and vulnerabilities.

### Scope

- Password hashing
- JWT validation
- MFA
- Rate limiting
- SQL injection prevention

### Example

```go
func TestSQLInjectionPrevention(t *testing.T) {
    db := setupTestDB(t)
    repo := NewUserRepository(db)
    
    // Attempt SQL injection
    maliciousUsername := "admin' OR '1'='1"
    
    user, err := repo.GetByUsername(context.Background(), maliciousUsername, "tenant-123")
    
    // Should not find user, not execute injection
    assert.Error(t, err)
    assert.Nil(t, user)
}
```

## ⚡ Performance Tests

### Purpose

Verify performance targets are met.

### Scope

- Login latency
- Token issuance latency
- Database query performance
- Concurrent request handling

### Tools

- **k6**: Load testing
- **go test -bench**: Benchmarking

### Example

```go
func BenchmarkLogin(b *testing.B) {
    app := setupTestApp(b)
    
    req := LoginRequest{
        Username: "testuser",
        Password: "password123",
        TenantID: "tenant-123",
    }
    
    body, _ := json.Marshal(req)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        w := httptest.NewRecorder()
        httpReq := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
        app.ServeHTTP(w, httpReq)
    }
}
```

## 🧩 Contract Tests

### Purpose

Test repository interfaces and service contracts.

### Scope

- Repository interface compliance
- Service interface compliance
- API contract compliance

### Example

```go
func TestUserRepositoryContract(t *testing.T) {
    tests := []struct {
        name string
        repo UserRepository
    }{
        {"PostgreSQL", NewPostgreSQLUserRepository(db)},
        {"MySQL", NewMySQLUserRepository(db)},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            testUserRepository(t, tt.repo)
        })
    }
}
```

## 🔄 E2E Tests

### Purpose

Test complete user flows.

### Scope

- Login flow
- MFA flow
- Token refresh flow
- User management flow

### Example

```go
func TestE2ELoginFlow(t *testing.T) {
    // Setup
    app := setupTestApp(t)
    createTestUser(t, "testuser", "password123")
    
    // 1. Login
    loginResp := login(t, app, "testuser", "password123")
    assert.NotEmpty(t, loginResp.AccessToken)
    
    // 2. Use token to access protected endpoint
    userResp := getUser(t, app, loginResp.AccessToken)
    assert.Equal(t, "testuser", userResp.Username)
    
    // 3. Refresh token
    refreshResp := refreshToken(t, app, loginResp.RefreshToken)
    assert.NotEmpty(t, refreshResp.AccessToken)
}
```

## 📋 Test Coverage

### Coverage Goals

- **Overall**: > 80%
- **Business Logic**: > 90%
- **API Handlers**: > 80%
- **Repositories**: > 85%

### Coverage Tools

```bash
# Generate coverage
go test -coverprofile=coverage.out ./...

# View coverage
go tool cover -html=coverage.out

# Coverage in CI
go test -cover -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## 🛠️ Testing Tools

### Frameworks

- **testing**: Standard library
- **testify**: Assertions and mocks
- **testcontainers**: Docker containers for testing

### Mocking

```go
// Generate mocks
//go:generate mockgen -source=repository.go -destination=mock_repository.go

type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*User, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*User), args.Error(1)
}
```

## 🚀 CI/CD Testing

### Continuous Integration

**Pipeline Stages**:
1. Lint and format check
2. Unit tests
3. Integration tests
4. Security scan
5. Coverage report

**Example**:
```yaml
# .github/workflows/test.yml
- name: Run tests
  run: go test -v -coverprofile=coverage.out ./...

- name: Upload coverage
  uses: codecov/codecov-action@v3
  with:
    file: ./coverage.out
```

## 📊 Test Metrics

### Key Metrics

- **Test Coverage**: > 80%
- **Test Execution Time**: < 5 minutes
- **Flaky Test Rate**: < 1%
- **Test Pass Rate**: > 99%

### Monitoring

- Track test execution time
- Monitor flaky tests
- Track coverage trends

## 📚 Related Documentation

- [Development Strategy](../planning/strategy.md)
- [Technical Stack](./tech-stack.md)
- [API Design](./api-design.md)

