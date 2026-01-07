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

func TestTenantRepository_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewTenantRepository(db)

	tenant := &models.Tenant{
		ID:        uuid.New(),
		Name:      "Test Tenant",
		Domain:    "test.example.com",
		Status:    models.TenantStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), tenant)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, tenant.ID)
}

func TestTenantRepository_GetByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewTenantRepository(db)

	tenant := &models.Tenant{
		ID:        uuid.New(),
		Name:      "Get Tenant",
		Domain:    "get.example.com",
		Status:    models.TenantStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), tenant)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(context.Background(), tenant.ID)
	require.NoError(t, err)
	assert.Equal(t, tenant.ID, retrieved.ID)
	assert.Equal(t, tenant.Name, retrieved.Name)
	assert.Equal(t, tenant.Domain, retrieved.Domain)
}

func TestTenantRepository_GetByDomain(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewTenantRepository(db)

	tenant := &models.Tenant{
		ID:        uuid.New(),
		Name:      "Domain Tenant",
		Domain:    "domain.example.com",
		Status:    models.TenantStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), tenant)
	require.NoError(t, err)

	retrieved, err := repo.GetByDomain(context.Background(), tenant.Domain)
	require.NoError(t, err)
	assert.Equal(t, tenant.ID, retrieved.ID)
	assert.Equal(t, tenant.Domain, retrieved.Domain)
}

func TestTenantRepository_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewTenantRepository(db)

	tenant := &models.Tenant{
		ID:        uuid.New(),
		Name:      "Original",
		Domain:    "original.example.com",
		Status:    models.TenantStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), tenant)
	require.NoError(t, err)

	// Update tenant
	newName := "Updated"
	tenant.Name = newName
	tenant.UpdatedAt = time.Now()
	err = repo.Update(context.Background(), tenant)
	require.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetByID(context.Background(), tenant.ID)
	require.NoError(t, err)
	assert.Equal(t, newName, retrieved.Name)
}

func TestTenantRepository_List(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewTenantRepository(db)

	// Create multiple tenants
	for i := 0; i < 3; i++ {
		tenant := &models.Tenant{
			ID:        uuid.New(),
			Name:      "Tenant" + string(rune(i+'0')),
			Domain:    "tenant" + string(rune(i+'0')) + ".example.com",
			Status:    models.TenantStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := repo.Create(context.Background(), tenant)
		require.NoError(t, err)
	}

	// List tenants
	filters := &interfaces.TenantFilters{
		Page:     1,
		PageSize: 10,
	}
	tenants, err := repo.List(context.Background(), filters)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(tenants), 3)
}

