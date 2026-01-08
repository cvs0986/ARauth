-- Rollback: Drop system settings and tenant configurations tables

DROP INDEX IF EXISTS idx_tenant_configurations_key;
DROP INDEX IF EXISTS idx_tenant_configurations_tenant_id;

DROP TABLE IF EXISTS tenant_configurations;
DROP TABLE IF EXISTS system_settings;

