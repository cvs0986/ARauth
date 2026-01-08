package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
)

// TenantCapabilityRepository defines operations for tenant capabilities
type TenantCapabilityRepository interface {
	// GetByTenantIDAndKey retrieves a tenant capability by tenant ID and key
	GetByTenantIDAndKey(ctx context.Context, tenantID uuid.UUID, key string) (*models.TenantCapability, error)

	// GetByTenantID retrieves all capabilities for a tenant
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*models.TenantCapability, error)

	// GetEnabledByTenantID retrieves all enabled capabilities for a tenant
	GetEnabledByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*models.TenantCapability, error)

	// Create creates a new tenant capability
	Create(ctx context.Context, capability *models.TenantCapability) error

	// Update updates an existing tenant capability
	Update(ctx context.Context, capability *models.TenantCapability) error

	// Delete deletes a tenant capability
	Delete(ctx context.Context, tenantID uuid.UUID, key string) error

	// DeleteByTenantID deletes all capabilities for a tenant
	DeleteByTenantID(ctx context.Context, tenantID uuid.UUID) error
}

