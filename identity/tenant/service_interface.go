package tenant

import (
	"context"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/identity/models"
	"github.com/nuage-identity/iam/storage/interfaces"
)

// ServiceInterface defines the interface for tenant service operations
type ServiceInterface interface {
	Create(ctx context.Context, req *CreateTenantRequest) (*models.Tenant, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error)
	GetByDomain(ctx context.Context, domain string) (*models.Tenant, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdateTenantRequest) (*models.Tenant, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filters *interfaces.TenantFilters) ([]*models.Tenant, error)
}

