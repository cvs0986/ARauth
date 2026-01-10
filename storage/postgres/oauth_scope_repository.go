package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// OAuthScopeRepository implements OAuthScopeRepository interface
type OAuthScopeRepository struct {
	db *sql.DB
}

// NewOAuthScopeRepository creates a new OAuth scope repository
func NewOAuthScopeRepository(db *sql.DB) interfaces.OAuthScopeRepository {
	return &OAuthScopeRepository{db: db}
}

// Create creates a new OAuth scope
func (r *OAuthScopeRepository) Create(ctx context.Context, scope *models.OAuthScope) error {
	query := `
		INSERT INTO oauth_scopes (
			id, tenant_id, name, description, permissions, is_default,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	now := time.Now()
	if scope.ID == uuid.Nil {
		scope.ID = uuid.New()
	}
	if scope.CreatedAt.IsZero() {
		scope.CreatedAt = now
	}
	if scope.UpdatedAt.IsZero() {
		scope.UpdatedAt = now
	}

	_, err := r.db.ExecContext(ctx, query,
		scope.ID,
		scope.TenantID,
		scope.Name,
		scope.Description,
		pq.Array(scope.Permissions),
		scope.IsDefault,
		scope.CreatedAt,
		scope.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create OAuth scope: %w", err)
	}

	return nil
}

// GetByID retrieves an OAuth scope by ID
func (r *OAuthScopeRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.OAuthScope, error) {
	query := `
		SELECT id, tenant_id, name, description, permissions, is_default,
		       created_at, updated_at, deleted_at
		FROM oauth_scopes
		WHERE id = $1 AND deleted_at IS NULL
	`

	scope := &models.OAuthScope{}
	var permissions pq.StringArray
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&scope.ID,
		&scope.TenantID,
		&scope.Name,
		&scope.Description,
		&permissions,
		&scope.IsDefault,
		&scope.CreatedAt,
		&scope.UpdatedAt,
		&deletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("OAuth scope not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth scope: %w", err)
	}

	scope.Permissions = []string(permissions)
	if deletedAt.Valid {
		scope.DeletedAt = &deletedAt.Time
	}

	return scope, nil
}

// GetByName retrieves an OAuth scope by tenant ID and name
func (r *OAuthScopeRepository) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.OAuthScope, error) {
	query := `
		SELECT id, tenant_id, name, description, permissions, is_default,
		       created_at, updated_at, deleted_at
		FROM oauth_scopes
		WHERE tenant_id = $1 AND name = $2 AND deleted_at IS NULL
	`

	scope := &models.OAuthScope{}
	var permissions pq.StringArray
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, tenantID, name).Scan(
		&scope.ID,
		&scope.TenantID,
		&scope.Name,
		&scope.Description,
		&permissions,
		&scope.IsDefault,
		&scope.CreatedAt,
		&scope.UpdatedAt,
		&deletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("OAuth scope not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth scope: %w", err)
	}

	scope.Permissions = []string(permissions)
	if deletedAt.Valid {
		scope.DeletedAt = &deletedAt.Time
	}

	return scope, nil
}

// List lists OAuth scopes for a tenant
func (r *OAuthScopeRepository) List(ctx context.Context, tenantID uuid.UUID, filters *interfaces.OAuthScopeFilters) ([]*models.OAuthScope, error) {
	query := `
		SELECT id, tenant_id, name, description, permissions, is_default,
		       created_at, updated_at, deleted_at
		FROM oauth_scopes
		WHERE tenant_id = $1 AND deleted_at IS NULL
	`
	args := []interface{}{tenantID}
	argPos := 2

	if filters != nil && filters.IsDefault != nil {
		query += fmt.Sprintf(" AND is_default = $%d", argPos)
		args = append(args, *filters.IsDefault)
		argPos++
	}

	query += " ORDER BY name ASC"

	if filters != nil && filters.PageSize > 0 {
		offset := (filters.Page - 1) * filters.PageSize
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
		args = append(args, filters.PageSize, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query OAuth scopes: %w", err)
	}
	defer rows.Close()

	var scopes []*models.OAuthScope
	for rows.Next() {
		scope := &models.OAuthScope{}
		var permissions pq.StringArray
		var deletedAt sql.NullTime

		err := rows.Scan(
			&scope.ID,
			&scope.TenantID,
			&scope.Name,
			&scope.Description,
			&permissions,
			&scope.IsDefault,
			&scope.CreatedAt,
			&scope.UpdatedAt,
			&deletedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan OAuth scope: %w", err)
		}

		scope.Permissions = []string(permissions)
		if deletedAt.Valid {
			scope.DeletedAt = &deletedAt.Time
		}

		scopes = append(scopes, scope)
	}

	return scopes, nil
}

// Update updates an OAuth scope
func (r *OAuthScopeRepository) Update(ctx context.Context, scope *models.OAuthScope) error {
	query := `
		UPDATE oauth_scopes
		SET name = $1, description = $2, permissions = $3, is_default = $4,
		    updated_at = $5
		WHERE id = $6 AND deleted_at IS NULL
	`

	scope.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		scope.Name,
		scope.Description,
		pq.Array(scope.Permissions),
		scope.IsDefault,
		scope.UpdatedAt,
		scope.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update OAuth scope: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("OAuth scope not found or already deleted")
	}

	return nil
}

// Delete soft-deletes an OAuth scope
func (r *OAuthScopeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE oauth_scopes
		SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete OAuth scope: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("OAuth scope not found or already deleted")
	}

	return nil
}

// GetDefaultScopes retrieves all default scopes for a tenant
func (r *OAuthScopeRepository) GetDefaultScopes(ctx context.Context, tenantID uuid.UUID) ([]*models.OAuthScope, error) {
	query := `
		SELECT id, tenant_id, name, description, permissions, is_default,
		       created_at, updated_at, deleted_at
		FROM oauth_scopes
		WHERE tenant_id = $1 AND is_default = true AND deleted_at IS NULL
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query default OAuth scopes: %w", err)
	}
	defer rows.Close()

	var scopes []*models.OAuthScope
	for rows.Next() {
		scope := &models.OAuthScope{}
		var permissions pq.StringArray
		var deletedAt sql.NullTime

		err := rows.Scan(
			&scope.ID,
			&scope.TenantID,
			&scope.Name,
			&scope.Description,
			&permissions,
			&scope.IsDefault,
			&scope.CreatedAt,
			&scope.UpdatedAt,
			&deletedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan OAuth scope: %w", err)
		}

		scope.Permissions = []string(permissions)
		if deletedAt.Valid {
			scope.DeletedAt = &deletedAt.Time
		}

		scopes = append(scopes, scope)
	}

	return scopes, nil
}

// GetScopesByPermissions retrieves all scopes that include any of the given permissions
func (r *OAuthScopeRepository) GetScopesByPermissions(ctx context.Context, tenantID uuid.UUID, permissions []string) ([]*models.OAuthScope, error) {
	if len(permissions) == 0 {
		return []*models.OAuthScope{}, nil
	}

	// Build query to find scopes where permissions array overlaps with the given permissions
	query := `
		SELECT DISTINCT id, tenant_id, name, description, permissions, is_default,
		       created_at, updated_at, deleted_at
		FROM oauth_scopes
		WHERE tenant_id = $1 
		  AND deleted_at IS NULL
		  AND permissions && $2
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, pq.Array(permissions))
	if err != nil {
		return nil, fmt.Errorf("failed to query OAuth scopes by permissions: %w", err)
	}
	defer rows.Close()

	var scopes []*models.OAuthScope
	for rows.Next() {
		scope := &models.OAuthScope{}
		var permissions pq.StringArray
		var deletedAt sql.NullTime

		err := rows.Scan(
			&scope.ID,
			&scope.TenantID,
			&scope.Name,
			&scope.Description,
			&permissions,
			&scope.IsDefault,
			&scope.CreatedAt,
			&scope.UpdatedAt,
			&deletedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan OAuth scope: %w", err)
		}

		scope.Permissions = []string(permissions)
		if deletedAt.Valid {
			scope.DeletedAt = &deletedAt.Time
		}

		scopes = append(scopes, scope)
	}

	return scopes, nil
}

