package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

// SetupTestDB creates a test database connection
func SetupTestDB(t *testing.T) *sql.DB {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost port=5432 user=iam_user password=test_password dbname=iam_test sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
	}

	return db
}

// CleanupTestDB cleans up test data
func CleanupTestDB(t *testing.T, db *sql.DB) {
	// Truncate all tables
	tables := []string{
		"audit_logs",
		"mfa_recovery_codes",
		"role_permissions",
		"user_roles",
		"permissions",
		"roles",
		"credentials",
		"users",
		"tenants",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			t.Logf("Warning: Failed to truncate table %s: %v", table, err)
		}
	}
}

// TeardownTestDB closes the test database connection
func TeardownTestDB(t *testing.T, db *sql.DB) {
	if db != nil {
		if err := db.Close(); err != nil {
			t.Logf("Warning: Failed to close test database: %v", err)
		}
	}
}

