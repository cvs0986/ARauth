package scim

import (
	"context"

	"github.com/arauth-identity/iam/identity/models"
	"github.com/google/uuid"
)

// TokenServiceInterface defines the interface for SCIM token management
type TokenServiceInterface interface {
	// CreateToken creates a new SCIM token
	CreateToken(ctx context.Context, tenantID uuid.UUID, req *CreateTokenRequest) (*models.SCIMToken, string, error) // Returns token and plaintext token

	// GetToken retrieves a SCIM token by ID
	GetToken(ctx context.Context, id uuid.UUID) (*models.SCIMToken, error)

	// ListTokens lists SCIM tokens for a tenant
	ListTokens(ctx context.Context, tenantID uuid.UUID) ([]*models.SCIMToken, error)

	// UpdateToken updates a SCIM token
	UpdateToken(ctx context.Context, id uuid.UUID, req *UpdateTokenRequest) (*models.SCIMToken, error)

	// DeleteToken deletes a SCIM token
	DeleteToken(ctx context.Context, id uuid.UUID) error

	// ValidateToken validates a SCIM token and returns the token if valid
	ValidateToken(ctx context.Context, tokenString string) (*models.SCIMToken, error)
}

// CreateTokenRequest represents a request to create a SCIM token
type CreateTokenRequest struct {
	Name      string    `json:"name" binding:"required"`
	Scopes    []string  `json:"scopes" binding:"required,min=1"`
	ExpiresAt *string   `json:"expires_at,omitempty"` // ISO 8601 format
}

// UpdateTokenRequest represents a request to update a SCIM token
type UpdateTokenRequest struct {
	Name      *string   `json:"name,omitempty"`
	Scopes    []string  `json:"scopes,omitempty"`
	ExpiresAt *string   `json:"expires_at,omitempty"` // ISO 8601 format
}

