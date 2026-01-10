# MFA Authentication Flow & Architecture

## Overview
The MFA system in ARauth ensures robust, two-factor authentication using Time-based One-Time Passwords (TOTP) and Recovery Codes. It adheres to strict security invariants including short-lived sessions, single-use verification, and tenant isolation.

## Core Invariants (Non-Negotiable)
1.  **Strict Plane Separation**: System/Platform admins are distinct from Tenant users.
2.  **No Implicit Trust**: First factor (Password) success does NOT grant access tokens if MFA is required.
3.  **Stateful Flow**: MFA verification requires a valid, bound, short-lived session ID created during primary authentication.
4.  **Single-Use**: MFA sessions are destroyed immediately upon successful verification or maximum failure attempts.
5.  **No Bypass**: MFA cannot be bypassed via token refresh, API calls, or context switching.

## State Machine

### 1. Primary Authentication (Login)
- **Input**: Username, Password, TenantID (optional for System).
- **Process**:
    - Validate credentials.
    - Check `MFAEnabled` on User and `MFARequired` on Tenant/System.
- **Output**:
    - If `MFARequired == false`: Issue Access/Refresh Tokens.
    - If `MFARequired == true`: 
        - Create **MFA Session** (TTL 5 mins).
        - Return `mfa_required: true`, `mfa_session_id: <uuid>`.
        - **NO TOKENS ISSUED**.

### 2. MFA Verification
- **Input**: `mfa_session_id`, `totp_code` OR `recovery_code`.
- **Process**:
    - Validate Session (exists, not expired, matches user/tenant).
    - Validate TOTP/Recovery Code.
- **Output**:
    - **Success**: 
        - Destroy MFA Session.
        - Issue Access/Refresh Tokens.
        - Log `mfa.challenge.verified`.
    - **Failure**:
        - Increment attempt counter.
        - If Max Attempts > 5: Destroy Session.
        - Return 401 Unauthorized.

## Session Lifecycle
- **Creation**: At `AuthHandler.Login` upon password success.
- **Storage**: Redis (or Memory in dev), Key: `mfa:session:<id>`.
- **TTL**: 5 minutes hard limit.
- **Destruction**: 
    - Explicit delete on Verification Success.
    - Explicit delete on Verification Limit Reached.
    - Auto-expire by TTL.

## Auditing
All significant events are audit logged:
- `mfa.enrolled`: User enables MFA.
- `mfa.challenge.created`: Session started.
- `mfa.challenge.verified`: Session completed successfully.
- `mfa.challenge.failed`: Verification failed (bad code).
- `mfa.disabled`: User disabled MFA.
- `mfa.reset`: Admin reset MFA for user.

## Data Model
- **User**: Stores `MFAEnabled` (bool), `MFASecretEncrypted` (string), `RecoveryCodes` (hashed).
- **Session**: `SessionID`, `UserID`, `TenantID`, `ExpiresAt`, `Attempts`.

## Future Improvements
- WebAuthn/Passkey support.
- SMS/Email OTP (low priority, insecure).
- Push Notifications.

## Refresh Token Preservation
To prevent bypass of MFA via refresh tokens:
1.  **MFAVerified Flag**: The refresh token table (`refresh_tokens`) includes an `mfa_verified` (boolean) column.
    - Set to `true` ONLY when a refresh token is issued after a successful MFA verification.
    - Set to `false` for standard password-only logins (if MFA is not required).
2.  **Rotation**: When a refresh token is rotated (used to get a new pair), the new refresh token **inherits** the `mfa_verified` status of the parent token.
3.  **Enforcement**: The Refresh Service checks: `IF User.MFAEnabled == true AND RefreshToken.MFAVerified == false THEN DENY`.

## JWT Claims (AMR)
The system implements the `amr` (Authentication Methods References) claim in Access Tokens:
- `["pwd"]`: Standard password authentication.
- `["pwd", "mfa"]`: Multi-factor authentication verified.
- The `amr` claim is set based on the `mfa_verified` status of the refresh token during refresh flows.
