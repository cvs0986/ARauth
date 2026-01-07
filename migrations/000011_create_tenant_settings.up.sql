-- Create tenant_settings table for configurable token lifetimes
CREATE TABLE tenant_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Token Lifetime Settings (in minutes/days)
    access_token_ttl_minutes INT NOT NULL DEFAULT 15,
    refresh_token_ttl_days INT NOT NULL DEFAULT 30,
    id_token_ttl_minutes INT NOT NULL DEFAULT 60,
    
    -- Remember Me Settings
    remember_me_enabled BOOLEAN NOT NULL DEFAULT true,
    remember_me_refresh_token_ttl_days INT NOT NULL DEFAULT 90,
    remember_me_access_token_ttl_minutes INT NOT NULL DEFAULT 60,
    
    -- Security Settings
    token_rotation_enabled BOOLEAN NOT NULL DEFAULT true,
    require_mfa_for_extended_sessions BOOLEAN NOT NULL DEFAULT false,
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    UNIQUE(tenant_id)
);

CREATE INDEX idx_tenant_settings_tenant_id ON tenant_settings(tenant_id);

-- Add constraints for minimum/maximum values
ALTER TABLE tenant_settings ADD CONSTRAINT chk_access_token_ttl_min 
    CHECK (access_token_ttl_minutes >= 5 AND access_token_ttl_minutes <= 1440);
ALTER TABLE tenant_settings ADD CONSTRAINT chk_refresh_token_ttl_min 
    CHECK (refresh_token_ttl_days >= 1 AND refresh_token_ttl_days <= 365);
ALTER TABLE tenant_settings ADD CONSTRAINT chk_remember_me_refresh_ttl_min 
    CHECK (remember_me_refresh_token_ttl_days >= 1 AND remember_me_refresh_token_ttl_days <= 365);

