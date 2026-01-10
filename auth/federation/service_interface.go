package federation

import (
	"context"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/federation"
)

// ServiceInterface defines the interface for federation service operations
type ServiceInterface interface {
	// Identity Provider Management
	CreateIdentityProvider(ctx context.Context, tenantID uuid.UUID, req *CreateIdPRequest) (*federation.IdentityProvider, error)
	GetIdentityProvider(ctx context.Context, id uuid.UUID) (*federation.IdentityProvider, error)
	GetIdentityProvidersByTenant(ctx context.Context, tenantID uuid.UUID) ([]*federation.IdentityProvider, error)
	UpdateIdentityProvider(ctx context.Context, id uuid.UUID, req *UpdateIdPRequest) (*federation.IdentityProvider, error)
	DeleteIdentityProvider(ctx context.Context, id uuid.UUID) error

	// OIDC Flow
	InitiateOIDCLogin(ctx context.Context, tenantID uuid.UUID, providerID uuid.UUID, redirectURI string) (string, string, error) // Returns auth URL and state
	HandleOIDCCallback(ctx context.Context, providerID uuid.UUID, code, state, redirectURI string) (*LoginResponse, error)

	// SAML Flow
	InitiateSAMLLogin(ctx context.Context, tenantID uuid.UUID, providerID uuid.UUID, acsURL string) (string, error) // Returns redirect URL
	HandleSAMLCallback(ctx context.Context, providerID uuid.UUID, samlResponse, relayState string) (*LoginResponse, error)
}

// CreateIdPRequest represents a request to create an identity provider
type CreateIdPRequest struct {
	Name            string                 `json:"name" binding:"required"`
	Type            federation.IdentityProviderType `json:"type" binding:"required"`
	Enabled         bool                   `json:"enabled"`
	Configuration   map[string]interface{} `json:"configuration" binding:"required"`
	AttributeMapping map[string]interface{} `json:"attribute_mapping,omitempty"`
}

// UpdateIdPRequest represents a request to update an identity provider
type UpdateIdPRequest struct {
	Name            *string                `json:"name,omitempty"`
	Enabled         *bool                  `json:"enabled,omitempty"`
	Configuration   map[string]interface{} `json:"configuration,omitempty"`
	AttributeMapping map[string]interface{} `json:"attribute_mapping,omitempty"`
}

// LoginResponse represents the response from a federated login
type LoginResponse struct {
	UserID      uuid.UUID `json:"user_id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	FirstName   string    `json:"first_name,omitempty"`
	LastName    string    `json:"last_name,omitempty"`
	TenantID    uuid.UUID `json:"tenant_id,omitempty"`
	IsNewUser   bool      `json:"is_new_user"`
	AccessToken string    `json:"access_token,omitempty"`
	IDToken     string    `json:"id_token,omitempty"`
}

