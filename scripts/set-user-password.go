package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"database/sql"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/google/uuid"
	"github.com/nuage-identity/iam/identity/credential"
	"github.com/nuage-identity/iam/security/password"
	"github.com/nuage-identity/iam/storage/postgres"
)

func main() {
	var (
		dbHost     = flag.String("host", "127.0.0.1", "Database host")
		dbPort     = flag.String("port", "5433", "Database port")
		dbUser     = flag.String("user", "dcim_user", "Database user")
		dbPassword = flag.String("password", "dcim_password", "Database password")
		dbName     = flag.String("dbname", "iam", "Database name")
		username   = flag.String("username", "", "Username to set password for")
		userPassword = flag.String("password-value", "", "Password to set")
	)
	flag.Parse()

	if *username == "" || *userPassword == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -username <username> -password-value <password>\n", os.Args[0])
		os.Exit(1)
	}

	// Connect to database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		*dbHost, *dbPort, *dbUser, *dbPassword, *dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to ping database: %v\n", err)
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Get user ID
	var userID uuid.UUID
	err = db.QueryRow("SELECT id FROM users WHERE username = $1", *username).Scan(&userID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "User not found: %v\n", err)
		os.Exit(1)
	}

	// Hash password
	hasher := password.NewHasher()
	passwordHash, err := hasher.Hash(*userPassword)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to hash password: %v\n", err)
		os.Exit(1)
	}

	// Create credential repository
	credRepo := postgres.NewCredentialRepository(db)

	// Check if credential exists
	ctx := context.Background()
	existingCred, err := credRepo.GetByUserID(ctx, userID)
	if err != nil {
		// Credential doesn't exist, create it
		cred := &credential.Credential{
			UserID:       userID,
			PasswordHash: passwordHash,
		}
		if err := credRepo.Create(ctx, cred); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create credential: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Created credentials for user '%s' (ID: %s)\n", *username, userID)
	} else {
		// Credential exists, update it
		existingCred.PasswordHash = passwordHash
		if err := credRepo.Update(ctx, existingCred); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to update credential: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Updated password for user '%s' (ID: %s)\n", *username, userID)
	}
}

