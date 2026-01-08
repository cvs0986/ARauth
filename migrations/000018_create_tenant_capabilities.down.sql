-- Rollback: Drop tenant_capabilities table

DROP INDEX IF EXISTS idx_tenant_capabilities_enabled;
DROP INDEX IF EXISTS idx_tenant_capabilities_key;
DROP INDEX IF EXISTS idx_tenant_capabilities_tenant_id;
DROP TABLE IF EXISTS tenant_capabilities;

