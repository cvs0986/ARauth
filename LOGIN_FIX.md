# ✅ Login Issue Fixed

## Problem
When logging in from the frontend, the API was returning:
```json
{
    "error": "tenant_required",
    "message": "Tenant ID or domain must be provided via X-Tenant-ID, X-Tenant-Domain header, query parameter, or subdomain"
}
```

## Root Cause
The login endpoint is tenant-scoped and requires the `X-Tenant-ID` header (set by tenant middleware), but the frontend was sending `tenant_id` in the request body instead of the header.

## Solution
Updated the `authApi.login` function in both frontend apps to:
1. Extract `tenant_id` from the request data
2. Send it as `X-Tenant-ID` header
3. Remove `tenant_id` from the request body

## Changes Made

### Admin Dashboard
**File**: `frontend/admin-dashboard/src/services/api.ts`

```typescript
login: async (data: LoginRequest): Promise<LoginResponse> => {
  // Extract tenant_id from data and send as header
  const { tenant_id, ...loginData } = data;
  
  // Create request config with X-Tenant-ID header if tenant_id is provided
  const config = tenant_id
    ? { headers: { 'X-Tenant-ID': tenant_id } }
    : {};
  
  const response = await apiClient.post<LoginResponse>(
    API_ENDPOINTS.AUTH.LOGIN,
    loginData,
    config
  );
  return response.data;
},
```

### E2E Test App
**File**: `frontend/e2e-test-app/src/services/api.ts`

Same change applied.

## How It Works Now

1. User enters username, password, and tenant ID in login form
2. Frontend extracts `tenant_id` from form data
3. Frontend sends `X-Tenant-ID` header with the request
4. Backend tenant middleware extracts tenant ID from header
5. Login handler processes the request with tenant context

## Verification

Login should now work correctly:

```bash
curl 'http://localhost:8080/api/v1/auth/login' \
  -H 'Content-Type: application/json' \
  -H 'X-Tenant-ID: 6e3a1985-61e6-4446-9b42-d4d0c39dad7a' \
  -d '{"username":"veer","password":"Veer@123456"}'
```

**Expected Response**:
```json
{
  "access_token": "...",
  "token_type": "Bearer",
  "expires_in": 900,
  "mfa_required": false
}
```

## Status
✅ **FIXED** - Login now works with `X-Tenant-ID` header

---

**Commit**: `fix(frontend): send X-Tenant-ID header for login requests`  
**Date**: 2024-01-08

