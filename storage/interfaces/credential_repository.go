package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/identity/credential"
)

// CredentialRepository defines the interface for credential data access
type CredentialRepository interface {
	// Create creates a new credential
	Create(ctx context.Context, cred *credential.Credential) error

	// GetByUserID retrieves credentials by user ID
	GetByUserID(ctx context.Context, userID uuid.UUID) (*credential.Credential, error)

	// Update updates existing credentials
	Update(ctx context.Context, cred *credential.Credential) error

	// Delete deletes credentials
	Delete(ctx context.Context, userID uuid.UUID) error
}

