# Changelog

All notable changes to this project will be documented in this file.

## [1.0.0] - 2026-01-10 (GA Release)

### Security (Critical)
- **MFA Refresh Token Bypass Fixed**: Implemented `mfa_verified` column in database and enforced checks in `RefreshService` to prevent non-MFA refresh tokens from bypassing MFA requirements (#55).
- **Token Blacklist**: Implemented Redis-based token revocation blacklist to support immediate token invalidation (#56).
- **Authorization Enforcement**: Fixed `AuthorizationMiddleware` impersonation vulnerability.

### Added
- **MFA Enforcement**: Strict session binding and validation for Multi-Factor Authentication.
- **Audit Logging**: Comprehensive audit logs for login, token issuance, and MFA events.
- **MFA Capability**: Capability-based MFA enforcement for Tenant and System users.
- **Documentation**: Added `MFA_FLOW.md`, `TOKEN_REFRESH.md`, and `TOKEN_REVOCATION.md` architecture docs.
- **Tests**: Extensive integration test suite for MFA enforcement and refresh flows.

### Changed
- **JWT Claims**: Added `amr` (Authentication Methods References) claim to Access Tokens.
- **Refresh Token Rotation**: Preserves `mfa_verified` status across token rotation.

### Fixed
- Missing error logging in permission service.
