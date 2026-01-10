-- Check if migration 000023 has been applied
-- Run this query in your database to check migration status

-- 1. Check current migration version
SELECT version, dirty FROM schema_migrations;

-- 2. Check if permissions table has tenant_id column
SELECT 
    column_name, 
    data_type, 
    is_nullable
FROM information_schema.columns 
WHERE table_name = 'permissions' 
  AND column_name IN ('tenant_id', 'updated_at', 'deleted_at')
ORDER BY column_name;

-- 3. Check if indexes exist
SELECT 
    indexname, 
    indexdef
FROM pg_indexes 
WHERE tablename = 'permissions' 
  AND indexname IN (
    'idx_permissions_tenant_id',
    'idx_permissions_tenant_resource_action',
    'idx_permissions_global_resource_action',
    'idx_permissions_deleted_at'
  )
ORDER BY indexname;

-- Expected results if migration 000023 is applied:
-- 1. schema_migrations.version should be 23 (or higher)
-- 2. permissions table should have: tenant_id (uuid, nullable), updated_at (timestamp), deleted_at (timestamp)
-- 3. All 4 indexes should exist

