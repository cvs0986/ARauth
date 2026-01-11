package interfaces

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// OAuthClient represents an OAuth2 client for machine-to-machine authentication
type OAuthClient struct {
	ID               uuid.UUID  `db:"id"`
	TenantID         uuid.UUID  `db:"tenant_id"`
	Name             string     `db:"name"`
	ClientID         string     `db:"client_id"`
	ClientSecretHash string     `db:"client_secret_hash"` // bcrypt hash, never plaintext
	Description      *string    `db:"description"`
	RedirectURIs     []string   `db:"redirect_uris"` // PostgreSQL array
	GrantTypes       []string   `db:"grant_types"`   // PostgreSQL array
	Scopes           []string   `db:"scopes"`        // PostgreSQL array
	IsConfidential   bool       `db:"is_confidential"`
	IsActive         bool       `db:"is_active"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`
	CreatedBy        *uuid.UUID `db:"created_by"`
}

// OAuthClientRepository defines operations for OAuth2 client management
type OAuthClientRepository interface {
	// Create creates a new OAuth2 client
	Create(ctx context.Context, client *OAuthClient) error

	// GetByID retrieves a client by ID
	GetByID(ctx context.Context, id uuid.UUID) (*OAuthClient, error)

	// GetByClientID retrieves a client by client_id (used during OAuth2 auth)
	GetByClientID(ctx context.Context, clientID string) (*OAuthClient, error)

	// ListByTenant retrieves all clients for a tenant
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*OAuthClient, error)

	// Update updates an existing client
	Update(ctx context.Context, client *OAuthClient) error

	// Delete deletes a client by ID
	Delete(ctx context.Context, id uuid.UUID) error
}
