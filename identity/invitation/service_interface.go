package invitation

import (
	"context"

	"github.com/arauth-identity/iam/identity/models"
	"github.com/google/uuid"
)

// ServiceInterface defines the interface for invitation management
type ServiceInterface interface {
	// CreateInvitation creates a new user invitation
	CreateInvitation(ctx context.Context, tenantID uuid.UUID, invitedBy uuid.UUID, req *CreateInvitationRequest) (*models.UserInvitation, error)

	// GetInvitation retrieves an invitation by ID
	GetInvitation(ctx context.Context, id uuid.UUID) (*models.UserInvitation, error)

	// GetInvitationByToken retrieves an invitation by token
	GetInvitationByToken(ctx context.Context, token string) (*models.UserInvitation, error)

	// ListInvitations lists invitations for a tenant
	ListInvitations(ctx context.Context, tenantID uuid.UUID, filters *ListInvitationsFilters) ([]*models.UserInvitation, int, error)

	// ResendInvitation resends an invitation email
	ResendInvitation(ctx context.Context, id uuid.UUID) error

	// AcceptInvitation accepts an invitation and creates a user account
	AcceptInvitation(ctx context.Context, token string, req *AcceptInvitationRequest) (*models.User, error)

	// DeleteInvitation deletes an invitation
	DeleteInvitation(ctx context.Context, id uuid.UUID) error
}

// CreateInvitationRequest represents a request to create an invitation
type CreateInvitationRequest struct {
	Email     string      `json:"email" binding:"required,email"`
	RoleIDs   []uuid.UUID `json:"role_ids,omitempty"`
	ExpiresIn int         `json:"expires_in,omitempty"` // Days until expiration (default: 7)
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// AcceptInvitationRequest represents a request to accept an invitation
type AcceptInvitationRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=255"`
	Password  string `json:"password" binding:"required,min=12"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
}

// ListInvitationsFilters defines filters for listing invitations
type ListInvitationsFilters struct {
	Email     string
	Status    string // "pending", "accepted", "expired"
	InvitedBy *uuid.UUID
	Page      int
	PageSize  int
}

