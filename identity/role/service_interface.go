package role

import (
	"context"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// ServiceInterface defines the interface for role service operations
type ServiceInterface interface {
	Create(ctx context.Context, req *CreateRoleRequest) (*models.Role, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Role, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdateRoleRequest) (*models.Role, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, filters *interfaces.RoleFilters) ([]*models.Role, error)
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*models.Role, error)
	AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*models.Permission, error)
	AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error
}

