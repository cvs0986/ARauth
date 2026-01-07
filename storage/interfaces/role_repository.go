package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
)

// RoleRepository defines the interface for role data access
type RoleRepository interface {
	// Create creates a new role
	Create(ctx context.Context, role *models.Role) error

	// GetByID retrieves a role by ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.Role, error)

	// GetByName retrieves a role by name and tenant ID
	GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Role, error)

	// Update updates an existing role
	Update(ctx context.Context, role *models.Role) error

	// Delete soft deletes a role
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves a list of roles with filters
	List(ctx context.Context, tenantID uuid.UUID, filters *RoleFilters) ([]*models.Role, error)

	// GetUserRoles retrieves all roles for a user
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*models.Role, error)

	// AssignRoleToUser assigns a role to a user
	AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error

	// RemoveRoleFromUser removes a role from a user
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
}

// RoleFilters represents filters for role queries
type RoleFilters struct {
	IsSystem *bool
	Search   *string
	Page     int
	PageSize int
}

