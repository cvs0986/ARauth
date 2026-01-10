package tenant

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// Service provides tenant management business logic
type Service struct {
	repo        interfaces.TenantRepository
	initializer *Initializer
}

// NewService creates a new tenant service
func NewService(repo interfaces.TenantRepository, initializer *Initializer) *Service {
	return &Service{
		repo:        repo,
		initializer: initializer,
	}
}

// CreateTenantRequest represents a request to create a tenant
type CreateTenantRequest struct {
	Name     string                 `json:"name" binding:"required,min=3,max=255"`
	Domain   string                 `json:"domain" binding:"required,min=3,max=255"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateTenantRequest represents a request to update a tenant
type UpdateTenantRequest struct {
	Name     *string                `json:"name,omitempty"`
	Domain   *string                `json:"domain,omitempty"`
	Status   *string                `json:"status,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Create creates a new tenant
func (s *Service) Create(ctx context.Context, req *CreateTenantRequest) (*models.Tenant, error) {
	// Normalize domain (lowercase, trim)
	domain := strings.ToLower(strings.TrimSpace(req.Domain))
	if domain == "" {
		return nil, fmt.Errorf("domain cannot be empty")
	}

	// Validate domain format (basic validation)
	if !isValidDomain(domain) {
		return nil, fmt.Errorf("invalid domain format")
	}

	// Check if domain already exists
	existing, _ := s.repo.GetByDomain(ctx, domain)
	if existing != nil {
		return nil, fmt.Errorf("tenant with domain %s already exists", domain)
	}

	// Create tenant
	tenant := &models.Tenant{
		ID:        uuid.New(),
		Name:      strings.TrimSpace(req.Name),
		Domain:    domain,
		Status:    models.TenantStatusActive,
		Metadata:  req.Metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Initialize predefined roles and permissions for the new tenant
	if s.initializer != nil {
		initResult, err := s.initializer.InitializeTenant(ctx, tenant.ID)
		if err != nil {
			// Log error but don't fail tenant creation
			// In production, you might want to rollback or handle this differently
			// For now, we'll allow tenant creation to succeed even if initialization fails
			// The initialization can be retried later if needed
			return nil, fmt.Errorf("tenant created but failed to initialize roles and permissions: %w", err)
		}
		// Store initialization result in tenant metadata for reference
		if tenant.Metadata == nil {
			tenant.Metadata = make(map[string]interface{})
		}
		tenant.Metadata["initialized_roles"] = true
		tenant.Metadata["tenant_owner_role_id"] = initResult.TenantOwnerRoleID.String()
		_ = initResult // Can be used later for assigning roles to first user
	}

	return tenant, nil
}

// GetByID retrieves a tenant by ID
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	tenant, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	if !tenant.IsActive() {
		return nil, fmt.Errorf("tenant is not active")
	}

	return tenant, nil
}

// GetByDomain retrieves a tenant by domain
func (s *Service) GetByDomain(ctx context.Context, domain string) (*models.Tenant, error) {
	domain = strings.ToLower(strings.TrimSpace(domain))
	tenant, err := s.repo.GetByDomain(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	if !tenant.IsActive() {
		return nil, fmt.Errorf("tenant is not active")
	}

	return tenant, nil
}

// Update updates an existing tenant
func (s *Service) Update(ctx context.Context, id uuid.UUID, req *UpdateTenantRequest) (*models.Tenant, error) {
	// Get existing tenant
	tenant, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	// Update fields
	if req.Name != nil {
		tenant.Name = strings.TrimSpace(*req.Name)
	}

	if req.Domain != nil {
		domain := strings.ToLower(strings.TrimSpace(*req.Domain))
		if domain == "" {
			return nil, fmt.Errorf("domain cannot be empty")
		}

		if !isValidDomain(domain) {
			return nil, fmt.Errorf("invalid domain format")
		}

		// Check if domain is already taken by another tenant
		existing, _ := s.repo.GetByDomain(ctx, domain)
		if existing != nil && existing.ID != id {
			return nil, fmt.Errorf("domain %s is already taken", domain)
		}

		tenant.Domain = domain
	}

	if req.Status != nil {
		// Validate status
		validStatuses := []string{
			models.TenantStatusActive,
			models.TenantStatusSuspended,
			models.TenantStatusDeleted,
		}
		valid := false
		for _, vs := range validStatuses {
			if *req.Status == vs {
				valid = true
				break
			}
		}
		if !valid {
			return nil, fmt.Errorf("invalid status: %s", *req.Status)
		}
		tenant.Status = *req.Status
	}

	if req.Metadata != nil {
		tenant.Metadata = req.Metadata
	}

	tenant.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	return tenant, nil
}

// Delete soft deletes a tenant
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	// Verify tenant exists
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}

	// Soft delete
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	// Note: In production, you might want to:
	// - Check if tenant has active users
	// - Archive tenant data
	// - Send notifications

	return nil
}

// List retrieves a list of tenants with filters
func (s *Service) List(ctx context.Context, filters *interfaces.TenantFilters) ([]*models.Tenant, error) {
	tenants, err := s.repo.List(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	return tenants, nil
}

// isValidDomain performs basic domain validation
func isValidDomain(domain string) bool {
	if len(domain) < 3 || len(domain) > 255 {
		return false
	}

	// Basic validation: alphanumeric, dots, hyphens
	for _, char := range domain {
		if !((char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '.' || char == '-') {
			return false
		}
	}

	// Must contain at least one dot
	if !strings.Contains(domain, ".") {
		return false
	}

	return true
}

