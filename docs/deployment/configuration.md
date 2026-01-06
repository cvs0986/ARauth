# Configuration Management

This document describes configuration management for Nuage Identity.

## 🎯 Configuration Principles

1. **Environment-based**: Different configs for dev/staging/prod
2. **Secure**: Secrets never in code
3. **Validated**: Config validation on startup
4. **Documented**: All config options documented

## 📋 Configuration Structure

### Config File (config.yaml)

```yaml
server:
  port: 8080
  host: "0.0.0.0"
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s

database:
  host: "localhost"
  port: 5432
  name: "iam"
  user: "iam_user"
  password: "${DB_PASSWORD}"  # From environment
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m
  ssl_mode: "disable"

redis:
  host: "localhost"
  port: 6379
  password: "${REDIS_PASSWORD}"  # From environment
  db: 0
  pool_size: 10
  min_idle_conns: 5

hydra:
  admin_url: "http://localhost:4445"
  public_url: "http://localhost:4444"
  timeout: 10s

security:
  jwt:
    issuer: "https://iam.example.com"
    access_token_ttl: 15m
    refresh_token_ttl: 30d
    id_token_ttl: 1h
    signing_key_path: "/etc/iam/jwt.key"
  password:
    min_length: 12
    require_uppercase: true
    require_lowercase: true
    require_number: true
    require_special: true
  mfa:
    issuer: "Nuage Identity"
    period: 30
    digits: 6
  rate_limit:
    login_attempts: 5
    login_window: 1m
    mfa_attempts: 5
    mfa_window: 5m
    api_requests: 100
    api_window: 1m

logging:
  level: "info"  # debug, info, warn, error
  format: "json"  # json, text
  output: "stdout"  # stdout, file
  file_path: "/var/log/iam/api.log"
  max_size: 100  # MB
  max_backups: 5
  max_age: 30  # days

metrics:
  enabled: true
  path: "/metrics"
  port: 9090
```

## 🔐 Environment Variables

### Required Variables

```bash
# Database
DB_PASSWORD=your-database-password

# Redis
REDIS_PASSWORD=your-redis-password

# JWT
JWT_SECRET=your-jwt-secret-key

# Hydra
HYDRA_SECRETS_SYSTEM=your-hydra-secret
```

### Optional Variables

```bash
# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=iam
DB_USER=iam_user

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

## 🏗️ Configuration Loading

### Go Implementation

```go
type Config struct {
    Server   ServerConfig   `yaml:"server"`
    Database DatabaseConfig `yaml:"database"`
    Redis    RedisConfig    `yaml:"redis"`
    Hydra    HydraConfig    `yaml:"hydra"`
    Security SecurityConfig `yaml:"security"`
    Logging  LoggingConfig  `yaml:"logging"`
    Metrics  MetricsConfig  `yaml:"metrics"`
}

func LoadConfig(path string) (*Config, error) {
    // Load from file
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    // Expand environment variables
    expanded := os.ExpandEnv(string(data))
    
    var config Config
    if err := yaml.Unmarshal([]byte(expanded), &config); err != nil {
        return nil, err
    }
    
    // Validate
    if err := config.Validate(); err != nil {
        return nil, err
    }
    
    return &config, nil
}
```

## ✅ Configuration Validation

### Validation Rules

```go
func (c *Config) Validate() error {
    if c.Server.Port < 1 || c.Server.Port > 65535 {
        return errors.New("invalid server port")
    }
    
    if c.Database.Host == "" {
        return errors.New("database host required")
    }
    
    if c.Security.JWT.AccessTokenTTL < 1*time.Minute {
        return errors.New("access token TTL too short")
    }
    
    // ... more validations
    
    return nil
}
```

## 🔄 Configuration by Environment

### Development

```yaml
# config.dev.yaml
server:
  port: 8080
database:
  host: localhost
logging:
  level: debug
```

### Staging

```yaml
# config.staging.yaml
server:
  port: 8080
database:
  host: postgres-staging
logging:
  level: info
```

### Production

```yaml
# config.prod.yaml
server:
  port: 8080
database:
  host: postgres-prod
  ssl_mode: require
logging:
  level: warn
metrics:
  enabled: true
```

## 🔐 Secrets Management

### Kubernetes Secrets

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: iam-secrets
type: Opaque
stringData:
  database-password: <password>
  redis-password: <password>
  jwt-secret: <secret>
```

### Environment Variables

```bash
# Never commit to git
export DB_PASSWORD="secure-password"
export REDIS_PASSWORD="secure-password"
export JWT_SECRET="secure-secret"
```

### Secret Management Systems

- **HashiCorp Vault**: For on-premise
- **AWS Secrets Manager**: For AWS
- **Azure Key Vault**: For Azure
- **GCP Secret Manager**: For GCP

## 📊 Configuration Best Practices

### 1. Never Commit Secrets

- Use environment variables
- Use secret management systems
- Use `.env.example` for documentation

### 2. Validate on Startup

- Validate all required config
- Fail fast on invalid config
- Clear error messages

### 3. Use Defaults

- Sensible defaults for optional config
- Document all defaults
- Override via environment variables

### 4. Environment-Specific Configs

- Separate configs per environment
- Use environment variables for differences
- Version control config files (without secrets)

## 📚 Related Documentation

- [Kubernetes Deployment](./kubernetes.md) - K8s config
- [Docker Compose](./docker-compose.md) - Docker config

