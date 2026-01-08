package interfaces

import (
	"context"

	"github.com/google/uuid"
)

// SystemRole represents a system role (not tenant-scoped)
type SystemRole struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatedAt   string    `db:"created_at"`
	UpdatedAt   string    `db:"updated_at"`
}

// SystemPermission represents a system permission
type SystemPermission struct {
	ID          uuid.UUID `db:"id"`
	Resource    string    `db:"resource"`
	Action      string    `db:"action"`
	Description string    `db:"description"`
	CreatedAt   string    `db:"created_at"`
	UpdatedAt   string    `db:"updated_at"`
}

// SystemRoleRepository defines the interface for system role data access
type SystemRoleRepository interface {
	// GetByID retrieves a system role by ID
	GetByID(ctx context.Context, id uuid.UUID) (*SystemRole, error)

	// GetByName retrieves a system role by name
	GetByName(ctx context.Context, name string) (*SystemRole, error)

	// List retrieves all system roles
	List(ctx context.Context) ([]*SystemRole, error)

	// GetUserSystemRoles retrieves all system roles for a user
	GetUserSystemRoles(ctx context.Context, userID uuid.UUID) ([]*SystemRole, error)

	// AssignRoleToUser assigns a system role to a user
	AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID, assignedBy *uuid.UUID) error

	// RemoveRoleFromUser removes a system role from a user
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error

	// GetRolePermissions retrieves all permissions for a system role
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*SystemPermission, error)
}

// SystemPermissionRepository defines the interface for system permission data access
type SystemPermissionRepository interface {
	// GetByID retrieves a system permission by ID
	GetByID(ctx context.Context, id uuid.UUID) (*SystemPermission, error)

	// GetByResourceAction retrieves a system permission by resource and action
	GetByResourceAction(ctx context.Context, resource, action string) (*SystemPermission, error)

	// List retrieves all system permissions
	List(ctx context.Context) ([]*SystemPermission, error)
}

