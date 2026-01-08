-- Migration: Create system_capabilities table
-- This table stores global system-level capabilities (System layer)
-- Implements the "Global System Level" from the capability model

CREATE TABLE system_capabilities (
    capability_key VARCHAR(255) PRIMARY KEY,
    enabled BOOLEAN NOT NULL DEFAULT false,
    default_value JSONB, -- Default configuration for this capability
    description TEXT,
    updated_by UUID REFERENCES users(id),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index for enabled capabilities
CREATE INDEX idx_system_capabilities_enabled ON system_capabilities(enabled) WHERE enabled = true;

-- Insert default system capabilities
INSERT INTO system_capabilities (capability_key, enabled, default_value, description) VALUES
    ('mfa', true, '{}', 'Multi-factor authentication support'),
    ('totp', true, '{}', 'Time-based OTP support'),
    ('saml', false, '{}', 'SAML federation support'),
    ('oidc', true, '{}', 'OIDC protocol support'),
    ('oauth2', true, '{}', 'OAuth2 protocol support'),
    ('passwordless', false, '{}', 'Passwordless authentication support'),
    ('ldap', false, '{}', 'LDAP/AD integration support'),
    ('max_token_ttl', true, '{"value": "15m"}', 'Maximum token TTL (15 minutes)'),
    ('allowed_grant_types', true, '{"value": ["authorization_code", "refresh_token", "client_credentials"]}', 'Allowed OAuth grant types'),
    ('allowed_scope_namespaces', true, '{"value": ["openid", "profile", "users", "clients"]}', 'Allowed scope namespaces'),
    ('pkce_mandatory', true, '{"value": true}', 'PKCE mandatory for OAuth flows');

-- Add comment
COMMENT ON TABLE system_capabilities IS 'Stores global system-level capabilities (what exists at all)';
COMMENT ON COLUMN system_capabilities.capability_key IS 'Capability identifier (e.g., mfa, totp, saml, oidc, oauth2, passwordless, ldap, max_token_ttl, allowed_grant_types, allowed_scope_namespaces, pkce_mandatory)';
COMMENT ON COLUMN system_capabilities.enabled IS 'Whether this capability is supported by the system';
COMMENT ON COLUMN system_capabilities.default_value IS 'Default configuration for this capability as JSON';

