package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/credential"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/security/password"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// Service provides user management business logic
type Service struct {
	repo              interfaces.UserRepository
	credentialRepo    interfaces.CredentialRepository
	passwordValidator *password.Validator
	passwordHasher    *password.Hasher
}

// NewService creates a new user service
func NewService(repo interfaces.UserRepository, credentialRepo interfaces.CredentialRepository) *Service {
	// Default password policy: min 12 chars, require all complexity
	validator := password.NewValidator(12, true, true, true, true)
	return &Service{
		repo:              repo,
		credentialRepo:    credentialRepo,
		passwordValidator: validator,
		passwordHasher:    password.NewHasher(),
	}
}

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	TenantID  uuid.UUID              `json:"tenant_id"` // Set from context, not from request body
	Username  string                 `json:"username" binding:"required,min=3,max=255"`
	Email     string                 `json:"email" binding:"required,email"`
	Password  string                 `json:"password" binding:"required,min=12"` // Password is required
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

	// Validate password
	if req.Password == "" {
		return nil, fmt.Errorf("password is required")
	}
	if err := s.passwordValidator.Validate(req.Password, req.Email); err != nil {
		return nil, fmt.Errorf("password validation failed: %w", err)
	}

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
		ID:            uuid.New(),
		TenantID:      &req.TenantID, // Convert to pointer
		PrincipalType: models.PrincipalTypeTenant, // Default to TENANT
		Username:      req.Username,
		Email:         req.Email,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Status:        status,
		Metadata:      req.Metadata,
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create credentials for the user
	passwordHash, err := s.passwordHasher.Hash(req.Password)
	if err != nil {
		// If credential creation fails, we should rollback user creation
		// For now, we'll just return an error (in production, use transactions)
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	cred := &credential.Credential{
		ID:                uuid.New(),
		UserID:            u.ID,
		PasswordHash:      passwordHash,
		PasswordChangedAt: time.Now(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.credentialRepo.Create(ctx, cred); err != nil {
		// If credential creation fails, we should rollback user creation
		// For now, we'll just return an error (in production, use transactions)
		return nil, fmt.Errorf("failed to create credentials: %w", err)
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
		if u.TenantID != nil {
			existing, _ := s.repo.GetByUsername(ctx, username, *u.TenantID)
			if existing != nil && existing.ID != id {
				return nil, fmt.Errorf("username already exists")
			}
		}
		u.Username = username
	}

	if req.Email != nil {
		email := strings.ToLower(strings.TrimSpace(*req.Email))
		if !isValidEmail(email) {
			return nil, fmt.Errorf("invalid email format")
		}
		// Check if email is already taken by another user
		if u.TenantID != nil {
			existing, _ := s.repo.GetByEmail(ctx, email, *u.TenantID)
			if existing != nil && existing.ID != id {
				return nil, fmt.Errorf("email already exists")
			}
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

