package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/identity/models"
	"github.com/nuage-identity/iam/storage/interfaces"
)

// tenantRepository implements TenantRepository for PostgreSQL
type tenantRepository struct {
	db *sql.DB
}

// NewTenantRepository creates a new PostgreSQL tenant repository
func NewTenantRepository(db *sql.DB) interfaces.TenantRepository {
	return &tenantRepository{db: db}
}

// Create creates a new tenant
func (r *tenantRepository) Create(ctx context.Context, tenant *models.Tenant) error {
	query := `
		INSERT INTO tenants (id, name, domain, status, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	now := time.Now()
	if tenant.ID == uuid.Nil {
		tenant.ID = uuid.New()
	}
	if tenant.CreatedAt.IsZero() {
		tenant.CreatedAt = now
	}
	if tenant.UpdatedAt.IsZero() {
		tenant.UpdatedAt = now
	}
	if tenant.Status == "" {
		tenant.Status = models.TenantStatusActive
	}

	var metadataJSON interface{}
	if tenant.Metadata != nil {
		metadataJSON, _ = json.Marshal(tenant.Metadata)
	} else {
		metadataJSON = nil
	}

	_, err := r.db.ExecContext(ctx, query,
		tenant.ID, tenant.Name, tenant.Domain, tenant.Status,
		metadataJSON, tenant.CreatedAt, tenant.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	return nil
}

// GetByID retrieves a tenant by ID
func (r *tenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	query := `
		SELECT id, name, domain, status, metadata, created_at, updated_at, deleted_at
		FROM tenants
		WHERE id = $1 AND deleted_at IS NULL
	`

	tenant := &models.Tenant{}
	var metadataJSON []byte
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tenant.ID, &tenant.Name, &tenant.Domain, &tenant.Status,
		&metadataJSON, &tenant.CreatedAt, &tenant.UpdatedAt, &deletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	if len(metadataJSON) > 0 {
		_ = json.Unmarshal(metadataJSON, &tenant.Metadata) // Ignore unmarshal errors for optional metadata
	}
	if deletedAt.Valid {
		tenant.DeletedAt = &deletedAt.Time
	}

	return tenant, nil
}

// GetByDomain retrieves a tenant by domain
func (r *tenantRepository) GetByDomain(ctx context.Context, domain string) (*models.Tenant, error) {
	query := `
		SELECT id, name, domain, status, metadata, created_at, updated_at, deleted_at
		FROM tenants
		WHERE domain = $1 AND deleted_at IS NULL
	`

	tenant := &models.Tenant{}
	var metadataJSON []byte
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, domain).Scan(
		&tenant.ID, &tenant.Name, &tenant.Domain, &tenant.Status,
		&metadataJSON, &tenant.CreatedAt, &tenant.UpdatedAt, &deletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant by domain: %w", err)
	}

	if len(metadataJSON) > 0 {
		_ = json.Unmarshal(metadataJSON, &tenant.Metadata) // Ignore unmarshal errors for optional metadata
	}
	if deletedAt.Valid {
		tenant.DeletedAt = &deletedAt.Time
	}

	return tenant, nil
}

// Update updates an existing tenant
func (r *tenantRepository) Update(ctx context.Context, tenant *models.Tenant) error {
	query := `
		UPDATE tenants
		SET name = $2, domain = $3, status = $4, metadata = $5, updated_at = $6
		WHERE id = $1 AND deleted_at IS NULL
	`

	tenant.UpdatedAt = time.Now()

	var metadataJSON []byte
	if tenant.Metadata != nil {
		metadataJSON, _ = json.Marshal(tenant.Metadata)
	}

	_, err := r.db.ExecContext(ctx, query,
		tenant.ID, tenant.Name, tenant.Domain, tenant.Status,
		metadataJSON, tenant.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	return nil
}

// Delete soft deletes a tenant
func (r *tenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE tenants
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tenant not found")
	}

	return nil
}

// List retrieves a list of tenants
func (r *tenantRepository) List(ctx context.Context, filters *interfaces.TenantFilters) ([]*models.Tenant, error) {
	if filters == nil {
		filters = &interfaces.TenantFilters{
			Page:     1,
			PageSize: 20,
		}
	}

	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 || filters.PageSize > 100 {
		filters.PageSize = 20
	}

	offset := (filters.Page - 1) * filters.PageSize

	query := `
		SELECT id, name, domain, status, metadata, created_at, updated_at, deleted_at
		FROM tenants
		WHERE deleted_at IS NULL
	`
	args := []interface{}{}
	argPos := 1

	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, *filters.Status)
		argPos++
	}

	if filters.Search != nil {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR domain ILIKE $%d)", argPos, argPos)
		searchPattern := "%" + *filters.Search + "%"
		args = append(args, searchPattern)
		argPos++
	}

	query += " ORDER BY created_at DESC LIMIT $" + fmt.Sprintf("%d", argPos) + " OFFSET $" + fmt.Sprintf("%d", argPos+1)
	args = append(args, filters.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*models.Tenant
	for rows.Next() {
		tenant := &models.Tenant{}
		var metadataJSON []byte
		var deletedAt sql.NullTime

		err := rows.Scan(
			&tenant.ID, &tenant.Name, &tenant.Domain, &tenant.Status,
			&metadataJSON, &tenant.CreatedAt, &tenant.UpdatedAt, &deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant: %w", err)
		}

		if len(metadataJSON) > 0 {
			_ = json.Unmarshal(metadataJSON, &tenant.Metadata) // Ignore unmarshal errors for optional metadata
		}
		if deletedAt.Valid {
			tenant.DeletedAt = &deletedAt.Time
		}

		tenants = append(tenants, tenant)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tenants: %w", err)
	}

	return tenants, nil
}

