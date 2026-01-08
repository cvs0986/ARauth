package interfaces

import (
	"context"

	"github.com/arauth-identity/iam/identity/models"
)

// SystemCapabilityRepository defines operations for system capabilities
type SystemCapabilityRepository interface {
	// GetByKey retrieves a system capability by key
	GetByKey(ctx context.Context, key string) (*models.SystemCapability, error)

	// GetAll retrieves all system capabilities
	GetAll(ctx context.Context) ([]*models.SystemCapability, error)

	// GetEnabled retrieves all enabled system capabilities
	GetEnabled(ctx context.Context) ([]*models.SystemCapability, error)

	// Create creates a new system capability
	Create(ctx context.Context, capability *models.SystemCapability) error

	// Update updates an existing system capability
	Update(ctx context.Context, capability *models.SystemCapability) error

	// Delete deletes a system capability
	Delete(ctx context.Context, key string) error
}

