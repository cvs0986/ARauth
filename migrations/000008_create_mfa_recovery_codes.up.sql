CREATE TABLE mfa_recovery_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash TEXT NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_mfa_recovery_codes_user_id ON mfa_recovery_codes(user_id);
CREATE INDEX idx_mfa_recovery_codes_used ON mfa_recovery_codes(user_id, used_at) WHERE used_at IS NULL;

