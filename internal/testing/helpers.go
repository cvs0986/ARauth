package testing

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
)

// TestDB provides a test database connection
type TestDB struct {
	DB *sql.DB
}

// NewTestDB creates a new test database connection
func NewTestDB(dsn string) (*TestDB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &TestDB{DB: db}, nil
}

// Cleanup cleans up test data
func (tdb *TestDB) Cleanup(ctx context.Context) error {
	// Truncate tables in reverse dependency order
	tables := []string{
		"user_invitations",
		"impersonation_sessions",
		"oauth_scopes",
		"webhook_deliveries",
		"webhooks",
		"federated_identities",
		"identity_providers",
		"audit_events",
		"user_roles",
		"role_permissions",
		"user_capabilities",
		"credentials",
		"permissions",
		"roles",
		"users",
		"tenants",
		"scim_tokens",
	}

	for _, table := range tables {
		_, err := tdb.DB.ExecContext(ctx, "TRUNCATE TABLE "+table+" CASCADE")
		if err != nil {
			// Ignore errors for tables that don't exist
			continue
		}
	}

	return nil
}

// CreateTestTenant creates a test tenant
func CreateTestTenant() *models.Tenant {
	return &models.Tenant{
		ID:        uuid.New(),
		Name:      "Test Tenant",
		Domain:    "test.local",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateTestUser creates a test user
func CreateTestUser(tenantID *uuid.UUID) *models.User {
	userID := uuid.New()
	return &models.User{
		ID:            userID,
		TenantID:      tenantID,
		PrincipalType: models.PrincipalTypeTenant,
		Username:      "testuser",
		Email:         "test@example.com",
		FirstName:     stringPtr("Test"),
		LastName:      stringPtr("User"),
		Status:        models.UserStatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// CreateTestSystemUser creates a test system user
func CreateTestSystemUser() *models.User {
	userID := uuid.New()
	return &models.User{
		ID:            userID,
		TenantID:      nil,
		PrincipalType: models.PrincipalTypeSystem,
		Username:      "system_admin",
		Email:         "admin@system.local",
		FirstName:     stringPtr("System"),
		LastName:      stringPtr("Admin"),
		Status:        models.UserStatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// CreateTestRole creates a test role
func CreateTestRole(tenantID uuid.UUID) *models.Role {
	return &models.Role{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        "test_role",
		Description: stringPtr("Test Role"),
		IsSystem:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// uuidPtr returns a pointer to a UUID
func uuidPtr(u uuid.UUID) *uuid.UUID {
	return &u
}

