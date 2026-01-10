package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
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
// Supports tenant-scoped permissions (tenant_id can be NULL for backward compatibility)
func (r *permissionRepository) Create(ctx context.Context, permission *models.Permission) error {
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

	// Check if tenant_id column exists (for migration compatibility)
	// If tenant_id is set, include it in the insert
	var query string
	var err error
	if permission.TenantID != uuid.Nil {
		query = `
			INSERT INTO permissions (id, tenant_id, name, description, resource, action, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		_, err = r.db.ExecContext(ctx, query,
			permission.ID, permission.TenantID, permission.Name, permission.Description,
			permission.Resource, permission.Action, permission.CreatedAt, permission.UpdatedAt,
		)
	} else {
		// Fallback for backward compatibility (if tenant_id column doesn't exist yet)
		query = `
			INSERT INTO permissions (id, name, description, resource, action, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err = r.db.ExecContext(ctx, query,
			permission.ID, permission.Name, permission.Description,
			permission.Resource, permission.Action, permission.CreatedAt,
		)
	}

	if err != nil {
		return fmt.Errorf("failed to create permission: %w", err)
	}

	return nil
}

// GetByID retrieves a permission by ID
// Supports tenant-scoped permissions (tenant_id can be NULL for backward compatibility)
func (r *permissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	// Try with tenant_id column first (for new schema)
	query := `
		SELECT id, COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid) as tenant_id, 
		       name, description, resource, action, created_at, 
		       COALESCE(updated_at, created_at) as updated_at, deleted_at
		FROM permissions
		WHERE id = $1
	`

	permission := &models.Permission{}
	var description sql.NullString
	var deletedAt sql.NullTime
	var updatedAt sql.NullTime
	var tenantIDStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&permission.ID, &tenantIDStr, &permission.Name, &description,
		&permission.Resource, &permission.Action, &permission.CreatedAt, &updatedAt, &deletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("permission not found: %w", err)
	}
	if err != nil {
		// Fallback: try without tenant_id column (for old schema)
		query = `
			SELECT id, name, description, resource, action, created_at
			FROM permissions
			WHERE id = $1
		`
		err = r.db.QueryRowContext(ctx, query, id).Scan(
			&permission.ID, &permission.Name, &description,
			&permission.Resource, &permission.Action, &permission.CreatedAt,
		)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("permission not found: %w", err)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get permission: %w", err)
		}
		permission.TenantID = uuid.Nil // Global permission
	} else {
		// Parse tenant_id if not nil
		if tenantIDStr != "00000000-0000-0000-0000-000000000000" {
			tenantID, parseErr := uuid.Parse(tenantIDStr)
			if parseErr == nil {
				permission.TenantID = tenantID
			}
		}
		if updatedAt.Valid {
			permission.UpdatedAt = updatedAt.Time
		} else {
			permission.UpdatedAt = permission.CreatedAt
		}
		if deletedAt.Valid {
			permission.DeletedAt = &deletedAt.Time
		}
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
// Supports tenant-scoped permissions (tenant_id can be NULL for backward compatibility)
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

	// Try with tenant_id column first (for new schema)
	query := `
		SELECT id, COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid) as tenant_id,
		       name, description, resource, action, created_at,
		       COALESCE(updated_at, created_at) as updated_at, deleted_at
		FROM permissions
		WHERE deleted_at IS NULL
	`
	args := []interface{}{}
	argPos := 1

	// Filter by tenant_id if provided
	if tenantID != uuid.Nil {
		query += fmt.Sprintf(" AND (tenant_id = $%d OR tenant_id IS NULL)", argPos)
		args = append(args, tenantID)
		argPos++
	}

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

	query += fmt.Sprintf(" ORDER BY resource, action LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, filters.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		// Fallback: try without tenant_id column (for old schema)
		query = `
			SELECT id, name, description, resource, action, created_at
			FROM permissions
		`
		args = []interface{}{}
		argPos = 1

		if filters.Resource != nil {
			query += fmt.Sprintf(" WHERE resource = $%d", argPos)
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

		rows, err = r.db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to list permissions: %w", err)
		}
	}
	defer rows.Close()

	var permissions []*models.Permission
	for rows.Next() {
		permission := &models.Permission{}
		var description sql.NullString
		var deletedAt sql.NullTime
		var updatedAt sql.NullTime
		var tenantIDStr string

		// Try to scan with tenant_id first
		err := rows.Scan(
			&permission.ID, &tenantIDStr, &permission.Name, &description,
			&permission.Resource, &permission.Action, &permission.CreatedAt, &updatedAt, &deletedAt,
		)
		if err != nil {
			// Fallback: scan without tenant_id
			err = rows.Scan(
				&permission.ID, &permission.Name, &description,
				&permission.Resource, &permission.Action, &permission.CreatedAt,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to scan permission: %w", err)
			}
			permission.TenantID = uuid.Nil // Global permission
		} else {
			// Parse tenant_id if not nil
			if tenantIDStr != "00000000-0000-0000-0000-000000000000" {
				parsedTenantID, parseErr := uuid.Parse(tenantIDStr)
				if parseErr == nil {
					permission.TenantID = parsedTenantID
				}
			}
			if updatedAt.Valid {
				permission.UpdatedAt = updatedAt.Time
			} else {
				permission.UpdatedAt = permission.CreatedAt
			}
			if deletedAt.Valid {
				permission.DeletedAt = &deletedAt.Time
			}
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

