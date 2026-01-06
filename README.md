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

- Go 1.21+
- Docker and Docker Compose
- PostgreSQL 14+ (or use Docker)
- Redis 7+ (or use Docker)
- ORY Hydra v2.0+ (or use Docker)

### Installation

```bash
# Clone the repository
git clone https://github.com/cvs0986/ARauth.git
cd ARauth

# Set up environment
cp .env.example .env
# Edit .env with your configuration

# Start dependencies
docker-compose up -d postgres-iam postgres-hydra redis hydra

# Run migrations
go run cmd/migrate/main.go up

# Start application
go run cmd/server/main.go
```

## 📚 Documentation

Comprehensive documentation is available in the [`docs/`](./docs/) directory:

- **[Architecture Overview](./docs/architecture/overview.md)** - System architecture
- **[Getting Started](./docs/guides/getting-started.md)** - Quick start guide
- **[API Design](./docs/technical/api-design.md)** - API specifications
- **[Deployment Guide](./docs/deployment/kubernetes.md)** - Kubernetes deployment
- **[Integration Guide](./docs/guides/integration-guide.md)** - Client integration

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

## 📋 Project Status

This project is in active development. See the [Roadmap](./docs/planning/roadmap.md) for current status and planned features.

## 🤝 Contributing

Contributions are welcome! Please read our contributing guidelines and code of conduct before submitting pull requests.

## 📄 License

[Add your license here]

## 🔗 Links

- **Repository**: https://github.com/cvs0986/ARauth
- **Documentation**: [docs/](./docs/)
- **Requirements**: [requirement.md](./requirement.md)

## 📧 Contact

For questions or support, please open an issue on GitHub.

---

**Built with ❤️ for modern, headless authentication**

