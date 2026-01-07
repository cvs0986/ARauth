package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// ServiceInterface defines the interface for user service operations
type ServiceInterface interface {
	Create(ctx context.Context, req *CreateUserRequest) (*models.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByUsername(ctx context.Context, username string, tenantID uuid.UUID) (*models.User, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdateUserRequest) (*models.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, filters *interfaces.UserFilters) ([]*models.User, error)
	Count(ctx context.Context, tenantID uuid.UUID, filters *interfaces.UserFilters) (int, error)
}

