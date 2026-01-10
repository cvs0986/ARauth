# ARauth IAM Control Plane - Data Gaps Inventory

**Document Type**: Backend API Requirements  
**Status**: Active Tracking  
**Last Updated**: 2026-01-11

---

## Purpose

This document tracks UI features that require backend APIs that don't currently exist. Per Guardrail #4, we MUST NOT fake or approximate data in the frontend.

---

## 🔴 Critical Gaps (Block Core Features)

### 1. Cross-Tenant User Aggregation
**UI Feature**: System Dashboard - "Total Users Across All Tenants"  
**Current State**: ❌ No backend endpoint  
**Required API**: `GET /system/metrics/users/aggregate`  
**Response**:
```json
{
  "total_users": 1523,
  "active_users": 1401,
  "mfa_enrolled": 892,
  "by_tenant": [
    {"tenant_id": "uuid", "user_count": 45},
    ...
  ]
}
```
**GitHub Issue**: #TBD  
**Priority**: HIGH  
**Workaround**: Show "Coming Soon" placeholder

---

### 2. MFA Adoption Rate Calculation
**UI Feature**: System Dashboard - "MFA Adoption Rate"  
**Current State**: ❌ No backend endpoint  
**Required API**: `GET /system/metrics/mfa/adoption`  
**Response**:
```json
{
  "total_users": 1523,
  "mfa_enrolled": 892,
  "adoption_rate": 58.6,
  "by_tenant": [
    {"tenant_id": "uuid", "adoption_rate": 75.2},
    ...
  ]
}
```
**GitHub Issue**: #TBD  
**Priority**: HIGH  
**Workaround**: Show "Coming Soon" placeholder

---

### 3. Tenant Health Scoring
**UI Feature**: System Dashboard - "Tenant Health" table  
**Current State**: ❌ No backend logic  
**Required API**: `GET /system/tenants/health`  
**Response**:
```json
{
  "tenants": [
    {
      "tenant_id": "uuid",
      "name": "Acme Corp",
      "health_score": 85,
      "factors": {
        "mfa_enabled": true,
        "password_policy_strength": "strong",
        "active_users_ratio": 0.92,
        "security_incidents": 0
      }
    },
    ...
  ]
}
```
**GitHub Issue**: #TBD  
**Priority**: MEDIUM  
**Workaround**: Show basic tenant list without health scores

---

## 🟡 Important Gaps (Enhance Features)

### 4. Security Posture Metrics
**UI Feature**: System Dashboard - "Security Posture" section  
**Current State**: ❌ No backend endpoint  
**Required API**: `GET /system/security/posture`  
**Response**:
```json
{
  "overall_score": 78,
  "risk_indicators": {
    "high_risk_tenants": 3,
    "compliance_gaps": 5,
    "unpatched_vulnerabilities": 0
  },
  "recommendations": [
    "Enable MFA for 3 tenants",
    "Strengthen password policy for 2 tenants"
  ]
}
```
**GitHub Issue**: #TBD  
**Priority**: MEDIUM  
**Workaround**: Show basic security metrics (MFA adoption only)

---

### 5. Active Sessions Aggregation
**UI Feature**: Tenant Dashboard - "Active Sessions" count  
**Current State**: ⚠️ Partial (can query refresh_tokens table)  
**Required API**: `GET /api/v1/sessions/active`  
**Response**:
```json
{
  "total_sessions": 45,
  "sessions": [
    {
      "user_id": "uuid",
      "ip_address": "192.168.1.1",
      "user_agent": "Chrome/120",
      "started_at": "2026-01-11T00:00:00Z",
      "last_activity": "2026-01-11T00:30:00Z"
    },
    ...
  ]
}
```
**GitHub Issue**: #TBD  
**Priority**: MEDIUM  
**Workaround**: Query refresh_tokens table (not real-time)

---

### 6. User Activity Timeline
**UI Feature**: Tenant Dashboard - "User Activity (7 days)" chart  
**Current State**: ❌ No backend endpoint  
**Required API**: `GET /api/v1/metrics/user-activity?days=7`  
**Response**:
```json
{
  "data_points": [
    {"date": "2026-01-05", "logins": 120, "api_calls": 4500},
    {"date": "2026-01-06", "logins": 135, "api_calls": 5100},
    ...
  ]
}
```
**GitHub Issue**: #TBD  
**Priority**: LOW  
**Workaround**: Show "Coming Soon" chart placeholder

---

## 🟢 Nice-to-Have Gaps (Future Enhancements)

### 7. Audit Log Export
**UI Feature**: Audit Logs - "Export" button  
**Current State**: ❌ No backend endpoint  
**Required API**: `POST /api/v1/audit/export`  
**Request**:
```json
{
  "format": "csv",
  "filters": {
    "start_date": "2026-01-01",
    "end_date": "2026-01-11",
    "event_types": ["login", "mfa.verified"]
  }
}
```
**Response**: CSV file download  
**GitHub Issue**: #TBD  
**Priority**: LOW  
**Workaround**: Users can view logs in UI only

---

### 8. Webhook Delivery Logs
**UI Feature**: Webhooks - "View Logs" button  
**Current State**: ❌ No backend storage  
**Required**: Webhook delivery tracking table  
**Required API**: `GET /api/v1/webhooks/:id/logs`  
**Response**:
```json
{
  "deliveries": [
    {
      "id": "uuid",
      "event": "user.created",
      "status": "success",
      "response_code": 200,
      "delivered_at": "2026-01-11T00:00:00Z"
    },
    ...
  ]
}
```
**GitHub Issue**: #TBD  
**Priority**: LOW  
**Workaround**: No delivery tracking

---

## 📊 Gap Summary

| Priority | Count | Status |
|----------|-------|--------|
| HIGH     | 2     | Blocking core features |
| MEDIUM   | 3     | Degraded UX |
| LOW      | 3     | Future enhancements |
| **Total** | **8** | **Tracked** |

---

## 🔄 Resolution Process

For each gap:
1. ✅ Document in this file
2. ✅ Create GitHub issue (backend team)
3. ✅ Add "Coming Soon" placeholder in UI
4. ✅ Link UI component to GitHub issue
5. ⏳ Wait for backend implementation
6. ✅ Remove placeholder when API is ready
7. ✅ Mark gap as resolved

---

## ✅ Resolved Gaps

### ~~1. Token Revocation~~
**Resolved**: 2026-01-10  
**API**: `POST /api/v1/auth/revoke`  
**GitHub Issue**: #56

### ~~2. MFA Refresh Enforcement~~
**Resolved**: 2026-01-10  
**API**: `mfa_verified` column + enforcement logic  
**GitHub Issue**: #55

---

**This document is updated as gaps are discovered and resolved.**
