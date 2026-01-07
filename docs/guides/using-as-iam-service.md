# Using Nuage Identity as an IAM Service

This guide explains how to use Nuage Identity IAM in your applications, similar to how you would use Keycloak, Auth0, or other IAM solutions.

## 🎯 Overview

Nuage Identity can be used in multiple ways:
1. **Docker Image** - Pull and run (like Keycloak)
2. **Docker Compose** - Complete stack with dependencies
3. **Kubernetes** - Helm chart deployment
4. **Source Code** - Clone and build
5. **Cloud Services** - Deploy to AWS/GCP/Azure

## 🐳 Method 1: Docker Image (Recommended)

### Pull and Run

```bash
# Pull the image (when published to Docker Hub)
docker pull nuage-identity/iam-api:latest

# Run with environment variables
docker run -d \
  --name nuage-iam \
  -p 8080:8080 \
  -e DATABASE_HOST=postgres \
  -e DATABASE_PORT=5432 \
  -e DATABASE_NAME=iam \
  -e DATABASE_USER=iam_user \
  -e DATABASE_PASSWORD=your_password \
  -e JWT_SECRET=your-jwt-secret \
  -e ENCRYPTION_KEY=your-32-byte-encryption-key \
  nuage-identity/iam-api:latest
```

### With Docker Compose

```yaml
version: '3.8'

services:
  # Your PostgreSQL (or use existing)
  postgres:
    image: postgres:14-alpine
    environment:
      POSTGRES_DB: iam
      POSTGRES_USER: iam_user
      POSTGRES_PASSWORD: ${DATABASE_PASSWORD}
    volumes:
      - postgres-data:/var/lib/postgresql/data

  # Nuage Identity IAM
  iam-api:
    image: nuage-identity/iam-api:latest
    ports:
      - "8080:8080"
    environment:
      DATABASE_HOST: postgres
      DATABASE_PORT: 5432
      DATABASE_NAME: iam
      DATABASE_USER: iam_user
      DATABASE_PASSWORD: ${DATABASE_PASSWORD}
      JWT_SECRET: ${JWT_SECRET}
      ENCRYPTION_KEY: ${ENCRYPTION_KEY}
    depends_on:
      - postgres
    volumes:
      - ./migrations:/migrations  # If you have custom migrations

volumes:
  postgres-data:
```

Run:
```bash
docker-compose up -d
```

## 📦 Method 2: Docker Compose (Complete Stack)

### Quick Start with All Dependencies

```bash
# Clone the repository
git clone https://github.com/your-org/nuage-identity.git
cd nuage-identity

# Copy environment file
cp .env.example .env

# Edit .env with your configuration
# Set DATABASE_PASSWORD, JWT_SECRET, etc.

# Start everything (IAM API + PostgreSQL + Redis + Hydra)
docker-compose up -d

# Run migrations
docker-compose exec iam-api migrate -path /migrations -database "postgres://iam_user:${DATABASE_PASSWORD}@postgres:5432/iam?sslmode=disable" up

# Access API
curl http://localhost:8080/health
```

## ☸️ Method 3: Kubernetes (Helm Chart)

### Install via Helm

```bash
# Add Helm repository (when published)
helm repo add nuage-identity https://charts.nuage-identity.com
helm repo update

# Install
helm install nuage-iam nuage-identity/nuage-identity \
  --set database.host=postgres \
  --set database.password=your_password \
  --set secrets.jwtSecret=your-jwt-secret \
  --set secrets.encryptionKey=your-32-byte-key

# Or use values file
helm install nuage-iam nuage-identity/nuage-identity \
  -f my-values.yaml
```

### Example values.yaml

```yaml
replicaCount: 3

database:
  host: postgres-service
  port: 5432
  name: iam_db
  user: iam_user

secrets:
  databasePassword: "your-password"
  jwtSecret: "your-jwt-secret"
  encryptionKey: "your-32-byte-encryption-key"

ingress:
  enabled: true
  hosts:
    - host: iam.example.com
      paths:
        - path: /
          pathType: Prefix
```

## 🔨 Method 4: Build from Source

### Clone and Build

```bash
# Clone repository
git clone https://github.com/your-org/nuage-identity.git
cd nuage-identity

# Build Docker image
docker build -t nuage-identity/iam-api:latest .

# Or build binary
go build -o iam-api ./cmd/server

# Run
./iam-api
```

### Build and Push to Registry

```bash
# Build
docker build -t your-registry/nuage-identity:latest .

# Push
docker push your-registry/nuage-identity:latest

# Use in your apps
docker pull your-registry/nuage-identity:latest
```

## 🔌 Integration Methods

### 1. REST API Integration

Your application makes HTTP requests to the IAM API:

```javascript
// Login
const response = await fetch('http://iam-api:8080/api/v1/auth/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-Tenant-ID': 'your-tenant-id'
  },
  body: JSON.stringify({
    username: 'user@example.com',
    password: 'password'
  })
});

const { access_token, refresh_token } = await response.json();

// Use token for API calls
const userResponse = await fetch('http://iam-api:8080/api/v1/users', {
  headers: {
    'Authorization': `Bearer ${access_token}`,
    'X-Tenant-ID': 'your-tenant-id'
  }
});
```

### 2. OAuth2/OIDC Integration

Since Nuage Identity uses ORY Hydra, you can use standard OAuth2/OIDC flows:

```javascript
// OAuth2 Authorization Code Flow
const authUrl = `http://hydra:4444/oauth2/auth?` +
  `client_id=your-client-id&` +
  `redirect_uri=http://your-app.com/callback&` +
  `response_type=code&` +
  `scope=openid profile email`;

// Redirect user to authUrl
// Handle callback with authorization code
// Exchange code for tokens
```

### 3. SDK Integration (Future)

```javascript
// JavaScript SDK (when available)
import { NuageIdentity } from '@nuage-identity/sdk';

const iam = new NuageIdentity({
  baseURL: 'http://iam-api:8080',
  tenantId: 'your-tenant-id'
});

// Login
const tokens = await iam.auth.login({
  username: 'user@example.com',
  password: 'password'
});

// Get users
const users = await iam.users.list();
```

## 📊 Comparison with Keycloak

| Feature | Nuage Identity | Keycloak |
|---------|---------------|----------|
| **Deployment** | Docker, K8s, Source | Docker, K8s, Source |
| **UI** | Headless (bring your own) | Built-in admin UI |
| **OAuth2/OIDC** | ✅ (via Hydra) | ✅ |
| **Multi-tenant** | ✅ Native | ⚠️ Realm-based |
| **Database** | PostgreSQL, MySQL, etc. | PostgreSQL, MySQL, etc. |
| **API-First** | ✅ | ⚠️ UI-first |
| **Lightweight** | ✅ | ⚠️ Heavier |
| **Customization** | ✅ Full control | ⚠️ Limited |
| **MFA** | ✅ TOTP | ✅ Multiple methods |
| **RBAC** | ✅ | ✅ |

## 🚀 Quick Start Examples

### Example 1: Simple Docker Setup

```bash
# 1. Pull image
docker pull nuage-identity/iam-api:latest

# 2. Run with your database
docker run -d \
  --name iam \
  -p 8080:8080 \
  -e DATABASE_HOST=your-postgres-host \
  -e DATABASE_PORT=5432 \
  -e DATABASE_NAME=iam \
  -e DATABASE_USER=iam_user \
  -e DATABASE_PASSWORD=your_password \
  -e JWT_SECRET=your-secret \
  nuage-identity/iam-api:latest

# 3. Test
curl http://localhost:8080/health
```

### Example 2: Integration in Your App

```python
# Python example
import requests

IAM_API_URL = "http://iam-api:8080"
TENANT_ID = "your-tenant-id"

# Login
response = requests.post(
    f"{IAM_API_URL}/api/v1/auth/login",
    headers={"X-Tenant-ID": TENANT_ID},
    json={
        "username": "user@example.com",
        "password": "password"
    }
)

tokens = response.json()
access_token = tokens["access_token"]

# Use token
users_response = requests.get(
    f"{IAM_API_URL}/api/v1/users",
    headers={
        "Authorization": f"Bearer {access_token}",
        "X-Tenant-ID": TENANT_ID
    }
)

users = users_response.json()
```

### Example 3: Using Existing Database

```bash
# Use your existing PostgreSQL
docker run -d \
  --name nuage-iam \
  -p 8080:8080 \
  -e DATABASE_HOST=your-existing-postgres \
  -e DATABASE_PORT=5432 \
  -e DATABASE_NAME=your-db \
  -e DATABASE_USER=your-user \
  -e DATABASE_PASSWORD=your-password \
  nuage-identity/iam-api:latest
```

## 🔧 Configuration

### Environment Variables

All configuration via environment variables:

```bash
# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# Database
DATABASE_HOST=postgres
DATABASE_PORT=5432
DATABASE_NAME=iam
DATABASE_USER=iam_user
DATABASE_PASSWORD=your_password

# Security
JWT_SECRET=your-jwt-secret
ENCRYPTION_KEY=your-32-byte-key

# Redis (optional)
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=your-redis-password

# Hydra (OAuth2)
HYDRA_ADMIN_URL=http://hydra:4445
```

### Config File

Or use YAML config file:

```yaml
server:
  port: 8080

database:
  host: postgres
  port: 5432
  name: iam
  user: iam_user
  password: "${DATABASE_PASSWORD}"

security:
  jwt:
    secret: "${JWT_SECRET}"
  encryption_key: "${ENCRYPTION_KEY}"
```

Mount config file:
```bash
docker run -v $(pwd)/config.yaml:/config/config.yaml \
  nuage-identity/iam-api:latest
```

## 📚 API Documentation

### Base URL
- Local: `http://localhost:8080`
- Production: `https://iam.example.com`

### Endpoints

```
POST   /api/v1/auth/login          # Login
GET    /api/v1/users               # List users
POST   /api/v1/users               # Create user
GET    /api/v1/users/:id           # Get user
PUT    /api/v1/users/:id           # Update user
DELETE /api/v1/users/:id           # Delete user

POST   /api/v1/tenants             # Create tenant
GET    /api/v1/tenants             # List tenants

POST   /api/v1/roles               # Create role
GET    /api/v1/roles               # List roles

POST   /api/v1/mfa/enroll          # Enroll MFA
POST   /api/v1/mfa/verify         # Verify MFA
```

See [API Documentation](../api/README.md) for complete API reference.

## 🔐 Security Best Practices

1. **Use HTTPS in Production**
   ```bash
   # Use reverse proxy (nginx, traefik) with SSL
   ```

2. **Secure Secrets**
   ```bash
   # Use secrets management (Kubernetes secrets, AWS Secrets Manager)
   # Never commit secrets to version control
   ```

3. **Configure CORS**
   ```go
   // Allow only your frontend domains
   c.Writer.Header().Set("Access-Control-Allow-Origin", "https://your-app.com")
   ```

4. **Rate Limiting**
   - Already configured in the API
   - Adjust limits in config

## 🎯 Use Cases

### 1. Microservices Architecture
```yaml
# Deploy IAM as a shared service
services:
  iam-api:
    image: nuage-identity/iam-api:latest
    # ... config
  
  your-service-1:
    # Uses IAM for authentication
    environment:
      IAM_API_URL: http://iam-api:8080
  
  your-service-2:
    # Uses IAM for authentication
    environment:
      IAM_API_URL: http://iam-api:8080
```

### 2. Multi-Tenant SaaS
- Each tenant gets isolated data
- Single IAM instance serves all tenants
- Tenant ID in headers for isolation

### 3. API Gateway Integration
- Deploy behind API Gateway (Kong, AWS API Gateway)
- Use for authentication/authorization
- Validate tokens at gateway level

## 📦 Distribution Methods

### 1. Docker Hub
```bash
# Publish to Docker Hub
docker tag nuage-identity/iam-api:latest your-org/nuage-identity:latest
docker push your-org/nuage-identity:latest

# Others can pull
docker pull your-org/nuage-identity:latest
```

### 2. GitHub Container Registry
```bash
# Publish to GHCR
docker tag nuage-identity/iam-api:latest ghcr.io/your-org/nuage-identity:latest
docker push ghcr.io/your-org/nuage-identity:latest
```

### 3. Helm Chart Repository
```bash
# Package and publish Helm chart
helm package helm/nuage-identity
helm push nuage-identity-1.0.0.tgz oci://your-registry/charts
```

### 4. Source Code
```bash
# Others can clone and build
git clone https://github.com/your-org/nuage-identity.git
cd nuage-identity
docker build -t nuage-identity:latest .
```

## 🆘 Support & Resources

- **Documentation**: [Full Documentation](../README.md)
- **API Reference**: [API Documentation](../api/README.md)
- **Examples**: [Integration Examples](../guides/integration-guide.md)
- **GitHub**: [Repository](https://github.com/your-org/nuage-identity)

## ✅ Summary

**Nuage Identity can be used just like Keycloak:**

1. ✅ **Docker Image** - Pull and run
2. ✅ **Docker Compose** - Complete stack
3. ✅ **Kubernetes** - Helm chart
4. ✅ **Source Code** - Clone and build
5. ✅ **REST API** - Standard HTTP integration
6. ✅ **OAuth2/OIDC** - Standard protocols
7. ✅ **Multi-tenant** - Native support
8. ✅ **API-First** - Headless architecture

**Key Difference**: Nuage Identity is headless (no built-in UI), giving you full control over the user experience while providing all IAM functionality via API.

---

**Last Updated**: 2024

