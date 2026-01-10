-- Migration: Create SCIM tokens table
-- Purpose: Store OAuth Bearer tokens for SCIM API authentication

CREATE TABLE IF NOT EXISTS scim_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL, -- Token name/description
    token_hash VARCHAR(255) NOT NULL, -- Hashed token value (bcrypt)
    lookup_hash VARCHAR(64) NOT NULL, -- SHA256 hash for fast lookup
    scopes TEXT[] NOT NULL, -- Array of allowed SCIM scopes (e.g., ["users", "groups"])
    expires_at TIMESTAMP,
    last_used_at TIMESTAMP,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT uq_scim_tokens_tenant_name UNIQUE (tenant_id, name) WHERE deleted_at IS NULL
);

-- Indexes
CREATE INDEX idx_scim_tokens_tenant_id ON scim_tokens(tenant_id);
CREATE INDEX idx_scim_tokens_lookup_hash ON scim_tokens(lookup_hash);
CREATE INDEX idx_scim_tokens_deleted_at ON scim_tokens(deleted_at) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE scim_tokens IS 'OAuth Bearer tokens for SCIM API authentication';
COMMENT ON COLUMN scim_tokens.token_hash IS 'BCrypt hash of the SCIM token (for verification)';
COMMENT ON COLUMN scim_tokens.lookup_hash IS 'SHA256 hash of the SCIM token (for fast lookup)';
COMMENT ON COLUMN scim_tokens.scopes IS 'Array of SCIM scopes (users, groups, etc.)';

