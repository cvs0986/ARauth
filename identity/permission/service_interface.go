package permission

import (
	"context"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/identity/models"
	"github.com/nuage-identity/iam/storage/interfaces"
)

// ServiceInterface defines the interface for permission service operations
type ServiceInterface interface {
	Create(ctx context.Context, req *CreatePermissionRequest) (*models.Permission, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Permission, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdatePermissionRequest) (*models.Permission, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, filters *interfaces.PermissionFilters) ([]*models.Permission, error)
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*models.Permission, error)
	AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error
}

