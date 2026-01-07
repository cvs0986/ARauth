# Database Configuration Guide

## ✅ Confirmation: Database is NOT Hardcoded

The Nuage Identity IAM application uses a **fully configurable** database connection. Nothing is hardcoded!

## 🔧 Configuration Methods

The application supports **three ways** to configure the database (in priority order):

1. **Environment Variables** (Highest Priority) - Overrides everything
2. **YAML Config Files** - With environment variable expansion
3. **Default Values** - Only used if nothing else is set

## 📋 Configuration Options

### Environment Variables

You can set these environment variables to configure your database:

```bash
export DATABASE_HOST="localhost"
export DATABASE_PORT="5433"        # Your PostgreSQL port
export DATABASE_NAME="iam"          # Your database name
export DATABASE_USER="your_user"    # Your PostgreSQL user
export DATABASE_PASSWORD="your_password"  # Your PostgreSQL password
export DATABASE_SSL_MODE="disable"  # or "require", "verify-full", etc.
```

### YAML Config File

Edit `config/config.yaml` or `config/config.dev.yaml`:

```yaml
database:
  host: "localhost"
  port: 5433                    # Your PostgreSQL port
  name: "iam"                    # Your database name
  user: "your_user"              # Your PostgreSQL user
  password: "${DATABASE_PASSWORD}" # From environment variable
  ssl_mode: "disable"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m
```

### Default Values (if nothing is set)

- Host: `localhost`
- Port: `5432`
- Name: `iam`
- User: `iam_user`
- SSL Mode: `disable`

## 🚀 Quick Setup for Your PostgreSQL (Port 5433)

### Option 1: Using Environment Variables (Recommended)

```bash
# Set your database configuration
export DATABASE_HOST="localhost"
export DATABASE_PORT="5433"
export DATABASE_NAME="iam"  # or your database name
export DATABASE_USER="your_postgres_user"
export DATABASE_PASSWORD="your_postgres_password"
export DATABASE_SSL_MODE="disable"

# Run the application
go run cmd/server/main.go
```

### Option 2: Using Config File

Edit `config/config.dev.yaml`:

```yaml
database:
  host: "localhost"
  port: 5433
  name: "iam"
  user: "your_postgres_user"
  password: "${DATABASE_PASSWORD}"
  ssl_mode: "disable"
```

Then set the password:
```bash
export DATABASE_PASSWORD="your_postgres_password"
go run cmd/server/main.go
```

### Option 3: Create a Custom Config File

Create `config/config.local.yaml`:

```yaml
database:
  host: "localhost"
  port: 5433
  name: "iam"
  user: "your_postgres_user"
  password: "your_password_here"  # Or use ${DATABASE_PASSWORD}
  ssl_mode: "disable"
```

Run with custom config:
```bash
export CONFIG_PATH="config/config.local.yaml"
go run cmd/server/main.go
```

## 🔍 How It Works

The configuration loader (`config/loader/loader.go`) works as follows:

1. **Loads YAML file** (if `CONFIG_PATH` is set or default `config/config.yaml`)
2. **Expands environment variables** in YAML (e.g., `${DATABASE_PASSWORD}`)
3. **Overrides with environment variables** (if set)
4. **Applies defaults** (only for missing values)

### Priority Order:
```
Environment Variables > YAML Config > Defaults
```

## ✅ Verification

### Check Current Configuration

The application logs the database connection on startup:

```bash
go run cmd/server/main.go
# Look for: "Database connection established"
```

### Test Connection

```bash
# Test with psql
psql -h localhost -p 5433 -U your_user -d iam

# Or test from application
curl http://localhost:8080/health
```

## 📝 Example: Complete Setup

```bash
# 1. Set environment variables
export DATABASE_HOST="localhost"
export DATABASE_PORT="5433"
export DATABASE_NAME="iam"
export DATABASE_USER="postgres"  # or your user
export DATABASE_PASSWORD="your_password"
export DATABASE_SSL_MODE="disable"

# 2. (Optional) Set other required secrets
export JWT_SECRET="your-jwt-secret-key"
export REDIS_PASSWORD=""  # if using Redis

# 3. Run migrations (if needed)
export DATABASE_URL="postgres://${DATABASE_USER}:${DATABASE_PASSWORD}@${DATABASE_HOST}:${DATABASE_PORT}/${DATABASE_NAME}?sslmode=${DATABASE_SSL_MODE}"
make migrate-up

# 4. Start the application
go run cmd/server/main.go
```

## 🔐 Security Best Practices

1. **Never commit passwords** to version control
2. **Use environment variables** for sensitive data
3. **Use `${VAR}` syntax** in YAML for environment variable expansion
4. **Use `.env` files** (with `.gitignore`) for local development
5. **Use secrets management** in production (Kubernetes secrets, etc.)

## 🐳 Docker/Production

In production, use environment variables or mounted config files:

```yaml
# docker-compose.yml
services:
  iam-api:
    environment:
      DATABASE_HOST: "postgres"
      DATABASE_PORT: "5432"
      DATABASE_NAME: "iam"
      DATABASE_USER: "iam_user"
      DATABASE_PASSWORD: "${DATABASE_PASSWORD}"
```

## ❓ Troubleshooting

### Connection Refused
- Check PostgreSQL is running: `psql -h localhost -p 5433 -U your_user -d iam`
- Verify port is correct: `netstat -tuln | grep 5433`
- Check firewall settings

### Authentication Failed
- Verify username and password
- Check PostgreSQL user permissions
- Verify database exists

### SSL Mode Issues
- For local development: `ssl_mode: "disable"`
- For production: `ssl_mode: "require"` or `"verify-full"`

## 📚 Related Documentation

- [Configuration Guide](../technical/configuration.md)
- [Deployment Guide](../deployment/production-guide.md)
- [Quick Start](../guides/getting-started.md)

---

**Last Updated**: 2024

