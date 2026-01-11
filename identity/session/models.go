package session

import (
	"time"

	"github.com/google/uuid"
)

// Session represents an active user session (mapped from RefreshToken)
type Session struct {
	ID              uuid.UUID  `json:"id"`
	UserID          uuid.UUID  `json:"user_id"`
	Username        string     `json:"username"`
	DeviceInfo      string     `json:"device_info,omitempty"`
	IPAddress       string     `json:"ip_address,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	ExpiresAt       time.Time  `json:"expires_at"`
	LastUsedAt      *time.Time `json:"last_used_at,omitempty"`
	IsImpersonation bool       `json:"is_impersonation"`
	ImpersonatorID  *uuid.UUID `json:"impersonator_id,omitempty"`
	RememberMe      bool       `json:"remember_me"`
	MFAVerified     bool       `json:"mfa_verified"`
}
