package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/identity/models"
)

// createTestTenant creates a test tenant for use in tests
func createTestTenant(ctx context.Context, db interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) error
}, tenantID uuid.UUID) error {
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

