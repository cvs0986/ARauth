package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/identity/models"
)

// PermissionRepository defines the interface for permission data access
type PermissionRepository interface {
	// Create creates a new permission
	Create(ctx context.Context, permission *models.Permission) error

	// GetByID retrieves a permission by ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.Permission, error)

	// GetByName retrieves a permission by name and tenant ID
	GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Permission, error)

	// Update updates an existing permission
	Update(ctx context.Context, permission *models.Permission) error

	// Delete soft deletes a permission
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves a list of permissions with filters
	List(ctx context.Context, tenantID uuid.UUID, filters *PermissionFilters) ([]*models.Permission, error)

	// GetRolePermissions retrieves all permissions for a role
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*models.Permission, error)

	// AssignPermissionToRole assigns a permission to a role
	AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error

	// RemovePermissionFromRole removes a permission from a role
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error
}

// PermissionFilters represents filters for permission queries
type PermissionFilters struct {
	Resource *string
	Action   *string
	Search   *string
	Page     int
	PageSize int
}

