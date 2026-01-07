// +build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/identity/models"
	"github.com/nuage-identity/iam/identity/role"
	"github.com/nuage-identity/iam/internal/cache"
	"github.com/nuage-identity/iam/internal/testutil"
	"github.com/nuage-identity/iam/storage/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_RBACFlow(t *testing.T) {
	// Setup test database
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)
	defer testutil.CleanupTestDB(t, db)

	var cacheClient *cache.Cache

	// Setup repositories
	userRepo := postgres.NewUserRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)
	roleRepo := postgres.NewRoleRepository(db)
	permissionRepo := postgres.NewPermissionRepository(db)

	// Create test tenant
	tenant := &models.Tenant{
		ID:     uuid.New(),
		Name:   "Test Tenant",
		Domain: "test.example.com",
		Status: models.TenantStatusActive,
	}
	require.NoError(t, tenantRepo.Create(context.Background(), tenant))

	// Create test user
	user := &models.User{
		ID:       uuid.New(),
		TenantID: tenant.ID,
		Username: "testuser",
		Email:    "test@example.com",
		Status:   models.UserStatusActive,
	}
	require.NoError(t, userRepo.Create(context.Background(), user))

	// Setup test server
	server, _ := testutil.SetupTestServer(db, cacheClient)
	defer server.Close()

	// Step 1: Create permission
	permissionReq := map[string]interface{}{
		"tenant_id": tenant.ID.String(),
		"name":      "users:read",
		"resource":  "users",
		"action":    "read",
	}
	body, _ := json.Marshal(permissionReq)

	req, _ := http.NewRequest("POST", server.URL+"/api/v1/permissions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenant.ID.String())

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var permissionResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&permissionResp)
	require.NoError(t, err)

	permissionID := permissionResp["id"].(string)
	assert.NotEmpty(t, permissionID)

	// Step 2: Create role
	roleReq := map[string]interface{}{
		"tenant_id": tenant.ID.String(),
		"name":      "UserReader",
	}
	body, _ = json.Marshal(roleReq)

	req, _ = http.NewRequest("POST", server.URL+"/api/v1/roles", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenant.ID.String())

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var roleResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&roleResp)
	require.NoError(t, err)

	roleID := roleResp["id"].(string)
	assert.NotEmpty(t, roleID)

	// Step 3: Assign permission to role (via service, as there's no API endpoint)
	permissionUUID, _ := uuid.Parse(permissionID)
	roleUUID, _ := uuid.Parse(roleID)

	// Get role service
	roleService := role.NewService(roleRepo, permissionRepo)
	err = roleService.AssignPermissionToRole(context.Background(), roleUUID, permissionUUID)
	require.NoError(t, err)

	// Step 4: Assign role to user (via service)
	err = roleService.AssignRoleToUser(context.Background(), user.ID, roleUUID)
	require.NoError(t, err)

	// Step 5: Verify user has role
	userRoles, err := roleRepo.GetUserRoles(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Len(t, userRoles, 1)
	assert.Equal(t, roleUUID, userRoles[0].ID)

	// Step 6: Verify role has permission
	rolePermissions, err := permissionRepo.GetRolePermissions(context.Background(), roleUUID)
	require.NoError(t, err)
	assert.Len(t, rolePermissions, 1)
	assert.Equal(t, permissionUUID, rolePermissions[0].ID)
	assert.Equal(t, "users:read", rolePermissions[0].Name)
}

