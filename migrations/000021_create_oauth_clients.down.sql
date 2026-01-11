-- Drop oauth_clients table and related objects
DROP INDEX IF EXISTS idx_oauth_clients_is_active;

DROP INDEX IF EXISTS idx_oauth_clients_client_id;

DROP INDEX IF EXISTS idx_oauth_clients_tenant_id;

DROP TABLE IF EXISTS oauth_clients;