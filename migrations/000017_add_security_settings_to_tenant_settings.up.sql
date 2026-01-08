-- Migration: Add security settings to tenant_settings table
-- This allows each tenant to have their own password policies, MFA requirements, and rate limiting

ALTER TABLE tenant_settings ADD COLUMN IF NOT EXISTS
    -- Password Policy Settings
    min_password_length INT NOT NULL DEFAULT 12,
    require_uppercase BOOLEAN NOT NULL DEFAULT true,
    require_lowercase BOOLEAN NOT NULL DEFAULT true,
    require_numbers BOOLEAN NOT NULL DEFAULT true,
    require_special_chars BOOLEAN NOT NULL DEFAULT true,
    password_expiry_days INT DEFAULT NULL, -- NULL means never expires
    
    -- MFA Settings
    mfa_required BOOLEAN NOT NULL DEFAULT false,
    
    -- Rate Limiting Settings
    rate_limit_requests INT NOT NULL DEFAULT 100,
    rate_limit_window_seconds INT NOT NULL DEFAULT 60;

-- Add constraints
ALTER TABLE tenant_settings ADD CONSTRAINT chk_min_password_length 
    CHECK (min_password_length >= 8 AND min_password_length <= 128);
ALTER TABLE tenant_settings ADD CONSTRAINT chk_password_expiry_days 
    CHECK (password_expiry_days IS NULL OR password_expiry_days >= 0);
ALTER TABLE tenant_settings ADD CONSTRAINT chk_rate_limit_requests 
    CHECK (rate_limit_requests >= 1 AND rate_limit_requests <= 10000);
ALTER TABLE tenant_settings ADD CONSTRAINT chk_rate_limit_window 
    CHECK (rate_limit_window_seconds >= 1 AND rate_limit_window_seconds <= 3600);

