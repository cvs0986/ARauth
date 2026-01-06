# Docker Compose Deployment

This document describes how to deploy Nuage Identity using Docker Compose for local development and on-premise deployments.

## 🎯 Prerequisites

- Docker 20.10+
- Docker Compose 2.0+
- 4GB+ RAM available
- Ports available: 8080, 5432, 6379, 4444, 4445

## 📦 Docker Compose Structure

```
docker-compose.yml
.env.example
```

## 🔧 Configuration

### docker-compose.yml

```yaml
version: '3.8'

services:
  postgres-iam:
    image: postgres:14-alpine
    container_name: postgres-iam
    environment:
      POSTGRES_DB: iam
      POSTGRES_USER: iam_user
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres-iam-data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U iam_user"]
      interval: 10s
      timeout: 5s
      retries: 5

  postgres-hydra:
    image: postgres:14-alpine
    container_name: postgres-hydra
    environment:
      POSTGRES_DB: hydra
      POSTGRES_USER: hydra
      POSTGRES_PASSWORD: ${HYDRA_DB_PASSWORD}
    volumes:
      - postgres-hydra-data:/var/lib/postgresql/data
    ports:
      - "5433:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U hydra"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: redis-iam
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis-data:/data
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "--raw", "incr", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5

  hydra-migrate:
    image: oryd/hydra:v2.0.0
    container_name: hydra-migrate
    command: migrate sql -e --yes
    environment:
      DSN: postgres://hydra:${HYDRA_DB_PASSWORD}@postgres-hydra:5432/hydra?sslmode=disable
    depends_on:
      postgres-hydra:
        condition: service_healthy

  hydra:
    image: oryd/hydra:v2.0.0
    container_name: hydra
    command: serve all --dev
    environment:
      DSN: postgres://hydra:${HYDRA_DB_PASSWORD}@postgres-hydra:5432/hydra?sslmode=disable
      URLS_SELF_ISSUER: http://localhost:4444
      URLS_CONSENT: http://localhost:3000/consent
      URLS_LOGIN: http://localhost:3000/login
    ports:
      - "4444:4444"  # Public
      - "4445:4445"  # Admin
    depends_on:
      hydra-migrate:
        condition: service_completed_successfully
      postgres-hydra:
        condition: service_healthy

  iam-api:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: iam-api
    environment:
      - SERVER_PORT=8080
      - SERVER_HOST=0.0.0.0
      - DATABASE_HOST=postgres-iam
      - DATABASE_PORT=5432
      - DATABASE_NAME=iam
      - DATABASE_USER=iam_user
      - DATABASE_PASSWORD=${POSTGRES_PASSWORD}
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=${REDIS_PASSWORD}
      - HYDRA_ADMIN_URL=http://hydra:4445
      - HYDRA_PUBLIC_URL=http://hydra:4444
      - JWT_SECRET=${JWT_SECRET}
    ports:
      - "8080:8080"
    depends_on:
      postgres-iam:
        condition: service_healthy
      redis:
        condition: service_healthy
      hydra:
        condition: service_started
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  postgres-iam-data:
  postgres-hydra-data:
  redis-data:
```

### .env.example

```env
# Database
POSTGRES_PASSWORD=change-me
HYDRA_DB_PASSWORD=change-me

# Redis
REDIS_PASSWORD=change-me

# JWT
JWT_SECRET=change-me-to-random-secret

# Hydra
HYDRA_SECRETS_SYSTEM=change-me-to-random-secret
```

## 🚀 Deployment Steps

### 1. Clone Repository

```bash
git clone <repository-url>
cd nuage-identity
```

### 2. Configure Environment

```bash
cp .env.example .env
# Edit .env with your values
```

### 3. Start Services

```bash
docker-compose up -d
```

### 4. Run Migrations

```bash
# IAM database migrations
docker-compose exec iam-api ./migrate up

# Hydra migrations (already run by hydra-migrate service)
```

### 5. Verify Deployment

```bash
# Check services
docker-compose ps

# Check logs
docker-compose logs -f iam-api

# Health check
curl http://localhost:8080/health
```

## 🔧 Development Mode

### Hot Reload

For development with hot reload:

```yaml
iam-api:
  volumes:
    - .:/app
  command: air  # Using air for hot reload
```

### Debug Mode

```yaml
iam-api:
  environment:
    - LOG_LEVEL=debug
```

## 📊 Service URLs

- **IAM API**: http://localhost:8080
- **Hydra Public**: http://localhost:4444
- **Hydra Admin**: http://localhost:4445
- **PostgreSQL (IAM)**: localhost:5432
- **PostgreSQL (Hydra)**: localhost:5433
- **Redis**: localhost:6379

## 🔄 Updates

### Rebuild and Restart

```bash
docker-compose up -d --build
```

### Restart Single Service

```bash
docker-compose restart iam-api
```

## 🧹 Cleanup

### Stop Services

```bash
docker-compose down
```

### Remove Volumes

```bash
docker-compose down -v
```

## 📚 Related Documentation

- [Kubernetes Deployment](./kubernetes.md) - Production deployment
- [Configuration](./configuration.md) - Configuration management

