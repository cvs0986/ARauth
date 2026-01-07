// +build integration

package role

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/identity/models"
	"github.com/nuage-identity/iam/identity/permission"
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

	roleRepo := postgres.NewRoleRepository(db)
	permissionRepo := postgres.NewPermissionRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)

	// Create tenant
	tenantID := uuid.New()
	tenant := &models.Tenant{
		ID:     tenantID,
		Name:   "Test Tenant",
		Domain: "test.example.com",
		Status: "active",
	}
	err := tenantRepo.Create(context.Background(), tenant)
	require.NoError(t, err)

	service := NewService(roleRepo, permissionRepo)

	req := &CreateRoleRequest{
		TenantID:    tenantID,
		Name:        "Admin",
		Description: &[]string{"Administrator role"}[0],
	}

	role, err := service.Create(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, role)
	assert.Equal(t, req.Name, role.Name)
	assert.Equal(t, tenantID, role.TenantID)
}

func TestService_AssignRoleToUser_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	roleRepo := postgres.NewRoleRepository(db)
	permissionRepo := postgres.NewPermissionRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)
	userRepo := postgres.NewUserRepository(db)

	// Create tenant
	tenantID := uuid.New()
	tenant := &models.Tenant{
		ID:     tenantID,
		Name:   "Test Tenant",
		Domain: "test.example.com",
		Status: "active",
	}
	err := tenantRepo.Create(context.Background(), tenant)
	require.NoError(t, err)

	// Create user
	userID := uuid.New()
	user := &models.User{
		ID:       userID,
		TenantID: tenantID,
		Username: "rbacuser",
		Email:    "rbac@example.com",
		Status:   "active",
	}
	err = userRepo.Create(context.Background(), user)
	require.NoError(t, err)

	service := NewService(roleRepo, permissionRepo)

	// Create role
	createReq := &CreateRoleRequest{
		TenantID: tenantID,
		Name:     "User",
	}
	role, err := service.Create(context.Background(), createReq)
	require.NoError(t, err)

	// Assign role to user
	err = service.AssignRoleToUser(context.Background(), userID, role.ID)
	require.NoError(t, err)

	// Verify role assignment
	userRoles, err := service.GetUserRoles(context.Background(), userID)
	require.NoError(t, err)
	assert.Len(t, userRoles, 1)
	assert.Equal(t, role.ID, userRoles[0].ID)
}

func TestService_AssignPermissionToRole_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	roleRepo := postgres.NewRoleRepository(db)
	permissionRepo := postgres.NewPermissionRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)

	// Create tenant
	tenantID := uuid.New()
	tenant := &models.Tenant{
		ID:     tenantID,
		Name:   "Test Tenant",
		Domain: "test.example.com",
		Status: "active",
	}
	err := tenantRepo.Create(context.Background(), tenant)
	require.NoError(t, err)

	roleService := NewService(roleRepo, permissionRepo)
	
	// Import permission service
	permissionService := permission.NewService(permissionRepo)

	// Create role
	createRoleReq := &CreateRoleRequest{
		TenantID: tenantID,
		Name:     "Editor",
	}
	role, err := roleService.Create(context.Background(), createRoleReq)
	require.NoError(t, err)

	// Create permission
	createPermReq := &permission.CreatePermissionRequest{
		TenantID: tenantID,
		Name:     "users:write",
		Resource: "users",
		Action:   "write",
	}
	permission, err := permissionService.Create(context.Background(), createPermReq)
	require.NoError(t, err)

	// Assign permission to role
	err = roleService.AssignPermissionToRole(context.Background(), role.ID, permission.ID)
	require.NoError(t, err)

	// Verify permission assignment
	rolePermissions, err := roleService.GetRolePermissions(context.Background(), role.ID)
	require.NoError(t, err)
	assert.Len(t, rolePermissions, 1)
	assert.Equal(t, permission.ID, rolePermissions[0].ID)
}

