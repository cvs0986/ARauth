package postgres

import (
	"context"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/google/uuid"
	"github.com/nuage-identity/iam/identity/models"
	"github.com/nuage-identity/iam/storage/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPermissionRepository_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPermissionRepository(db)

	permission := &models.Permission{
		ID:        uuid.New(),
		Name:      "users:read",
		Resource:  "users",
		Action:    "read",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), permission)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, permission.ID)
}

func TestPermissionRepository_GetByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPermissionRepository(db)

	tenantID := uuid.New()
	permission := &models.Permission{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Name:      "users:write",
		Resource:  "users",
		Action:    "write",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), permission)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(context.Background(), permission.ID)
	require.NoError(t, err)
	assert.Equal(t, permission.ID, retrieved.ID)
	assert.Equal(t, permission.Name, retrieved.Name)
}

func TestPermissionRepository_GetByName(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPermissionRepository(db)

	permission := &models.Permission{
		ID:        uuid.New(),
		Name:      "users:delete",
		Resource:  "users",
		Action:    "delete",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), permission)
	require.NoError(t, err)

	retrieved, err := repo.GetByName(context.Background(), tenantID, permission.Name)
	require.NoError(t, err)
	assert.Equal(t, permission.ID, retrieved.ID)
	assert.Equal(t, permission.Name, retrieved.Name)
}

func TestPermissionRepository_List(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPermissionRepository(db)

	tenantID := uuid.New()

	// Create multiple permissions
	for i := 0; i < 3; i++ {
		permission := &models.Permission{
			ID:        uuid.New(),
			TenantID:  tenantID,
			Name:      "resource" + string(rune(i+'0')) + ":action",
			Resource:  "resource" + string(rune(i+'0')),
			Action:    "action",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := repo.Create(context.Background(), permission)
		require.NoError(t, err)
	}

	// List permissions
	filters := &interfaces.PermissionFilters{
		Page:     1,
		PageSize: 10,
	}
	permissions, err := repo.List(context.Background(), tenantID, filters)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(permissions), 3)
}

