# MFA Enforcement Verification Plan

## Objective
Verify strict enforcement of MFA across all authentication paths, ensuring no bypass is possible.

## 1. Login Path Verification
- [ ] **MFA Required Detection**: Login with enabled user returns `mfa_required=true` and NO tokens.
- [ ] **MFA Session Creation**: Ensure session ID is returned.
- [ ] **Session Binding**: Verify session is bound to correct User/Tenant.

## 2. MFA Verification Logic
- [ ] **Challenge Success**: Valid TOTP + Session -> Access/Refresh Tokens.
- [ ] **Challenge Failure**: Invalid TOTP -> 401 Unauthorized.
- [ ] **Session Replay**: Attempt to reuse same session ID -> 401 Unauthorized (Session not found/expired).
- [ ] **Cross-Tenant Attack**: Attempt verify with correct Session ID but wrong Tenant context -> 401/403.

## 3. Token Issuance & Refresh
- [ ] **Refresh Token Bypass**: Attempt to use `POST /token/refresh` with a refresh token obtained *before* MFA was enabled (if applicable) or verify new refresh tokens are only issued after MFA.
- [ ] **No Pre-MFA Tokens**: Confirm Login response strictly contains NO tokens when MFA is pending.

## 4. API Access Control
- [ ] **JWT Enforcement**: Verify `IsMFA` claim (if implemented) or reliance on Token/Blacklist service.
- [ ] **Middleware Check**: Ensure `JWTAuthMiddleware` correctly parses and validates tokens. (Note: MFA status is usually baked into the act of issuing the token itself; if you have the token, you passed MFA. Verification is focused on *getting* the token).

## 5. Negative Scenarios
- [ ] **Bypass Attempt**: Login -> Ignore MFA response -> Call API (should fail as no token).
- [ ] **Session Injection**: Try to pass arbitrary session ID to Verify endpoint.

## Execution Strategy
- Create `api/handlers/mfa_enforcement_test.go`.
- Use `httptest` with `gin` router to simulate full E2E flows within the test suite.
- Mock `MFAService`, `TokenService`, and `AuditService` carefully to reflect real behavior, or use real implementations with mocked repositories where possible for "integration-like" fidelity.
