package config

import (
	"time"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Hydra    HydraConfig    `yaml:"hydra"`
	Security SecurityConfig  `yaml:"security"`
	Logging  LoggingConfig  `yaml:"logging"`
	Metrics  MetricsConfig  `yaml:"metrics"`
	Bootstrap BootstrapConfig `yaml:"bootstrap"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port        int           `yaml:"port" env:"SERVER_PORT" envDefault:"8080"`
	Host        string        `yaml:"host" env:"SERVER_HOST" envDefault:"0.0.0.0"`
	ReadTimeout time.Duration `yaml:"read_timeout" env:"SERVER_READ_TIMEOUT" envDefault:"30s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env:"SERVER_WRITE_TIMEOUT" envDefault:"30s"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" env:"SERVER_IDLE_TIMEOUT" envDefault:"120s"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string        `yaml:"host" env:"DATABASE_HOST" envDefault:"localhost"`
	Port            int           `yaml:"port" env:"DATABASE_PORT" envDefault:"5432"`
	Name            string        `yaml:"name" env:"DATABASE_NAME" envDefault:"iam"`
	User            string        `yaml:"user" env:"DATABASE_USER" envDefault:"iam_user"`
	Password        string        `yaml:"password" env:"DATABASE_PASSWORD"`
	MaxOpenConns    int           `yaml:"max_open_conns" env:"DATABASE_MAX_OPEN_CONNS" envDefault:"25"`
	MaxIdleConns    int           `yaml:"max_idle_conns" env:"DATABASE_MAX_IDLE_CONNS" envDefault:"5"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env:"DATABASE_CONN_MAX_LIFETIME" envDefault:"5m"`
	SSLMode         string        `yaml:"ssl_mode" env:"DATABASE_SSL_MODE" envDefault:"disable"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host         string `yaml:"host" env:"REDIS_HOST" envDefault:"localhost"`
	Port         int    `yaml:"port" env:"REDIS_PORT" envDefault:"6379"`
	Password     string `yaml:"password" env:"REDIS_PASSWORD"`
	DB           int    `yaml:"db" env:"REDIS_DB" envDefault:"0"`
	PoolSize     int    `yaml:"pool_size" env:"REDIS_POOL_SIZE" envDefault:"10"`
	MinIdleConns int    `yaml:"min_idle_conns" env:"REDIS_MIN_IDLE_CONNS" envDefault:"5"`
}

// HydraConfig holds Hydra configuration
type HydraConfig struct {
	AdminURL string        `yaml:"admin_url" env:"HYDRA_ADMIN_URL" envDefault:"http://localhost:4445"`
	PublicURL string        `yaml:"public_url" env:"HYDRA_PUBLIC_URL" envDefault:"http://localhost:4444"`
	Timeout  time.Duration `yaml:"timeout" env:"HYDRA_TIMEOUT" envDefault:"10s"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	EncryptionKey string         `yaml:"encryption_key" env:"ENCRYPTION_KEY"` // 32-byte key for AES-256
	TOTPIssuer    string         `yaml:"totp_issuer" env:"TOTP_ISSUER" envDefault:"ARauth Identity"`
	JWT           JWTConfig      `yaml:"jwt"`
	Password      PasswordConfig `yaml:"password"`
	MFA           MFAConfig      `yaml:"mfa"`
	RateLimit     RateLimitConfig `yaml:"rate_limit"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Issuer          string        `yaml:"issuer" env:"JWT_ISSUER" envDefault:"https://iam.example.com"`
	AccessTokenTTL  time.Duration `yaml:"access_token_ttl" env:"JWT_ACCESS_TOKEN_TTL" envDefault:"15m"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl" env:"JWT_REFRESH_TOKEN_TTL" envDefault:"30d"`
	IDTokenTTL      time.Duration `yaml:"id_token_ttl" env:"JWT_ID_TOKEN_TTL" envDefault:"1h"`
	SigningKeyPath  string        `yaml:"signing_key_path" env:"JWT_SIGNING_KEY_PATH"`
	Secret          string        `yaml:"secret" env:"JWT_SECRET"`
	RememberMe      RememberMeConfig `yaml:"remember_me"`
	TokenRotation   bool          `yaml:"token_rotation" env:"JWT_TOKEN_ROTATION" envDefault:"true"`
	RequireMFAForExtendedSessions bool `yaml:"require_mfa_for_extended_sessions" env:"JWT_REQUIRE_MFA_EXTENDED" envDefault:"false"`
}

// RememberMeConfig holds Remember Me configuration
type RememberMeConfig struct {
	Enabled          bool          `yaml:"enabled" env:"JWT_REMEMBER_ME_ENABLED" envDefault:"true"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl" env:"JWT_REMEMBER_ME_REFRESH_TTL" envDefault:"90d"`
	AccessTokenTTL  time.Duration `yaml:"access_token_ttl" env:"JWT_REMEMBER_ME_ACCESS_TTL" envDefault:"60m"`
}

// PasswordConfig holds password policy configuration
type PasswordConfig struct {
	MinLength      int  `yaml:"min_length" env:"PASSWORD_MIN_LENGTH" envDefault:"12"`
	RequireUpper   bool `yaml:"require_uppercase" env:"PASSWORD_REQUIRE_UPPERCASE" envDefault:"true"`
	RequireLower   bool `yaml:"require_lowercase" env:"PASSWORD_REQUIRE_LOWERCASE" envDefault:"true"`
	RequireNumber  bool `yaml:"require_number" env:"PASSWORD_REQUIRE_NUMBER" envDefault:"true"`
	RequireSpecial bool `yaml:"require_special" env:"PASSWORD_REQUIRE_SPECIAL" envDefault:"true"`
}

// MFAConfig holds MFA configuration
type MFAConfig struct {
	Issuer string `yaml:"issuer" env:"MFA_ISSUER" envDefault:"ARauth Identity"`
	Period int    `yaml:"period" env:"MFA_PERIOD" envDefault:"30"`
	Digits int    `yaml:"digits" env:"MFA_DIGITS" envDefault:"6"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	LoginAttempts int           `yaml:"login_attempts" env:"RATE_LIMIT_LOGIN_ATTEMPTS" envDefault:"5"`
	LoginWindow   time.Duration `yaml:"login_window" env:"RATE_LIMIT_LOGIN_WINDOW" envDefault:"1m"`
	MFAAttempts   int           `yaml:"mfa_attempts" env:"RATE_LIMIT_MFA_ATTEMPTS" envDefault:"5"`
	MFAWindow     time.Duration `yaml:"mfa_window" env:"RATE_LIMIT_MFA_WINDOW" envDefault:"5m"`
	APIRequests   int           `yaml:"api_requests" env:"RATE_LIMIT_API_REQUESTS" envDefault:"100"`
	APIWindow     time.Duration `yaml:"api_window" env:"RATE_LIMIT_API_WINDOW" envDefault:"1m"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level     string `yaml:"level" env:"LOG_LEVEL" envDefault:"info"`
	Format    string `yaml:"format" env:"LOG_FORMAT" envDefault:"json"`
	Output    string `yaml:"output" env:"LOG_OUTPUT" envDefault:"stdout"`
	FilePath  string `yaml:"file_path" env:"LOG_FILE_PATH" envDefault:"/var/log/iam/api.log"`
	MaxSize   int    `yaml:"max_size" env:"LOG_MAX_SIZE" envDefault:"100"`
	MaxBackups int   `yaml:"max_backups" env:"LOG_MAX_BACKUPS" envDefault:"5"`
	MaxAge    int    `yaml:"max_age" env:"LOG_MAX_AGE" envDefault:"30"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled" env:"METRICS_ENABLED" envDefault:"true"`
	Path    string `yaml:"path" env:"METRICS_PATH" envDefault:"/metrics"`
	Port    int    `yaml:"port" env:"METRICS_PORT" envDefault:"9090"`
}

