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
// Note: permissions table doesn't have tenant_id or updated_at columns - permissions are global
func (r *permissionRepository) Create(ctx context.Context, permission *models.Permission) error {
	query := `
		INSERT INTO permissions (id, name, description, resource, action, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	now := time.Now()
	if permission.ID == uuid.Nil {
		permission.ID = uuid.New()
	}
	if permission.CreatedAt.IsZero() {
		permission.CreatedAt = now
	}

	_, err := r.db.ExecContext(ctx, query,
		permission.ID, permission.Name, permission.Description,
		permission.Resource, permission.Action, permission.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create permission: %w", err)
	}

	return nil
}

// GetByID retrieves a permission by ID
// Note: permissions table doesn't have tenant_id, updated_at, or deleted_at columns - permissions are global
func (r *permissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	query := `
		SELECT id, name, description, resource, action, created_at
		FROM permissions
		WHERE id = $1
	`

	permission := &models.Permission{}
	var description sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&permission.ID, &permission.Name, &description,
		&permission.Resource, &permission.Action, &permission.CreatedAt,
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

	return permission, nil
}

// GetByName retrieves a permission by name
// Note: permissions table doesn't have tenant_id, updated_at, or deleted_at columns - permissions are global
// tenantID parameter is kept for interface compatibility but not used
func (r *permissionRepository) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Permission, error) {
	query := `
		SELECT id, name, description, resource, action, created_at
		FROM permissions
		WHERE name = $1
	`

	permission := &models.Permission{}
	var description sql.NullString

	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&permission.ID, &permission.Name, &description,
		&permission.Resource, &permission.Action, &permission.CreatedAt,
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

	return permission, nil
}

// Update updates an existing permission
// Note: permissions table doesn't have updated_at or deleted_at columns
func (r *permissionRepository) Update(ctx context.Context, permission *models.Permission) error {
	query := `
		UPDATE permissions
		SET name = $2, description = $3, resource = $4, action = $5
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		permission.ID, permission.Name, permission.Description,
		permission.Resource, permission.Action,
	)

	if err != nil {
		return fmt.Errorf("failed to update permission: %w", err)
	}

	return nil
}

// Delete deletes a permission (hard delete - permissions table doesn't have deleted_at)
func (r *permissionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM permissions WHERE id = $1`

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
// Note: permissions table doesn't have tenant_id column - permissions are global
// tenantID parameter is kept for interface compatibility but not used
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
		SELECT id, name, description, resource, action, created_at
		FROM permissions
	`
	args := []interface{}{}
	argPos := 1

	if filters.Resource != nil {
		if len(args) == 0 {
			query += " WHERE resource = $1"
		} else {
			query += fmt.Sprintf(" AND resource = $%d", argPos)
		}
		args = append(args, *filters.Resource)
		argPos++
	}

	if filters.Action != nil {
		if len(args) == 0 {
			query += " WHERE action = $1"
		} else {
			query += fmt.Sprintf(" AND action = $%d", argPos)
		}
		args = append(args, *filters.Action)
		argPos++
	}

	if filters.Search != nil {
		if len(args) == 0 {
			query += " WHERE (name ILIKE $1 OR description ILIKE $1 OR resource ILIKE $1)"
		} else {
			query += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d OR resource ILIKE $%d)", argPos, argPos, argPos)
		}
		searchPattern := "%" + *filters.Search + "%"
		args = append(args, searchPattern)
		argPos++
	}

	query += fmt.Sprintf(" ORDER BY resource, action LIMIT $%d OFFSET $%d", argPos, argPos+1)
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

		err := rows.Scan(
			&permission.ID, &permission.Name, &description,
			&permission.Resource, &permission.Action, &permission.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}

		if description.Valid {
			permission.Description = &description.String
		}

		permissions = append(permissions, permission)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating permissions: %w", err)
	}

	return permissions, nil
}

// GetRolePermissions retrieves all permissions for a role
// Note: permissions table doesn't have tenant_id, updated_at, or deleted_at columns - permissions are global
func (r *permissionRepository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*models.Permission, error) {
	query := `
		SELECT p.id, p.name, p.description, p.resource, p.action, p.created_at
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
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

		err := rows.Scan(
			&permission.ID, &permission.Name, &description,
			&permission.Resource, &permission.Action, &permission.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}

		if description.Valid {
			permission.Description = &description.String
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

