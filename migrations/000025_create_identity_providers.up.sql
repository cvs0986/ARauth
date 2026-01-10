-- Migration: Create identity_providers table for OIDC/SAML federation
-- This table stores configuration for external identity providers

CREATE TABLE identity_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('oidc', 'saml')),
    enabled BOOLEAN NOT NULL DEFAULT true,
    configuration JSONB NOT NULL,
    attribute_mapping JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(tenant_id, name)
);

-- Indexes
CREATE INDEX idx_identity_providers_tenant_id ON identity_providers(tenant_id);
CREATE INDEX idx_identity_providers_type ON identity_providers(type);
CREATE INDEX idx_identity_providers_enabled ON identity_providers(enabled);
CREATE INDEX idx_identity_providers_deleted_at ON identity_providers(deleted_at) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE identity_providers IS 'Stores configuration for external identity providers (OIDC/SAML)';
COMMENT ON COLUMN identity_providers.type IS 'Type of identity provider: oidc or saml';
COMMENT ON COLUMN identity_providers.configuration IS 'Provider-specific configuration (JSON)';
COMMENT ON COLUMN identity_providers.attribute_mapping IS 'Mapping from provider attributes to ARauth user attributes';

