package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/identity/models"
	"github.com/nuage-identity/iam/storage/interfaces"
)

// permissionRepository implements PermissionRepository for PostgreSQL
type permissionRepository struct {
	db *sql.DB
}

// NewPermissionRepository creates a new PostgreSQL permission repository
func NewPermissionRepository(db *sql.DB) interfaces.PermissionRepository {
	return &permissionRepository{db: db}
}

// Create creates a new permission
func (r *permissionRepository) Create(ctx context.Context, permission *models.Permission) error {
	query := `
		INSERT INTO permissions (id, tenant_id, name, description, resource, action, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	now := time.Now()
	if permission.ID == uuid.Nil {
		permission.ID = uuid.New()
	}
	if permission.CreatedAt.IsZero() {
		permission.CreatedAt = now
	}
	if permission.UpdatedAt.IsZero() {
		permission.UpdatedAt = now
	}

	_, err := r.db.ExecContext(ctx, query,
		permission.ID, permission.TenantID, permission.Name, permission.Description,
		permission.Resource, permission.Action, permission.CreatedAt, permission.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create permission: %w", err)
	}

	return nil
}

// GetByID retrieves a permission by ID
func (r *permissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	query := `
		SELECT id, tenant_id, name, description, resource, action, created_at, updated_at, deleted_at
		FROM permissions
		WHERE id = $1 AND deleted_at IS NULL
	`

	permission := &models.Permission{}
	var description sql.NullString
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&permission.ID, &permission.TenantID, &permission.Name, &description,
		&permission.Resource, &permission.Action, &permission.CreatedAt,
		&permission.UpdatedAt, &deletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("permission not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	if description.Valid {
		permission.Description = &description.String
	}
	if deletedAt.Valid {
		permission.DeletedAt = &deletedAt.Time
	}

	return permission, nil
}

// GetByName retrieves a permission by name and tenant ID
func (r *permissionRepository) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Permission, error) {
	query := `
		SELECT id, tenant_id, name, description, resource, action, created_at, updated_at, deleted_at
		FROM permissions
		WHERE tenant_id = $1 AND name = $2 AND deleted_at IS NULL
	`

	permission := &models.Permission{}
	var description sql.NullString
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, tenantID, name).Scan(
		&permission.ID, &permission.TenantID, &permission.Name, &description,
		&permission.Resource, &permission.Action, &permission.CreatedAt,
		&permission.UpdatedAt, &deletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("permission not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get permission by name: %w", err)
	}

	if description.Valid {
		permission.Description = &description.String
	}
	if deletedAt.Valid {
		permission.DeletedAt = &deletedAt.Time
	}

	return permission, nil
}

// Update updates an existing permission
func (r *permissionRepository) Update(ctx context.Context, permission *models.Permission) error {
	query := `
		UPDATE permissions
		SET name = $2, description = $3, resource = $4, action = $5, updated_at = $6
		WHERE id = $1 AND deleted_at IS NULL
	`

	permission.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		permission.ID, permission.Name, permission.Description,
		permission.Resource, permission.Action, permission.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update permission: %w", err)
	}

	return nil
}

// Delete soft deletes a permission
func (r *permissionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE permissions
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("permission not found")
	}

	return nil
}

// List retrieves a list of permissions with filters
func (r *permissionRepository) List(ctx context.Context, tenantID uuid.UUID, filters *interfaces.PermissionFilters) ([]*models.Permission, error) {
	if filters == nil {
		filters = &interfaces.PermissionFilters{
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
		SELECT id, tenant_id, name, description, resource, action, created_at, updated_at, deleted_at
		FROM permissions
		WHERE tenant_id = $1 AND deleted_at IS NULL
	`
	args := []interface{}{tenantID}
	argPos := 2

	if filters.Resource != nil {
		query += fmt.Sprintf(" AND resource = $%d", argPos)
		args = append(args, *filters.Resource)
		argPos++
	}

	if filters.Action != nil {
		query += fmt.Sprintf(" AND action = $%d", argPos)
		args = append(args, *filters.Action)
		argPos++
	}

	if filters.Search != nil {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d OR resource ILIKE $%d)", argPos, argPos, argPos)
		searchPattern := "%" + *filters.Search + "%"
		args = append(args, searchPattern)
		argPos++
	}

	query += " ORDER BY resource, action LIMIT $" + fmt.Sprintf("%d", argPos) + " OFFSET $" + fmt.Sprintf("%d", argPos+1)
	args = append(args, filters.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}
	defer rows.Close()

	var permissions []*models.Permission
	for rows.Next() {
		permission := &models.Permission{}
		var description sql.NullString
		var deletedAt sql.NullTime

		err := rows.Scan(
			&permission.ID, &permission.TenantID, &permission.Name, &description,
			&permission.Resource, &permission.Action, &permission.CreatedAt,
			&permission.UpdatedAt, &deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}

		if description.Valid {
			permission.Description = &description.String
		}
		if deletedAt.Valid {
			permission.DeletedAt = &deletedAt.Time
		}

		permissions = append(permissions, permission)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating permissions: %w", err)
	}

	return permissions, nil
}

// GetRolePermissions retrieves all permissions for a role
func (r *permissionRepository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*models.Permission, error) {
	query := `
		SELECT p.id, p.tenant_id, p.name, p.description, p.resource, p.action, p.created_at, p.updated_at, p.deleted_at
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1 AND p.deleted_at IS NULL
	`

	rows, err := r.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}
	defer rows.Close()

	var permissions []*models.Permission
	for rows.Next() {
		permission := &models.Permission{}
		var description sql.NullString
		var deletedAt sql.NullTime

		err := rows.Scan(
			&permission.ID, &permission.TenantID, &permission.Name, &description,
			&permission.Resource, &permission.Action, &permission.CreatedAt,
			&permission.UpdatedAt, &deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}

		if description.Valid {
			permission.Description = &description.String
		}
		if deletedAt.Valid {
			permission.DeletedAt = &deletedAt.Time
		}

		permissions = append(permissions, permission)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating permissions: %w", err)
	}

	return permissions, nil
}

// AssignPermissionToRole assigns a permission to a role
func (r *permissionRepository) AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	query := `
		INSERT INTO role_permissions (id, role_id, permission_id, created_at)
		VALUES (gen_random_uuid(), $1, $2, NOW())
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to assign permission to role: %w", err)
	}

	return nil
}

// RemovePermissionFromRole removes a permission from a role
func (r *permissionRepository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	query := `DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2`

	result, err := r.db.ExecContext(ctx, query, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to remove permission from role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("permission assignment not found")
	}

	return nil
}

