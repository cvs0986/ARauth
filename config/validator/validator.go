package validator

import (
	"fmt"
	"time"

	"github.com/arauth-identity/iam/config"
)

// Validate validates the configuration
func Validate(cfg *config.Config) error {
	// Server validation
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d (must be 1-65535)", cfg.Server.Port)
	}
	if cfg.Server.Host == "" {
		return fmt.Errorf("server host is required")
	}

	// Database validation
	if cfg.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if cfg.Database.Port < 1 || cfg.Database.Port > 65535 {
		return fmt.Errorf("invalid database port: %d (must be 1-65535)", cfg.Database.Port)
	}
	if cfg.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if cfg.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if cfg.Database.Password == "" {
		return fmt.Errorf("database password is required")
	}
	if cfg.Database.MaxOpenConns < 1 {
		return fmt.Errorf("database max_open_conns must be >= 1")
	}
	if cfg.Database.MaxIdleConns < 0 {
		return fmt.Errorf("database max_idle_conns must be >= 0")
	}

	// Redis validation
	if cfg.Redis.Host == "" {
		return fmt.Errorf("redis host is required")
	}
	if cfg.Redis.Port < 1 || cfg.Redis.Port > 65535 {
		return fmt.Errorf("invalid redis port: %d (must be 1-65535)", cfg.Redis.Port)
	}

	// Hydra validation
	if cfg.Hydra.AdminURL == "" {
		return fmt.Errorf("hydra admin_url is required")
	}
	if cfg.Hydra.PublicURL == "" {
		return fmt.Errorf("hydra public_url is required")
	}

	// Security validation
	if cfg.Security.JWT.AccessTokenTTL < 1*time.Minute {
		return fmt.Errorf("jwt access_token_ttl too short (minimum 1 minute)")
	}
	if cfg.Security.JWT.RefreshTokenTTL < 1*time.Hour {
		return fmt.Errorf("jwt refresh_token_ttl too short (minimum 1 hour)")
	}
	if cfg.Security.Password.MinLength < 8 {
		return fmt.Errorf("password min_length must be >= 8")
	}

	// Logging validation
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[cfg.Logging.Level] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", cfg.Logging.Level)
	}

	validFormats := map[string]bool{
		"json": true,
		"text": true,
	}
	if !validFormats[cfg.Logging.Format] {
		return fmt.Errorf("invalid log format: %s (must be json or text)", cfg.Logging.Format)
	}

	return nil
}

