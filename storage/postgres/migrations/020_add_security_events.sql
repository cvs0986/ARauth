-- Migration: Add security_events table
-- Created: 2026-01-11
-- Purpose: Store security events for audit and monitoring

CREATE TABLE IF NOT EXISTS security_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    event_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    tenant_id UUID,
    user_id UUID,
    ip VARCHAR(45),
    resource VARCHAR(255),
    action VARCHAR(100),
    result VARCHAR(50),
    details JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_security_events_event_type ON security_events (event_type);

CREATE INDEX IF NOT EXISTS idx_security_events_severity ON security_events (severity);

CREATE INDEX IF NOT EXISTS idx_security_events_tenant_id ON security_events (tenant_id);

CREATE INDEX IF NOT EXISTS idx_security_events_user_id ON security_events (user_id);

CREATE INDEX IF NOT EXISTS idx_security_events_created_at ON security_events (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_security_events_ip ON security_events (ip);

-- Composite index for common filter combinations
CREATE INDEX IF NOT EXISTS idx_security_events_tenant_severity ON security_events (
    tenant_id,
    severity,
    created_at DESC
);

-- Comment on table
COMMENT ON TABLE security_events IS 'Security events for audit and monitoring purposes';

COMMENT ON COLUMN security_events.event_type IS 'Type of security event (auth_failure, permission_denied, etc.)';

COMMENT ON COLUMN security_events.severity IS 'Severity level: info, warning, critical';

COMMENT ON COLUMN security_events.details IS 'Additional event details in JSON format';