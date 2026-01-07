// +build integration

package tenant

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/identity/models"
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

	repo := postgres.NewTenantRepository(db)
	service := NewService(repo)

	req := &CreateTenantRequest{
		Name:   "Integration Tenant",
		Domain: "integration.example.com",
	}

	tenant, err := service.Create(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, tenant)
	assert.Equal(t, req.Name, tenant.Name)
	assert.Equal(t, req.Domain, tenant.Domain)
	assert.NotEqual(t, uuid.Nil, tenant.ID)
}

func TestService_GetByID_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := postgres.NewTenantRepository(db)
	service := NewService(repo)

	// Create tenant first
	createReq := &CreateTenantRequest{
		Name:   "Get Tenant",
		Domain: "get.example.com",
	}
	createdTenant, err := service.Create(context.Background(), createReq)
	require.NoError(t, err)

	// Retrieve by ID
	retrievedTenant, err := service.GetByID(context.Background(), createdTenant.ID)
	require.NoError(t, err)
	assert.Equal(t, createdTenant.ID, retrievedTenant.ID)
	assert.Equal(t, createdTenant.Name, retrievedTenant.Name)
	assert.Equal(t, createdTenant.Domain, retrievedTenant.Domain)
}

func TestService_GetByDomain_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := postgres.NewTenantRepository(db)
	service := NewService(repo)

	// Create tenant first
	createReq := &CreateTenantRequest{
		Name:   "Domain Tenant",
		Domain: "domain.example.com",
	}
	createdTenant, err := service.Create(context.Background(), createReq)
	require.NoError(t, err)

	// Retrieve by domain
	retrievedTenant, err := service.GetByDomain(context.Background(), createdTenant.Domain)
	require.NoError(t, err)
	assert.Equal(t, createdTenant.ID, retrievedTenant.ID)
	assert.Equal(t, createdTenant.Domain, retrievedTenant.Domain)
}

func TestService_Isolation_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	tenantRepo := postgres.NewTenantRepository(db)
	userRepo := postgres.NewUserRepository(db)
	tenantService := NewService(tenantRepo)
	userService := user.NewService(userRepo)

	// Create two tenants
	tenant1Req := &CreateTenantRequest{
		Name:   "Tenant 1",
		Domain: "tenant1.example.com",
	}
	tenant1, err := tenantService.Create(context.Background(), tenant1Req)
	require.NoError(t, err)

	tenant2Req := &CreateTenantRequest{
		Name:   "Tenant 2",
		Domain: "tenant2.example.com",
	}
	tenant2, err := tenantService.Create(context.Background(), tenant2Req)
	require.NoError(t, err)

	// Create users in each tenant
	user1Req := &user.CreateUserRequest{
		TenantID: tenant1.ID,
		Username: "user1",
		Email:    "user1@tenant1.com",
		Status:   "active",
	}
	user1, err := userService.Create(context.Background(), user1Req)
	require.NoError(t, err)

	user2Req := &user.CreateUserRequest{
		TenantID: tenant2.ID,
		Username: "user2",
		Email:    "user2@tenant2.com",
		Status:   "active",
	}
	user2, err := userService.Create(context.Background(), user2Req)
	require.NoError(t, err)

	// Verify tenant isolation - user1 should not be accessible from tenant2 context
	retrievedUser1, err := userService.GetByID(context.Background(), user1.ID)
	require.NoError(t, err)
	assert.Equal(t, tenant1.ID, retrievedUser1.TenantID)
	assert.NotEqual(t, tenant2.ID, retrievedUser1.TenantID)

	// Verify user2 belongs to tenant2
	retrievedUser2, err := userService.GetByID(context.Background(), user2.ID)
	require.NoError(t, err)
	assert.Equal(t, tenant2.ID, retrievedUser2.TenantID)
	assert.NotEqual(t, tenant1.ID, retrievedUser2.TenantID)
}

