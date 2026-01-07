package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
)

// createTestTenant creates a test tenant for use in tests
func createTestTenant(ctx context.Context, db *sql.DB, tenantID uuid.UUID) error {
	query := `
		INSERT INTO tenants (id, name, domain, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := db.ExecContext(ctx, query,
		tenantID,
		"Test Tenant",
		"test-"+tenantID.String()+".example.com",
		models.TenantStatusActive,
		time.Now(),
		time.Now(),
	)
	return err
}

