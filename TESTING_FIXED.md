# ✅ User Creation Issue Fixed

## Problem
When creating a user with `X-Tenant-ID` header, the API was returning:
```json
{
    "error": "invalid_request",
    "message": "Request validation failed",
    "details": [
        {
            "field": "TenantID",
            "message": "TenantID is required"
        }
    ]
}
```

## Root Cause
The `CreateUserRequest` struct had `TenantID` with `binding:"required"` tag, which caused Gin's validation to fail **before** the handler could set the tenant ID from the context (extracted from `X-Tenant-ID` header).

## Solution
Removed the `binding:"required"` tag from `TenantID` in `CreateUserRequest` because:
1. Tenant ID is set from context by tenant middleware, not from request body
2. This prevents tenant spoofing (users can't specify a different tenant ID)
3. The handler sets `req.TenantID = tenantID` after binding succeeds

## Fix Applied
**File**: `identity/user/service.go`
```go
// Before
TenantID  uuid.UUID `json:"tenant_id" binding:"required"`

// After
TenantID  uuid.UUID `json:"tenant_id"` // Set from context, not from request body
```

## Verification
✅ User creation now works correctly with `X-Tenant-ID` header:

```bash
curl --location 'http://localhost:8080/api/v1/users' \
  --header 'Content-Type: application/json' \
  --header 'X-Tenant-ID: 6e3a1985-61e6-4446-9b42-d4d0c39dad7a' \
  --data-raw '{
    "username": "veer",
    "email": "veer@test.local",
    "password": "Veer@123456",
    "first_name": "Veer",
    "last_name": "Singh"
  }'
```

**Response**:
```json
{
  "id": "79f3d597-5631-4f03-8b85-aefb41d4f363",
  "tenant_id": "6e3a1985-61e6-4446-9b42-d4d0c39dad7a",
  "username": "veer3",
  "email": "veer3@test.local",
  "first_name": "Veer",
  "last_name": "Singh",
  "status": "active",
  "mfa_enabled": false,
  "created_at": "2026-01-08T01:08:36.928636169+05:30",
  "updated_at": "2026-01-08T01:08:36.928636169+05:30"
}
```

## Status
✅ **FIXED** - User creation works correctly with `X-Tenant-ID` header

---

**Commit**: `fix(api): remove required binding from TenantID in CreateUserRequest`  
**Date**: 2024-01-08

