-- Migration: Create structured audit_events table
-- This table stores structured audit events with actor, target, and metadata
-- It complements the existing audit_logs table with more detailed event tracking

CREATE TABLE audit_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(100) NOT NULL,
    actor_user_id UUID NOT NULL,
    actor_username VARCHAR(255) NOT NULL,
    actor_principal_type VARCHAR(20) NOT NULL,
    target_type VARCHAR(50),
    target_id UUID,
    target_identifier VARCHAR(255),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    source_ip INET,
    user_agent TEXT,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    metadata JSONB,
    result VARCHAR(20) NOT NULL, -- "success", "failure", "denied"
    error TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_audit_events_event_type ON audit_events(event_type);
CREATE INDEX idx_audit_events_actor_user_id ON audit_events(actor_user_id);
CREATE INDEX idx_audit_events_target_id ON audit_events(target_id);
CREATE INDEX idx_audit_events_tenant_id ON audit_events(tenant_id);
CREATE INDEX idx_audit_events_timestamp ON audit_events(timestamp DESC);
CREATE INDEX idx_audit_events_result ON audit_events(result);

-- Composite index for common queries
CREATE INDEX idx_audit_events_tenant_timestamp ON audit_events(tenant_id, timestamp DESC);
CREATE INDEX idx_audit_events_actor_timestamp ON audit_events(actor_user_id, timestamp DESC);

