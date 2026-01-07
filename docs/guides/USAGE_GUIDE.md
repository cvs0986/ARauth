# ARauth Identity - Usage Guide

## 🎯 How Others Can Use This IAM Solution

ARauth Identity can be used in your applications just like **Keycloak**, **Auth0**, or other IAM solutions.

## 🚀 Quick Start Methods

### Method 1: Docker Image (Like Keycloak)

```bash
# Pull and run - just like Keycloak!
docker pull arauth-identity/iam-api:latest

docker run -d \
  --name nuage-iam \
  -p 8080:8080 \
  -e DATABASE_HOST=postgres \
  -e DATABASE_PORT=5432 \
  -e DATABASE_NAME=iam \
  -e DATABASE_USER=iam_user \
  -e DATABASE_PASSWORD=your_password \
  -e JWT_SECRET=your-secret \
  arauth-identity/iam-api:latest
```

### Method 2: Docker Compose

```bash
# Clone repository
git clone https://github.com/your-org/arauth-identity.git
cd arauth-identity

# Start everything
docker-compose up -d
```

### Method 3: Kubernetes (Helm)

```bash
# Install via Helm
helm install nuage-iam arauth-identity/arauth-identity \
  --set database.host=postgres \
  --set database.password=your_password
```

### Method 4: Build from Source

```bash
# Clone and build
git clone https://github.com/your-org/arauth-identity.git
cd arauth-identity
docker build -t arauth-identity/iam-api:latest .
```

## 📊 Comparison with Keycloak

| Feature | ARauth Identity | Keycloak |
|---------|---------------|----------|
| **Docker Image** | ✅ `docker pull arauth-identity/iam-api` | ✅ `docker pull quay.io/keycloak` |
| **Kubernetes** | ✅ Helm chart | ✅ Operator |
| **OAuth2/OIDC** | ✅ (via Hydra) | ✅ |
| **Multi-tenant** | ✅ Native | ⚠️ Realm-based |
| **UI** | Headless (bring your own) | Built-in admin UI |
| **API-First** | ✅ | ⚠️ UI-first |
| **Lightweight** | ✅ | ⚠️ Heavier |

## 🔌 Integration Examples

### REST API Integration

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

const { access_token } = await response.json();

// Use token
const users = await fetch('http://iam-api:8080/api/v1/users', {
  headers: {
    'Authorization': `Bearer ${access_token}`,
    'X-Tenant-ID': 'your-tenant-id'
  }
});
```

### OAuth2/OIDC Integration

```javascript
// Standard OAuth2 flow (via Hydra)
const authUrl = `http://hydra:4444/oauth2/auth?` +
  `client_id=your-client-id&` +
  `redirect_uri=http://your-app.com/callback&` +
  `response_type=code&` +
  `scope=openid profile email`;
```

## 📦 Distribution Methods

### 1. Docker Hub
```bash
# Publish
docker push your-org/arauth-identity:latest

# Others use
docker pull your-org/arauth-identity:latest
```

### 2. GitHub Container Registry
```bash
# Publish
docker push ghcr.io/your-org/arauth-identity:latest

# Others use
docker pull ghcr.io/your-org/arauth-identity:latest
```

### 3. Helm Chart Repository
```bash
# Package
helm package helm/arauth-identity

# Publish
helm push arauth-identity-1.0.0.tgz oci://your-registry/charts
```

### 4. Source Code
```bash
# Others clone and build
git clone https://github.com/your-org/arauth-identity.git
cd arauth-identity
docker build -t arauth-identity:latest .
```

## 🎯 Use Cases

### 1. Microservices
```yaml
services:
  iam-api:
    image: arauth-identity/iam-api:latest
  
  your-service:
    environment:
      IAM_API_URL: http://iam-api:8080
```

### 2. Multi-Tenant SaaS
- Single IAM instance
- Multiple tenants
- Isolated data per tenant

### 3. API Gateway
- Deploy behind API Gateway
- Use for authentication
- Validate tokens at gateway

## 🔧 Configuration

All via environment variables:

```bash
DATABASE_HOST=postgres
DATABASE_PORT=5432
DATABASE_NAME=iam
DATABASE_USER=iam_user
DATABASE_PASSWORD=your_password
JWT_SECRET=your-secret
ENCRYPTION_KEY=your-32-byte-key
```

## 📚 Documentation

- **[Using as IAM Service](./docs/guides/using-as-iam-service.md)** - Complete guide
- **[Docker Image Usage](./docs/deployment/docker-image-usage.md)** - Docker quick start
- **[API Documentation](./docs/api/README.md)** - API reference

## ✅ Summary

**Yes! Others can use this just like Keycloak:**

1. ✅ **Docker Image** - Pull and run
2. ✅ **Docker Compose** - Complete stack
3. ✅ **Kubernetes** - Helm chart
4. ✅ **Source Code** - Clone and build
5. ✅ **REST API** - Standard HTTP integration
6. ✅ **OAuth2/OIDC** - Standard protocols

**Key Advantage**: Headless architecture gives you full control over UI while providing all IAM functionality via API.

---

**Last Updated**: 2024

