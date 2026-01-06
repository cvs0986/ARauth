# Getting Started Guide

This guide will help you get started with Nuage Identity development.

## 🎯 Prerequisites

- Go 1.21+
- Docker and Docker Compose
- PostgreSQL 14+ (or use Docker)
- Redis 7+ (or use Docker)
- ORY Hydra v2.0+ (or use Docker)

## 🚀 Quick Start

### 1. Clone Repository

```bash
git clone <repository-url>
cd nuage-identity
```

### 2. Set Up Environment

```bash
cp .env.example .env
# Edit .env with your configuration
```

### 3. Start Dependencies

```bash
docker-compose up -d postgres-iam postgres-hydra redis hydra
```

### 4. Run Migrations

```bash
# IAM database migrations
go run cmd/migrate/main.go up

# Hydra migrations (if needed)
# Usually handled by Hydra container
```

### 5. Start Application

```bash
go run cmd/server/main.go
```

### 6. Verify

```bash
curl http://localhost:8080/health
```

## 📚 Next Steps

1. Read [Architecture Overview](../architecture/overview.md)
2. Review [API Design](../technical/api-design.md)
3. Check [Development Strategy](../planning/strategy.md)

## 🔧 Development Setup

### IDE Setup

**VS Code**:
- Install Go extension
- Install Go tools

**GoLand**:
- Configure Go SDK
- Set up run configurations

### Code Quality Tools

```bash
# Format code
go fmt ./...

# Lint code
golangci-lint run

# Run tests
go test ./...
```

## 📖 Related Documentation

- [Integration Guide](./integration-guide.md) - Client integration
- [Troubleshooting](./troubleshooting.md) - Common issues

