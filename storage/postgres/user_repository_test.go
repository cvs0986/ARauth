package postgres

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/google/uuid"
	"github.com/nuage-identity/iam/identity/models"
	"github.com/nuage-identity/iam/internal/testutil"
	"github.com/nuage-identity/iam/storage/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) (*sql.DB, func()) {
	// Use test database URL from environment or skip
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
		return nil, func() {}
	}

	// Create DB connection using sql.Open directly for testing
	dbConn, err := sql.Open("postgres", testDBURL)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Test connection
	if err := dbConn.Ping(); err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
		return nil, func() {}
	}

	// Clean up before test
	testutil.CleanupTestDB(t, dbConn)

	// Return cleanup function
	cleanup := func() {
		testutil.CleanupTestDB(t, dbConn)
		testutil.TeardownTestDB(t, dbConn)
	}

	return dbConn, cleanup
}

func TestUserRepository_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	tenantID := uuid.New()
	user := &models.User{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Username:  "testuser",
		Email:     "test@example.com",
		Status:    models.UserStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)
}

func TestUserRepository_GetByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	tenantID := uuid.New()
	user := &models.User{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Username:  "getuser",
		Email:     "get@example.com",
		Status:    models.UserStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.Username, retrieved.Username)
	assert.Equal(t, user.Email, retrieved.Email)
}

func TestUserRepository_GetByUsername(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	tenantID := uuid.New()
	user := &models.User{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Username:  "usernameuser",
		Email:     "username@example.com",
		Status:    models.UserStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	retrieved, err := repo.GetByUsername(context.Background(), user.Username, tenantID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.Username, retrieved.Username)
}

func TestUserRepository_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	tenantID := uuid.New()
	user := &models.User{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Username:  "updateuser",
		Email:     "update@example.com",
		Status:    models.UserStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	// Update user
	user.Email = "updated@example.com"
	user.UpdatedAt = time.Now()
	err = repo.Update(context.Background(), user)
	require.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetByID(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Equal(t, "updated@example.com", retrieved.Email)
}

func TestUserRepository_List(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	tenantID := uuid.New()
	
	// Create multiple users
	for i := 0; i < 3; i++ {
		user := &models.User{
			ID:        uuid.New(),
			TenantID:  tenantID,
			Username:  fmt.Sprintf("listuser%d", i),
			Email:     fmt.Sprintf("list%d@example.com", i),
			Status:    models.UserStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := repo.Create(context.Background(), user)
		require.NoError(t, err)
	}

	// List users
	filters := &interfaces.UserFilters{
		TenantID: &tenantID,
		Limit:    10,
		Offset:   0,
	}
	users, err := repo.List(context.Background(), filters)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(users), 3)
} {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	tenantID := uuid.New()
	user := &models.User{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Username:  "testuser",
		Email:     "test@example.com",
		Status:    models.UserStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.Username, retrieved.Username)
	assert.Equal(t, user.Email, retrieved.Email)
}

func TestUserRepository_GetByUsername(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	tenantID := uuid.New()
	user := &models.User{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Username:  "testuser",
		Email:     "test@example.com",
		Status:    models.UserStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	retrieved, err := repo.GetByUsername(context.Background(), user.Username, tenantID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.Username, retrieved.Username)
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	tenantID := uuid.New()
	user := &models.User{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Username:  "testuser",
		Email:     "test@example.com",
		Status:    models.UserStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	retrieved, err := repo.GetByEmail(context.Background(), user.Email, tenantID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.Email, retrieved.Email)
}

func TestUserRepository_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	tenantID := uuid.New()
	user := &models.User{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Username:  "testuser",
		Email:     "test@example.com",
		Status:    models.UserStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	newEmail := "updated@example.com"
	user.Email = newEmail
	err = repo.Update(context.Background(), user)
	require.NoError(t, err)

	updated, err := repo.GetByID(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Equal(t, newEmail, updated.Email)
}

func TestUserRepository_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	tenantID := uuid.New()
	user := &models.User{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Username:  "testuser",
		Email:     "test@example.com",
		Status:    models.UserStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	err = repo.Delete(context.Background(), user.ID)
	require.NoError(t, err)

	// Verify soft delete - GetByID should return error
	deleted, err := repo.GetByID(context.Background(), user.ID)
	assert.Error(t, err)
	assert.Nil(t, deleted)
}

func TestUserRepository_List(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	tenantID := uuid.New()

	// Create multiple users
	for i := 0; i < 5; i++ {
		user := &models.User{
			ID:        uuid.New(),
			TenantID:  tenantID,
			Username:  "testuser" + string(rune(i+'0')),
			Email:     "test" + string(rune(i+'0')) + "@example.com",
			Status:    models.UserStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := repo.Create(context.Background(), user)
		require.NoError(t, err)
	}

	filters := &interfaces.UserFilters{
		Page:     1,
		PageSize: 10,
	}

	users, err := repo.List(context.Background(), tenantID, filters)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(users), 5)
}

