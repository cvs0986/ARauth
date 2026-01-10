package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/federation"
)

// IdentityProviderRepository defines the interface for identity provider data access
type IdentityProviderRepository interface {
	// Create creates a new identity provider
	Create(ctx context.Context, provider *federation.IdentityProvider) error

	// GetByID retrieves an identity provider by ID
	GetByID(ctx context.Context, id uuid.UUID) (*federation.IdentityProvider, error)

	// GetByTenantID retrieves all identity providers for a tenant
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*federation.IdentityProvider, error)

	// GetByName retrieves an identity provider by tenant ID and name
	GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*federation.IdentityProvider, error)

	// Update updates an existing identity provider
	Update(ctx context.Context, provider *federation.IdentityProvider) error

	// Delete soft deletes an identity provider
	Delete(ctx context.Context, id uuid.UUID) error
}

// FederatedIdentityRepository defines the interface for federated identity data access
type FederatedIdentityRepository interface {
	// Create creates a new federated identity
	Create(ctx context.Context, identity *federation.FederatedIdentity) error

	// GetByID retrieves a federated identity by ID
	GetByID(ctx context.Context, id uuid.UUID) (*federation.FederatedIdentity, error)

	// GetByUserID retrieves all federated identities for a user
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*federation.FederatedIdentity, error)

	// GetByProviderAndExternalID retrieves a federated identity by provider and external ID
	GetByProviderAndExternalID(ctx context.Context, providerID uuid.UUID, externalID string) (*federation.FederatedIdentity, error)

	// GetByProviderID retrieves all federated identities for a provider
	GetByProviderID(ctx context.Context, providerID uuid.UUID) ([]*federation.FederatedIdentity, error)

	// Update updates an existing federated identity
	Update(ctx context.Context, identity *federation.FederatedIdentity) error

	// Delete deletes a federated identity
	Delete(ctx context.Context, id uuid.UUID) error

	// SetPrimary sets a federated identity as primary (and unsets others for the user)
	SetPrimary(ctx context.Context, userID uuid.UUID, identityID uuid.UUID) error
}

