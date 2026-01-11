-- Create oauth_clients table for OAuth2 client credential management
CREATE TABLE oauth_clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    client_id VARCHAR(255) NOT NULL UNIQUE,
    client_secret_hash TEXT NOT NULL,
    description TEXT,
    redirect_uris TEXT[], -- Array of allowed redirect URIs for OAuth2 flows
    grant_types TEXT[] NOT NULL DEFAULT ARRAY['authorization_code'], -- Supported OAuth2 grant types
    scopes TEXT[] NOT NULL DEFAULT ARRAY['openid'], -- Allowed OAuth2 scopes
    is_confidential BOOLEAN NOT NULL DEFAULT true, -- Public (SPA) vs confidential (backend) client
    is_active BOOLEAN NOT NULL DEFAULT true, -- Soft disable without deletion
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users(id),

-- Ensure unique client names per tenant
CONSTRAINT oauth_clients_tenant_name_unique UNIQUE(tenant_id, name)
);

-- Index for tenant-based queries (most common access pattern)
CREATE INDEX idx_oauth_clients_tenant_id ON oauth_clients (tenant_id);

-- Index for client_id lookups (used during OAuth2 authentication)
CREATE INDEX idx_oauth_clients_client_id ON oauth_clients (client_id);

-- Index for filtering active clients
CREATE INDEX idx_oauth_clients_is_active ON oauth_clients (is_active);

-- Comments for documentation
COMMENT ON TABLE oauth_clients IS 'OAuth2 client credentials for machine-to-machine authentication';

COMMENT ON COLUMN oauth_clients.client_id IS 'Unique client identifier (format: client_<hex>)';

COMMENT ON COLUMN oauth_clients.client_secret_hash IS 'bcrypt hash of client secret (cost 12, never plaintext)';

COMMENT ON COLUMN oauth_clients.redirect_uris IS 'Allowed redirect URIs for authorization code flow';

COMMENT ON COLUMN oauth_clients.grant_types IS 'Supported OAuth2 grant types (authorization_code, client_credentials, refresh_token)';

COMMENT ON COLUMN oauth_clients.scopes IS 'Allowed OAuth2 scopes for this client';

COMMENT ON COLUMN oauth_clients.is_confidential IS 'true = confidential client (can keep secrets), false = public client (SPA, mobile)';

COMMENT ON COLUMN oauth_clients.is_active IS 'Soft disable flag - inactive clients cannot authenticate';