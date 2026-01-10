package oauth_scope

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// Service provides OAuth scope management
type Service struct {
	scopeRepo interfaces.OAuthScopeRepository
}

// NewService creates a new OAuth scope service
func NewService(scopeRepo interfaces.OAuthScopeRepository) ServiceInterface {
	return &Service{
		scopeRepo: scopeRepo,
	}
}

// CreateScope creates a new OAuth scope
func (s *Service) CreateScope(ctx context.Context, tenantID uuid.UUID, req *CreateScopeRequest) (*models.OAuthScope, error) {
	// Validate scope name format (should be lowercase, alphanumeric with dots/underscores)
	if !isValidScopeName(req.Name) {
		return nil, fmt.Errorf("invalid scope name: must be lowercase alphanumeric with dots, underscores, or colons")
	}

	// Check if scope with same name already exists
	existing, err := s.scopeRepo.GetByName(ctx, tenantID, req.Name)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("scope with name '%s' already exists", req.Name)
	}

	// Validate permissions are not empty
	if len(req.Permissions) == 0 {
		return nil, fmt.Errorf("at least one permission is required")
	}

	scope := &models.OAuthScope{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
		IsDefault:   req.IsDefault,
	}

	if err := s.scopeRepo.Create(ctx, scope); err != nil {
		return nil, fmt.Errorf("failed to create scope: %w", err)
	}

	return scope, nil
}

// GetScope retrieves an OAuth scope by ID
func (s *Service) GetScope(ctx context.Context, id uuid.UUID) (*models.OAuthScope, error) {
	return s.scopeRepo.GetByID(ctx, id)
}

// GetScopeByName retrieves an OAuth scope by tenant ID and name
func (s *Service) GetScopeByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.OAuthScope, error) {
	return s.scopeRepo.GetByName(ctx, tenantID, name)
}

// ListScopes lists OAuth scopes for a tenant
func (s *Service) ListScopes(ctx context.Context, tenantID uuid.UUID, filters *ScopeFilters) ([]*models.OAuthScope, error) {
	repoFilters := &interfaces.OAuthScopeFilters{
		IsDefault: filters.IsDefault,
		Page:      filters.Page,
		PageSize:  filters.PageSize,
	}

	return s.scopeRepo.List(ctx, tenantID, repoFilters)
}

// UpdateScope updates an OAuth scope
func (s *Service) UpdateScope(ctx context.Context, id uuid.UUID, req *UpdateScopeRequest) (*models.OAuthScope, error) {
	// Get existing scope
	scope, err := s.scopeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("scope not found: %w", err)
	}

	// Update fields if provided
	if req.Name != nil {
		if !isValidScopeName(*req.Name) {
			return nil, fmt.Errorf("invalid scope name: must be lowercase alphanumeric with dots, underscores, or colons")
		}
		// Check if new name conflicts with existing scope
		if *req.Name != scope.Name {
			existing, err := s.scopeRepo.GetByName(ctx, scope.TenantID, *req.Name)
			if err == nil && existing != nil {
				return nil, fmt.Errorf("scope with name '%s' already exists", *req.Name)
			}
		}
		scope.Name = *req.Name
	}

	if req.Description != nil {
		scope.Description = req.Description
	}

	if req.Permissions != nil {
		if len(req.Permissions) == 0 {
			return nil, fmt.Errorf("at least one permission is required")
		}
		scope.Permissions = req.Permissions
	}

	if req.IsDefault != nil {
		scope.IsDefault = *req.IsDefault
	}

	// Update in database
	if err := s.scopeRepo.Update(ctx, scope); err != nil {
		return nil, fmt.Errorf("failed to update scope: %w", err)
	}

	return scope, nil
}

// DeleteScope deletes an OAuth scope
func (s *Service) DeleteScope(ctx context.Context, id uuid.UUID) error {
	return s.scopeRepo.Delete(ctx, id)
}

// GetScopesForPermissions returns all scopes that include any of the given permissions
func (s *Service) GetScopesForPermissions(ctx context.Context, tenantID uuid.UUID, permissions []string) ([]*models.OAuthScope, error) {
	return s.scopeRepo.GetScopesByPermissions(ctx, tenantID, permissions)
}

// GetDefaultScopes returns all default scopes for a tenant
func (s *Service) GetDefaultScopes(ctx context.Context, tenantID uuid.UUID) ([]*models.OAuthScope, error) {
	return s.scopeRepo.GetDefaultScopes(ctx, tenantID)
}

// isValidScopeName validates scope name format
// Scope names should be lowercase alphanumeric with dots, underscores, or colons
// Examples: "users.read", "roles.manage", "tenant:admin"
func isValidScopeName(name string) bool {
	if name == "" {
		return false
	}

	// Check if all characters are valid
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '.' ||
			char == '_' ||
			char == ':') {
			return false
		}
	}

	// Must start with a letter or number
	if !((name[0] >= 'a' && name[0] <= 'z') || (name[0] >= '0' && name[0] <= '9')) {
		return false
	}

	return true
}

