package linking

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/federation"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// Service provides identity linking functionality
type Service struct {
	fedIdRepo interfaces.FederatedIdentityRepository
	idpRepo   interfaces.IdentityProviderRepository
}

// NewService creates a new identity linking service
func NewService(
	fedIdRepo interfaces.FederatedIdentityRepository,
	idpRepo interfaces.IdentityProviderRepository,
) ServiceInterface {
	return &Service{
		fedIdRepo: fedIdRepo,
		idpRepo:   idpRepo,
	}
}

// LinkIdentity links a federated identity to a user
func (s *Service) LinkIdentity(ctx context.Context, userID uuid.UUID, providerID uuid.UUID, externalID string, attributes map[string]interface{}) error {
	// Check if identity already exists
	existing, err := s.fedIdRepo.GetByProviderAndExternalID(ctx, providerID, externalID)
	if err == nil && existing != nil {
		// Identity already linked to another user
		if existing.UserID != userID {
			return fmt.Errorf("identity already linked to another user")
		}
		// Already linked to this user, nothing to do
		return nil
	}

	// Get provider to check if it exists
	_, err = s.idpRepo.GetByID(ctx, providerID)
	if err != nil {
		return fmt.Errorf("identity provider not found: %w", err)
	}

	// Check if this will be the first identity for the user
	identities, err := s.fedIdRepo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user identities: %w", err)
	}

	// Create federated identity
	fedId := &federation.FederatedIdentity{
		ID:          uuid.New(),
		UserID:      userID,
		ProviderID:  providerID,
		ExternalID:  externalID,
		Attributes:  attributes,
		IsPrimary:   len(identities) == 0, // First identity is primary
		Verified:    false,
		CreatedAt:   time.Now(),
	}

	return s.fedIdRepo.Create(ctx, fedId)
}

// UnlinkIdentity unlinks a federated identity from a user
func (s *Service) UnlinkIdentity(ctx context.Context, userID uuid.UUID, federatedIdentityID uuid.UUID) error {
	// Get the federated identity
	fedId, err := s.fedIdRepo.GetByID(ctx, federatedIdentityID)
	if err != nil {
		return fmt.Errorf("federated identity not found: %w", err)
	}

	// Verify it belongs to the user
	if fedId.UserID != userID {
		return fmt.Errorf("federated identity does not belong to this user")
	}

	// Check if it's the primary identity
	if fedId.IsPrimary {
		// Get all other identities for this user
		identities, err := s.fedIdRepo.GetByUserID(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to get user identities: %w", err)
		}

		// Find a non-primary identity to promote
		var newPrimary *federation.FederatedIdentity
		for _, id := range identities {
			if id.ID != federatedIdentityID && !id.IsPrimary {
				newPrimary = id
				break
			}
		}

		// If there's another identity, make it primary
		if newPrimary != nil {
			newPrimary.IsPrimary = true
			if err := s.fedIdRepo.Update(ctx, newPrimary); err != nil {
				return fmt.Errorf("failed to set new primary identity: %w", err)
			}
		}
	}

	// Delete the federated identity
	return s.fedIdRepo.Delete(ctx, federatedIdentityID)
}

// SetPrimaryIdentity sets a federated identity as the primary identity for a user
func (s *Service) SetPrimaryIdentity(ctx context.Context, userID uuid.UUID, federatedIdentityID uuid.UUID) error {
	// Get the federated identity
	fedId, err := s.fedIdRepo.GetByID(ctx, federatedIdentityID)
	if err != nil {
		return fmt.Errorf("federated identity not found: %w", err)
	}

	// Verify it belongs to the user
	if fedId.UserID != userID {
		return fmt.Errorf("federated identity does not belong to this user")
	}

	// If already primary, nothing to do
	if fedId.IsPrimary {
		return nil
	}

	// Get all identities for this user
	identities, err := s.fedIdRepo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user identities: %w", err)
	}

	// Unset primary on all other identities
	for _, id := range identities {
		if id.IsPrimary && id.ID != federatedIdentityID {
			id.IsPrimary = false
			if err := s.fedIdRepo.Update(ctx, id); err != nil {
				return fmt.Errorf("failed to unset primary identity: %w", err)
			}
		}
	}

	// Set this identity as primary
	fedId.IsPrimary = true
	return s.fedIdRepo.Update(ctx, fedId)
}

// GetUserIdentities retrieves all linked identities for a user
func (s *Service) GetUserIdentities(ctx context.Context, userID uuid.UUID) ([]*IdentityInfo, error) {
	// Get all federated identities for the user
	fedIds, err := s.fedIdRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user identities: %w", err)
	}

	// Convert to IdentityInfo
	identities := make([]*IdentityInfo, 0, len(fedIds))
	for _, fedId := range fedIds {
		// Get provider info
		provider, err := s.idpRepo.GetByID(ctx, fedId.ProviderID)
		if err != nil {
			// Skip if provider not found
			continue
		}

		var verifiedAt *string
		if fedId.VerifiedAt != nil {
			verifiedAtStr := fedId.VerifiedAt.Format(time.RFC3339)
			verifiedAt = &verifiedAtStr
		}

		identities = append(identities, &IdentityInfo{
			ID:           fedId.ID,
			ProviderID:   fedId.ProviderID,
			ProviderName: provider.Name,
			ProviderType: string(provider.Type),
			ExternalID:   fedId.ExternalID,
			IsPrimary:    fedId.IsPrimary,
			Verified:     fedId.Verified,
			VerifiedAt:   verifiedAt,
			Attributes:   fedId.Attributes,
			CreatedAt:    fedId.CreatedAt.Format(time.RFC3339),
		})
	}

	return identities, nil
}

// VerifyIdentity marks a federated identity as verified
func (s *Service) VerifyIdentity(ctx context.Context, federatedIdentityID uuid.UUID) error {
	// Get the federated identity
	fedId, err := s.fedIdRepo.GetByID(ctx, federatedIdentityID)
	if err != nil {
		return fmt.Errorf("federated identity not found: %w", err)
	}

	// Mark as verified
	fedId.Verified = true
	now := time.Now()
	fedId.VerifiedAt = &now

	return s.fedIdRepo.Update(ctx, fedId)
}

