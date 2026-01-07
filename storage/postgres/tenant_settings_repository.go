package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/storage/interfaces"
)

// tenantSettingsRepository implements TenantSettingsRepository for PostgreSQL
type tenantSettingsRepository struct {
	db *sql.DB
}

// NewTenantSettingsRepository creates a new PostgreSQL tenant settings repository
func NewTenantSettingsRepository(db *sql.DB) interfaces.TenantSettingsRepository {
	return &tenantSettingsRepository{db: db}
}

// GetByTenantID retrieves settings for a tenant
func (r *tenantSettingsRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*interfaces.TenantSettings, error) {
	query := `
		SELECT id, tenant_id, access_token_ttl_minutes, refresh_token_ttl_days,
		       id_token_ttl_minutes, remember_me_enabled, remember_me_refresh_token_ttl_days,
		       remember_me_access_token_ttl_minutes, token_rotation_enabled,
		       require_mfa_for_extended_sessions
		FROM tenant_settings
		WHERE tenant_id = $1
	`

	settings := &interfaces.TenantSettings{}
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&settings.ID, &settings.TenantID, &settings.AccessTokenTTLMinutes,
		&settings.RefreshTokenTTLDays, &settings.IDTokenTTLMinutes,
		&settings.RememberMeEnabled, &settings.RememberMeRefreshTokenTTLDays,
		&settings.RememberMeAccessTokenTTLMinutes, &settings.TokenRotationEnabled,
		&settings.RequireMFAForExtendedSessions,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant settings not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant settings: %w", err)
	}

	return settings, nil
}

// Create creates new tenant settings
func (r *tenantSettingsRepository) Create(ctx context.Context, settings *interfaces.TenantSettings) error {
	query := `
		INSERT INTO tenant_settings (
			id, tenant_id, access_token_ttl_minutes, refresh_token_ttl_days,
			id_token_ttl_minutes, remember_me_enabled, remember_me_refresh_token_ttl_days,
			remember_me_access_token_ttl_minutes, token_rotation_enabled,
			require_mfa_for_extended_sessions, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	now := time.Now()
	if settings.ID == uuid.Nil {
		settings.ID = uuid.New()
	}

	_, err := r.db.ExecContext(ctx, query,
		settings.ID, settings.TenantID, settings.AccessTokenTTLMinutes,
		settings.RefreshTokenTTLDays, settings.IDTokenTTLMinutes,
		settings.RememberMeEnabled, settings.RememberMeRefreshTokenTTLDays,
		settings.RememberMeAccessTokenTTLMinutes, settings.TokenRotationEnabled,
		settings.RequireMFAForExtendedSessions, now, now,
	)

	if err != nil {
		return fmt.Errorf("failed to create tenant settings: %w", err)
	}

	return nil
}

// Update updates existing tenant settings
func (r *tenantSettingsRepository) Update(ctx context.Context, settings *interfaces.TenantSettings) error {
	query := `
		UPDATE tenant_settings
		SET access_token_ttl_minutes = $2, refresh_token_ttl_days = $3,
		    id_token_ttl_minutes = $4, remember_me_enabled = $5,
		    remember_me_refresh_token_ttl_days = $6, remember_me_access_token_ttl_minutes = $7,
		    token_rotation_enabled = $8, require_mfa_for_extended_sessions = $9,
		    updated_at = $10
		WHERE tenant_id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		settings.TenantID, settings.AccessTokenTTLMinutes, settings.RefreshTokenTTLDays,
		settings.IDTokenTTLMinutes, settings.RememberMeEnabled,
		settings.RememberMeRefreshTokenTTLDays, settings.RememberMeAccessTokenTTLMinutes,
		settings.TokenRotationEnabled, settings.RequireMFAForExtendedSessions,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update tenant settings: %w", err)
	}

	return nil
}

// Delete deletes tenant settings
func (r *tenantSettingsRepository) Delete(ctx context.Context, tenantID uuid.UUID) error {
	query := `DELETE FROM tenant_settings WHERE tenant_id = $1`

	_, err := r.db.ExecContext(ctx, query, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete tenant settings: %w", err)
	}

	return nil
}

