package impersonation

import (
	"context"

	"github.com/arauth-identity/iam/identity/models"
	"github.com/google/uuid"
)

// ServiceInterface defines the interface for impersonation service
type ServiceInterface interface {
	// StartImpersonation starts an impersonation session and generates an impersonation token
	StartImpersonation(ctx context.Context, impersonatorID uuid.UUID, targetUserID uuid.UUID, reason *string) (*ImpersonationResult, error)

	// EndImpersonation ends an active impersonation session
	EndImpersonation(ctx context.Context, sessionID uuid.UUID, endedBy uuid.UUID) error

	// EndImpersonationByToken ends an impersonation session by token JTI
	EndImpersonationByToken(ctx context.Context, tokenJTI uuid.UUID, endedBy uuid.UUID) error

	// GetActiveSessions retrieves active impersonation sessions
	GetActiveSessions(ctx context.Context, filters *ImpersonationFilters) ([]*models.ImpersonationSession, error)

	// GetSession retrieves an impersonation session by ID
	GetSession(ctx context.Context, sessionID uuid.UUID) (*models.ImpersonationSession, error)
}

// ImpersonationResult contains the result of starting an impersonation session
type ImpersonationResult struct {
	Session       *models.ImpersonationSession `json:"session"`
	AccessToken   string                        `json:"access_token"`
	RefreshToken  string                        `json:"refresh_token"`
	IDToken       string                        `json:"id_token"`
	ExpiresIn     int                           `json:"expires_in"`
	TokenType     string                        `json:"token_type"`
}

// ImpersonationFilters defines filters for listing impersonation sessions
type ImpersonationFilters struct {
	ImpersonatorID *uuid.UUID
	TargetUserID   *uuid.UUID
	TenantID       *uuid.UUID
	ActiveOnly     bool
	Page           int
	PageSize       int
}

