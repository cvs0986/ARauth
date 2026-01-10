-- Migration: Create webhooks table
-- This table stores webhook configurations for tenants

CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE NOT NULL,
    name VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    secret VARCHAR(255) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    events TEXT[] NOT NULL, -- Array of event types to subscribe to
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(tenant_id, name)
);

-- Indexes
CREATE INDEX idx_webhooks_tenant_id ON webhooks(tenant_id);
CREATE INDEX idx_webhooks_enabled ON webhooks(enabled) WHERE enabled = true;
CREATE INDEX idx_webhooks_deleted_at ON webhooks(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_webhooks_events ON webhooks USING GIN(events); -- GIN index for array search

-- Comments
COMMENT ON TABLE webhooks IS 'Stores webhook configurations for tenants';
COMMENT ON COLUMN webhooks.events IS 'Array of event types this webhook subscribes to (e.g., user.created, role.assigned)';
COMMENT ON COLUMN webhooks.secret IS 'Secret used to sign webhook payloads (HMAC-SHA256)';

