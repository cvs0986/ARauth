-- Rollback: Make tenant_id NOT NULL again

ALTER TABLE refresh_tokens
ALTER COLUMN tenant_id SET NOT NULL;

