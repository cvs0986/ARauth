package scim

import (
	"context"

	"github.com/arauth-identity/iam/identity/models"
	"github.com/google/uuid"
)

// MockSCIMTokenRepository is a mock implementation of interfaces.SCIMTokenRepository
type MockSCIMTokenRepository struct {
	CreateFunc          func(ctx context.Context, token *models.SCIMToken) error
	GetByIDFunc         func(ctx context.Context, id uuid.UUID) (*models.SCIMToken, error)
	GetByLookupHashFunc func(ctx context.Context, lookupHash string) (*models.SCIMToken, error)
	ListFunc            func(ctx context.Context, tenantID uuid.UUID) ([]*models.SCIMToken, error)
	UpdateFunc          func(ctx context.Context, token *models.SCIMToken) error
	DeleteFunc          func(ctx context.Context, id uuid.UUID) error
	UpdateLastUsedFunc  func(ctx context.Context, id uuid.UUID) error
}

func (m *MockSCIMTokenRepository) Create(ctx context.Context, token *models.SCIMToken) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, token)
	}
	return nil
}

func (m *MockSCIMTokenRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.SCIMToken, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockSCIMTokenRepository) GetByLookupHash(ctx context.Context, lookupHash string) (*models.SCIMToken, error) {
	if m.GetByLookupHashFunc != nil {
		return m.GetByLookupHashFunc(ctx, lookupHash)
	}
	return nil, nil
}

func (m *MockSCIMTokenRepository) List(ctx context.Context, tenantID uuid.UUID) ([]*models.SCIMToken, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, tenantID)
	}
	return []*models.SCIMToken{}, nil
}

func (m *MockSCIMTokenRepository) Update(ctx context.Context, token *models.SCIMToken) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, token)
	}
	return nil
}

func (m *MockSCIMTokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockSCIMTokenRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	if m.UpdateLastUsedFunc != nil {
		return m.UpdateLastUsedFunc(ctx, id)
	}
	return nil
}
