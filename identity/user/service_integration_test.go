// +build integration

package user

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/internal/testutil"
	"github.com/nuage-identity/iam/storage/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Create_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := postgres.NewUserRepository(db)
	service := NewService(repo)

	tenantID := uuid.New()
	req := &CreateUserRequest{
		TenantID: tenantID,
		Username: "integrationuser",
		Email:    "integration@example.com",
		Status:   "active",
	}

	user, err := service.Create(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, req.Username, user.Username)
	assert.Equal(t, req.Email, user.Email)
	assert.Equal(t, tenantID, user.TenantID)
}

func TestService_GetByID_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := postgres.NewUserRepository(db)
	service := NewService(repo)

	tenantID := uuid.New()
	req := &CreateUserRequest{
		TenantID: tenantID,
		Username: "getuser",
		Email:    "get@example.com",
	}

	createdUser, err := service.Create(context.Background(), req)
	require.NoError(t, err)

	// Retrieve by ID
	retrievedUser, err := service.GetByID(context.Background(), createdUser.ID)
	require.NoError(t, err)
	assert.Equal(t, createdUser.ID, retrievedUser.ID)
	assert.Equal(t, createdUser.Username, retrievedUser.Username)
}

func TestService_GetByUsername_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := postgres.NewUserRepository(db)
	service := NewService(repo)

	tenantID := uuid.New()
	req := &CreateUserRequest{
		TenantID: tenantID,
		Username: "usernameuser",
		Email:    "username@example.com",
	}

	createdUser, err := service.Create(context.Background(), req)
	require.NoError(t, err)

	// Retrieve by username
	retrievedUser, err := service.GetByUsername(context.Background(), req.Username, tenantID)
	require.NoError(t, err)
	assert.Equal(t, createdUser.ID, retrievedUser.ID)
	assert.Equal(t, createdUser.Username, retrievedUser.Username)
}

