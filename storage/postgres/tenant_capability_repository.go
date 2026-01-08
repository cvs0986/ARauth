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

// tenantCapabilityRepository implements TenantCapabilityRepository for PostgreSQL
type tenantCapabilityRepository struct {
	db *sql.DB
}

// NewTenantCapabilityRepository creates a new PostgreSQL tenant capability repository
func NewTenantCapabilityRepository(db *sql.DB) interfaces.TenantCapabilityRepository {
	return &tenantCapabilityRepository{db: db}
}

// GetByTenantIDAndKey retrieves a tenant capability by tenant ID and key
func (r *tenantCapabilityRepository) GetByTenantIDAndKey(ctx context.Context, tenantID uuid.UUID, key string) (*models.TenantCapability, error) {
	query := `
		SELECT tenant_id, capability_key, enabled, value, configured_by, configured_at
		FROM tenant_capabilities
		WHERE tenant_id = $1 AND capability_key = $2
	`

	capability := &models.TenantCapability{}
	var value sql.NullString
	var configuredBy sql.NullString

	err := r.db.QueryRowContext(ctx, query, tenantID, key).Scan(
		&capability.TenantID,
		&capability.CapabilityKey,
		&capability.Enabled,
		&value,
		&configuredBy,
		&capability.ConfiguredAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant capability not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant capability: %w", err)
	}

	if value.Valid {
		capability.Value = json.RawMessage(value.String)
	}
	if configuredBy.Valid {
		id, err := uuid.Parse(configuredBy.String)
		if err == nil {
			capability.ConfiguredBy = &id
		}
	}

	return capability, nil
}

// GetByTenantID retrieves all capabilities for a tenant
func (r *tenantCapabilityRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*models.TenantCapability, error) {
	query := `
		SELECT tenant_id, capability_key, enabled, value, configured_by, configured_at
		FROM tenant_capabilities
		WHERE tenant_id = $1
		ORDER BY capability_key
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant capabilities: %w", err)
	}
	defer rows.Close()

	var capabilities []*models.TenantCapability
	for rows.Next() {
		capability := &models.TenantCapability{}
		var value sql.NullString
		var configuredBy sql.NullString

		err := rows.Scan(
			&capability.TenantID,
			&capability.CapabilityKey,
			&capability.Enabled,
			&value,
			&configuredBy,
			&capability.ConfiguredAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant capability: %w", err)
		}

		if value.Valid {
			capability.Value = json.RawMessage(value.String)
		}
		if configuredBy.Valid {
			id, err := uuid.Parse(configuredBy.String)
			if err == nil {
				capability.ConfiguredBy = &id
			}
		}

		capabilities = append(capabilities, capability)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tenant capabilities: %w", err)
	}

	return capabilities, nil
}

// GetEnabledByTenantID retrieves all enabled capabilities for a tenant
func (r *tenantCapabilityRepository) GetEnabledByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*models.TenantCapability, error) {
	query := `
		SELECT tenant_id, capability_key, enabled, value, configured_by, configured_at
		FROM tenant_capabilities
		WHERE tenant_id = $1 AND enabled = true
		ORDER BY capability_key
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled tenant capabilities: %w", err)
	}
	defer rows.Close()

	var capabilities []*models.TenantCapability
	for rows.Next() {
		capability := &models.TenantCapability{}
		var value sql.NullString
		var configuredBy sql.NullString

		err := rows.Scan(
			&capability.TenantID,
			&capability.CapabilityKey,
			&capability.Enabled,
			&value,
			&configuredBy,
			&capability.ConfiguredAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant capability: %w", err)
		}

		if value.Valid {
			capability.Value = json.RawMessage(value.String)
		}
		if configuredBy.Valid {
			id, err := uuid.Parse(configuredBy.String)
			if err == nil {
				capability.ConfiguredBy = &id
			}
		}

		capabilities = append(capabilities, capability)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating enabled tenant capabilities: %w", err)
	}

	return capabilities, nil
}

// Create creates a new tenant capability
func (r *tenantCapabilityRepository) Create(ctx context.Context, capability *models.TenantCapability) error {
	query := `
		INSERT INTO tenant_capabilities (tenant_id, capability_key, enabled, value, configured_by, configured_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	valueJSON := "{}"
	if len(capability.Value) > 0 {
		valueJSON = string(capability.Value)
	}

	_, err := r.db.ExecContext(ctx, query,
		capability.TenantID,
		capability.CapabilityKey,
		capability.Enabled,
		valueJSON,
		capability.ConfiguredBy,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to create tenant capability: %w", err)
	}

	return nil
}

// Update updates an existing tenant capability
func (r *tenantCapabilityRepository) Update(ctx context.Context, capability *models.TenantCapability) error {
	query := `
		UPDATE tenant_capabilities
		SET enabled = $3, value = $4, configured_by = $5, configured_at = $6
		WHERE tenant_id = $1 AND capability_key = $2
	`

	valueJSON := "{}"
	if len(capability.Value) > 0 {
		valueJSON = string(capability.Value)
	}

	_, err := r.db.ExecContext(ctx, query,
		capability.TenantID,
		capability.CapabilityKey,
		capability.Enabled,
		valueJSON,
		capability.ConfiguredBy,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update tenant capability: %w", err)
	}

	return nil
}

// Delete deletes a tenant capability
func (r *tenantCapabilityRepository) Delete(ctx context.Context, tenantID uuid.UUID, key string) error {
	query := `DELETE FROM tenant_capabilities WHERE tenant_id = $1 AND capability_key = $2`

	_, err := r.db.ExecContext(ctx, query, tenantID, key)
	if err != nil {
		return fmt.Errorf("failed to delete tenant capability: %w", err)
	}

	return nil
}

// DeleteByTenantID deletes all capabilities for a tenant
func (r *tenantCapabilityRepository) DeleteByTenantID(ctx context.Context, tenantID uuid.UUID) error {
	query := `DELETE FROM tenant_capabilities WHERE tenant_id = $1`

	_, err := r.db.ExecContext(ctx, query, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete tenant capabilities: %w", err)
	}

	return nil
}

