-- Migration: Remove tenant_id from permissions table (rollback)

-- Drop indexes
DROP INDEX IF EXISTS idx_permissions_tenant_id;
DROP INDEX IF EXISTS idx_permissions_tenant_resource_action;
DROP INDEX IF EXISTS idx_permissions_global_resource_action;
DROP INDEX IF EXISTS idx_permissions_deleted_at;

-- Remove columns
ALTER TABLE permissions 
DROP COLUMN IF EXISTS tenant_id,
DROP COLUMN IF EXISTS updated_at,
DROP COLUMN IF EXISTS deleted_at;

-- Restore original unique constraint
CREATE UNIQUE INDEX IF NOT EXISTS permissions_resource_action_key ON permissions(resource, action);

