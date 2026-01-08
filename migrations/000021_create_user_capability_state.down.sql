-- Rollback: Drop user_capability_state table

DROP INDEX IF EXISTS idx_user_capability_state_enrolled;
DROP INDEX IF EXISTS idx_user_capability_state_capability;
DROP INDEX IF EXISTS idx_user_capability_state_user_id;
DROP TABLE IF EXISTS user_capability_state;

