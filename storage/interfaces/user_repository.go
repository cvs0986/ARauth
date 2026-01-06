package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/identity/user"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, u *user.User) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uuid.UUID) (*user.User, error)

	// GetByUsername retrieves a user by username and tenant ID
	GetByUsername(ctx context.Context, username string, tenantID uuid.UUID) (*user.User, error)

	// GetByEmail retrieves a user by email and tenant ID
	GetByEmail(ctx context.Context, email string, tenantID uuid.UUID) (*user.User, error)

	// Update updates an existing user
	Update(ctx context.Context, u *user.User) error

	// Delete soft deletes a user
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves a list of users with filters
	List(ctx context.Context, tenantID uuid.UUID, filters *UserFilters) ([]*user.User, error)

	// Count returns the total count of users matching filters
	Count(ctx context.Context, tenantID uuid.UUID, filters *UserFilters) (int, error)
}

// UserFilters represents filters for user queries
type UserFilters struct {
	Status   *string
	Search   *string // Search in username, email, first_name, last_name
	Page     int
	PageSize int
}

