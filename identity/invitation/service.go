package invitation

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/identity/user"
	"github.com/arauth-identity/iam/identity/role"
	"github.com/arauth-identity/iam/internal/email"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// Service provides invitation management
type Service struct {
	invitationRepo interfaces.InvitationRepository
	userService    user.ServiceInterface
	roleService    role.ServiceInterface
	userRepo       interfaces.UserRepository
	emailService   email.ServiceInterface
	tenantRepo     interfaces.TenantRepository
}

// NewService creates a new invitation service
func NewService(
	invitationRepo interfaces.InvitationRepository,
	userService user.ServiceInterface,
	roleService role.ServiceInterface,
	userRepo interfaces.UserRepository,
	emailService email.ServiceInterface,
	tenantRepo interfaces.TenantRepository,
) ServiceInterface {
	return &Service{
		invitationRepo: invitationRepo,
		userService:    userService,
		roleService:    roleService,
		userRepo:       userRepo,
		emailService:   emailService,
		tenantRepo:     tenantRepo,
	}
}

// generateToken generates a secure random invitation token
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// hashToken creates a SHA256 hash for fast token lookup
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// CreateInvitation creates a new user invitation
func (s *Service) CreateInvitation(ctx context.Context, tenantID uuid.UUID, invitedBy uuid.UUID, req *CreateInvitationRequest) (*models.UserInvitation, error) {
	// Validate email
	email := req.Email
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	// Check if user already exists
	existingUser, _ := s.userRepo.GetByEmail(ctx, email, tenantID)
	if existingUser != nil {
		return nil, fmt.Errorf("user with this email already exists")
	}

	// Check if there's a pending invitation for this email
	existingInvitation, _ := s.invitationRepo.GetByEmail(ctx, tenantID, email)
	if existingInvitation != nil && existingInvitation.IsValid() {
		return nil, fmt.Errorf("pending invitation already exists for this email")
	}

	// Generate invitation token
	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate invitation token: %w", err)
	}

	tokenHash := hashToken(token)

	// Set expiration (default: 7 days)
	expiresIn := req.ExpiresIn
	if expiresIn == 0 {
		expiresIn = 7
	}
	expiresAt := time.Now().AddDate(0, 0, expiresIn)

	// Create invitation
	invitation := &models.UserInvitation{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Email:     email,
		InvitedBy: invitedBy,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		RoleIDs:   req.RoleIDs,
		Metadata:  req.Metadata,
	}

	if err := s.invitationRepo.Create(ctx, invitation); err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// Set token for response (only returned on creation)
	invitation.Token = token

	// Get tenant name for email
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err == nil && tenant != nil {
		// Send invitation email
		_ = s.emailService.SendInvitationEmail(ctx, email, token, tenant.Name, expiresAt.Format(time.RFC3339))
	}

	return invitation, nil
}

// GetInvitation retrieves an invitation by ID
func (s *Service) GetInvitation(ctx context.Context, id uuid.UUID) (*models.UserInvitation, error) {
	return s.invitationRepo.GetByID(ctx, id)
}

// GetInvitationByToken retrieves an invitation by token
func (s *Service) GetInvitationByToken(ctx context.Context, token string) (*models.UserInvitation, error) {
	tokenHash := hashToken(token)
	return s.invitationRepo.GetByTokenHash(ctx, tokenHash)
}

// ListInvitations lists invitations for a tenant
func (s *Service) ListInvitations(ctx context.Context, tenantID uuid.UUID, filters *ListInvitationsFilters) ([]*models.UserInvitation, int, error) {
	internalFilters := &interfaces.InvitationFilters{
		Email:     filters.Email,
		Status:    filters.Status,
		InvitedBy: filters.InvitedBy,
		Page:      filters.Page,
		PageSize:  filters.PageSize,
	}

	invitations, err := s.invitationRepo.List(ctx, tenantID, internalFilters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list invitations: %w", err)
	}

	total, err := s.invitationRepo.Count(ctx, tenantID, internalFilters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count invitations: %w", err)
	}

	return invitations, total, nil
}

// ResendInvitation resends an invitation email
func (s *Service) ResendInvitation(ctx context.Context, id uuid.UUID) error {
	invitation, err := s.invitationRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("invitation not found: %w", err)
	}

	if !invitation.IsValid() {
		return fmt.Errorf("invitation is not valid (expired, accepted, or deleted)")
	}

	// Generate new token
	token, err := generateToken()
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	tokenHash := hashToken(token)
	invitation.TokenHash = tokenHash

	// Update invitation
	if err := s.invitationRepo.Update(ctx, invitation); err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}

	// Get tenant name for email
	tenant, err := s.tenantRepo.GetByID(ctx, invitation.TenantID)
	if err == nil && tenant != nil {
		// Send invitation email
		_ = s.emailService.SendInvitationEmail(ctx, invitation.Email, token, tenant.Name, invitation.ExpiresAt.Format(time.RFC3339))
	}

	return nil
}

// AcceptInvitation accepts an invitation and creates a user account
func (s *Service) AcceptInvitation(ctx context.Context, token string, req *AcceptInvitationRequest) (*models.User, error) {
	// Get invitation by token
	invitation, err := s.GetInvitationByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid invitation token")
	}

	// Validate invitation
	if !invitation.IsValid() {
		return nil, fmt.Errorf("invitation is not valid (expired, accepted, or deleted)")
	}

	// Check if user already exists
	existingUser, _ := s.userRepo.GetByEmail(ctx, invitation.Email, invitation.TenantID)
	if existingUser != nil {
		return nil, fmt.Errorf("user with this email already exists")
	}

	// Create user
	createReq := &user.CreateUserRequest{
		TenantID:  invitation.TenantID,
		Username:  req.Username,
		Email:     invitation.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Status:    "active",
	}

	createdUser, err := s.userService.Create(ctx, createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Assign roles if specified
	for _, roleID := range invitation.RoleIDs {
		_ = s.roleService.AssignRoleToUser(ctx, createdUser.ID, roleID)
	}

	// Mark invitation as accepted
	now := time.Now()
	invitation.AcceptedAt = &now
	invitation.AcceptedBy = &createdUser.ID

	if err := s.invitationRepo.Update(ctx, invitation); err != nil {
		// Log error but don't fail the request
		_ = err
	}

	// Send welcome email
	_ = s.emailService.SendWelcomeEmail(ctx, createdUser.Email, createdUser.Username)

	return createdUser, nil
}

// DeleteInvitation deletes an invitation
func (s *Service) DeleteInvitation(ctx context.Context, id uuid.UUID) error {
	return s.invitationRepo.Delete(ctx, id)
}

