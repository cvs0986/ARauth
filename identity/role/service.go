package role

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// Service provides role management business logic
type Service struct {
	roleRepo       interfaces.RoleRepository
	permissionRepo interfaces.PermissionRepository
}

// NewService creates a new role service
func NewService(roleRepo interfaces.RoleRepository, permissionRepo interfaces.PermissionRepository) *Service {
	return &Service{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
	}
}

// CreateRoleRequest represents a request to create a role
type CreateRoleRequest struct {
	TenantID    uuid.UUID `json:"tenant_id" binding:"required"`
	Name        string    `json:"name" binding:"required,min=3,max=255"`
	Description *string   `json:"description,omitempty"`
	IsSystem    bool      `json:"is_system,omitempty"`
}

// UpdateRoleRequest represents a request to update a role
type UpdateRoleRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// Create creates a new role
func (s *Service) Create(ctx context.Context, req *CreateRoleRequest) (*models.Role, error) {
	// Validate tenant ID
	if req.TenantID == uuid.Nil {
		return nil, fmt.Errorf("tenant_id is required")
	}

	// Normalize name
	name := strings.TrimSpace(req.Name)
	if len(name) < 3 {
		return nil, fmt.Errorf("role name must be at least 3 characters")
	}

	// Check if role with same name already exists
	existing, _ := s.roleRepo.GetByName(ctx, req.TenantID, name)
	if existing != nil {
		return nil, fmt.Errorf("role with name %s already exists", name)
	}

	// Create role
	role := &models.Role{
		ID:          uuid.New(),
		TenantID:    req.TenantID,
		Name:        name,
		Description: req.Description,
		IsSystem:    req.IsSystem,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return role, nil
}

// GetByID retrieves a role by ID
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	return role, nil
}

// Update updates an existing role
func (s *Service) Update(ctx context.Context, id uuid.UUID, req *UpdateRoleRequest) (*models.Role, error) {
	// Get existing role
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	// Update fields
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if len(name) < 3 {
			return nil, fmt.Errorf("role name must be at least 3 characters")
		}

		// Check if name is already taken by another role
		existing, _ := s.roleRepo.GetByName(ctx, role.TenantID, name)
		if existing != nil && existing.ID != id {
			return nil, fmt.Errorf("role name %s is already taken", name)
		}

		role.Name = name
	}

	if req.Description != nil {
		role.Description = req.Description
	}

	role.UpdatedAt = time.Now()

	if err := s.roleRepo.Update(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	return role, nil
}

// Delete deletes a role
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.roleRepo.Delete(ctx, id)
}

// List retrieves a list of roles
func (s *Service) List(ctx context.Context, tenantID uuid.UUID, filters *interfaces.RoleFilters) ([]*models.Role, error) {
	if tenantID == uuid.Nil {
		return nil, fmt.Errorf("tenant_id is required")
	}

	roles, err := s.roleRepo.List(ctx, tenantID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	return roles, nil
}

// GetUserRoles retrieves all roles for a user
func (s *Service) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*models.Role, error) {
	roles, err := s.roleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	return roles, nil
}

// AssignRoleToUser assigns a role to a user
func (s *Service) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	// Verify role exists
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Assign role
	if err := s.roleRepo.AssignRoleToUser(ctx, userID, roleID); err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	// Note: In production, you might want to verify user belongs to same tenant as role
	_ = role // Use role for tenant validation if needed

	return nil
}

// RemoveRoleFromUser removes a role from a user
func (s *Service) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return s.roleRepo.RemoveRoleFromUser(ctx, userID, roleID)
}

// GetRolePermissions retrieves all permissions for a role
func (s *Service) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*models.Permission, error) {
	permissions, err := s.permissionRepo.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	return permissions, nil
}

// AssignPermissionToRole assigns a permission to a role
func (s *Service) AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	// Verify role exists
	_, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Verify permission exists
	_, err = s.permissionRepo.GetByID(ctx, permissionID)
	if err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	// Assign permission
	if err := s.permissionRepo.AssignPermissionToRole(ctx, roleID, permissionID); err != nil {
		return fmt.Errorf("failed to assign permission: %w", err)
	}

	return nil
}

// RemovePermissionFromRole removes a permission from a role
func (s *Service) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	return s.permissionRepo.RemovePermissionFromRole(ctx, roleID, permissionID)
}

