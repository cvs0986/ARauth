package user

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, u *models.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string, tenantID uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, username, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string, tenantID uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, email, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, u *models.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, tenantID uuid.UUID, filters *interfaces.UserFilters) ([]*models.User, error) {
	args := m.Called(ctx, tenantID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context, tenantID uuid.UUID, filters *interfaces.UserFilters) (int, error) {
	args := m.Called(ctx, tenantID, filters)
	return args.Int(0), args.Error(1)
}

func (m *MockUserRepository) GetByEmailSystem(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func TestService_Create(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewService(mockRepo)

	tenantID := uuid.New()
	firstName := "Test"
	lastName := "User"
	req := &CreateUserRequest{
		TenantID:  tenantID,
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: &firstName,
		LastName:  &lastName,
	}

	expectedUser := &models.User{
		ID:       uuid.New(),
		TenantID: &tenantID,
		Username: req.Username,
		Email:    req.Email,
	}

	mockRepo.On("GetByUsername", mock.Anything, req.Username, tenantID).Return(nil, assert.AnError)
	mockRepo.On("GetByEmail", mock.Anything, req.Email, tenantID).Return(nil, assert.AnError)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil).Run(func(args mock.Arguments) {
		user := args.Get(1).(*models.User)
		user.ID = expectedUser.ID
	})

	user, err := service.Create(context.Background(), req)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.Equal(t, req.Username, user.Username)
	assert.Equal(t, req.Email, user.Email)

	mockRepo.AssertExpectations(t)
}

func TestService_Create_DuplicateUsername(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewService(mockRepo)

	tenantID := uuid.New()
	req := &CreateUserRequest{
		TenantID: tenantID,
		Username: "testuser",
		Email:    "test@example.com",
	}

	existingUser := &models.User{
		ID:       uuid.New(),
		Username: req.Username,
	}

	mockRepo.On("GetByUsername", mock.Anything, req.Username, tenantID).Return(existingUser, nil)

	user, err := service.Create(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "already exists")

	mockRepo.AssertExpectations(t)
}

func TestService_GetByID(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewService(mockRepo)

	userID := uuid.New()
	expectedUser := &models.User{
		ID:       userID,
		Username: "testuser",
		Email:    "test@example.com",
	}

	mockRepo.On("GetByID", mock.Anything, userID).Return(expectedUser, nil)

	user, err := service.GetByID(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, expectedUser.ID, user.ID)
	assert.Equal(t, expectedUser.Username, user.Username)

	mockRepo.AssertExpectations(t)
}

func TestService_Update(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewService(mockRepo)

	userID := uuid.New()
	tenantID := uuid.New()
	existingUser := &models.User{
		ID:       userID,
		TenantID: &tenantID,
		Username: "testuser",
		Email:    "test@example.com",
	}

	updatedEmail := "updated@example.com"
	updatedFirstName := "Updated"
	req := &UpdateUserRequest{
		Email:     &updatedEmail,
		FirstName: &updatedFirstName,
	}

	mockRepo.On("GetByID", mock.Anything, userID).Return(existingUser, nil)
	mockRepo.On("GetByEmail", mock.Anything, updatedEmail, tenantID).Return(nil, assert.AnError) // Email not taken
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

	user, err := service.Update(context.Background(), userID, req)
	require.NoError(t, err)
	assert.Equal(t, updatedEmail, user.Email)
	if user.FirstName != nil {
		assert.Equal(t, updatedFirstName, *user.FirstName)
	}

	mockRepo.AssertExpectations(t)
}

func TestService_Delete(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewService(mockRepo)

	userID := uuid.New()

	mockRepo.On("Delete", mock.Anything, userID).Return(nil)

	err := service.Delete(context.Background(), userID)
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockRepo.AssertNumberOfCalls(t, "Delete", 1)
}

