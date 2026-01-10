package models

import (
	"time"

	"github.com/google/uuid"
)

// UserInvitation represents a user invitation
type UserInvitation struct {
	ID         uuid.UUID              `json:"id" db:"id"`
	TenantID   uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	Email      string                 `json:"email" db:"email"`
	InvitedBy  uuid.UUID              `json:"invited_by" db:"invited_by"`
	Token      string                 `json:"token,omitempty" db:"-"` // Only returned on creation
	TokenHash  string                 `json:"-" db:"token_hash"`
	ExpiresAt  time.Time              `json:"expires_at" db:"expires_at"`
	AcceptedAt *time.Time             `json:"accepted_at,omitempty" db:"accepted_at"`
	AcceptedBy *uuid.UUID             `json:"accepted_by,omitempty" db:"accepted_by"`
	RoleIDs    []uuid.UUID            `json:"role_ids,omitempty" db:"role_ids"`
	Metadata   map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at" db:"updated_at"`
	DeletedAt  *time.Time              `json:"deleted_at,omitempty" db:"deleted_at"`
}

// IsExpired returns true if the invitation has expired
func (i *UserInvitation) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

// IsAccepted returns true if the invitation has been accepted
func (i *UserInvitation) IsAccepted() bool {
	return i.AcceptedAt != nil
}

// IsDeleted returns true if the invitation is soft-deleted
func (i *UserInvitation) IsDeleted() bool {
	return i.DeletedAt != nil
}

// IsValid returns true if the invitation is valid (not expired, not accepted, not deleted)
func (i *UserInvitation) IsValid() bool {
	return !i.IsExpired() && !i.IsAccepted() && !i.IsDeleted()
}

