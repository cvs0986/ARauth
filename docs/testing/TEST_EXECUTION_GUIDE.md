# Test Execution Guide

## Overview

This guide explains how to execute the test cases documented in the `TEST_CASES/` directory. It provides execution patterns, best practices, and troubleshooting tips.

## Execution Philosophy

### Principles
1. **Sequential Execution**: Test cases build on each other - execute in order
2. **Clean State**: Start each test from a known clean state
3. **Documentation**: Document all results, especially failures
4. **Reproducibility**: Every test should be reproducible
5. **Security Focus**: Verify security guarantees at each step

### Execution Modes

#### Mode 1: Full Test Suite
Execute all test cases in sequence for complete validation.

#### Mode 2: Feature-Specific Testing
Execute test cases for a specific feature area.

#### Mode 3: Regression Testing
Execute test cases to verify no regressions after changes.

#### Mode 4: Security Testing
Focus on security-related test cases.

## Test Execution Workflow

### Phase 1: Preparation
1. **Environment Setup**: Follow `TEST_ENVIRONMENT_SETUP.md`
2. **Clean Database**: Ensure database is empty
3. **Start Services**: API, PostgreSQL, Redis
4. **Verify Health**: All health checks pass

### Phase 2: System Bootstrap
1. **Execute**: `TEST_CASES/SYSTEM_LEVEL.md`
2. **Verify**: System owner created
3. **Verify**: System roles and permissions exist
4. **Document**: Bootstrap results

### Phase 3: Core Features
1. **Tenant Lifecycle**: `TEST_CASES/TENANT_LIFECYCLE.md`
2. **Authentication**: `TEST_CASES/AUTHENTICATION.md`
3. **RBAC**: `TEST_CASES/RBAC_PERMISSIONS.md`

### Phase 4: Advanced Features
1. **MFA**: `TEST_CASES/MFA_TOTP.md`
2. **Capabilities**: `TEST_CASES/CAPABILITIES.md`
3. **Federation**: `TEST_CASES/FEDERATION_SAML_OIDC.md`
4. **SCIM**: `TEST_CASES/SCIM_PROVISIONING.md`

### Phase 5: Security & Compliance
1. **Security Tests**: `TEST_CASES/SECURITY_NEGATIVE_TESTS.md`
2. **Audit Logs**: `TEST_CASES/AUDIT_LOGS.md`
3. **Webhooks**: `TEST_CASES/WEBHOOKS_EVENTS.md`

### Phase 6: Failure & Recovery
1. **Failure Scenarios**: `TEST_CASES/FAILURE_RECOVERY.md`
2. **Performance**: `TEST_CASES/PERFORMANCE_LIMITS.md`

### Phase 7: Frontend
1. **Admin Console**: `TEST_CASES/ADMIN_CONSOLE.md`

### Phase 8: Validation
1. **Exit Criteria**: `TEST_EXIT_CRITERIA.md`
2. **Documentation**: Document all results

## Test Case Execution Pattern

### For Each Test Case

#### Step 1: Read Test Case
- Understand the feature being tested
- Review preconditions
- Understand expected behavior

#### Step 2: Prepare Environment
- Ensure preconditions are met
- Set up required data
- Configure as needed

#### Step 3: Execute Steps
- Follow step-by-step instructions
- Execute API calls
- Verify intermediate states

#### Step 4: Verify Results
- Check functional behavior
- Verify security guarantees
- Validate audit events
- Check negative cases

#### Step 5: Document Results
- Record pass/fail status
- Document any issues
- Note deviations from expected behavior

## API Testing Tools

### Using curl

#### Basic Pattern
```bash
# Login
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "AdminPassword123!",
    "tenant_id": null
  }' | jq -r '.access_token')

# Use token
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json"
```

#### Helper Functions
```bash
# Save to .bashrc or .zshrc
iam_login() {
  local username=$1
  local password=$2
  local tenant_id=$3
  
  curl -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d "{
      \"username\": \"$username\",
      \"password\": \"$password\",
      \"tenant_id\": \"$tenant_id\"
    }" | jq -r '.access_token'
}

iam_api() {
  local method=$1
  local endpoint=$2
  local token=$3
  local data=$4
  
  if [ -n "$data" ]; then
    curl -X $method "http://localhost:8080$endpoint" \
      -H "Authorization: Bearer $token" \
      -H "Content-Type: application/json" \
      -d "$data"
  else
    curl -X $method "http://localhost:8080$endpoint" \
      -H "Authorization: Bearer $token" \
      -H "Content-Type: application/json"
  fi
}
```

### Using Postman/Insomnia

1. **Import Collection**: Create collection from test cases
2. **Set Variables**: Base URL, tokens, tenant IDs
3. **Execute Requests**: Follow test case steps
4. **Verify Responses**: Check status codes and data

### Using Scripts

Create test scripts for repetitive operations:

```bash
#!/bin/bash
# test_tenant_creation.sh

BASE_URL="http://localhost:8080"
ADMIN_TOKEN=$(iam_login "admin" "AdminPassword123!" null)

# Create tenant
TENANT_RESPONSE=$(iam_api POST "$BASE_URL/api/v1/tenants" "$ADMIN_TOKEN" '{
  "name": "Test Tenant",
  "domain": "test-tenant"
}')

TENANT_ID=$(echo $TENANT_RESPONSE | jq -r '.id')
echo "Created tenant: $TENANT_ID"
```

## Database Verification

### Query Patterns

```bash
# Check tenant exists
psql -h localhost -U iam_user -d iam -c \
  "SELECT id, name, domain FROM tenants WHERE domain = 'test-tenant';"

# Check user exists
psql -h localhost -U iam_user -d iam -c \
  "SELECT id, username, email FROM users WHERE email = 'user@example.com';"

# Check audit events
psql -h localhost -U iam_user -d iam -c \
  "SELECT event_type, actor_username, result FROM audit_events ORDER BY created_at DESC LIMIT 10;"

# Check capabilities
psql -h localhost -U iam_user -d iam -c \
  "SELECT capability_key, enabled FROM system_capabilities;"
```

## Audit Log Verification

### Pattern
```bash
# Query audit events
curl -X GET "http://localhost:8080/api/v1/audit/events?event_type=user.created&limit=10" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Verify specific event
EVENT_ID="..."
curl -X GET "http://localhost:8080/api/v1/audit/events/$EVENT_ID" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

### Verification Checklist
- [ ] Event type matches action
- [ ] Actor information is correct
- [ ] Target information is correct
- [ ] Timestamp is recent
- [ ] Result is "success" for successful operations
- [ ] Metadata contains expected information

## Security Testing

### Authentication Testing
- Test invalid credentials
- Test expired tokens
- Test missing tokens
- Test malformed tokens

### Authorization Testing
- Test unauthorized access
- Test permission boundaries
- Test tenant isolation
- Test privilege escalation attempts

### Input Validation
- Test SQL injection
- Test XSS attempts
- Test path traversal
- Test command injection

## Negative Testing

### Common Patterns

#### Invalid Input
```bash
# Missing required field
curl -X POST http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com"}'
# Expected: 400 Bad Request
```

#### Unauthorized Access
```bash
# Missing token
curl -X GET http://localhost:8080/api/v1/users
# Expected: 401 Unauthorized

# Invalid token
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer invalid-token"
# Expected: 401 Unauthorized
```

#### Rate Limiting
```bash
# Rapid requests
for i in {1..10}; do
  curl -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username": "admin", "password": "wrong"}'
done
# Expected: 429 Too Many Requests after threshold
```

## Result Documentation

### Test Result Template

```markdown
## Test Case: [Name]

**Date**: [Date]
**Tester**: [Name]
**Environment**: [Description]

### Execution Steps
1. [Step 1] - ✅ Pass / ❌ Fail
2. [Step 2] - ✅ Pass / ❌ Fail
...

### Results
- **Functional**: ✅ Pass / ❌ Fail
- **Security**: ✅ Pass / ❌ Fail
- **Audit**: ✅ Pass / ❌ Fail

### Issues Found
- [Issue description]

### Notes
- [Additional notes]
```

## Troubleshooting

### Common Issues

#### API Returns 500 Error
1. Check server logs
2. Verify database connection
3. Check request payload format
4. Verify authentication token

#### Database Errors
1. Check database connection
2. Verify migrations applied
3. Check table existence
4. Verify permissions

#### Authentication Failures
1. Verify credentials
2. Check token expiration
3. Verify tenant context
4. Check user status

#### Missing Audit Events
1. Check audit service is running
2. Verify event logging code
3. Check database for events
4. Verify permissions for audit queries

## Best Practices

1. **Start Clean**: Always start from a clean database state
2. **Document Everything**: Record all results and issues
3. **Verify Security**: Check security guarantees at each step
4. **Test Negatives**: Always test failure cases
5. **Validate Audit**: Verify audit events for all actions
6. **Isolate Tests**: Don't let tests interfere with each other
7. **Use Variables**: Store tokens, IDs in variables for reuse
8. **Check Responses**: Always verify response codes and data
9. **Clean Up**: Clean up test data when done
10. **Report Issues**: Document all issues found

## Automation

### Converting to Automated Tests

Test cases can be converted to automated tests:

```go
func TestUserCreation(t *testing.T) {
    // Setup
    token := loginAsAdmin(t)
    
    // Execute
    user := createUser(t, token, CreateUserRequest{
        Username: "testuser",
        Email: "test@example.com",
        Password: "TestPassword123!",
    })
    
    // Verify
    assert.NotNil(t, user)
    assert.Equal(t, "testuser", user.Username)
    
    // Verify audit
    events := queryAuditEvents(t, token, "user.created")
    assert.Len(t, events, 1)
}
```

## Next Steps

1. **Read** test case files in `TEST_CASES/`
2. **Set up** environment per `TEST_ENVIRONMENT_SETUP.md`
3. **Execute** test cases following this guide
4. **Document** all results
5. **Validate** exit criteria in `TEST_EXIT_CRITERIA.md`

---

**Remember**: Testing is iterative. If a test fails, investigate, fix, and re-test.

