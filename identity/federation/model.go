package federation

import (
	"time"

	"github.com/google/uuid"
)

// IdentityProviderType represents the type of identity provider
type IdentityProviderType string

const (
	IdentityProviderTypeOIDC IdentityProviderType = "oidc"
	IdentityProviderTypeSAML IdentityProviderType = "saml"
)

// IdentityProvider represents an external identity provider configuration
type IdentityProvider struct {
	ID              uuid.UUID              `json:"id" db:"id"`
	TenantID        uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	Name            string                 `json:"name" db:"name"`
	Type            IdentityProviderType   `json:"type" db:"type"`
	Enabled         bool                   `json:"enabled" db:"enabled"`
	Configuration   map[string]interface{} `json:"configuration" db:"configuration"`
	AttributeMapping map[string]interface{} `json:"attribute_mapping,omitempty" db:"attribute_mapping"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
	DeletedAt       *time.Time             `json:"deleted_at,omitempty" db:"deleted_at"`
}

// OIDCConfiguration represents OIDC provider configuration
type OIDCConfiguration struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	IssuerURL    string   `json:"issuer_url"`
	AuthURL      string   `json:"auth_url"`
	TokenURL     string   `json:"token_url"`
	UserInfoURL  string   `json:"userinfo_url"`
	Scopes       []string `json:"scopes"`
}

// SAMLConfiguration represents SAML provider configuration
type SAMLConfiguration struct {
	EntityID          string `json:"entity_id"`
	SSOURL            string `json:"sso_url"`
	SLOURL            string `json:"slo_url,omitempty"`
	X509Certificate   string `json:"x509_certificate"`
	SignRequests      bool   `json:"sign_requests"`
	SignAssertions    bool   `json:"sign_assertions"`
	WantAssertionsSigned bool `json:"want_assertions_signed"`
}

// FederatedIdentity represents a link between a user and an external identity
type FederatedIdentity struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	UserID      uuid.UUID              `json:"user_id" db:"user_id"`
	ProviderID  uuid.UUID              `json:"provider_id" db:"provider_id"`
	ExternalID  string                 `json:"external_id" db:"external_id"`
	Attributes  map[string]interface{} `json:"attributes,omitempty" db:"attributes"`
	IsPrimary   bool                   `json:"is_primary" db:"is_primary"`
	Verified    bool                   `json:"verified" db:"verified"`
	VerifiedAt  *time.Time             `json:"verified_at,omitempty" db:"verified_at"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// AttributeMapping defines how to map provider attributes to user attributes
type AttributeMapping struct {
	Email       string `json:"email,omitempty"`
	Username    string `json:"username,omitempty"`
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	Phone       string `json:"phone,omitempty"`
}

