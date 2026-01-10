package interfaces

import (
	"context"

	"github.com/arauth-identity/iam/identity/models"
	"github.com/google/uuid"
)

// OAuthScopeRepository defines the interface for OAuth scope storage
type OAuthScopeRepository interface {
	// Create creates a new OAuth scope
	Create(ctx context.Context, scope *models.OAuthScope) error

	// GetByID retrieves an OAuth scope by ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.OAuthScope, error)

	// GetByName retrieves an OAuth scope by tenant ID and name
	GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.OAuthScope, error)

	// List lists OAuth scopes for a tenant
	List(ctx context.Context, tenantID uuid.UUID, filters *OAuthScopeFilters) ([]*models.OAuthScope, error)

	// Update updates an OAuth scope
	Update(ctx context.Context, scope *models.OAuthScope) error

	// Delete soft-deletes an OAuth scope
	Delete(ctx context.Context, id uuid.UUID) error

	// GetDefaultScopes retrieves all default scopes for a tenant
	GetDefaultScopes(ctx context.Context, tenantID uuid.UUID) ([]*models.OAuthScope, error)

	// GetScopesByPermissions retrieves all scopes that include any of the given permissions
	GetScopesByPermissions(ctx context.Context, tenantID uuid.UUID, permissions []string) ([]*models.OAuthScope, error)
}

// OAuthScopeFilters defines filters for listing OAuth scopes
type OAuthScopeFilters struct {
	IsDefault *bool
	Page      int
	PageSize  int
}

