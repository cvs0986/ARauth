-- Migration: Create system settings and tenant configurations tables
-- System settings are global configurations managed by system admins
-- Tenant configurations are tenant-specific settings that can be managed by system admins

-- System settings table (global settings)
CREATE TABLE system_settings (
    key VARCHAR(255) PRIMARY KEY,
    value JSONB NOT NULL,
    description TEXT,
    updated_by UUID REFERENCES users(id),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Tenant configurations table (tenant-specific settings)
CREATE TABLE tenant_configurations (
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL,
    value JSONB NOT NULL,
    configured_by UUID REFERENCES users(id),
    configured_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, key)
);

-- Create indexes
CREATE INDEX idx_tenant_configurations_tenant_id ON tenant_configurations(tenant_id);
CREATE INDEX idx_tenant_configurations_key ON tenant_configurations(key);

-- Insert default system settings
INSERT INTO system_settings (key, value, description) VALUES
    ('password_policy', '{"min_length": 12, "require_uppercase": true, "require_lowercase": true, "require_numbers": true, "require_special": true}', 'Global password policy for all tenants'),
    ('mfa_policy', '{"enforced_for_system_users": true, "enforced_for_tenant_admins": false, "enforced_for_all_users": false}', 'MFA enforcement policy'),
    ('session_policy', '{"max_session_duration": 3600, "idle_timeout": 900, "system_user_session_duration": 300}', 'Session management policy'),
    ('rate_limit_policy', '{"system_api_rpm": 1000, "tenant_api_rpm": 100, "login_attempts": 5}', 'Rate limiting policy'),
    ('token_policy', '{"access_token_ttl": 900, "refresh_token_ttl": 2592000, "system_access_token_ttl": 300}', 'Token lifetime policy');

