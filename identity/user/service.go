package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/identity/models"
	"github.com/nuage-identity/iam/security/password"
	"github.com/nuage-identity/iam/storage/interfaces"
)

// Service provides user management business logic
type Service struct {
	repo            interfaces.UserRepository
	passwordValidator *password.Validator
}

// NewService creates a new user service
func NewService(repo interfaces.UserRepository) *Service {
	// Default password policy: min 12 chars, require all complexity
	validator := password.NewValidator(12, true, true, true, true)
	return &Service{
		repo:              repo,
		passwordValidator: validator,
	}
}

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	TenantID  uuid.UUID              `json:"tenant_id" binding:"required"`
	Username  string                 `json:"username" binding:"required,min=3,max=255"`
	Email     string                 `json:"email" binding:"required,email"`
	FirstName *string                `json:"first_name,omitempty"`
	LastName  *string                `json:"last_name,omitempty"`
	Status    string                 `json:"status,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	Username  *string                `json:"username,omitempty"`
	Email     *string                `json:"email,omitempty"`
	FirstName *string                `json:"first_name,omitempty"`
	LastName  *string                `json:"last_name,omitempty"`
	Status    *string                `json:"status,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Create creates a new user
func (s *Service) Create(ctx context.Context, req *CreateUserRequest) (*models.User, error) {
	// Validate tenant ID is provided
	if req.TenantID == uuid.Nil {
		return nil, fmt.Errorf("tenant_id is required")
	}

	// Validate username
	req.Username = strings.TrimSpace(req.Username)
	if len(req.Username) < 3 {
		return nil, fmt.Errorf("username must be at least 3 characters")
	}

	// Validate email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if !isValidEmail(req.Email) {
		return nil, fmt.Errorf("invalid email format")
	}

	// Note: Password validation should be done when setting password
	// This service doesn't handle password directly (handled by credential service)

	// Check if user already exists
	existing, _ := s.repo.GetByUsername(ctx, req.Username, req.TenantID)
	if existing != nil {
		return nil, fmt.Errorf("username already exists")
	}

	existing, _ = s.repo.GetByEmail(ctx, req.Email, req.TenantID)
	if existing != nil {
		return nil, fmt.Errorf("email already exists")
	}

	// Set default status
	status := req.Status
	if status == "" {
		status = models.UserStatusActive
	}

	// Create user
	u := &models.User{
		ID:        uuid.New(),
		TenantID:  req.TenantID,
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Status:    status,
		Metadata:  req.Metadata,
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return u, nil
}

// GetByID retrieves a user by ID
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return u, nil
}

// GetByUsername retrieves a user by username
func (s *Service) GetByUsername(ctx context.Context, username string, tenantID uuid.UUID) (*models.User, error) {
	u, err := s.repo.GetByUsername(ctx, username, tenantID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return u, nil
}

// Update updates an existing user
func (s *Service) Update(ctx context.Context, id uuid.UUID, req *UpdateUserRequest) (*models.User, error) {
	// Get existing user
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Update fields if provided
	if req.Username != nil {
		username := strings.TrimSpace(*req.Username)
		if len(username) < 3 {
			return nil, fmt.Errorf("username must be at least 3 characters")
		}
		// Check if username is already taken by another user
		existing, _ := s.repo.GetByUsername(ctx, username, u.TenantID)
		if existing != nil && existing.ID != id {
			return nil, fmt.Errorf("username already exists")
		}
		u.Username = username
	}

	if req.Email != nil {
		email := strings.ToLower(strings.TrimSpace(*req.Email))
		if !isValidEmail(email) {
			return nil, fmt.Errorf("invalid email format")
		}
		// Check if email is already taken by another user
		existing, _ := s.repo.GetByEmail(ctx, email, u.TenantID)
		if existing != nil && existing.ID != id {
			return nil, fmt.Errorf("email already exists")
		}
		u.Email = email
	}

	if req.FirstName != nil {
		u.FirstName = req.FirstName
	}

	if req.LastName != nil {
		u.LastName = req.LastName
	}

	if req.Status != nil {
		u.Status = *req.Status
	}

	if req.Metadata != nil {
		u.Metadata = req.Metadata
	}

	// Save updates
	if err := s.repo.Update(ctx, u); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return u, nil
}

// Delete soft deletes a user
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// List retrieves a list of users (tenant-scoped)
func (s *Service) List(ctx context.Context, tenantID uuid.UUID, filters *interfaces.UserFilters) ([]*models.User, error) {
	// Ensure tenant ID is provided
	if tenantID == uuid.Nil {
		return nil, fmt.Errorf("tenant_id is required")
	}
	return s.repo.List(ctx, tenantID, filters)
}

// Count returns the total count of users
func (s *Service) Count(ctx context.Context, tenantID uuid.UUID, filters *interfaces.UserFilters) (int, error) {
	return s.repo.Count(ctx, tenantID, filters)
}

// isValidEmail performs basic email validation
func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

