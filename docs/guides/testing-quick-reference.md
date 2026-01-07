# Testing Quick Reference

Quick reference guide for testing Nuage Identity IAM locally.

## 🚀 Quick Start

### 1. Start Backend
```bash
# Ensure PostgreSQL is running on port 5433
# Start the API server
go run cmd/server/main.go
# API available at http://localhost:8080
```

### 2. Verify Backend
```bash
curl http://localhost:8080/health
```

### 3. Test API Endpoints

#### Create Tenant
```bash
curl -X POST http://localhost:8080/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Company",
    "domain": "test.com",
    "status": "active"
  }'
```

#### Create User
```bash
# Replace TENANT_ID with actual tenant ID
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: TENANT_ID" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "SecurePassword123!",
    "first_name": "Test",
    "last_name": "User"
  }'
```

#### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: TENANT_ID" \
  -d '{
    "username": "testuser",
    "password": "SecurePassword123!"
  }'
```

## 📋 Test Scenarios Checklist

### Authentication Flow
- [ ] Create tenant
- [ ] Create user
- [ ] Login with credentials
- [ ] Receive access token
- [ ] Use access token for API calls
- [ ] Refresh token
- [ ] Logout

### MFA Flow
- [ ] Enroll in MFA (POST /mfa/enroll)
- [ ] Get QR code/secret
- [ ] Verify MFA setup (POST /mfa/verify)
- [ ] Login with MFA challenge
- [ ] Verify MFA code (POST /mfa/challenge/verify)
- [ ] Use recovery code

### User Management
- [ ] List users (GET /users)
- [ ] Get user by ID (GET /users/:id)
- [ ] Update user (PUT /users/:id)
- [ ] Delete user (DELETE /users/:id)

### Role Management
- [ ] Create role (POST /roles)
- [ ] List roles (GET /roles)
- [ ] Assign permissions to role
- [ ] Assign role to user

### Permission Management
- [ ] Create permission (POST /permissions)
- [ ] List permissions (GET /permissions)
- [ ] Assign permission to role
- [ ] Test permission-based access

## 🧪 Running Tests

### Backend Tests
```bash
# All tests
make test

# With coverage
make test-coverage

# Specific package
go test ./api/handlers/...

# E2E tests
go test -tags=e2e ./api/e2e/...
```

### Frontend Tests (when implemented)
```bash
# Admin Dashboard
cd frontend/admin-dashboard && npm test

# E2E Testing App
cd frontend/e2e-test-app && npm test

# E2E browser tests
npm run test:e2e
```

## 🔍 Common Test Cases

### Test Case 1: Complete User Journey
1. Create tenant
2. Create user in tenant
3. Login as user
4. Enroll in MFA
5. Login with MFA
6. View profile
7. Update profile

### Test Case 2: RBAC Testing
1. Create tenant
2. Create admin user
3. Create regular user
4. Create role with permissions
5. Assign role to regular user
6. Test permission-based access
7. Verify unauthorized access is blocked

### Test Case 3: Multi-Tenant Isolation
1. Create tenant A
2. Create tenant B
3. Create user in tenant A
4. Create user in tenant B
5. Login as tenant A user
6. Verify cannot access tenant B data
7. Verify tenant isolation

### Test Case 4: Security Testing
1. Test rate limiting (multiple failed logins)
2. Test SQL injection protection
3. Test XSS protection
4. Test CSRF protection
5. Test input validation

## 🛠️ Testing Tools

### API Testing
- **curl**: Command-line HTTP client
- **Postman**: GUI API testing
- **httpie**: User-friendly HTTP client
- **REST Client**: VS Code extension

### Browser Testing
- **Playwright**: E2E browser automation
- **Cypress**: Alternative E2E testing
- **Browser DevTools**: Manual testing

### Load Testing
- **k6**: Modern load testing
- **Apache Bench**: Simple load testing
- **wrk**: HTTP benchmarking

## 📊 Expected Results

### API Response Times
- Health check: < 50ms
- Login: < 200ms
- User CRUD: < 100ms
- List operations: < 300ms

### Test Coverage Goals
- Backend unit tests: >80%
- Backend integration: 100% endpoints
- Frontend components: >70%
- E2E flows: 100% critical paths

## 🐛 Troubleshooting

### API Not Responding
```bash
# Check if server is running
curl http://localhost:8080/health

# Check logs
# Look for error messages in console
```

### Database Connection Issues
```bash
# Verify PostgreSQL is running
psql -h localhost -p 5433 -U iam_user -d iam

# Check connection string in config
cat config/config.yaml
```

### CORS Issues
- Check CORS middleware configuration
- Verify allowed origins
- Check request headers

### Authentication Issues
- Verify tenant ID in headers
- Check token expiration
- Verify token format
- Check Hydra connection (if using OAuth2)

## 📚 Additional Resources

- [Full Testing Strategy](../testing/e2e-testing-strategy.md)
- [Frontend Quick Start](./frontend-quick-start.md)
- [API Documentation](../api/README.md)
- [Implementation Plan](../planning/frontend-implementation-plan.md)

---

**Quick Reference Version**: 1.0  
**Last Updated**: 2024

