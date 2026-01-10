package linking

import (
	"context"

	"github.com/google/uuid"
)

// ServiceInterface defines the interface for identity linking operations
type ServiceInterface interface {
	// LinkIdentity links a federated identity to a user
	LinkIdentity(ctx context.Context, userID uuid.UUID, providerID uuid.UUID, externalID string, attributes map[string]interface{}) error

	// UnlinkIdentity unlinks a federated identity from a user
	UnlinkIdentity(ctx context.Context, userID uuid.UUID, federatedIdentityID uuid.UUID) error

	// SetPrimaryIdentity sets a federated identity as the primary identity for a user
	SetPrimaryIdentity(ctx context.Context, userID uuid.UUID, federatedIdentityID uuid.UUID) error

	// GetUserIdentities retrieves all linked identities for a user
	GetUserIdentities(ctx context.Context, userID uuid.UUID) ([]*IdentityInfo, error)

	// VerifyIdentity marks a federated identity as verified
	VerifyIdentity(ctx context.Context, federatedIdentityID uuid.UUID) error
}

// IdentityInfo represents information about a linked identity
type IdentityInfo struct {
	ID            uuid.UUID              `json:"id"`
	ProviderID    uuid.UUID              `json:"provider_id"`
	ProviderName  string                 `json:"provider_name"`
	ProviderType  string                 `json:"provider_type"`
	ExternalID    string                 `json:"external_id"`
	IsPrimary     bool                   `json:"is_primary"`
	Verified      bool                   `json:"verified"`
	VerifiedAt    *string                `json:"verified_at,omitempty"`
	Attributes    map[string]interface{} `json:"attributes,omitempty"`
	CreatedAt     string                 `json:"created_at"`
}

