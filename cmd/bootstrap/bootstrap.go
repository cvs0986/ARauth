package bootstrap

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/config"
	"github.com/arauth-identity/iam/identity/credential"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/security/password"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// BootstrapService handles system bootstrap operations
type BootstrapService struct {
	cfg            *config.BootstrapConfig
	userRepo       interfaces.UserRepository
	credentialRepo interfaces.CredentialRepository
	systemRoleRepo interfaces.SystemRoleRepository
}

// NewBootstrapService creates a new bootstrap service
func NewBootstrapService(
	cfg *config.BootstrapConfig,
	userRepo interfaces.UserRepository,
	credentialRepo interfaces.CredentialRepository,
	systemRoleRepo interfaces.SystemRoleRepository,
) *BootstrapService {
	return &BootstrapService{
		cfg:            cfg,
		userRepo:       userRepo,
		credentialRepo: credentialRepo,
		systemRoleRepo: systemRoleRepo,
	}
}

// Bootstrap creates the master system user if it doesn't exist
func (s *BootstrapService) Bootstrap(ctx context.Context) error {
	// Check if master user already exists
	existing, err := s.userRepo.GetByEmailSystem(ctx, s.cfg.MasterUser.Email)
	if err == nil && existing != nil && existing.PrincipalType == models.PrincipalTypeSystem {
		if !s.cfg.Force {
			return fmt.Errorf("master user already exists (use --force to re-bootstrap)")
		}
		// Force re-bootstrap: delete existing master user
		if err := s.userRepo.Delete(ctx, existing.ID); err != nil {
			return fmt.Errorf("failed to delete existing master user: %w", err)
		}
	}

	// Validate password is provided
	if s.cfg.MasterUser.Password == "" {
		return fmt.Errorf("master user password is required (set BOOTSTRAP_PASSWORD env var)")
	}

	// 1. Create Master User (tenant_id = NULL, principal_type = SYSTEM)
	masterUser := &models.User{
		ID:            uuid.New(),
		Username:      s.cfg.MasterUser.Username,
		Email:         s.cfg.MasterUser.Email,
		FirstName:     stringPtr(s.cfg.MasterUser.FirstName),
		LastName:      stringPtr(s.cfg.MasterUser.LastName),
		Status:        models.UserStatusActive,
		PrincipalType: models.PrincipalTypeSystem, // SYSTEM user
		TenantID:      nil,                         // No tenant
	}

	if err := s.userRepo.Create(ctx, masterUser); err != nil {
		return fmt.Errorf("failed to create master user: %w", err)
	}

	// 2. Set Master User Password
	hasher := password.NewHasher()
	passwordHash, err := hasher.Hash(s.cfg.MasterUser.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	cred := &credential.Credential{
		UserID:       masterUser.ID,
		PasswordHash: passwordHash,
	}

	if err := s.credentialRepo.Create(ctx, cred); err != nil {
		return fmt.Errorf("failed to create credentials: %w", err)
	}

	// 3. Assign system_owner role
	systemOwnerRole, err := s.systemRoleRepo.GetByName(ctx, "system_owner")
	if err != nil {
		return fmt.Errorf("failed to get system_owner role: %w", err)
	}

	if err := s.systemRoleRepo.AssignRoleToUser(ctx, masterUser.ID, systemOwnerRole.ID, nil); err != nil {
		return fmt.Errorf("failed to assign system_owner role: %w", err)
	}

	return nil
}

// stringPtr returns a pointer to the string
func stringPtr(s string) *string {
	return &s
}

