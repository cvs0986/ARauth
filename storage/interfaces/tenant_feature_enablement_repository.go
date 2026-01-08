package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
)

// TenantFeatureEnablementRepository defines operations for tenant feature enablement
type TenantFeatureEnablementRepository interface {
	// GetByTenantIDAndKey retrieves a tenant feature enablement by tenant ID and key
	GetByTenantIDAndKey(ctx context.Context, tenantID uuid.UUID, key string) (*models.TenantFeatureEnablement, error)

	// GetByTenantID retrieves all feature enablements for a tenant
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*models.TenantFeatureEnablement, error)

	// GetEnabledByTenantID retrieves all enabled features for a tenant
	GetEnabledByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*models.TenantFeatureEnablement, error)

	// Create creates a new tenant feature enablement
	Create(ctx context.Context, enablement *models.TenantFeatureEnablement) error

	// Update updates an existing tenant feature enablement
	Update(ctx context.Context, enablement *models.TenantFeatureEnablement) error

	// Delete deletes a tenant feature enablement
	Delete(ctx context.Context, tenantID uuid.UUID, key string) error

	// DeleteByTenantID deletes all feature enablements for a tenant
	DeleteByTenantID(ctx context.Context, tenantID uuid.UUID) error
}

