-- Rollback: Drop system roles and permissions tables

DROP INDEX IF EXISTS idx_user_system_roles_role_id;
DROP INDEX IF EXISTS idx_user_system_roles_user_id;
DROP INDEX IF EXISTS idx_system_role_permissions_permission_id;
DROP INDEX IF EXISTS idx_system_role_permissions_role_id;
DROP INDEX IF EXISTS idx_system_permissions_resource_action;
DROP INDEX IF EXISTS idx_system_roles_name;

DROP TABLE IF EXISTS user_system_roles;
DROP TABLE IF EXISTS system_role_permissions;
DROP TABLE IF EXISTS system_permissions;
DROP TABLE IF EXISTS system_roles;

