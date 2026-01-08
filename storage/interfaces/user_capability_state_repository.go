package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
)

// UserCapabilityStateRepository defines operations for user capability state
type UserCapabilityStateRepository interface {
	// GetByUserIDAndKey retrieves a user capability state by user ID and key
	GetByUserIDAndKey(ctx context.Context, userID uuid.UUID, key string) (*models.UserCapabilityState, error)

	// GetByUserID retrieves all capability states for a user
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.UserCapabilityState, error)

	// GetEnrolledByUserID retrieves all enrolled capabilities for a user
	GetEnrolledByUserID(ctx context.Context, userID uuid.UUID) ([]*models.UserCapabilityState, error)

	// Create creates a new user capability state
	Create(ctx context.Context, state *models.UserCapabilityState) error

	// Update updates an existing user capability state
	Update(ctx context.Context, state *models.UserCapabilityState) error

	// Delete deletes a user capability state
	Delete(ctx context.Context, userID uuid.UUID, key string) error

	// DeleteByUserID deletes all capability states for a user
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

