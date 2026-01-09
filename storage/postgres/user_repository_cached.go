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

// cachedUserRepository wraps UserRepository with caching
type cachedUserRepository struct {
	repo       interfaces.UserRepository
	cache      *cache.Cache
	cacheTTL   time.Duration
}

// NewCachedUserRepository creates a cached user repository
func NewCachedUserRepository(repo interfaces.UserRepository, cacheClient *cache.Cache) interfaces.UserRepository {
	if cacheClient == nil {
		return repo // Return unwrapped repository if no cache
	}

	return &cachedUserRepository{
		repo:     repo,
		cache:    cacheClient,
		cacheTTL: 5 * time.Minute, // Default cache TTL
	}
}

// cacheKey generates a cache key for user operations
func (r *cachedUserRepository) cacheKey(operation string, params ...interface{}) string {
	key := fmt.Sprintf("user:%s", operation)
	for _, p := range params {
		key += fmt.Sprintf(":%v", p)
	}
	return key
}

// Create creates a new user
func (r *cachedUserRepository) Create(ctx context.Context, user *models.User) error {
	err := r.repo.Create(ctx, user)
	if err != nil {
		return err
	}

	// Invalidate related cache entries
	var tenantID uuid.UUID
	if user.TenantID != nil {
		tenantID = *user.TenantID
	}
	r.invalidateUserCache(ctx, user.ID, tenantID, user.Username, user.Email)
	return nil
}

// GetByID retrieves a user by ID with caching
func (r *cachedUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	cacheKey := r.cacheKey("id", id.String())

	// Try to get from cache
	var cachedUser *models.User
	err := r.cache.Get(ctx, cacheKey, &cachedUser)
	if err == nil && cachedUser != nil {
		return cachedUser, nil
	}

	// Get from database
	user, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if user != nil {
		_ = r.cache.Set(ctx, cacheKey, user, r.cacheTTL) // Ignore cache errors
	}

	return user, nil
}

// GetByUsername retrieves a user by username with caching
func (r *cachedUserRepository) GetByUsername(ctx context.Context, username string, tenantID uuid.UUID) (*models.User, error) {
	cacheKey := r.cacheKey("username", tenantID.String(), username)

	// Try to get from cache
	var cachedUser *models.User
	err := r.cache.Get(ctx, cacheKey, &cachedUser)
	if err == nil && cachedUser != nil {
		return cachedUser, nil
	}

	// Get from database
	user, err := r.repo.GetByUsername(ctx, username, tenantID)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if user != nil {
		_ = r.cache.Set(ctx, cacheKey, user, r.cacheTTL) // Ignore cache errors
		// Also cache by ID
		_ = r.cache.Set(ctx, r.cacheKey("id", user.ID.String()), user, r.cacheTTL) // Ignore cache errors
	}

	return user, nil
}

// GetByEmail retrieves a user by email with caching
func (r *cachedUserRepository) GetByEmail(ctx context.Context, email string, tenantID uuid.UUID) (*models.User, error) {
	cacheKey := r.cacheKey("email", tenantID.String(), email)

	// Try to get from cache
	var cachedUser *models.User
	err := r.cache.Get(ctx, cacheKey, &cachedUser)
	if err == nil && cachedUser != nil {
		return cachedUser, nil
	}

	// Get from database
	user, err := r.repo.GetByEmail(ctx, email, tenantID)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if user != nil {
		_ = r.cache.Set(ctx, cacheKey, user, r.cacheTTL) // Ignore cache errors
		// Also cache by ID
		_ = r.cache.Set(ctx, r.cacheKey("id", user.ID.String()), user, r.cacheTTL) // Ignore cache errors
	}

	return user, nil
}

// Update updates an existing user
func (r *cachedUserRepository) Update(ctx context.Context, user *models.User) error {
	// Get old user to invalidate cache
	oldUser, _ := r.repo.GetByID(ctx, user.ID)

	err := r.repo.Update(ctx, user)
	if err != nil {
		return err
	}

	// Invalidate cache
	if oldUser != nil {
		var oldTenantID uuid.UUID
		if oldUser.TenantID != nil {
			oldTenantID = *oldUser.TenantID
		}
		r.invalidateUserCache(ctx, user.ID, oldTenantID, oldUser.Username, oldUser.Email)
	}
	var tenantID uuid.UUID
	if user.TenantID != nil {
		tenantID = *user.TenantID
	}
	r.invalidateUserCache(ctx, user.ID, tenantID, user.Username, user.Email)

	return nil
}

// Delete soft deletes a user
func (r *cachedUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Get user to invalidate cache
	user, _ := r.repo.GetByID(ctx, id)

	err := r.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Invalidate cache
	if user != nil {
		var tenantID uuid.UUID
		if user.TenantID != nil {
			tenantID = *user.TenantID
		}
		r.invalidateUserCache(ctx, id, tenantID, user.Username, user.Email)
	}

	return nil
}

// List retrieves a list of users (not cached due to pagination)
func (r *cachedUserRepository) List(ctx context.Context, tenantID uuid.UUID, filters *interfaces.UserFilters) ([]*models.User, error) {
	// List operations are not cached due to pagination and filtering complexity
	return r.repo.List(ctx, tenantID, filters)
}

// Count returns the total number of users (not cached)
func (r *cachedUserRepository) Count(ctx context.Context, tenantID uuid.UUID, filters *interfaces.UserFilters) (int, error) {
	return r.repo.Count(ctx, tenantID, filters)
}

// ListSystem retrieves a list of system users (principal_type = 'SYSTEM') with filters
func (r *cachedUserRepository) ListSystem(ctx context.Context, filters *interfaces.UserFilters) ([]*models.User, error) {
	// For system users, we might want to cache differently or not cache at all
	// For now, delegate to underlying repository
	return r.repo.ListSystem(ctx, filters)
}

// CountSystem returns the total count of system users matching filters
func (r *cachedUserRepository) CountSystem(ctx context.Context, filters *interfaces.UserFilters) (int, error) {
	// For system users, we might want to cache differently or not cache at all
	// For now, delegate to underlying repository
	return r.repo.CountSystem(ctx, filters)
}

// GetByEmailSystem retrieves a SYSTEM user by email (no tenant ID required)
func (r *cachedUserRepository) GetByEmailSystem(ctx context.Context, email string) (*models.User, error) {
	// SYSTEM users are not cached by email (they're rare and should be looked up directly)
	return r.repo.GetByEmailSystem(ctx, email)
}

// GetSystemUserByUsername retrieves a SYSTEM user by username (no tenant ID required)
func (r *cachedUserRepository) GetSystemUserByUsername(ctx context.Context, username string) (*models.User, error) {
	// SYSTEM users are not cached by username (they're rare and should be looked up directly)
	return r.repo.GetSystemUserByUsername(ctx, username)
}

// invalidateUserCache invalidates all cache entries for a user
func (r *cachedUserRepository) invalidateUserCache(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID, username, email string) {
	keys := []string{
		r.cacheKey("id", userID.String()),
		r.cacheKey("username", tenantID.String(), username),
		r.cacheKey("email", tenantID.String(), email),
	}

	for _, key := range keys {
		_ = r.cache.Delete(ctx, key) // Ignore cache errors
	}
}

