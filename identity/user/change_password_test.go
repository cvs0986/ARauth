package user

import (
	"context"
	"testing"

	"github.com/arauth-identity/iam/identity/credential"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestChangePassword(t *testing.T) {
	ctx := context.Background()

	userID := uuid.New()
	user := &models.User{ID: userID, Email: "test@example.com"}
	cred := &credential.Credential{UserID: userID, PasswordHash: "oldhash"}

	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockTokenRepo := new(MockRefreshTokenRepository)
		service := NewService(mockRepo, mockCredRepo, mockTokenRepo)

		mockRepo.On("GetByID", ctx, userID).Return(user, nil)
		mockCredRepo.On("GetByUserID", ctx, userID).Return(cred, nil)
		mockCredRepo.On("Update", ctx, mock.MatchedBy(func(c *credential.Credential) bool {
			return c.UserID == userID && c.PasswordHash != "oldhash"
		})).Return(nil)
		mockTokenRepo.On("RevokeAllForUser", ctx, userID).Return(nil)

		err := service.ChangePassword(ctx, userID, "NewSecurePass123!")
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockCredRepo.AssertExpectations(t)
		mockTokenRepo.AssertExpectations(t)
	})

	t.Run("fail_revocation", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockTokenRepo := new(MockRefreshTokenRepository)
		service := NewService(mockRepo, mockCredRepo, mockTokenRepo)

		mockRepo.On("GetByID", ctx, userID).Return(user, nil)
		mockCredRepo.On("GetByUserID", ctx, userID).Return(cred, nil)
		mockCredRepo.On("Update", ctx, mock.Anything).Return(nil)
		mockTokenRepo.On("RevokeAllForUser", ctx, userID).Return(assert.AnError)

		err := service.ChangePassword(ctx, userID, "NewSecurePass123!")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to revoke sessions") // Verify fail-fast
	})
}
