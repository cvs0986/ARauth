package models

import (
	"time"

	"github.com/google/uuid"
)

// ImpersonationSession represents an admin impersonation session
type ImpersonationSession struct {
	ID                uuid.UUID              `json:"id" db:"id"`
	ImpersonatorID   uuid.UUID              `json:"impersonator_user_id" db:"impersonator_user_id"`
	TargetUserID     uuid.UUID              `json:"target_user_id" db:"target_user_id"`
	TenantID         *uuid.UUID             `json:"tenant_id,omitempty" db:"tenant_id"`
	StartedAt        time.Time              `json:"started_at" db:"started_at"`
	EndedAt          *time.Time             `json:"ended_at,omitempty" db:"ended_at"`
	TokenJTI         *uuid.UUID             `json:"token_jti,omitempty" db:"token_jti"`
	Reason           *string                `json:"reason,omitempty" db:"reason"`
	Metadata         map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
}

// IsActive returns true if the impersonation session is currently active
func (s *ImpersonationSession) IsActive() bool {
	return s.EndedAt == nil
}

