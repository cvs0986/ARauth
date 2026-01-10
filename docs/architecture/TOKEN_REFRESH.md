# Token Refresh Architecture

## Overview
The Token Refresh flow allows clients to obtain new Access Tokens without re-prompting the user for credentials. This process must maintain the same security guarantees as the original authentication event, particularly regarding Multi-Factor Authentication (MFA).

## Security Invariants
1.  **MFA Preservation**: A refresh token MUST capture whether the original session was MFA-verified.
2.  **MFA Enforcement**: If a user enables MFA *after* obtaining a refresh token, that old non-MFA refresh token MUST NOT optionally be usable to obtain new tokens without MFA verification.
    - **Policy**: `IF User.MFAEnabled AND !RefreshToken.MFAVerified THEN REJECT`.
3.  **Rotation**: Refresh tokens are single-use (rotated on use). The new token inherits security properties (like MFA verification status) from the used token.

## Data Model
The `refresh_tokens` table includes:
- `user_id`: Link to user.
- `token_hash`: Hashed token value.
- `mfa_verified`: Boolean flag. `true` if issued via MFA flow, `false` otherwise.
- `expires_at`: Expiration timestamp.

## Flows

### 1. Initial Login (MFA Required)
1.  User enters password -> Validated.
2.  System requires MFA -> Session created.
3.  User enters TOTP -> Validated.
4.  System issues Refresh Token with `mfa_verified = true`.

### 2. Token Refresh
1.  Client sends `refresh_token`.
2.  System looks up token, verifies expiry and revocation.
3.  **MFA Check**:
    - Fetch User.
    - `IF user.MFAEnabled == true AND refresh_token.mfa_verified == false`:
        - **DENY** (Return 401 Unauthorized, "MFA required").
        - Client must redirect user to full login with MFA.
4.  **Token Rotation**:
    - Revoke old token.
    - Issue NEW refresh token.
    - Set `new_token.mfa_verified = old_token.mfa_verified`.
5.  **Claim Generation**:
    - If `mfa_verified == true`: Set `amr` claim to `["pwd", "mfa"]`.
    - Else: Set `amr` claim to `["pwd"]`.
