-- Rollback: Remove migrated capability data
-- This rollback removes the capability model data that was migrated from tenant_settings
-- Note: This does NOT restore tenant_settings data, as it was not deleted during migration

-- Step 1: Remove user capability states that were migrated
DELETE FROM user_capability_state
WHERE capability_key IN ('mfa', 'totp')
  AND enrolled = true
  AND enrolled_at >= (SELECT MAX(updated_at) FROM migrations WHERE version = 22);

-- Step 2: Remove tenant feature enablements that were migrated
DELETE FROM tenant_feature_enablement
WHERE feature_key = 'mfa'
  AND enabled = true
  AND enabled_at >= (SELECT MAX(updated_at) FROM migrations WHERE version = 22);

-- Step 3: Reset tenant capabilities to defaults (optional - can be kept)
-- Note: We keep tenant_capabilities as they represent system assignments
-- Only reset the values that were migrated from tenant_settings
UPDATE tenant_capabilities
SET value = NULL
WHERE capability_key = 'max_token_ttl'
  AND value IS NOT NULL;

UPDATE tenant_capabilities
SET enabled = false
WHERE capability_key = 'mfa'
  AND enabled = true;

-- Note: The migration does not delete tenant_capabilities entries
-- as they represent system-level assignments that should persist
-- If full rollback is needed, you can delete all tenant_capabilities:
-- DELETE FROM tenant_capabilities;

