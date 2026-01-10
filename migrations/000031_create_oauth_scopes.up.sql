-- Migration: Create oauth_scopes table
-- Purpose: Map permissions to OAuth scopes for fine-grained access control

CREATE TABLE IF NOT EXISTS oauth_scopes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL, -- Scope name (e.g., "users.read", "roles.manage")
    description TEXT,
    permissions TEXT[] NOT NULL, -- Array of permission names that map to this scope
    is_default BOOLEAN NOT NULL DEFAULT false, -- Whether this scope is included by default
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT uq_oauth_scopes_tenant_name UNIQUE (tenant_id, name) WHERE deleted_at IS NULL
);

-- Indexes
CREATE INDEX idx_oauth_scopes_tenant_id ON oauth_scopes(tenant_id);
CREATE INDEX idx_oauth_scopes_deleted_at ON oauth_scopes(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_oauth_scopes_default ON oauth_scopes(tenant_id, is_default) WHERE is_default = true AND deleted_at IS NULL;

-- Comments
COMMENT ON TABLE oauth_scopes IS 'Maps OAuth scopes to permissions for fine-grained access control';
COMMENT ON COLUMN oauth_scopes.name IS 'Scope name (e.g., "users.read", "roles.manage")';
COMMENT ON COLUMN oauth_scopes.permissions IS 'Array of permission names that map to this scope';
COMMENT ON COLUMN oauth_scopes.is_default IS 'Whether this scope is included by default in token claims';

