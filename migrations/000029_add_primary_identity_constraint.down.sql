-- Migration: Remove unique constraint for primary identity

DROP INDEX IF EXISTS idx_federated_identities_user_primary_unique;

