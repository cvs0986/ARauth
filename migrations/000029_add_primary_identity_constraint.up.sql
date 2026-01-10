-- Migration: Add unique constraint for primary identity
-- Ensures only one primary identity per user

-- Create unique constraint: only one primary identity per user
CREATE UNIQUE INDEX IF NOT EXISTS idx_federated_identities_user_primary_unique 
ON federated_identities(user_id) 
WHERE is_primary = true;

