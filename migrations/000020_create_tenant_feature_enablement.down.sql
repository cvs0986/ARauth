-- Rollback: Drop tenant_feature_enablement table

DROP INDEX IF EXISTS idx_tenant_feature_enablement_enabled;
DROP INDEX IF EXISTS idx_tenant_feature_enablement_key;
DROP INDEX IF EXISTS idx_tenant_feature_enablement_tenant_id;
DROP TABLE IF EXISTS tenant_feature_enablement;

