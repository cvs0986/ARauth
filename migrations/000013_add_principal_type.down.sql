-- Rollback: Remove principal_type and restore tenant_id NOT NULL constraint

DROP INDEX IF EXISTS idx_users_system_users;
DROP INDEX IF EXISTS idx_users_principal_type;

ALTER TABLE users
DROP CONSTRAINT IF EXISTS chk_principal_type_tenant_id;

ALTER TABLE users
ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE users
DROP COLUMN IF EXISTS principal_type;

