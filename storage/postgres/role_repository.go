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

// roleRepository implements RoleRepository for PostgreSQL
type roleRepository struct {
	db *sql.DB
}

// NewRoleRepository creates a new PostgreSQL role repository
func NewRoleRepository(db *sql.DB) interfaces.RoleRepository {
	return &roleRepository{db: db}
}

// Create creates a new role
func (r *roleRepository) Create(ctx context.Context, role *models.Role) error {
	query := `
		INSERT INTO roles (id, tenant_id, name, description, is_system, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	now := time.Now()
	if role.ID == uuid.Nil {
		role.ID = uuid.New()
	}
	if role.CreatedAt.IsZero() {
		role.CreatedAt = now
	}
	if role.UpdatedAt.IsZero() {
		role.UpdatedAt = now
	}

	_, err := r.db.ExecContext(ctx, query,
		role.ID, role.TenantID, role.Name, role.Description,
		role.IsSystem, role.CreatedAt, role.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	return nil
}

// GetByID retrieves a role by ID
func (r *roleRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	query := `
		SELECT id, tenant_id, name, description, is_system, created_at, updated_at, deleted_at
		FROM roles
		WHERE id = $1 AND deleted_at IS NULL
	`

	role := &models.Role{}
	var description sql.NullString
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&role.ID, &role.TenantID, &role.Name, &description,
		&role.IsSystem, &role.CreatedAt, &role.UpdatedAt, &deletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("role not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	if description.Valid {
		role.Description = &description.String
	}
	if deletedAt.Valid {
		role.DeletedAt = &deletedAt.Time
	}

	return role, nil
}

// GetByName retrieves a role by name and tenant ID
func (r *roleRepository) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Role, error) {
	query := `
		SELECT id, tenant_id, name, description, is_system, created_at, updated_at, deleted_at
		FROM roles
		WHERE tenant_id = $1 AND name = $2 AND deleted_at IS NULL
	`

	role := &models.Role{}
	var description sql.NullString
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, tenantID, name).Scan(
		&role.ID, &role.TenantID, &role.Name, &description,
		&role.IsSystem, &role.CreatedAt, &role.UpdatedAt, &deletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("role not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get role by name: %w", err)
	}

	if description.Valid {
		role.Description = &description.String
	}
	if deletedAt.Valid {
		role.DeletedAt = &deletedAt.Time
	}

	return role, nil
}

// Update updates an existing role
func (r *roleRepository) Update(ctx context.Context, role *models.Role) error {
	query := `
		UPDATE roles
		SET name = $2, description = $3, updated_at = $4
		WHERE id = $1 AND deleted_at IS NULL AND is_system = false
	`

	role.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		role.ID, role.Name, role.Description, role.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	return nil
}

// Delete soft deletes a role
func (r *roleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Check if role is system role (cannot be deleted)
	var isSystem bool
	err := r.db.QueryRowContext(ctx, "SELECT is_system FROM roles WHERE id = $1", id).Scan(&isSystem)
	if err == nil && isSystem {
		return fmt.Errorf("cannot delete system role")
	}

	query := `
		UPDATE roles
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("role not found")
	}

	return nil
}

// List retrieves a list of roles with filters
func (r *roleRepository) List(ctx context.Context, tenantID uuid.UUID, filters *interfaces.RoleFilters) ([]*models.Role, error) {
	if filters == nil {
		filters = &interfaces.RoleFilters{
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
		SELECT id, tenant_id, name, description, is_system, created_at, updated_at, deleted_at
		FROM roles
		WHERE tenant_id = $1 AND deleted_at IS NULL
	`
	args := []interface{}{tenantID}
	argPos := 2

	if filters.IsSystem != nil {
		query += fmt.Sprintf(" AND is_system = $%d", argPos)
		args = append(args, *filters.IsSystem)
		argPos++
	}

	if filters.Search != nil {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argPos, argPos)
		searchPattern := "%" + *filters.Search + "%"
		args = append(args, searchPattern)
		argPos++
	}

	query += " ORDER BY created_at DESC LIMIT $" + fmt.Sprintf("%d", argPos) + " OFFSET $" + fmt.Sprintf("%d", argPos+1)
	args = append(args, filters.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	defer rows.Close()

	var roles []*models.Role
	for rows.Next() {
		role := &models.Role{}
		var description sql.NullString
		var deletedAt sql.NullTime

		err := rows.Scan(
			&role.ID, &role.TenantID, &role.Name, &description,
			&role.IsSystem, &role.CreatedAt, &role.UpdatedAt, &deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}

		if description.Valid {
			role.Description = &description.String
		}
		if deletedAt.Valid {
			role.DeletedAt = &deletedAt.Time
		}

		roles = append(roles, role)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating roles: %w", err)
	}

	return roles, nil
}

// GetUserRoles retrieves all roles for a user
func (r *roleRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*models.Role, error) {
	query := `
		SELECT r.id, r.tenant_id, r.name, r.description, r.is_system, r.created_at, r.updated_at, r.deleted_at
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND r.deleted_at IS NULL
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	defer rows.Close()

	var roles []*models.Role
	for rows.Next() {
		role := &models.Role{}
		var description sql.NullString
		var deletedAt sql.NullTime

		err := rows.Scan(
			&role.ID, &role.TenantID, &role.Name, &description,
			&role.IsSystem, &role.CreatedAt, &role.UpdatedAt, &deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}

		if description.Valid {
			role.Description = &description.String
		}
		if deletedAt.Valid {
			role.DeletedAt = &deletedAt.Time
		}

		roles = append(roles, role)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating roles: %w", err)
	}

	return roles, nil
}

// AssignRoleToUser assigns a role to a user
func (r *roleRepository) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	query := `
		INSERT INTO user_roles (id, user_id, role_id, assigned_at)
		VALUES (gen_random_uuid(), $1, $2, NOW())
		ON CONFLICT (user_id, role_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to assign role to user: %w", err)
	}

	return nil
}

// RemoveRoleFromUser removes a role from a user
func (r *roleRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	query := `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`

	result, err := r.db.ExecContext(ctx, query, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to remove role from user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("role assignment not found")
	}

	return nil
}

