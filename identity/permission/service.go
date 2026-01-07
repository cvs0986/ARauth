package permission

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// Service provides permission management business logic
type Service struct {
	repo interfaces.PermissionRepository
}

// NewService creates a new permission service
func NewService(repo interfaces.PermissionRepository) *Service {
	return &Service{repo: repo}
}

// CreatePermissionRequest represents a request to create a permission
type CreatePermissionRequest struct {
	TenantID    uuid.UUID `json:"tenant_id" binding:"required"`
	Name        string    `json:"name" binding:"required,min=3,max=255"`
	Description *string   `json:"description,omitempty"`
	Resource    string    `json:"resource" binding:"required"`
	Action      string    `json:"action" binding:"required"`
}

// UpdatePermissionRequest represents a request to update a permission
type UpdatePermissionRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Resource    *string `json:"resource,omitempty"`
	Action      *string `json:"action,omitempty"`
}

// Create creates a new permission
func (s *Service) Create(ctx context.Context, req *CreatePermissionRequest) (*models.Permission, error) {
	// Validate tenant ID
	if req.TenantID == uuid.Nil {
		return nil, fmt.Errorf("tenant_id is required")
	}

	// Normalize name
	name := strings.TrimSpace(req.Name)
	if len(name) < 3 {
		return nil, fmt.Errorf("permission name must be at least 3 characters")
	}

	// Validate resource and action
	resource := strings.TrimSpace(req.Resource)
	action := strings.TrimSpace(req.Action)
	if resource == "" || action == "" {
		return nil, fmt.Errorf("resource and action are required")
	}

	// Check if permission with same name already exists
	existing, _ := s.repo.GetByName(ctx, req.TenantID, name)
	if existing != nil {
		return nil, fmt.Errorf("permission with name %s already exists", name)
	}

	// Create permission
	permission := &models.Permission{
		ID:          uuid.New(),
		TenantID:    req.TenantID,
		Name:        name,
		Description: req.Description,
		Resource:    resource,
		Action:      action,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, permission); err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	return permission, nil
}

// GetByID retrieves a permission by ID
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	permission, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("permission not found: %w", err)
	}

	return permission, nil
}

// Update updates an existing permission
func (s *Service) Update(ctx context.Context, id uuid.UUID, req *UpdatePermissionRequest) (*models.Permission, error) {
	// Get existing permission
	permission, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("permission not found: %w", err)
	}

	// Update fields
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if len(name) < 3 {
			return nil, fmt.Errorf("permission name must be at least 3 characters")
		}

		// Check if name is already taken
		existing, _ := s.repo.GetByName(ctx, permission.TenantID, name)
		if existing != nil && existing.ID != id {
			return nil, fmt.Errorf("permission name %s is already taken", name)
		}

		permission.Name = name
	}

	if req.Description != nil {
		permission.Description = req.Description
	}

	if req.Resource != nil {
		permission.Resource = strings.TrimSpace(*req.Resource)
	}

	if req.Action != nil {
		permission.Action = strings.TrimSpace(*req.Action)
	}

	permission.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, permission); err != nil {
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}

	return permission, nil
}

// Delete deletes a permission
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// List retrieves a list of permissions
func (s *Service) List(ctx context.Context, tenantID uuid.UUID, filters *interfaces.PermissionFilters) ([]*models.Permission, error) {
	if tenantID == uuid.Nil {
		return nil, fmt.Errorf("tenant_id is required")
	}

	permissions, err := s.repo.List(ctx, tenantID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	return permissions, nil
}

