package oauth_scope

import (
	"context"

	"github.com/arauth-identity/iam/identity/models"
	"github.com/google/uuid"
)

// ServiceInterface defines the interface for OAuth scope service
type ServiceInterface interface {
	// CreateScope creates a new OAuth scope
	CreateScope(ctx context.Context, tenantID uuid.UUID, req *CreateScopeRequest) (*models.OAuthScope, error)

	// GetScope retrieves an OAuth scope by ID
	GetScope(ctx context.Context, id uuid.UUID) (*models.OAuthScope, error)

	// GetScopeByName retrieves an OAuth scope by tenant ID and name
	GetScopeByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.OAuthScope, error)

	// ListScopes lists OAuth scopes for a tenant
	ListScopes(ctx context.Context, tenantID uuid.UUID, filters *ScopeFilters) ([]*models.OAuthScope, error)

	// UpdateScope updates an OAuth scope
	UpdateScope(ctx context.Context, id uuid.UUID, req *UpdateScopeRequest) (*models.OAuthScope, error)

	// DeleteScope deletes an OAuth scope
	DeleteScope(ctx context.Context, id uuid.UUID) error

	// GetScopesForPermissions returns all scopes that include any of the given permissions
	GetScopesForPermissions(ctx context.Context, tenantID uuid.UUID, permissions []string) ([]*models.OAuthScope, error)

	// GetDefaultScopes returns all default scopes for a tenant
	GetDefaultScopes(ctx context.Context, tenantID uuid.UUID) ([]*models.OAuthScope, error)
}

// CreateScopeRequest represents a request to create an OAuth scope
type CreateScopeRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description *string  `json:"description,omitempty"`
	Permissions []string `json:"permissions" binding:"required,min=1"`
	IsDefault   bool     `json:"is_default"`
}

// UpdateScopeRequest represents a request to update an OAuth scope
type UpdateScopeRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	IsDefault   *bool    `json:"is_default,omitempty"`
}

// ScopeFilters defines filters for listing OAuth scopes
type ScopeFilters struct {
	IsDefault *bool
	Page      int
	PageSize  int
}

