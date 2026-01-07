package models

import (
	"time"

	"github.com/google/uuid"
)

// Permission represents a permission in the system
type Permission struct {
	ID          uuid.UUID `json:"id" db:"id"`
	TenantID    uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description,omitempty" db:"description"`
	Resource    string    `json:"resource" db:"resource"` // e.g., "user", "tenant", "role"
	Action      string    `json:"action" db:"action"`     // e.g., "create", "read", "update", "delete"
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"-" db:"deleted_at"`
}

// IsActive checks if permission is active (not deleted)
func (p *Permission) IsActive() bool {
	return p.DeletedAt == nil
}

// String returns a string representation of the permission
func (p *Permission) String() string {
	return p.Resource + ":" + p.Action
}

