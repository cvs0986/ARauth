package loader

import (
	"fmt"
	"os"
	"strings"

	"github.com/arauth-identity/iam/config"
	"gopkg.in/yaml.v3"
)

// LoadConfig loads configuration from file and environment variables
func LoadConfig(path string) (*config.Config, error) {
	cfg := &config.Config{}

	// Load from YAML file if it exists
	if path != "" {
		if _, err := os.Stat(path); err == nil {
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}

			// Expand environment variables in YAML
			expanded := os.ExpandEnv(string(data))

			if err := yaml.Unmarshal([]byte(expanded), cfg); err != nil {
				return nil, fmt.Errorf("failed to parse config file: %w", err)
			}
		}
	}

	// Override with environment variables
	loadFromEnv(cfg)

	// Set defaults
	setDefaults(cfg)

	return cfg, nil
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(cfg *config.Config) {
	// Server
	if port := os.Getenv("SERVER_PORT"); port != "" {
		_, _ = fmt.Sscanf(port, "%d", &cfg.Server.Port)
	}
	if host := os.Getenv("SERVER_HOST"); host != "" {
		cfg.Server.Host = host
	}

	// Database
	if host := os.Getenv("DATABASE_HOST"); host != "" {
		cfg.Database.Host = host
	}
	if port := os.Getenv("DATABASE_PORT"); port != "" {
		_, _ = fmt.Sscanf(port, "%d", &cfg.Database.Port)
	}
	if name := os.Getenv("DATABASE_NAME"); name != "" {
		cfg.Database.Name = name
	}
	if user := os.Getenv("DATABASE_USER"); user != "" {
		cfg.Database.User = user
	}
	if password := os.Getenv("DATABASE_PASSWORD"); password != "" {
		cfg.Database.Password = password
	}

	// Redis
	if host := os.Getenv("REDIS_HOST"); host != "" {
		cfg.Redis.Host = host
	}
	if port := os.Getenv("REDIS_PORT"); port != "" {
		_, _ = fmt.Sscanf(port, "%d", &cfg.Redis.Port)
	}
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		cfg.Redis.Password = password
	}

	// Hydra
	if adminURL := os.Getenv("HYDRA_ADMIN_URL"); adminURL != "" {
		cfg.Hydra.AdminURL = adminURL
	}
	if publicURL := os.Getenv("HYDRA_PUBLIC_URL"); publicURL != "" {
		cfg.Hydra.PublicURL = publicURL
	}

	// JWT
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		cfg.Security.JWT.Secret = secret
	}
	if issuer := os.Getenv("JWT_ISSUER"); issuer != "" {
		cfg.Security.JWT.Issuer = issuer
	}

	// Logging
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		cfg.Logging.Level = strings.ToLower(level)
	}
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		cfg.Logging.Format = strings.ToLower(format)
	}
}

// setDefaults sets default values for configuration
func setDefaults(cfg *config.Config) {
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Database.Host == "" {
		cfg.Database.Host = "localhost"
	}
	if cfg.Database.Port == 0 {
		cfg.Database.Port = 5432
	}
	if cfg.Database.Name == "" {
		cfg.Database.Name = "iam"
	}
	if cfg.Database.User == "" {
		cfg.Database.User = "iam_user"
	}
	if cfg.Redis.Host == "" {
		cfg.Redis.Host = "localhost"
	}
	if cfg.Redis.Port == 0 {
		cfg.Redis.Port = 6379
	}
	if cfg.Hydra.AdminURL == "" {
		cfg.Hydra.AdminURL = "http://localhost:4445"
	}
	if cfg.Hydra.PublicURL == "" {
		cfg.Hydra.PublicURL = "http://localhost:4444"
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "json"
	}
}

