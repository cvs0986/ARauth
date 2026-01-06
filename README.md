# Nuage Identity - Headless IAM Platform

A lightweight, headless Identity & Access Management (IAM) platform powered by ORY Hydra, designed for modern applications that bring their own login UI.

## 🎯 Overview

Nuage Identity is a production-grade, API-first IAM solution that provides:
- **Headless Authentication** - No hosted login UI, apps bring their own
- **OAuth2/OIDC Compliance** - Powered by ORY Hydra
- **Stateless & Scalable** - Horizontally scalable architecture
- **Database Agnostic** - Support for PostgreSQL, MySQL, MSSQL, MongoDB
- **Enterprise Ready** - MFA, rate limiting, security best practices

## 🏗️ Architecture

```
Client App (Web/Mobile)
 └── Custom Login UI
       └── IAM API (/auth/login)
             ├── Identity Service
             ├── Credential Validation
             ├── MFA (optional)
             ├── Claims Builder
             └── ORY Hydra Admin API
                    └── OAuth2 / OIDC Tokens
```

## 🚀 Quick Start

### Prerequisites

- Go 1.21+ (will be installed automatically)
- Docker and Docker Compose
- PostgreSQL 14+ (via Docker)
- Redis 7+ (via Docker)
- ORY Hydra v2.0+ (via Docker)

### Installation

#### 1. Install Go (if not already installed)

```bash
# Run the setup script
bash scripts/setup-go.sh

# Add Go to your PATH (if needed)
export PATH=$PATH:~/go-install/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

#### 2. Install Development Tools

```bash
# Install all required development tools
bash scripts/install-dev-tools.sh
```

#### 3. Set Up Environment

```bash
# Copy environment template
cp .env.example .env

# Edit .env with your configuration
# Set DATABASE_PASSWORD, REDIS_PASSWORD, JWT_SECRET, etc.
```

#### 4. Start Dependencies

```bash
# Start PostgreSQL, Redis, and Hydra
make docker-up

# Or manually:
docker-compose up -d postgres-iam postgres-hydra redis hydra
```

#### 5. Run Database Migrations

```bash
# Set database URL
export DATABASE_URL="postgres://iam_user:change-me@localhost:5432/iam?sslmode=disable"

# Run migrations
make migrate-up

# Or use the script:
bash scripts/migrate.sh up
```

#### 6. Start Application

```bash
# Run the application
make run

# Or directly:
go run cmd/server/main.go
```

#### 7. Verify

```bash
# Health check
curl http://localhost:8080/health
```

## 📚 Documentation

Comprehensive documentation is available in the [`docs/`](./docs/) directory:

- **[Architecture Overview](./docs/architecture/overview.md)** - System architecture
- **[Getting Started](./docs/guides/getting-started.md)** - Quick start guide
- **[API Design](./docs/technical/api-design.md)** - API specifications
- **[Deployment Guide](./docs/deployment/kubernetes.md)** - Kubernetes deployment
- **[Integration Guide](./docs/guides/integration-guide.md)** - Client integration
- **[Development Strategy](./docs/planning/strategy.md)** - Development approach

## 🛠️ Development

### Build

```bash
make build
```

### Run Tests

```bash
make test
make test-coverage  # With coverage report
```

### Code Quality

```bash
make lint   # Run linters
make fmt    # Format code
```

### Database Migrations

```bash
make migrate-up      # Run migrations
make migrate-down    # Rollback migrations
make migrate-create  # Create new migration
```

### Docker Commands

```bash
make docker-up      # Start services
make docker-down    # Stop services
make docker-clean   # Stop and remove volumes
```

## 🧩 Key Components

1. **IAM API** - Core authentication and authorization service
2. **Identity Service** - User, tenant, role, and permission management
3. **Auth Service** - Headless authentication endpoints
4. **OAuth2/OIDC** - ORY Hydra integration for token issuance
5. **Claims Builder** - JWT claims generation with tenant, roles, permissions

## 🔐 Security Features

- ✅ Argon2id password hashing
- ✅ MFA support (TOTP + recovery codes)
- ✅ Rate limiting
- ✅ Refresh token rotation
- ✅ Short-lived access tokens
- ✅ Key rotation via JWKS

## 📁 Project Structure

```
iam/
├── cmd/server/          # Application entry point
├── api/                 # HTTP API layer
├── auth/                # Authentication service
├── identity/            # Identity management
├── policy/              # Authorization
├── storage/             # Data access layer
├── security/             # Security utilities
├── config/               # Configuration
├── internal/             # Internal utilities
├── migrations/           # Database migrations
└── scripts/              # Development scripts
```

## 📋 Project Status

This project is in active development. See the [Roadmap](./docs/planning/roadmap.md) for current status and planned features.

## 🤝 Contributing

Contributions are welcome! Please read our contributing guidelines and code of conduct before submitting pull requests.

## 📄 License

[Add your license here]

## 🔗 Links

- **Documentation**: [docs/](./docs/)
- **Requirements**: [requirement.md](./requirement.md)

## 📧 Contact

For questions or support, please open an issue on GitHub.

---

**Built with ❤️ for modern, headless authentication**
