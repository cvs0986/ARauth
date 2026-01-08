// +build ignore

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/arauth-identity/iam/cmd/bootstrap"
	"github.com/arauth-identity/iam/config/loader"
	"github.com/arauth-identity/iam/storage/postgres"
	"go.uber.org/zap"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "Path to config file")
	username := flag.String("username", "", "Master user username (overrides config)")
	email := flag.String("email", "", "Master user email (overrides config)")
	password := flag.String("password", "", "Master user password (required, overrides config)")
	force := flag.Bool("force", false, "Force bootstrap even if master user exists")
	flag.Parse()

	// Load configuration
	cfg, err := loader.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Override with CLI flags
	if *username != "" {
		cfg.Bootstrap.MasterUser.Username = *username
	}
	if *email != "" {
		cfg.Bootstrap.MasterUser.Email = *email
	}
	if *password != "" {
		cfg.Bootstrap.MasterUser.Password = *password
	}
	cfg.Bootstrap.Force = *force

	// Validate password is provided
	if cfg.Bootstrap.MasterUser.Password == "" {
		log.Fatal("Password is required. Use --password flag or set BOOTSTRAP_PASSWORD env var")
	}

	// Initialize logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Connect to database
	db, err := postgres.NewConnection(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	logger.Info("Database connection established")

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	credentialRepo := postgres.NewCredentialRepository(db)
	systemRoleRepo := postgres.NewSystemRoleRepository(db)

	// Initialize bootstrap service
	bootstrapService := bootstrap.NewBootstrapService(
		&cfg.Bootstrap,
		userRepo,
		credentialRepo,
		systemRoleRepo,
	)

	ctx := context.Background()

	// Run bootstrap
	if err := bootstrapService.Bootstrap(ctx); err != nil {
		if !*force && err.Error() == "master user already exists (use --force to re-bootstrap)" {
			fmt.Println("⚠️  System already bootstrapped. Use --force to re-bootstrap.")
			os.Exit(0)
		}
		logger.Fatal("Bootstrap failed", zap.Error(err))
	}

	fmt.Println("✅ System bootstrapped successfully!")
	fmt.Printf("   Master User ID: %s\n", "created")
	fmt.Printf("   Username: %s\n", cfg.Bootstrap.MasterUser.Username)
	fmt.Printf("   Email: %s\n", cfg.Bootstrap.MasterUser.Email)
	fmt.Println("\n⚠️  IMPORTANT: Change the master user password on first login!")
}

