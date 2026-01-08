-- Migration: Create tenant_capabilities table
-- This table stores which capabilities are allowed for each tenant (System → Tenant layer)
-- Implements the "System → Tenant Capability Assignment" from the capability model

CREATE TABLE tenant_capabilities (
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    capability_key VARCHAR(255) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT false,
    value JSONB, -- For capability-specific configuration (e.g., max_token_ttl: "10m", allowed_grant_types: ["authorization_code"])
    configured_by UUID REFERENCES users(id),
    configured_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, capability_key)
);

-- Create indexes for performance
CREATE INDEX idx_tenant_capabilities_tenant_id ON tenant_capabilities(tenant_id);
CREATE INDEX idx_tenant_capabilities_key ON tenant_capabilities(capability_key);
CREATE INDEX idx_tenant_capabilities_enabled ON tenant_capabilities(enabled) WHERE enabled = true;

-- Add comment
COMMENT ON TABLE tenant_capabilities IS 'Stores which capabilities are allowed for each tenant (System → Tenant assignment)';
COMMENT ON COLUMN tenant_capabilities.capability_key IS 'Capability identifier (e.g., mfa, totp, saml, oidc, oauth2, passwordless, ldap, max_token_ttl, allowed_grant_types, allowed_scope_namespaces)';
COMMENT ON COLUMN tenant_capabilities.enabled IS 'Whether this capability is allowed for the tenant';
COMMENT ON COLUMN tenant_capabilities.value IS 'Capability-specific configuration as JSON (e.g., max_token_ttl duration, allowed grant types array)';

