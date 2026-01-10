package introspection

import (
	"context"

	"github.com/google/uuid"
)

// ServiceInterface defines the interface for token introspection
type ServiceInterface interface {
	// IntrospectToken introspects a token and returns its metadata
	// Implements RFC 7662 OAuth 2.0 Token Introspection
	IntrospectToken(ctx context.Context, token string, tokenTypeHint string) (*TokenInfo, error)
}

// TokenInfo represents token introspection response (RFC 7662)
type TokenInfo struct {
	Active    bool   `json:"active"`              // REQUIRED: Whether the token is active
	Scope     string `json:"scope,omitempty"`      // OPTIONAL: Space-separated list of scopes
	ClientID  string `json:"client_id,omitempty"` // OPTIONAL: Client identifier
	Username  string `json:"username,omitempty"`   // OPTIONAL: Username
	ExpiresAt int64  `json:"exp,omitempty"`       // OPTIONAL: Expiration timestamp
	IssuedAt  int64  `json:"iat,omitempty"`       // OPTIONAL: Issuance timestamp
	NotBefore int64  `json:"nbf,omitempty"`       // OPTIONAL: Not before timestamp
	Subject   string `json:"sub,omitempty"`      // OPTIONAL: Subject (user ID)
	Audience  string `json:"aud,omitempty"`      // OPTIONAL: Audience
	Issuer    string `json:"iss,omitempty"`      // OPTIONAL: Issuer
	JTI       string `json:"jti,omitempty"`      // OPTIONAL: JWT ID

	// ARauth-specific extensions
	TenantID      string   `json:"tenant_id,omitempty"`       // Tenant ID (if tenant user)
	PrincipalType string   `json:"principal_type,omitempty"`  // SYSTEM or TENANT
	Roles         []string `json:"roles,omitempty"`           // User roles
	Permissions   []string `json:"permissions,omitempty"`     // User permissions
	SystemRoles   []string `json:"system_roles,omitempty"`    // System roles (SYSTEM users)
	SystemPerms   []string `json:"system_permissions,omitempty"` // System permissions (SYSTEM users)
}

