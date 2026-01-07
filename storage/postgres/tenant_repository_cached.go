package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/internal/cache"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// cachedTenantRepository wraps TenantRepository with caching
type cachedTenantRepository struct {
	repo     interfaces.TenantRepository
	cache    *cache.Cache
	cacheTTL time.Duration
}

// NewCachedTenantRepository creates a cached tenant repository
func NewCachedTenantRepository(repo interfaces.TenantRepository, cacheClient *cache.Cache) interfaces.TenantRepository {
	if cacheClient == nil {
		return repo // Return unwrapped repository if no cache
	}

	return &cachedTenantRepository{
		repo:     repo,
		cache:    cacheClient,
		cacheTTL: 10 * time.Minute, // Tenants change less frequently
	}
}

// cacheKey generates a cache key for tenant operations
func (r *cachedTenantRepository) cacheKey(operation string, params ...interface{}) string {
	key := fmt.Sprintf("tenant:%s", operation)
	for _, p := range params {
		key += fmt.Sprintf(":%v", p)
	}
	return key
}

// Create creates a new tenant
func (r *cachedTenantRepository) Create(ctx context.Context, tenant *models.Tenant) error {
	err := r.repo.Create(ctx, tenant)
	if err != nil {
		return err
	}

	// Invalidate related cache entries
	r.invalidateTenantCache(ctx, tenant.ID, tenant.Domain)
	return nil
}

// GetByID retrieves a tenant by ID with caching
func (r *cachedTenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	cacheKey := r.cacheKey("id", id.String())

	// Try to get from cache
	var cachedTenant *models.Tenant
	err := r.cache.Get(ctx, cacheKey, &cachedTenant)
	if err == nil && cachedTenant != nil {
		return cachedTenant, nil
	}

	// Get from database
	tenant, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if tenant != nil {
		_ = r.cache.Set(ctx, cacheKey, tenant, r.cacheTTL) // Ignore cache errors
		// Also cache by domain
		if tenant.Domain != "" {
			_ = r.cache.Set(ctx, r.cacheKey("domain", tenant.Domain), tenant, r.cacheTTL) // Ignore cache errors
		}
	}

	return tenant, nil
}

// GetByDomain retrieves a tenant by domain with caching
func (r *cachedTenantRepository) GetByDomain(ctx context.Context, domain string) (*models.Tenant, error) {
	cacheKey := r.cacheKey("domain", domain)

	// Try to get from cache
	var cachedTenant *models.Tenant
	err := r.cache.Get(ctx, cacheKey, &cachedTenant)
	if err == nil && cachedTenant != nil {
		return cachedTenant, nil
	}

	// Get from database
	tenant, err := r.repo.GetByDomain(ctx, domain)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if tenant != nil {
		_ = r.cache.Set(ctx, cacheKey, tenant, r.cacheTTL) // Ignore cache errors
		// Also cache by ID
		_ = r.cache.Set(ctx, r.cacheKey("id", tenant.ID.String()), tenant, r.cacheTTL) // Ignore cache errors
	}

	return tenant, nil
}

// Update updates an existing tenant
func (r *cachedTenantRepository) Update(ctx context.Context, tenant *models.Tenant) error {
	// Get old tenant to invalidate cache
	oldTenant, _ := r.repo.GetByID(ctx, tenant.ID)

	err := r.repo.Update(ctx, tenant)
	if err != nil {
		return err
	}

	// Invalidate cache
	if oldTenant != nil {
		r.invalidateTenantCache(ctx, tenant.ID, oldTenant.Domain)
	}
	r.invalidateTenantCache(ctx, tenant.ID, tenant.Domain)

	return nil
}

// Delete soft deletes a tenant
func (r *cachedTenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Get tenant to invalidate cache
	tenant, _ := r.repo.GetByID(ctx, id)

	err := r.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Invalidate cache
	if tenant != nil {
		r.invalidateTenantCache(ctx, id, tenant.Domain)
	}

	return nil
}

// List retrieves a list of tenants (not cached due to pagination)
func (r *cachedTenantRepository) List(ctx context.Context, filters *interfaces.TenantFilters) ([]*models.Tenant, error) {
	// List operations are not cached due to pagination and filtering complexity
	return r.repo.List(ctx, filters)
}

// invalidateTenantCache invalidates all cache entries for a tenant
func (r *cachedTenantRepository) invalidateTenantCache(ctx context.Context, tenantID uuid.UUID, domain string) {
	keys := []string{
		r.cacheKey("id", tenantID.String()),
	}

	if domain != "" {
		keys = append(keys, r.cacheKey("domain", domain))
	}

	for _, key := range keys {
		_ = r.cache.Delete(ctx, key) // Ignore cache errors
	}
}

