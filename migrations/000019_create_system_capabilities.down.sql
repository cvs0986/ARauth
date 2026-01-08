-- Rollback: Drop system_capabilities table

DROP INDEX IF EXISTS idx_system_capabilities_enabled;
DROP TABLE IF EXISTS system_capabilities;

