package models

import (
	"time"

	"github.com/google/uuid"
)

// OAuthScope represents an OAuth scope that maps to permissions
type OAuthScope struct {
	ID          uuid.UUID `json:"id" db:"id"`
	TenantID    uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description,omitempty" db:"description"`
	Permissions []string  `json:"permissions" db:"permissions"` // Array of permission names
	IsDefault   bool      `json:"is_default" db:"is_default"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// IsDeleted returns true if the scope is soft-deleted
func (s *OAuthScope) IsDeleted() bool {
	return s.DeletedAt != nil
}

