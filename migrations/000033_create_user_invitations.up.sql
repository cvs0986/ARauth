-- Migration: Create user invitations table
-- Purpose: Store user invitation records for invite-based onboarding

CREATE TABLE IF NOT EXISTS user_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    invited_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE, -- Invitation token (hashed)
    token_hash VARCHAR(255) NOT NULL UNIQUE, -- SHA256 hash for lookup
    expires_at TIMESTAMP NOT NULL,
    accepted_at TIMESTAMP,
    accepted_by UUID REFERENCES users(id) ON DELETE SET NULL,
    role_ids UUID[], -- Array of role IDs to assign upon acceptance
    metadata JSONB, -- Additional invitation metadata
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT uq_user_invitations_tenant_email UNIQUE (tenant_id, email) WHERE deleted_at IS NULL AND accepted_at IS NULL
);

-- Indexes
CREATE INDEX idx_user_invitations_tenant_id ON user_invitations(tenant_id);
CREATE INDEX idx_user_invitations_token_hash ON user_invitations(token_hash);
CREATE INDEX idx_user_invitations_email ON user_invitations(email);
CREATE INDEX idx_user_invitations_expires_at ON user_invitations(expires_at) WHERE accepted_at IS NULL AND deleted_at IS NULL;
CREATE INDEX idx_user_invitations_deleted_at ON user_invitations(deleted_at) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE user_invitations IS 'User invitation records for invite-based onboarding';
COMMENT ON COLUMN user_invitations.token IS 'Plaintext invitation token (stored temporarily, then hashed)';
COMMENT ON COLUMN user_invitations.token_hash IS 'SHA256 hash of the invitation token for fast lookup';
COMMENT ON COLUMN user_invitations.role_ids IS 'Array of role IDs to assign to the user upon invitation acceptance';
COMMENT ON COLUMN user_invitations.metadata IS 'Additional metadata (e.g., custom message, redirect URL)';

