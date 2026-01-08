-- Rollback: Remove security settings from tenant_settings table

ALTER TABLE tenant_settings DROP CONSTRAINT IF EXISTS chk_rate_limit_window;
ALTER TABLE tenant_settings DROP CONSTRAINT IF EXISTS chk_rate_limit_requests;
ALTER TABLE tenant_settings DROP CONSTRAINT IF EXISTS chk_password_expiry_days;
ALTER TABLE tenant_settings DROP CONSTRAINT IF EXISTS chk_min_password_length;

ALTER TABLE tenant_settings DROP COLUMN IF EXISTS rate_limit_window_seconds;
ALTER TABLE tenant_settings DROP COLUMN IF EXISTS rate_limit_requests;
ALTER TABLE tenant_settings DROP COLUMN IF EXISTS mfa_required;
ALTER TABLE tenant_settings DROP COLUMN IF EXISTS password_expiry_days;
ALTER TABLE tenant_settings DROP COLUMN IF EXISTS require_special_chars;
ALTER TABLE tenant_settings DROP COLUMN IF EXISTS require_numbers;
ALTER TABLE tenant_settings DROP COLUMN IF EXISTS require_lowercase;
ALTER TABLE tenant_settings DROP COLUMN IF EXISTS require_uppercase;
ALTER TABLE tenant_settings DROP COLUMN IF EXISTS min_password_length;

