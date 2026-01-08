package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// tenantFeatureEnablementRepository implements TenantFeatureEnablementRepository for PostgreSQL
type tenantFeatureEnablementRepository struct {
	db *sql.DB
}

// NewTenantFeatureEnablementRepository creates a new PostgreSQL tenant feature enablement repository
func NewTenantFeatureEnablementRepository(db *sql.DB) interfaces.TenantFeatureEnablementRepository {
	return &tenantFeatureEnablementRepository{db: db}
}

// GetByTenantIDAndKey retrieves a tenant feature enablement by tenant ID and key
func (r *tenantFeatureEnablementRepository) GetByTenantIDAndKey(ctx context.Context, tenantID uuid.UUID, key string) (*models.TenantFeatureEnablement, error) {
	query := `
		SELECT tenant_id, feature_key, enabled, configuration, enabled_by, enabled_at
		FROM tenant_feature_enablement
		WHERE tenant_id = $1 AND feature_key = $2
	`

	enablement := &models.TenantFeatureEnablement{}
	var configuration sql.NullString
	var enabledBy sql.NullString

	err := r.db.QueryRowContext(ctx, query, tenantID, key).Scan(
		&enablement.TenantID,
		&enablement.FeatureKey,
		&enablement.Enabled,
		&configuration,
		&enabledBy,
		&enablement.EnabledAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant feature enablement not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant feature enablement: %w", err)
	}

	if configuration.Valid {
		enablement.Configuration = json.RawMessage(configuration.String)
	}
	if enabledBy.Valid {
		id, err := uuid.Parse(enabledBy.String)
		if err == nil {
			enablement.EnabledBy = &id
		}
	}

	return enablement, nil
}

// GetByTenantID retrieves all feature enablements for a tenant
func (r *tenantFeatureEnablementRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*models.TenantFeatureEnablement, error) {
	query := `
		SELECT tenant_id, feature_key, enabled, configuration, enabled_by, enabled_at
		FROM tenant_feature_enablement
		WHERE tenant_id = $1
		ORDER BY feature_key
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant feature enablements: %w", err)
	}
	defer rows.Close()

	var enablements []*models.TenantFeatureEnablement
	for rows.Next() {
		enablement := &models.TenantFeatureEnablement{}
		var configuration sql.NullString
		var enabledBy sql.NullString

		err := rows.Scan(
			&enablement.TenantID,
			&enablement.FeatureKey,
			&enablement.Enabled,
			&configuration,
			&enabledBy,
			&enablement.EnabledAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant feature enablement: %w", err)
		}

		if configuration.Valid {
			enablement.Configuration = json.RawMessage(configuration.String)
		}
		if enabledBy.Valid {
			id, err := uuid.Parse(enabledBy.String)
			if err == nil {
				enablement.EnabledBy = &id
			}
		}

		enablements = append(enablements, enablement)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tenant feature enablements: %w", err)
	}

	return enablements, nil
}

// GetEnabledByTenantID retrieves all enabled features for a tenant
func (r *tenantFeatureEnablementRepository) GetEnabledByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*models.TenantFeatureEnablement, error) {
	query := `
		SELECT tenant_id, feature_key, enabled, configuration, enabled_by, enabled_at
		FROM tenant_feature_enablement
		WHERE tenant_id = $1 AND enabled = true
		ORDER BY feature_key
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled tenant feature enablements: %w", err)
	}
	defer rows.Close()

	var enablements []*models.TenantFeatureEnablement
	for rows.Next() {
		enablement := &models.TenantFeatureEnablement{}
		var configuration sql.NullString
		var enabledBy sql.NullString

		err := rows.Scan(
			&enablement.TenantID,
			&enablement.FeatureKey,
			&enablement.Enabled,
			&configuration,
			&enabledBy,
			&enablement.EnabledAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant feature enablement: %w", err)
		}

		if configuration.Valid {
			enablement.Configuration = json.RawMessage(configuration.String)
		}
		if enabledBy.Valid {
			id, err := uuid.Parse(enabledBy.String)
			if err == nil {
				enablement.EnabledBy = &id
			}
		}

		enablements = append(enablements, enablement)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating enabled tenant feature enablements: %w", err)
	}

	return enablements, nil
}

// Create creates a new tenant feature enablement
func (r *tenantFeatureEnablementRepository) Create(ctx context.Context, enablement *models.TenantFeatureEnablement) error {
	query := `
		INSERT INTO tenant_feature_enablement (tenant_id, feature_key, enabled, configuration, enabled_by, enabled_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	configJSON := "{}"
	if len(enablement.Configuration) > 0 {
		configJSON = string(enablement.Configuration)
	}

	_, err := r.db.ExecContext(ctx, query,
		enablement.TenantID,
		enablement.FeatureKey,
		enablement.Enabled,
		configJSON,
		enablement.EnabledBy,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to create tenant feature enablement: %w", err)
	}

	return nil
}

// Update updates an existing tenant feature enablement
func (r *tenantFeatureEnablementRepository) Update(ctx context.Context, enablement *models.TenantFeatureEnablement) error {
	query := `
		UPDATE tenant_feature_enablement
		SET enabled = $3, configuration = $4, enabled_by = $5, enabled_at = $6
		WHERE tenant_id = $1 AND feature_key = $2
	`

	configJSON := "{}"
	if len(enablement.Configuration) > 0 {
		configJSON = string(enablement.Configuration)
	}

	_, err := r.db.ExecContext(ctx, query,
		enablement.TenantID,
		enablement.FeatureKey,
		enablement.Enabled,
		configJSON,
		enablement.EnabledBy,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update tenant feature enablement: %w", err)
	}

	return nil
}

// Delete deletes a tenant feature enablement
func (r *tenantFeatureEnablementRepository) Delete(ctx context.Context, tenantID uuid.UUID, key string) error {
	query := `DELETE FROM tenant_feature_enablement WHERE tenant_id = $1 AND feature_key = $2`

	_, err := r.db.ExecContext(ctx, query, tenantID, key)
	if err != nil {
		return fmt.Errorf("failed to delete tenant feature enablement: %w", err)
	}

	return nil
}

// DeleteByTenantID deletes all feature enablements for a tenant
func (r *tenantFeatureEnablementRepository) DeleteByTenantID(ctx context.Context, tenantID uuid.UUID) error {
	query := `DELETE FROM tenant_feature_enablement WHERE tenant_id = $1`

	_, err := r.db.ExecContext(ctx, query, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete tenant feature enablements: %w", err)
	}

	return nil
}

