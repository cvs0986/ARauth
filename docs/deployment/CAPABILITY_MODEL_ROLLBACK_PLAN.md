# Capability Model Rollback Plan

This document outlines the rollback procedures for the ARauth Capability Model deployment.

**Last Updated**: 2025-01-27  
**Status**: Ready for Use

---

## 🎯 Rollback Scenarios

### Scenario 1: Immediate Rollback (< 1 hour after deployment)

**Trigger**: Critical bugs, data corruption, or service unavailability

**Steps**:

1. **Stop New Deployments**
   ```bash
   # Pause CI/CD pipeline
   # Notify team
   ```

2. **Revert Backend**
   ```bash
   # Kubernetes
   kubectl rollout undo deployment/iam-api
   
   # Docker Compose
   docker-compose down
   git checkout <previous-version>
   docker-compose up -d
   ```

3. **Revert Frontend**
   ```bash
   # Kubernetes
   kubectl rollout undo deployment/admin-dashboard
   
   # Docker Compose
   cd frontend/admin-dashboard
   git checkout <previous-version>
   npm run build
   # Deploy previous build
   ```

4. **Database Rollback** (if needed)
   ```bash
   # Rollback data migration
   migrate -path migrations -database "$DATABASE_URL" down 1
   
   # Verify rollback
   psql $DATABASE_URL -c "SELECT COUNT(*) FROM tenant_capabilities;"
   ```

5. **Verify Services**
   ```bash
   # Check health endpoints
   curl http://localhost:8080/health
   
   # Test authentication
   curl -X POST http://localhost:8080/api/v1/auth/login ...
   ```

**Time to Rollback**: 5-10 minutes

---

### Scenario 2: Partial Rollback (Feature Flags)

**Trigger**: Specific features causing issues, but core system works

**Steps**:

1. **Disable Capability Middleware**
   ```go
   // In routes.go, comment out capability middleware
   // router.Use(capability.RequireCapability(...))
   ```

2. **Disable Specific Capabilities**
   ```sql
   -- Disable problematic capability at system level
   UPDATE system_capabilities 
   SET enabled = false 
   WHERE capability_key = 'problematic_capability';
   ```

3. **Redeploy**
   ```bash
   # Deploy with feature flags disabled
   # Continue monitoring
   ```

**Time to Rollback**: 15-30 minutes

---

### Scenario 3: Database-Only Rollback

**Trigger**: Data migration issues, but code works

**Steps**:

1. **Rollback Data Migration**
   ```bash
   migrate -path migrations -database "$DATABASE_URL" down 1
   ```

2. **Verify Data**
   ```sql
   -- Check tenant_settings still has original data
   SELECT * FROM tenant_settings LIMIT 5;
   
   -- Verify capability tables are empty or reset
   SELECT COUNT(*) FROM tenant_capabilities;
   SELECT COUNT(*) FROM tenant_feature_enablement;
   SELECT COUNT(*) FROM user_capability_state;
   ```

3. **Keep Code Changes**
   - Backend code can remain deployed
   - Capability service will work with empty tables
   - System will fall back to tenant_settings

**Time to Rollback**: 2-5 minutes

---

## 🔄 Rollback Procedures by Component

### Database Rollback

**Full Rollback**:
```bash
# Rollback all capability migrations
migrate -path migrations -database "$DATABASE_URL" down 5

# This will rollback:
# - 000022_migrate_existing_capabilities
# - 000021_create_user_capability_state
# - 000020_create_tenant_feature_enablement
# - 000019_create_system_capabilities
# - 000018_create_tenant_capabilities
```

**Partial Rollback** (data only):
```bash
# Rollback only data migration
migrate -path migrations -database "$DATABASE_URL" down 1
```

**Verification**:
```sql
-- Check tables are dropped
SELECT table_name FROM information_schema.tables 
WHERE table_name IN (
  'system_capabilities',
  'tenant_capabilities',
  'tenant_feature_enablement',
  'user_capability_state'
);
-- Should return 0 rows
```

---

### Backend Rollback

**Option 1: Git Revert**
```bash
# Revert to previous commit
git revert <commit-hash>
git push origin main

# Redeploy
kubectl rollout restart deployment/iam-api
```

**Option 2: Version Tag**
```bash
# Deploy previous version
kubectl set image deployment/iam-api iam-api=registry/iam-api:v1.0.0
kubectl rollout status deployment/iam-api
```

**Option 3: Feature Flag**
```go
// Disable capability checks in code
const EnableCapabilityModel = false

if EnableCapabilityModel {
    router.Use(capability.RequireCapability(...))
}
```

---

### Frontend Rollback

**Option 1: Git Revert**
```bash
cd frontend/admin-dashboard
git revert <commit-hash>
npm run build
# Deploy previous build
```

**Option 2: Version Tag**
```bash
# Deploy previous version
kubectl set image deployment/admin-dashboard admin-dashboard=registry/admin-dashboard:v1.0.0
```

**Option 3: Hide UI Elements**
```typescript
// Hide capability management UI
const SHOW_CAPABILITY_UI = false;

{SHOW_CAPABILITY_UI && <CapabilityManagement />}
```

---

## 📋 Rollback Decision Matrix

| Issue Type | Severity | Rollback Type | Time |
|------------|----------|---------------|------|
| Data corruption | Critical | Full | 5-10 min |
| Service unavailable | Critical | Full | 5-10 min |
| High error rate | High | Partial | 15-30 min |
| Performance degradation | Medium | Partial | 15-30 min |
| UI issues | Low | Frontend only | 5-10 min |
| Migration issues | High | Database only | 2-5 min |

---

## ✅ Rollback Verification

After rollback, verify:

1. **Services Running**
   ```bash
   curl http://localhost:8080/health
   # Should return 200 OK
   ```

2. **Authentication Works**
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username": "test", "password": "test", "tenant_id": "..."}'
   # Should return tokens
   ```

3. **Database State**
   ```sql
   -- Verify tenant_settings still works
   SELECT * FROM tenant_settings LIMIT 1;
   
   -- Verify capability tables are gone or empty
   SELECT COUNT(*) FROM tenant_capabilities;
   ```

4. **No Data Loss**
   ```sql
   -- Compare user counts
   SELECT COUNT(*) FROM users;
   
   -- Compare tenant counts
   SELECT COUNT(*) FROM tenants;
   ```

---

## 🚨 Emergency Contacts

- **Database Admin**: [TBD]
- **DevOps Lead**: [TBD]
- **Backend Lead**: [TBD]
- **On-Call Engineer**: [TBD]

---

## 📝 Rollback Log Template

```
Rollback Date: [DATE]
Rollback Time: [TIME]
Triggered By: [NAME]
Reason: [REASON]
Rollback Type: [Full/Partial/Database]
Components Rolled Back: [LIST]
Time to Rollback: [DURATION]
Verification: [PASS/FAIL]
Notes: [NOTES]
```

---

## 🔗 Related Documents

- [Deployment Plan](./CAPABILITY_MODEL_DEPLOYMENT_PLAN.md)
- [Migration Script](../migrations/000022_migrate_existing_capabilities.down.sql)
- [Architecture Documentation](../architecture/CAPABILITY_MODEL.md)

---

**Rollback Plan Owner**: [TBD]  
**Last Reviewed**: [TBD]  
**Next Review**: [TBD]

