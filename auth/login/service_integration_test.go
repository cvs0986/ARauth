// +build integration

package login

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/identity/models"
	"github.com/nuage-identity/iam/internal/testutil"
	"github.com/nuage-identity/iam/security/password"
	"github.com/nuage-identity/iam/storage/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Login_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test database
	db, err := testutil.SetupTestDB(t)
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}
	defer testutil.TeardownTestDB(t, db)

	// Create repositories
	userRepo := postgres.NewUserRepository(db)
	credentialRepo := postgres.NewCredentialRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)

	// Create password hasher
	hasher := password.NewHasher()

	// Create tenant
	tenantID := uuid.New()
	tenant := &models.Tenant{
		ID:     tenantID,
		Name:   "Test Tenant",
		Domain: "test.example.com",
		Status: "active",
	}
	err = tenantRepo.Create(context.Background(), tenant)
	require.NoError(t, err)

	// Create user
	userID := uuid.New()
	user := &models.User{
		ID:       userID,
		TenantID: tenantID,
		Username: "testuser",
		Email:    "test@example.com",
		Status:   "active",
	}
	err = userRepo.Create(context.Background(), user)
	require.NoError(t, err)

	// Create credential
	hashedPassword, err := hasher.Hash("password123")
	require.NoError(t, err)

	credential := &models.Credential{
		UserID:       userID,
		PasswordHash: hashedPassword,
	}
	err = credentialRepo.Create(context.Background(), credential)
	require.NoError(t, err)

	// Create service
	service := NewService(userRepo, credentialRepo, nil, nil, nil)

	// Test login
	req := &LoginRequest{
		TenantID: tenantID,
		Username: "testuser",
		Password: "password123",
	}

	response, err := service.Login(context.Background(), req)
	require.NoError(t, err)
	assert.NotEmpty(t, response.AccessToken)
	assert.Equal(t, "Bearer", response.TokenType)
}

func TestService_Login_InvalidPassword(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, err := testutil.SetupTestDB(t)
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}
	defer testutil.TeardownTestDB(t, db)

	userRepo := postgres.NewUserRepository(db)
	credentialRepo := postgres.NewCredentialRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)
	hasher := password.NewHasher()

	tenantID := uuid.New()
	tenant := &models.Tenant{
		ID:     tenantID,
		Name:   "Test Tenant",
		Domain: "test.example.com",
		Status: "active",
	}
	err = tenantRepo.Create(context.Background(), tenant)
	require.NoError(t, err)

	userID := uuid.New()
	user := &models.User{
		ID:       userID,
		TenantID: tenantID,
		Username: "testuser",
		Email:    "test@example.com",
		Status:   "active",
	}
	err = userRepo.Create(context.Background(), user)
	require.NoError(t, err)

	hashedPassword, err := hasher.Hash("password123")
	require.NoError(t, err)

	credential := &models.Credential{
		UserID:       userID,
		PasswordHash: hashedPassword,
	}
	err = credentialRepo.Create(context.Background(), credential)
	require.NoError(t, err)

	service := NewService(userRepo, credentialRepo, nil, nil, nil)

	req := &LoginRequest{
		TenantID: tenantID,
		Username: "testuser",
		Password: "wrongpassword",
	}

	_, err = service.Login(context.Background(), req)
	assert.Error(t, err)
}

