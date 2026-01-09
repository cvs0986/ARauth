-- Rollback: Drop audit_events table

DROP INDEX IF EXISTS idx_audit_events_actor_timestamp;
DROP INDEX IF EXISTS idx_audit_events_tenant_timestamp;
DROP INDEX IF EXISTS idx_audit_events_result;
DROP INDEX IF EXISTS idx_audit_events_timestamp;
DROP INDEX IF EXISTS idx_audit_events_tenant_id;
DROP INDEX IF EXISTS idx_audit_events_target_id;
DROP INDEX IF EXISTS idx_audit_events_actor_user_id;
DROP INDEX IF EXISTS idx_audit_events_event_type;

DROP TABLE IF EXISTS audit_events;

