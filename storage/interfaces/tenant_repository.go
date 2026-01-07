package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
)

// TenantRepository defines the interface for tenant data access
type TenantRepository interface {
	// Create creates a new tenant
	Create(ctx context.Context, tenant *models.Tenant) error

	// GetByID retrieves a tenant by ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error)

	// GetByDomain retrieves a tenant by domain
	GetByDomain(ctx context.Context, domain string) (*models.Tenant, error)

	// Update updates an existing tenant
	Update(ctx context.Context, tenant *models.Tenant) error

	// Delete soft deletes a tenant
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves a list of tenants
	List(ctx context.Context, filters *TenantFilters) ([]*models.Tenant, error)
}

// TenantFilters represents filters for tenant queries
type TenantFilters struct {
	Status *string
	Search *string
	Page   int
	PageSize int
}

