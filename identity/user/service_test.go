package user

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/arauth-identity/iam/identity/models"
)

// MockRepository is a mock implementation of UserRepository
type MockRepository struct {
	users map[uuid.UUID]*models.User
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		users: make(map[uuid.UUID]*models.User),
	}
}

func (m *MockRepository) Create(ctx context.Context, user *models.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, assert.AnError
	}
	return user, nil
}

func (m *MockRepository) GetByUsername(ctx context.Context, username string, tenantID uuid.UUID) (*models.User, error) {
	for _, user := range m.users {
		if user.Username == username && user.TenantID != nil && *user.TenantID == tenantID {
			return user, nil
		}
	}
	return nil, assert.AnError
}

func (m *MockRepository) GetByEmail(ctx context.Context, email string, tenantID uuid.UUID) (*models.User, error) {
	for _, user := range m.users {
		if user.Email == email && user.TenantID != nil && *user.TenantID == tenantID {
			return user, nil
		}
	}
	return nil, assert.AnError
}

func (m *MockRepository) Update(ctx context.Context, user *models.User) error {
	if _, ok := m.users[user.ID]; !ok {
		return assert.AnError
	}
	m.users[user.ID] = user
	return nil
}

func (m *MockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.users, id)
	return nil
}

func (m *MockRepository) List(ctx context.Context, tenantID uuid.UUID, filters interface{}) ([]*models.User, error) {
	var users []*models.User
	for _, user := range m.users {
		if user.TenantID != nil && *user.TenantID == tenantID {
			users = append(users, user)
		}
	}
	return users, nil
}

func (m *MockRepository) Count(ctx context.Context, tenantID uuid.UUID, filters interface{}) (int, error) {
	count := 0
	for _, user := range m.users {
		if user.TenantID != nil && *user.TenantID == tenantID {
			count++
		}
	}
	return count, nil
}

// MockCredentialRepository is a mock implementation of CredentialRepository
type MockCredentialRepository struct{}

func (m *MockCredentialRepository) Create(ctx context.Context, credential *models.Credential) error {
	return nil
}

func (m *MockCredentialRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Credential, error) {
	return nil, nil
}

func (m *MockCredentialRepository) Update(ctx context.Context, credential *models.Credential) error {
	return nil
}

func (m *MockCredentialRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	return nil
}

// TestCreateUser tests user creation
func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()

	mockRepo := NewMockRepository()
	mockCredRepo := &MockCredentialRepository{}
	service := NewService(mockRepo, mockCredRepo)

	tests := []struct {
		name    string
		req     *CreateUserRequest
		wantErr bool
	}{
		{
			name: "valid user creation",
			req: &CreateUserRequest{
				TenantID:  tenantID,
				Username:  "testuser",
				Email:     "test@example.com",
				Password:  "SecurePassword123!@#",
				FirstName: stringPtr("Test"),
				LastName:  stringPtr("User"),
			},
			wantErr: false,
		},
		{
			name: "missing username",
			req: &CreateUserRequest{
				TenantID: tenantID,
				Email:    "test@example.com",
				Password: "SecurePassword123!@#",
			},
			wantErr: true,
		},
		{
			name: "missing email",
			req: &CreateUserRequest{
				TenantID: tenantID,
				Username: "testuser",
				Password: "SecurePassword123!@#",
			},
			wantErr: true,
		},
		{
			name: "weak password",
			req: &CreateUserRequest{
				TenantID: tenantID,
				Username: "testuser",
				Email:    "test@example.com",
				Password: "weak",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.Create(ctx, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.req.Username, user.Username)
				assert.Equal(t, tt.req.Email, user.Email)
				assert.Equal(t, models.PrincipalTypeTenant, user.PrincipalType)
				assert.Equal(t, tenantID, *user.TenantID)
			}
		})
	}
}

// TestGetUserByID tests user retrieval by ID
func TestGetUserByID(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()

	mockRepo := NewMockRepository()
	mockCredRepo := &MockCredentialRepository{}
	service := NewService(mockRepo, mockCredRepo)

	// Create a test user
	createReq := &CreateUserRequest{
		TenantID: tenantID,
		Username: "testuser",
		Email:    "test@example.com",
		Password: "SecurePassword123!@#",
	}
	createdUser, err := service.Create(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, createdUser)

	// Test retrieval
	user, err := service.GetByID(ctx, createdUser.ID)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, createdUser.ID, user.ID)
	assert.Equal(t, createdUser.Username, user.Username)

	// Test non-existent user
	nonExistentID := uuid.New()
	user, err = service.GetByID(ctx, nonExistentID)
	assert.Error(t, err)
	assert.Nil(t, user)
}

// TestUpdateUser tests user updates
func TestUpdateUser(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()

	mockRepo := NewMockRepository()
	mockCredRepo := &MockCredentialRepository{}
	service := NewService(mockRepo, mockCredRepo)

	// Create a test user
	createReq := &CreateUserRequest{
		TenantID: tenantID,
		Username: "testuser",
		Email:    "test@example.com",
		Password: "SecurePassword123!@#",
	}
	createdUser, err := service.Create(ctx, createReq)
	require.NoError(t, err)

	// Update user
	newFirstName := "Updated"
	newLastName := "Name"
	updateReq := &UpdateUserRequest{
		FirstName: &newFirstName,
		LastName:  &newLastName,
	}

	updatedUser, err := service.Update(ctx, createdUser.ID, updateReq)
	assert.NoError(t, err)
	assert.NotNil(t, updatedUser)
	assert.Equal(t, newFirstName, *updatedUser.FirstName)
	assert.Equal(t, newLastName, *updatedUser.LastName)
}

// TestDeleteUser tests user deletion
func TestDeleteUser(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()

	mockRepo := NewMockRepository()
	mockCredRepo := &MockCredentialRepository{}
	service := NewService(mockRepo, mockCredRepo)

	// Create a test user
	createReq := &CreateUserRequest{
		TenantID: tenantID,
		Username: "testuser",
		Email:    "test@example.com",
		Password: "SecurePassword123!@#",
	}
	createdUser, err := service.Create(ctx, createReq)
	require.NoError(t, err)

	// Delete user
	err = service.Delete(ctx, createdUser.ID)
	assert.NoError(t, err)

	// Verify deletion
	user, err := service.GetByID(ctx, createdUser.ID)
	assert.Error(t, err)
	assert.Nil(t, user)
}

func stringPtr(s string) *string {
	return &s
}
