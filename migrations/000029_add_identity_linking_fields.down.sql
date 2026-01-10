-- Migration: Remove identity linking fields from federated_identities

-- Drop indexes
DROP INDEX IF EXISTS idx_federated_identities_user_primary_unique;
DROP INDEX IF EXISTS idx_federated_identities_user_primary;

-- Drop columns
ALTER TABLE federated_identities 
DROP COLUMN IF EXISTS verified_at,
DROP COLUMN IF EXISTS verified,
DROP COLUMN IF EXISTS is_primary;

