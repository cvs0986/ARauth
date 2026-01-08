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

// systemCapabilityRepository implements SystemCapabilityRepository for PostgreSQL
type systemCapabilityRepository struct {
	db *sql.DB
}

// NewSystemCapabilityRepository creates a new PostgreSQL system capability repository
func NewSystemCapabilityRepository(db *sql.DB) interfaces.SystemCapabilityRepository {
	return &systemCapabilityRepository{db: db}
}

// GetByKey retrieves a system capability by key
func (r *systemCapabilityRepository) GetByKey(ctx context.Context, key string) (*models.SystemCapability, error) {
	query := `
		SELECT capability_key, enabled, default_value, description, updated_by, updated_at
		FROM system_capabilities
		WHERE capability_key = $1
	`

	capability := &models.SystemCapability{}
	var defaultValue sql.NullString
	var description sql.NullString
	var updatedBy sql.NullString

	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&capability.CapabilityKey,
		&capability.Enabled,
		&defaultValue,
		&description,
		&updatedBy,
		&capability.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("system capability not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get system capability: %w", err)
	}

	if defaultValue.Valid {
		capability.DefaultValue = json.RawMessage(defaultValue.String)
	}
	if description.Valid {
		capability.Description = &description.String
	}
	if updatedBy.Valid {
		id, err := uuid.Parse(updatedBy.String)
		if err == nil {
			capability.UpdatedBy = &id
		}
	}

	return capability, nil
}

// GetAll retrieves all system capabilities
func (r *systemCapabilityRepository) GetAll(ctx context.Context) ([]*models.SystemCapability, error) {
	query := `
		SELECT capability_key, enabled, default_value, description, updated_by, updated_at
		FROM system_capabilities
		ORDER BY capability_key
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get system capabilities: %w", err)
	}
	defer rows.Close()

	var capabilities []*models.SystemCapability
	for rows.Next() {
		capability := &models.SystemCapability{}
		var defaultValue sql.NullString
		var description sql.NullString
		var updatedBy sql.NullString

		err := rows.Scan(
			&capability.CapabilityKey,
			&capability.Enabled,
			&defaultValue,
			&description,
			&updatedBy,
			&capability.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan system capability: %w", err)
		}

		if defaultValue.Valid {
			capability.DefaultValue = json.RawMessage(defaultValue.String)
		}
		if description.Valid {
			capability.Description = &description.String
		}
		if updatedBy.Valid {
			id, err := uuid.Parse(updatedBy.String)
			if err == nil {
				capability.UpdatedBy = &id
			}
		}

		capabilities = append(capabilities, capability)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating system capabilities: %w", err)
	}

	return capabilities, nil
}

// GetEnabled retrieves all enabled system capabilities
func (r *systemCapabilityRepository) GetEnabled(ctx context.Context) ([]*models.SystemCapability, error) {
	query := `
		SELECT capability_key, enabled, default_value, description, updated_by, updated_at
		FROM system_capabilities
		WHERE enabled = true
		ORDER BY capability_key
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled system capabilities: %w", err)
	}
	defer rows.Close()

	var capabilities []*models.SystemCapability
	for rows.Next() {
		capability := &models.SystemCapability{}
		var defaultValue sql.NullString
		var description sql.NullString
		var updatedBy sql.NullString

		err := rows.Scan(
			&capability.CapabilityKey,
			&capability.Enabled,
			&defaultValue,
			&description,
			&updatedBy,
			&capability.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan system capability: %w", err)
		}

		if defaultValue.Valid {
			capability.DefaultValue = json.RawMessage(defaultValue.String)
		}
		if description.Valid {
			capability.Description = &description.String
		}
		if updatedBy.Valid {
			id, err := uuid.Parse(updatedBy.String)
			if err == nil {
				capability.UpdatedBy = &id
			}
		}

		capabilities = append(capabilities, capability)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating enabled system capabilities: %w", err)
	}

	return capabilities, nil
}

// Create creates a new system capability
func (r *systemCapabilityRepository) Create(ctx context.Context, capability *models.SystemCapability) error {
	query := `
		INSERT INTO system_capabilities (capability_key, enabled, default_value, description, updated_by, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	defaultValueJSON := "{}"
	if len(capability.DefaultValue) > 0 {
		defaultValueJSON = string(capability.DefaultValue)
	}

	_, err := r.db.ExecContext(ctx, query,
		capability.CapabilityKey,
		capability.Enabled,
		defaultValueJSON,
		capability.Description,
		capability.UpdatedBy,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to create system capability: %w", err)
	}

	return nil
}

// Update updates an existing system capability
func (r *systemCapabilityRepository) Update(ctx context.Context, capability *models.SystemCapability) error {
	query := `
		UPDATE system_capabilities
		SET enabled = $2, default_value = $3, description = $4, updated_by = $5, updated_at = $6
		WHERE capability_key = $1
	`

	defaultValueJSON := "{}"
	if len(capability.DefaultValue) > 0 {
		defaultValueJSON = string(capability.DefaultValue)
	}

	_, err := r.db.ExecContext(ctx, query,
		capability.CapabilityKey,
		capability.Enabled,
		defaultValueJSON,
		capability.Description,
		capability.UpdatedBy,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update system capability: %w", err)
	}

	return nil
}

// Delete deletes a system capability
func (r *systemCapabilityRepository) Delete(ctx context.Context, key string) error {
	query := `DELETE FROM system_capabilities WHERE capability_key = $1`

	_, err := r.db.ExecContext(ctx, query, key)
	if err != nil {
		return fmt.Errorf("failed to delete system capability: %w", err)
	}

	return nil
}

