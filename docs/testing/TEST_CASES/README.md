# Test Cases Directory

This directory contains comprehensive test cases for all ARauth IAM features, organized by feature area.

## Test Case Files

### Core Features
- **SYSTEM_LEVEL.md** ✅ - System bootstrap, system owner, system roles/permissions/capabilities
- **TENANT_LIFECYCLE.md** - Tenant creation, initialization, management, suspension, deletion
- **AUTHENTICATION.md** - Login, logout, token management, refresh, revocation
- **MFA_TOTP.md** - MFA enrollment, verification, recovery codes, reset
- **RBAC_PERMISSIONS.md** - Roles, permissions, assignments, last-owner protection

### Advanced Features
- **CAPABILITIES.md** - Capability model (System → Tenant → User), evaluation, enrollment
- **OAUTH_OIDC.md** - OAuth2/OIDC flows, scopes, grant types, PKCE
- **FEDERATION_SAML_OIDC.md** - SAML/OIDC federation, identity providers, identity linking
- **SCIM_PROVISIONING.md** - SCIM 2.0 user/group provisioning, bulk operations

### Admin & Operations
- **ADMIN_CONSOLE.md** - Frontend admin console, all UI features
- **AUDIT_LOGS.md** - Audit event generation, querying, immutability
- **WEBHOOKS_EVENTS.md** - Webhook creation, delivery, retry, signing

### Security & Reliability
- **SECURITY_NEGATIVE_TESTS.md** - Attack vectors, abuse cases, security validation
- **FAILURE_RECOVERY.md** - Database failures, Redis failures, recovery behavior
- **PERFORMANCE_LIMITS.md** - Rate limiting, concurrent users, response times

## Test Case Template

Each test case file follows this structure:

1. **Feature Name** - Clear description
2. **Feature Source** - Code files, modules, endpoints
3. **Why This Feature Exists** - Business need
4. **Preconditions** - Required setup
5. **Step-by-Step Test Execution** - Detailed instructions
6. **Expected Functional Behavior** - What should happen
7. **Expected Security Behavior** - Security guarantees
8. **Negative / Abuse Test Cases** - Failure scenarios
9. **Audit Events Expected** - Audit validation
10. **Recovery / Rollback Behavior** - Failure handling
11. **Pass / Fail Criteria** - Success criteria

## Execution Order

Recommended execution order:

1. **SYSTEM_LEVEL.md** - Must run first (creates system owner)
2. **TENANT_LIFECYCLE.md** - Creates test tenants
3. **AUTHENTICATION.md** - Tests login flows
4. **RBAC_PERMISSIONS.md** - Sets up roles/permissions
5. **MFA_TOTP.md** - Tests MFA features
6. **CAPABILITIES.md** - Tests capability model
7. **OAUTH_OIDC.md** - Tests OAuth flows
8. **FEDERATION_SAML_OIDC.md** - Tests federation
9. **SCIM_PROVISIONING.md** - Tests SCIM
10. **ADMIN_CONSOLE.md** - Tests frontend
11. **AUDIT_LOGS.md** - Validates audit
12. **WEBHOOKS_EVENTS.md** - Tests webhooks
13. **SECURITY_NEGATIVE_TESTS.md** - Security validation
14. **FAILURE_RECOVERY.md** - Failure scenarios
15. **PERFORMANCE_LIMITS.md** - Performance testing

## Status

- ✅ **Complete**: Full test cases with detailed steps
- 📝 **Template**: Skeleton structure, needs detailed test steps
- ⏳ **Pending**: Not yet created

## Notes

- All test cases should start from a clean database state
- Test cases build on each other - execute in order
- Document all results, especially failures
- Verify security guarantees at each step
- Validate audit events for all actions

---

**See**: `TEST_EXECUTION_GUIDE.md` for how to execute these test cases.

