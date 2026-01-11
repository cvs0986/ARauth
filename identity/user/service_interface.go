package user

import (
	"context"

	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
	"github.com/google/uuid"
)

// ServiceInterface defines the interface for user service operations
type ServiceInterface interface {
	Create(ctx context.Context, req *CreateUserRequest) (*models.User, error)
	CreateSystem(ctx context.Context, req *CreateUserRequest) (*models.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByUsername(ctx context.Context, username string, tenantID uuid.UUID) (*models.User, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdateUserRequest) (*models.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, filters *interfaces.UserFilters) ([]*models.User, error)
	Count(ctx context.Context, tenantID uuid.UUID, filters *interfaces.UserFilters) (int, error)
	ListSystem(ctx context.Context, filters *interfaces.UserFilters) ([]*models.User, error)
	CountSystem(ctx context.Context, filters *interfaces.UserFilters) (int, error)

	// ChangePassword changes a user's password and revokes all active sessions
	ChangePassword(ctx context.Context, userID uuid.UUID, newPassword string) error
}
