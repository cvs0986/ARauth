-- Migration: Create user_capability_state table
-- This table stores user-level capability enrollment state (User layer)
-- Implements the "User Enrollment/State" from the capability model

CREATE TABLE user_capability_state (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    capability_key VARCHAR(255) NOT NULL,
    enrolled BOOLEAN NOT NULL DEFAULT false,
    state_data JSONB, -- e.g., TOTP secret, recovery codes, enrollment metadata
    enrolled_at TIMESTAMP,
    last_used_at TIMESTAMP,
    PRIMARY KEY (user_id, capability_key)
);

-- Create indexes for performance
CREATE INDEX idx_user_capability_state_user_id ON user_capability_state(user_id);
CREATE INDEX idx_user_capability_state_capability ON user_capability_state(capability_key);
CREATE INDEX idx_user_capability_state_enrolled ON user_capability_state(enrolled) WHERE enrolled = true;

-- Add comment
COMMENT ON TABLE user_capability_state IS 'Stores user-level capability enrollment state (what user has enrolled in)';
COMMENT ON COLUMN user_capability_state.capability_key IS 'Capability identifier (e.g., totp, mfa, passwordless)';
COMMENT ON COLUMN user_capability_state.enrolled IS 'Whether the user is enrolled in this capability';
COMMENT ON COLUMN user_capability_state.state_data IS 'Capability-specific state data as JSON (e.g., TOTP secret, recovery codes, enrollment metadata)';

