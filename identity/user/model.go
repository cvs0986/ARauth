package user

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	TenantID        uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	Username        string           `json:"username" db:"username"`
	Email           string           `json:"email" db:"email"`
	FirstName       *string          `json:"first_name,omitempty" db:"first_name"`
	LastName        *string          `json:"last_name,omitempty" db:"last_name"`
	Status          string           `json:"status" db:"status"`
	MFAEnabled      bool             `json:"mfa_enabled" db:"mfa_enabled"`
	MFASecretEncrypted *string       `json:"-" db:"mfa_secret_encrypted"`
	LastLoginAt     *time.Time       `json:"last_login_at,omitempty" db:"last_login_at"`
	Metadata        map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt       time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at" db:"updated_at"`
	DeletedAt       *time.Time       `json:"-" db:"deleted_at"`
}

// UserStatus represents user status values
const (
	UserStatusActive   = "active"
	UserStatusSuspended = "suspended"
	UserStatusDeleted  = "deleted"
)

// IsActive checks if user is active
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive && u.DeletedAt == nil
}

// FullName returns the full name of the user
func (u *User) FullName() string {
	if u.FirstName != nil && u.LastName != nil {
		return *u.FirstName + " " + *u.LastName
	}
	if u.FirstName != nil {
		return *u.FirstName
	}
	if u.LastName != nil {
		return *u.LastName
	}
	return ""
}

