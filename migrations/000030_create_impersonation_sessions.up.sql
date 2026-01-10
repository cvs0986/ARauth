-- Migration: Create impersonation_sessions table
-- Purpose: Track admin impersonation sessions for audit and security

CREATE TABLE IF NOT EXISTS impersonation_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    impersonator_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP,
    token_jti TEXT, -- JWT ID of the impersonation token (stored as text for UUID string)
    reason TEXT, -- Optional reason for impersonation
    metadata JSONB, -- Additional metadata
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_impersonation_sessions_impersonator ON impersonation_sessions(impersonator_user_id);
CREATE INDEX idx_impersonation_sessions_target ON impersonation_sessions(target_user_id);
CREATE INDEX idx_impersonation_sessions_tenant ON impersonation_sessions(tenant_id);
CREATE INDEX idx_impersonation_sessions_active ON impersonation_sessions(ended_at) WHERE ended_at IS NULL;
CREATE INDEX idx_impersonation_sessions_token_jti ON impersonation_sessions(token_jti);

-- Comments
COMMENT ON TABLE impersonation_sessions IS 'Tracks admin impersonation sessions for audit and security purposes';
COMMENT ON COLUMN impersonation_sessions.impersonator_user_id IS 'User who initiated the impersonation (admin)';
COMMENT ON COLUMN impersonation_sessions.target_user_id IS 'User being impersonated';
COMMENT ON COLUMN impersonation_sessions.token_jti IS 'JWT ID of the impersonation token';
COMMENT ON COLUMN impersonation_sessions.reason IS 'Optional reason for impersonation (e.g., support ticket, debugging)';

