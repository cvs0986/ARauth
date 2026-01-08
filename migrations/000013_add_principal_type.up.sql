-- Migration: Add principal_type to users table and make tenant_id nullable for SYSTEM users
-- This enables the two-plane architecture: Platform Control Plane (SYSTEM) vs Tenant Plane (TENANT)

-- Add principal_type column
ALTER TABLE users
ADD COLUMN principal_type VARCHAR(50) DEFAULT 'TENANT' NOT NULL 
  CHECK (principal_type IN ('SYSTEM', 'TENANT', 'SERVICE'));

-- Make tenant_id nullable (required for SYSTEM users)
-- First, drop the existing NOT NULL constraint if it exists
ALTER TABLE users
ALTER COLUMN tenant_id DROP NOT NULL;

-- Add constraint: SYSTEM users must have tenant_id = NULL, others must have tenant_id NOT NULL
ALTER TABLE users
ADD CONSTRAINT chk_principal_type_tenant_id 
  CHECK (
    (principal_type = 'SYSTEM' AND tenant_id IS NULL) OR
    (principal_type != 'SYSTEM' AND tenant_id IS NOT NULL)
  );

-- Create indexes for principal type lookups
CREATE INDEX idx_users_principal_type ON users(principal_type);
CREATE INDEX idx_users_system_users ON users(principal_type) WHERE principal_type = 'SYSTEM';

-- Update existing users to ensure they have tenant_id (they should already)
UPDATE users SET principal_type = 'TENANT' WHERE principal_type IS NULL OR tenant_id IS NOT NULL;

