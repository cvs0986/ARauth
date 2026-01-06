package credential

import (
	"time"

	"github.com/google/uuid"
)

// Credential represents user credentials
type Credential struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	UserID              uuid.UUID  `json:"user_id" db:"user_id"`
	PasswordHash        string     `json:"-" db:"password_hash"`
	PasswordChangedAt   time.Time  `json:"password_changed_at" db:"password_changed_at"`
	PasswordExpiresAt   *time.Time `json:"password_expires_at,omitempty" db:"password_expires_at"`
	FailedLoginAttempts int        `json:"-" db:"failed_login_attempts"`
	LockedUntil         *time.Time `json:"locked_until,omitempty" db:"locked_until"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}

// IsLocked checks if the credential is currently locked
func (c *Credential) IsLocked() bool {
	if c.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*c.LockedUntil)
}

// IncrementFailedAttempts increments the failed login attempts counter
func (c *Credential) IncrementFailedAttempts() {
	c.FailedLoginAttempts++
	// Lock account after 5 failed attempts for 30 minutes
	if c.FailedLoginAttempts >= 5 {
		lockUntil := time.Now().Add(30 * time.Minute)
		c.LockedUntil = &lockUntil
	}
}

// ResetFailedAttempts resets the failed login attempts counter
func (c *Credential) ResetFailedAttempts() {
	c.FailedLoginAttempts = 0
	c.LockedUntil = nil
}

