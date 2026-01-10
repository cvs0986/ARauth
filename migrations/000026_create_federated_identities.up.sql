-- Migration: Create federated_identities table
-- This table links users to their external identities from identity providers

CREATE TABLE federated_identities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    provider_id UUID REFERENCES identity_providers(id) ON DELETE CASCADE NOT NULL,
    external_id VARCHAR(255) NOT NULL,
    attributes JSONB,
    is_primary BOOLEAN NOT NULL DEFAULT false,
    verified BOOLEAN NOT NULL DEFAULT false,
    verified_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(provider_id, external_id)
);

-- Indexes
CREATE INDEX idx_federated_identities_user_id ON federated_identities(user_id);
CREATE INDEX idx_federated_identities_provider_id ON federated_identities(provider_id);
CREATE INDEX idx_federated_identities_external_id ON federated_identities(external_id);
CREATE INDEX idx_federated_identities_is_primary ON federated_identities(is_primary) WHERE is_primary = true;

-- Comments
COMMENT ON TABLE federated_identities IS 'Links users to their external identities from identity providers';
COMMENT ON COLUMN federated_identities.external_id IS 'User ID from the external identity provider';
COMMENT ON COLUMN federated_identities.attributes IS 'Additional attributes from the identity provider';
COMMENT ON COLUMN federated_identities.is_primary IS 'Whether this is the primary identity for the user';
COMMENT ON COLUMN federated_identities.verified IS 'Whether this identity has been verified';

