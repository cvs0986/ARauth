// +build integration

package login

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/auth/claims"
	"github.com/nuage-identity/iam/identity/credential"
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
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	// Create repositories
	userRepo := postgres.NewUserRepository(db)
	credentialRepo := postgres.NewCredentialRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)
	roleRepo := postgres.NewRoleRepository(db)
	permissionRepo := postgres.NewPermissionRepository(db)

	// Create password hasher
	hasher := password.NewHasher()

	// Create tenant
	tenantID := uuid.New()
	tenant := &models.Tenant{
		ID:     tenantID,
		Name:   "Test Tenant",
		Domain: "test-" + tenantID.String() + ".example.com",
		Status: "active",
	}
	err := tenantRepo.Create(context.Background(), tenant)
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
	hashedPassword, err := hasher.Hash("TestPassword123!")
	require.NoError(t, err)

	cred := &credential.Credential{
		UserID:       userID,
		PasswordHash: hashedPassword,
	}
	err = credentialRepo.Create(context.Background(), cred)
	require.NoError(t, err)

	// Create claims builder
	claimsBuilder := claims.NewBuilder(roleRepo, permissionRepo)

	// Create service (with nil hydra client for now - can be mocked)
	service := NewService(userRepo, credentialRepo, nil, claimsBuilder)

	// Test login
	req := &LoginRequest{
		TenantID: tenantID,
		Username: "testuser",
		Password: "TestPassword123!",
	}

	response, err := service.Login(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, response)
	// MFA should not be required for this user
	assert.False(t, response.MFARequired)
}

func TestService_Login_InvalidPassword_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	userRepo := postgres.NewUserRepository(db)
	credentialRepo := postgres.NewCredentialRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)
	roleRepo := postgres.NewRoleRepository(db)
	permissionRepo := postgres.NewPermissionRepository(db)
	hasher := password.NewHasher()

	tenantID := uuid.New()
	tenant := &models.Tenant{
		ID:     tenantID,
		Name:   "Test Tenant",
		Domain: "test-" + tenantID.String() + ".example.com",
		Status: "active",
	}
	err := tenantRepo.Create(context.Background(), tenant)
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

	hashedPassword, err := hasher.Hash("TestPassword123!")
	require.NoError(t, err)

	cred := &credential.Credential{
		UserID:       userID,
		PasswordHash: hashedPassword,
	}
	err = credentialRepo.Create(context.Background(), cred)
	require.NoError(t, err)

	claimsBuilder := claims.NewBuilder(roleRepo, permissionRepo)
	service := NewService(userRepo, credentialRepo, nil, claimsBuilder)

	req := &LoginRequest{
		TenantID: tenantID,
		Username: "testuser",
		Password: "WrongPassword123!",
	}

	_, err = service.Login(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid credentials")
}

func TestService_Login_UserNotFound_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	userRepo := postgres.NewUserRepository(db)
	credentialRepo := postgres.NewCredentialRepository(db)
	roleRepo := postgres.NewRoleRepository(db)
	permissionRepo := postgres.NewPermissionRepository(db)

	claimsBuilder := claims.NewBuilder(roleRepo, permissionRepo)
	service := NewService(userRepo, credentialRepo, nil, claimsBuilder)

	req := &LoginRequest{
		TenantID: uuid.New(),
		Username: "nonexistent",
		Password: "TestPassword123!",
	}

	_, err := service.Login(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid credentials")
}

