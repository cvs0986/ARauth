-- Migration: Create tenant_feature_enablement table
-- This table stores which features tenants have actually enabled (Tenant layer)
-- Implements the "Tenant Enablement" from the capability model

CREATE TABLE tenant_feature_enablement (
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    feature_key VARCHAR(255) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT false,
    configuration JSONB, -- Feature-specific configuration (e.g., MFA enforcement rules)
    enabled_by UUID REFERENCES users(id),
    enabled_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, feature_key)
);

-- Create indexes for performance
CREATE INDEX idx_tenant_feature_enablement_tenant_id ON tenant_feature_enablement(tenant_id);
CREATE INDEX idx_tenant_feature_enablement_key ON tenant_feature_enablement(feature_key);
CREATE INDEX idx_tenant_feature_enablement_enabled ON tenant_feature_enablement(enabled) WHERE enabled = true;

-- Add comment
COMMENT ON TABLE tenant_feature_enablement IS 'Stores which features tenants have actually enabled (Tenant choice)';
COMMENT ON COLUMN tenant_feature_enablement.feature_key IS 'Feature identifier (e.g., mfa, totp, saml, oidc, oauth2, passwordless, ldap)';
COMMENT ON COLUMN tenant_feature_enablement.enabled IS 'Whether this feature is enabled by the tenant';
COMMENT ON COLUMN tenant_feature_enablement.configuration IS 'Feature-specific configuration as JSON (e.g., MFA enforcement rules, OIDC client settings)';

