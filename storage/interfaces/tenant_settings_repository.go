package interfaces

import (
	"context"

	"github.com/google/uuid"
)

// TenantSettings represents tenant-specific settings
type TenantSettings struct {
	ID                                uuid.UUID `db:"id"`
	TenantID                          uuid.UUID `db:"tenant_id"`
	AccessTokenTTLMinutes            int       `db:"access_token_ttl_minutes"`
	RefreshTokenTTLDays              int       `db:"refresh_token_ttl_days"`
	IDTokenTTLMinutes                int       `db:"id_token_ttl_minutes"`
	RememberMeEnabled                bool      `db:"remember_me_enabled"`
	RememberMeRefreshTokenTTLDays    int       `db:"remember_me_refresh_token_ttl_days"`
	RememberMeAccessTokenTTLMinutes  int       `db:"remember_me_access_token_ttl_minutes"`
	TokenRotationEnabled             bool      `db:"token_rotation_enabled"`
	RequireMFAForExtendedSessions   bool      `db:"require_mfa_for_extended_sessions"`
}

// TenantSettingsRepository defines operations for tenant settings
type TenantSettingsRepository interface {
	// GetByTenantID retrieves settings for a tenant
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*TenantSettings, error)

	// Create creates new tenant settings
	Create(ctx context.Context, settings *TenantSettings) error

	// Update updates existing tenant settings
	Update(ctx context.Context, settings *TenantSettings) error

	// Delete deletes tenant settings
	Delete(ctx context.Context, tenantID uuid.UUID) error
}

