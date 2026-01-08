package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// SystemRoleRepository implements interfaces.SystemRoleRepository
type SystemRoleRepository struct {
	db *sql.DB
}

// NewSystemRoleRepository creates a new system role repository
func NewSystemRoleRepository(db *sql.DB) *SystemRoleRepository {
	return &SystemRoleRepository{db: db}
}

// GetByID retrieves a system role by ID
func (r *SystemRoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*interfaces.SystemRole, error) {
	query := `SELECT id, name, description, created_at, updated_at 
	          FROM system_roles 
	          WHERE id = $1`
	
	var role interfaces.SystemRole
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.CreatedAt,
		&role.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("system role not found")
		}
		return nil, fmt.Errorf("failed to get system role: %w", err)
	}
	return &role, nil
}

// GetByName retrieves a system role by name
func (r *SystemRoleRepository) GetByName(ctx context.Context, name string) (*interfaces.SystemRole, error) {
	query := `SELECT id, name, description, created_at, updated_at 
	          FROM system_roles 
	          WHERE name = $1`
	
	var role interfaces.SystemRole
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.CreatedAt,
		&role.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("system role not found")
		}
		return nil, fmt.Errorf("failed to get system role: %w", err)
	}
	return &role, nil
}

// List retrieves all system roles
func (r *SystemRoleRepository) List(ctx context.Context) ([]*interfaces.SystemRole, error) {
	query := `SELECT id, name, description, created_at, updated_at 
	          FROM system_roles 
	          ORDER BY name`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list system roles: %w", err)
	}
	defer rows.Close()

	var roles []*interfaces.SystemRole
	for rows.Next() {
		var role interfaces.SystemRole
		if err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.Description,
			&role.CreatedAt,
			&role.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan system role: %w", err)
		}
		roles = append(roles, &role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating system roles: %w", err)
	}

	return roles, nil
}

// GetUserSystemRoles retrieves all system roles for a user
func (r *SystemRoleRepository) GetUserSystemRoles(ctx context.Context, userID uuid.UUID) ([]*interfaces.SystemRole, error) {
	query := `SELECT sr.id, sr.name, sr.description, sr.created_at, sr.updated_at
	          FROM system_roles sr
	          INNER JOIN user_system_roles usr ON sr.id = usr.role_id
	          WHERE usr.user_id = $1
	          ORDER BY sr.name`
	
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user system roles: %w", err)
	}
	defer rows.Close()

	var roles []*interfaces.SystemRole
	for rows.Next() {
		var role interfaces.SystemRole
		if err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.Description,
			&role.CreatedAt,
			&role.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan system role: %w", err)
		}
		roles = append(roles, &role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user system roles: %w", err)
	}

	return roles, nil
}

// AssignRoleToUser assigns a system role to a user
func (r *SystemRoleRepository) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID, assignedBy *uuid.UUID) error {
	query := `INSERT INTO user_system_roles (user_id, role_id, assigned_by, assigned_at)
	          VALUES ($1, $2, $3, NOW())
	          ON CONFLICT (user_id, role_id) DO NOTHING`
	
	_, err := r.db.ExecContext(ctx, query, userID, roleID, assignedBy)
	if err != nil {
		return fmt.Errorf("failed to assign system role to user: %w", err)
	}
	return nil
}

// RemoveRoleFromUser removes a system role from a user
func (r *SystemRoleRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	query := `DELETE FROM user_system_roles 
	          WHERE user_id = $1 AND role_id = $2`
	
	_, err := r.db.ExecContext(ctx, query, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to remove system role from user: %w", err)
	}
	return nil
}

// GetRolePermissions retrieves all permissions for a system role
func (r *SystemRoleRepository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*interfaces.SystemPermission, error) {
	query := `SELECT sp.id, sp.resource, sp.action, sp.description, sp.created_at, sp.updated_at
	          FROM system_permissions sp
	          INNER JOIN system_role_permissions srp ON sp.id = srp.permission_id
	          WHERE srp.role_id = $1
	          ORDER BY sp.resource, sp.action`
	
	rows, err := r.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get system role permissions: %w", err)
	}
	defer rows.Close()

	var permissions []*interfaces.SystemPermission
	for rows.Next() {
		var perm interfaces.SystemPermission
		if err := rows.Scan(
			&perm.ID,
			&perm.Resource,
			&perm.Action,
			&perm.Description,
			&perm.CreatedAt,
			&perm.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan system permission: %w", err)
		}
		permissions = append(permissions, &perm)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating system role permissions: %w", err)
	}

	return permissions, nil
}

