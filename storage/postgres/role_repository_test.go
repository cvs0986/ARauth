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

func TestRoleRepository_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRoleRepository(db)

	tenantID := uuid.New()
	role := &models.Role{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Name:      "Admin",
		IsSystem:  false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), role)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, role.ID)
}

func TestRoleRepository_GetByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRoleRepository(db)

	tenantID := uuid.New()
	role := &models.Role{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Name:      "Editor",
		IsSystem:  false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), role)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(context.Background(), role.ID)
	require.NoError(t, err)
	assert.Equal(t, role.ID, retrieved.ID)
	assert.Equal(t, role.Name, retrieved.Name)
}

func TestRoleRepository_GetByName(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRoleRepository(db)

	tenantID := uuid.New()
	role := &models.Role{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Name:      "Viewer",
		IsSystem:  false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), role)
	require.NoError(t, err)

	retrieved, err := repo.GetByName(context.Background(), tenantID, role.Name)
	require.NoError(t, err)
	assert.Equal(t, role.ID, retrieved.ID)
	assert.Equal(t, role.Name, retrieved.Name)
}

func TestRoleRepository_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRoleRepository(db)

	tenantID := uuid.New()
	role := &models.Role{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Name:      "Original",
		IsSystem:  false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), role)
	require.NoError(t, err)

	// Update role
	newName := "Updated"
	role.Name = newName
	role.UpdatedAt = time.Now()
	err = repo.Update(context.Background(), role)
	require.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetByID(context.Background(), role.ID)
	require.NoError(t, err)
	assert.Equal(t, newName, retrieved.Name)
}

func TestRoleRepository_List(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRoleRepository(db)

	tenantID := uuid.New()

	// Create multiple roles
	for i := 0; i < 3; i++ {
		role := &models.Role{
			ID:        uuid.New(),
			TenantID:  tenantID,
			Name:      "Role" + string(rune(i+'0')),
			IsSystem:  false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := repo.Create(context.Background(), role)
		require.NoError(t, err)
	}

	// List roles
	filters := &interfaces.RoleFilters{
		Page:     1,
		PageSize: 10,
	}
	roles, err := repo.List(context.Background(), tenantID, filters)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(roles), 3)
}

