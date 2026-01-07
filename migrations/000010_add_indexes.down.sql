-- Drop indexes

DROP INDEX IF EXISTS idx_audit_logs_tenant_action;
DROP INDEX IF EXISTS idx_audit_logs_created_at;
DROP INDEX IF EXISTS idx_audit_logs_action;
DROP INDEX IF EXISTS idx_audit_logs_actor_id;
DROP INDEX IF EXISTS idx_audit_logs_tenant_id;

DROP INDEX IF EXISTS idx_mfa_recovery_codes_user_used;
DROP INDEX IF EXISTS idx_mfa_recovery_codes_user_id;

DROP INDEX IF EXISTS idx_role_permissions_unique;
DROP INDEX IF EXISTS idx_role_permissions_permission_id;
DROP INDEX IF EXISTS idx_role_permissions_role_id;

DROP INDEX IF EXISTS idx_user_roles_unique;
DROP INDEX IF EXISTS idx_user_roles_role_id;
DROP INDEX IF EXISTS idx_user_roles_user_id;

DROP INDEX IF EXISTS idx_permissions_name;
DROP INDEX IF EXISTS idx_permissions_resource_action;

DROP INDEX IF EXISTS idx_roles_is_system;
DROP INDEX IF EXISTS idx_roles_name;
DROP INDEX IF EXISTS idx_roles_tenant_id;

DROP INDEX IF EXISTS idx_tenants_status;
DROP INDEX IF EXISTS idx_tenants_domain;

DROP INDEX IF EXISTS idx_credentials_last_attempt_at;
DROP INDEX IF EXISTS idx_credentials_user_id;

DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_users_username_filtered;
DROP INDEX IF EXISTS idx_users_email_filtered;
DROP INDEX IF EXISTS idx_users_tenant_id;

