package interfaces

import (
	"context"

	"github.com/arauth-identity/iam/identity/models"
	"github.com/google/uuid"
)

// SCIMTokenRepository defines the interface for SCIM token storage
type SCIMTokenRepository interface {
	// Create creates a new SCIM token
	Create(ctx context.Context, token *models.SCIMToken) error

	// GetByID retrieves a SCIM token by ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.SCIMToken, error)

	// GetByLookupHash retrieves a SCIM token by its lookup hash (SHA256)
	GetByLookupHash(ctx context.Context, lookupHash string) (*models.SCIMToken, error)

	// List lists SCIM tokens for a tenant
	List(ctx context.Context, tenantID uuid.UUID) ([]*models.SCIMToken, error)

	// Update updates a SCIM token
	Update(ctx context.Context, token *models.SCIMToken) error

	// Delete soft-deletes a SCIM token
	Delete(ctx context.Context, id uuid.UUID) error

	// UpdateLastUsed updates the last_used_at timestamp
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error
}

