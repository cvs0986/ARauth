package interfaces

import (
	"context"

	"github.com/arauth-identity/iam/identity/models"
	"github.com/google/uuid"
)

// ImpersonationRepository defines the interface for impersonation session storage
type ImpersonationRepository interface {
	// Create creates a new impersonation session
	Create(ctx context.Context, session *models.ImpersonationSession) error

	// GetByID retrieves an impersonation session by ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.ImpersonationSession, error)

	// GetByTokenJTI retrieves an impersonation session by token JTI
	GetByTokenJTI(ctx context.Context, tokenJTI uuid.UUID) (*models.ImpersonationSession, error)

	// GetActiveByImpersonator retrieves active impersonation sessions for an impersonator
	GetActiveByImpersonator(ctx context.Context, impersonatorID uuid.UUID) ([]*models.ImpersonationSession, error)

	// GetActiveByTarget retrieves active impersonation sessions for a target user
	GetActiveByTarget(ctx context.Context, targetUserID uuid.UUID) ([]*models.ImpersonationSession, error)

	// EndSession ends an impersonation session
	EndSession(ctx context.Context, id uuid.UUID) error

	// List lists impersonation sessions with filters
	List(ctx context.Context, filters *ImpersonationFilters) ([]*models.ImpersonationSession, error)
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

