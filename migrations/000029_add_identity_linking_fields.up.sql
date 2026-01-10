-- Migration: Add identity linking fields to federated_identities
-- This migration adds fields to support identity linking features:
-- - is_primary: marks the primary identity for a user
-- - verified: whether the identity has been verified
-- - verified_at: timestamp of verification

-- Add new columns if they don't exist
DO $$
BEGIN
    -- Add is_primary column
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'federated_identities' 
        AND column_name = 'is_primary'
    ) THEN
        ALTER TABLE federated_identities 
        ADD COLUMN is_primary BOOLEAN NOT NULL DEFAULT false;
    END IF;

    -- Add verified column
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'federated_identities' 
        AND column_name = 'verified'
    ) THEN
        ALTER TABLE federated_identities 
        ADD COLUMN verified BOOLEAN NOT NULL DEFAULT false;
    END IF;

    -- Add verified_at column
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'federated_identities' 
        AND column_name = 'verified_at'
    ) THEN
        ALTER TABLE federated_identities 
        ADD COLUMN verified_at TIMESTAMP WITH TIME ZONE;
    END IF;
END $$;

-- Create index for primary identity lookups
CREATE INDEX IF NOT EXISTS idx_federated_identities_user_primary 
ON federated_identities(user_id, is_primary) 
WHERE is_primary = true;

-- Create unique constraint: only one primary identity per user
CREATE UNIQUE INDEX IF NOT EXISTS idx_federated_identities_user_primary_unique 
ON federated_identities(user_id) 
WHERE is_primary = true;

-- Comments
COMMENT ON COLUMN federated_identities.is_primary IS 'Marks this as the primary identity for the user';
COMMENT ON COLUMN federated_identities.verified IS 'Whether this identity has been verified';
COMMENT ON COLUMN federated_identities.verified_at IS 'Timestamp when the identity was verified';

