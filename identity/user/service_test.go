package user

import (
	"context"
	"testing"

	"github.com/arauth-identity/iam/identity/credential"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FakeUserRepository is a fake implementation of UserRepository
type FakeUserRepository struct {
	users map[uuid.UUID]*models.User
}

func NewFakeUserRepository() *FakeUserRepository {
	return &FakeUserRepository{
		users: make(map[uuid.UUID]*models.User),
	}
}

func (m *FakeUserRepository) Create(ctx context.Context, user *models.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *FakeUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, assert.AnError
	}
	return user, nil
}

func (m *FakeUserRepository) GetByUsername(ctx context.Context, username string, tenantID uuid.UUID) (*models.User, error) {
	for _, user := range m.users {
		if user.Username == username && user.TenantID != nil && *user.TenantID == tenantID {
			return user, nil
		}
	}
	return nil, assert.AnError
}

func (m *FakeUserRepository) GetByEmail(ctx context.Context, email string, tenantID uuid.UUID) (*models.User, error) {
	for _, user := range m.users {
		if user.Email == email && user.TenantID != nil && *user.TenantID == tenantID {
			return user, nil
		}
	}
	return nil, assert.AnError
}

func (m *FakeUserRepository) Update(ctx context.Context, user *models.User) error {
	if _, ok := m.users[user.ID]; !ok {
		return assert.AnError
	}
	m.users[user.ID] = user
	return nil
}

func (m *FakeUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.users, id)
	return nil
}

func (m *FakeUserRepository) List(ctx context.Context, tenantID uuid.UUID, filters *interfaces.UserFilters) ([]*models.User, error) {
	var users []*models.User
	for _, user := range m.users {
		if user.TenantID != nil && *user.TenantID == tenantID {
			users = append(users, user)
		}
	}
	return users, nil
}

func (m *FakeUserRepository) Count(ctx context.Context, tenantID uuid.UUID, filters *interfaces.UserFilters) (int, error) {
	count := 0
	for _, user := range m.users {
		if user.TenantID != nil && *user.TenantID == tenantID {
			count++
		}
	}
	return count, nil
}

// Implement missing methods for FakeUserRepository (if any required by interface)
func (m *FakeUserRepository) GetSystemUserByUsername(ctx context.Context, username string) (*models.User, error) {
	return nil, assert.AnError
}
func (m *FakeUserRepository) GetByEmailSystem(ctx context.Context, email string) (*models.User, error) {
	return nil, assert.AnError
}
func (m *FakeUserRepository) ListSystem(ctx context.Context, filters *interfaces.UserFilters) ([]*models.User, error) {
	return nil, nil
}
func (m *FakeUserRepository) CountSystem(ctx context.Context, filters *interfaces.UserFilters) (int, error) {
	return 0, nil
}

// FakeCredentialRepository is a mock implementation of CredentialRepository
type FakeCredentialRepository struct{}

// Update signature to match interface: Create(ctx, *credential.Credential) error
func (m *FakeCredentialRepository) Create(ctx context.Context, cred *credential.Credential) error {
	return nil
}

// But wait, the interface likely imports credential package.
// "github.com/arauth-identity/iam/identity/credential"
// I need to update the signature to match.

func (m *FakeCredentialRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*credential.Credential, error) {
	return nil, nil
}

// Update also might have changed? Lint only complained about Create.
func (m *FakeCredentialRepository) Update(ctx context.Context, cred *credential.Credential) error {
	return nil
}

func (m *FakeCredentialRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	return nil
}

// TestCreateUser tests user creation
func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()

	mockRepo := NewFakeUserRepository()
	mockCredRepo := &FakeCredentialRepository{}
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	service := NewService(mockRepo, mockCredRepo, mockRefreshTokenRepo)

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

	mockRepo := NewFakeUserRepository()
	mockCredRepo := &FakeCredentialRepository{}
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	service := NewService(mockRepo, mockCredRepo, mockRefreshTokenRepo)

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

	mockRepo := NewFakeUserRepository()
	mockCredRepo := &FakeCredentialRepository{}
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	service := NewService(mockRepo, mockCredRepo, mockRefreshTokenRepo)

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

	mockRepo := NewFakeUserRepository()
	mockCredRepo := &FakeCredentialRepository{}
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	service := NewService(mockRepo, mockCredRepo, mockRefreshTokenRepo)

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

// stringPtr is already defined in service_error_test.go, so we don't redeclare it
