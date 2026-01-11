package scim

import (
	"context"
	"testing"

	"github.com/arauth-identity/iam/identity/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTokenService_CreateToken(t *testing.T) {
	mockRepo := &MockSCIMTokenRepository{}
	service := NewTokenService(mockRepo)

	t.Run("success", func(t *testing.T) {
		tenantID := uuid.New()
		req := &CreateTokenRequest{
			Name:   "Test Token",
			Scopes: []string{"users.read"},
		}

		mockRepo.CreateFunc = func(ctx context.Context, token *models.SCIMToken) error {
			assert.Equal(t, tenantID, token.TenantID)
			assert.Equal(t, req.Name, token.Name)
			assert.NotEmpty(t, token.TokenHash)
			assert.NotEmpty(t, token.LookupHash)
			return nil
		}

		token, plaintext, err := service.CreateToken(context.Background(), tenantID, req)
		assert.NoError(t, err)
		assert.NotNil(t, token)
		assert.NotEmpty(t, plaintext)
	})
}

func TestTokenService_RotateToken(t *testing.T) {
	mockRepo := &MockSCIMTokenRepository{}
	service := NewTokenService(mockRepo)

	t.Run("success", func(t *testing.T) {
		tokenID := uuid.New()
		existingToken := &models.SCIMToken{
			ID:        tokenID,
			Name:      "Old Token",
			TokenHash: "old-hash",
		}

		// Mock GetByID
		mockRepo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.SCIMToken, error) {
			assert.Equal(t, tokenID, id)
			return existingToken, nil
		}

		// Mock Update
		mockRepo.UpdateFunc = func(ctx context.Context, token *models.SCIMToken) error {
			assert.Equal(t, tokenID, token.ID)
			assert.NotEqual(t, "old-hash", token.TokenHash)
			assert.NotEmpty(t, token.LookupHash)
			return nil
		}

		rotatedToken, plaintext, err := service.RotateToken(context.Background(), tokenID)
		assert.NoError(t, err)
		assert.NotNil(t, rotatedToken)
		assert.NotEmpty(t, plaintext)
		assert.NotEqual(t, "old-hash", rotatedToken.TokenHash)
	})

	t.Run("not_found", func(t *testing.T) {
		tokenID := uuid.New()
		mockRepo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.SCIMToken, error) {
			return nil, assert.AnError
		}

		_, _, err := service.RotateToken(context.Background(), tokenID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token not found")
	})
}
