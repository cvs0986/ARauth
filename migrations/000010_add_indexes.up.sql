-- Add indexes for performance optimization

-- Users table indexes
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(tenant_id, email) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_username ON users(tenant_id, username) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at DESC);

-- Credentials table indexes
CREATE INDEX IF NOT EXISTS idx_credentials_user_id ON credentials(user_id);
CREATE INDEX IF NOT EXISTS idx_credentials_last_attempt_at ON credentials(last_attempt_at);

-- Tenants table indexes
CREATE INDEX IF NOT EXISTS idx_tenants_domain ON tenants(domain) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status) WHERE deleted_at IS NULL;

-- Roles table indexes
CREATE INDEX IF NOT EXISTS idx_roles_tenant_id ON roles(tenant_id);
CREATE INDEX IF NOT EXISTS idx_roles_name ON roles(tenant_id, name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_roles_is_system ON roles(is_system) WHERE deleted_at IS NULL;

-- Permissions table indexes
CREATE INDEX IF NOT EXISTS idx_permissions_tenant_id ON permissions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_permissions_resource_action ON permissions(resource, action) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_permissions_name ON permissions(tenant_id, name) WHERE deleted_at IS NULL;

-- User roles table indexes
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_roles_unique ON user_roles(user_id, role_id);

-- Role permissions table indexes
CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id ON role_permissions(permission_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_role_permissions_unique ON role_permissions(role_id, permission_id);

-- MFA recovery codes table indexes
CREATE INDEX IF NOT EXISTS idx_mfa_recovery_codes_user_id ON mfa_recovery_codes(user_id);
CREATE INDEX IF NOT EXISTS idx_mfa_recovery_codes_used ON mfa_recovery_codes(user_id, used);

-- Audit logs table indexes
CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_id ON audit_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_event_type ON audit_logs(event_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_event ON audit_logs(tenant_id, event_type, created_at DESC);

