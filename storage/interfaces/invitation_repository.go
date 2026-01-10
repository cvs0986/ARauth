package interfaces

import (
	"context"

	"github.com/arauth-identity/iam/identity/models"
	"github.com/google/uuid"
)

// InvitationRepository defines the interface for invitation storage
type InvitationRepository interface {
	// Create creates a new invitation
	Create(ctx context.Context, invitation *models.UserInvitation) error

	// GetByID retrieves an invitation by ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.UserInvitation, error)

	// GetByTokenHash retrieves an invitation by token hash
	GetByTokenHash(ctx context.Context, tokenHash string) (*models.UserInvitation, error)

	// GetByEmail retrieves an invitation by email and tenant ID
	GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*models.UserInvitation, error)

	// List lists invitations for a tenant
	List(ctx context.Context, tenantID uuid.UUID, filters *InvitationFilters) ([]*models.UserInvitation, error)

	// Count counts invitations for a tenant
	Count(ctx context.Context, tenantID uuid.UUID, filters *InvitationFilters) (int, error)

	// Update updates an invitation
	Update(ctx context.Context, invitation *models.UserInvitation) error

	// Delete soft-deletes an invitation
	Delete(ctx context.Context, id uuid.UUID) error
}

// InvitationFilters defines filters for listing invitations
type InvitationFilters struct {
	Email      string
	Status     string // "pending", "accepted", "expired"
	InvitedBy  *uuid.UUID
	Page       int
	PageSize   int
}

