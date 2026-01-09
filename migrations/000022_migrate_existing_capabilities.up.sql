-- Migration: Migrate existing tenant settings to capability model
-- This migration converts existing tenant_settings data to the new capability model
-- It preserves all existing configurations while moving them to the capability structure

-- Step 1: Assign default capabilities to all existing tenants
-- For each tenant, assign commonly used capabilities based on system defaults
INSERT INTO tenant_capabilities (tenant_id, capability_key, enabled, value, configured_at)
SELECT 
    t.id AS tenant_id,
    sc.capability_key,
    sc.enabled AS enabled, -- Use system default enabled state
    sc.default_value AS value, -- Use system default value
    NOW() AS configured_at
FROM tenants t
CROSS JOIN system_capabilities sc
WHERE sc.enabled = true -- Only assign enabled system capabilities
  AND NOT EXISTS (
    SELECT 1 FROM tenant_capabilities tc 
    WHERE tc.tenant_id = t.id AND tc.capability_key = sc.capability_key
  )
ON CONFLICT (tenant_id, capability_key) DO NOTHING;

-- Step 2: Migrate token TTL settings from tenant_settings to max_token_ttl capability
-- Convert access_token_ttl_minutes to max_token_ttl capability value
UPDATE tenant_capabilities tc
SET value = jsonb_build_object(
    'max_access_token_ttl_minutes', ts.access_token_ttl_minutes,
    'max_refresh_token_ttl_days', ts.refresh_token_ttl_days,
    'max_id_token_ttl_minutes', ts.id_token_ttl_minutes,
    'remember_me_enabled', ts.remember_me_enabled,
    'remember_me_refresh_token_ttl_days', ts.remember_me_refresh_token_ttl_days,
    'remember_me_access_token_ttl_minutes', ts.remember_me_access_token_ttl_minutes
)
FROM tenant_settings ts
WHERE tc.tenant_id = ts.tenant_id
  AND tc.capability_key = 'max_token_ttl'
  AND ts.access_token_ttl_minutes IS NOT NULL;

-- Step 3: Migrate MFA settings from tenant_settings to mfa capability
-- If tenant has mfa_required = true, enable mfa capability and feature
UPDATE tenant_capabilities tc
SET enabled = true
FROM tenant_settings ts
WHERE tc.tenant_id = ts.tenant_id
  AND tc.capability_key = 'mfa'
  AND ts.mfa_required = true
  AND tc.enabled = false;

-- Step 4: Enable MFA feature for tenants that have mfa_required = true
INSERT INTO tenant_feature_enablement (tenant_id, feature_key, enabled, enabled_at)
SELECT 
    ts.tenant_id,
    'mfa' AS feature_key,
    true AS enabled,
    NOW() AS enabled_at
FROM tenant_settings ts
WHERE ts.mfa_required = true
  AND NOT EXISTS (
    SELECT 1 FROM tenant_feature_enablement tfe
    WHERE tfe.tenant_id = ts.tenant_id AND tfe.feature_key = 'mfa'
  )
ON CONFLICT (tenant_id, feature_key) DO NOTHING;

-- Step 5: Migrate user MFA enrollment state
-- If a user has MFA enabled (mfa_enabled = true), create user capability state
INSERT INTO user_capability_state (user_id, capability_key, enrolled, enrolled_at)
SELECT 
    u.id AS user_id,
    'mfa' AS capability_key,
    true AS enrolled,
    COALESCE(u.updated_at, u.created_at) AS enrolled_at
FROM users u
WHERE u.mfa_enabled = true
  AND NOT EXISTS (
    SELECT 1 FROM user_capability_state ucs
    WHERE ucs.user_id = u.id AND ucs.capability_key = 'mfa'
  )
ON CONFLICT (user_id, capability_key) DO NOTHING;

-- Step 6: Migrate TOTP enrollment for users with MFA enabled
-- If user has encrypted MFA secret, they are enrolled in TOTP
INSERT INTO user_capability_state (user_id, capability_key, enrolled, enrolled_at, state_data)
SELECT 
    u.id AS user_id,
    'totp' AS capability_key,
    true AS enrolled,
    COALESCE(u.updated_at, u.created_at) AS enrolled_at,
    jsonb_build_object('has_secret', true) AS state_data
FROM users u
WHERE u.mfa_enabled = true
  AND u.mfa_secret_encrypted IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM user_capability_state ucs
    WHERE ucs.user_id = u.id AND ucs.capability_key = 'totp'
  )
ON CONFLICT (user_id, capability_key) DO NOTHING;

-- Step 7: Add comments for documentation
COMMENT ON TABLE tenant_capabilities IS 'Migrated from tenant_settings - see migration 000022';
COMMENT ON TABLE tenant_feature_enablement IS 'Migrated from tenant_settings.mfa_required - see migration 000022';
COMMENT ON TABLE user_capability_state IS 'Migrated from users.mfa_enabled and users.mfa_secret_encrypted - see migration 000022';

