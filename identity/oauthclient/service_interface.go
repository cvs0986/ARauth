package oauthclient

import (
	"context"

	"github.com/google/uuid"
)

// ServiceInterface defines operations for OAuth2 client management
type ServiceInterface interface {
	// CreateClient creates a new OAuth2 client with generated credentials
	CreateClient(ctx context.Context, tenantID uuid.UUID, req *CreateClientRequest, createdBy uuid.UUID) (*CreateClientResponse, error)

	// ListClients retrieves all clients for a tenant (WITHOUT secrets)
	ListClients(ctx context.Context, tenantID uuid.UUID) ([]*Client, error)

	// GetClient retrieves a single client (WITHOUT secret)
	GetClient(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*Client, error)

	// RotateSecret generates a new secret and invalidates the old one
	RotateSecret(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*RotateSecretResponse, error)

	// DeleteClient deletes a client
	DeleteClient(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error
}
