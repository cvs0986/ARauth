-- Migration: Make tenant_id nullable in refresh_tokens for SYSTEM users
-- SYSTEM users don't have a tenant_id, so refresh tokens should support NULL

ALTER TABLE refresh_tokens
ALTER COLUMN tenant_id DROP NOT NULL;

-- Add comment
COMMENT ON COLUMN refresh_tokens.tenant_id IS 'Tenant ID (NULL for SYSTEM users)';

