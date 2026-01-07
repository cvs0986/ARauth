# Frontend-Backend Integration Guide

This document explains how the Admin Dashboard (frontend) and IAM API (backend) work together in different deployment scenarios: local development, Kubernetes, and cloud deployments.

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Client Browser                            │
│  ┌──────────────────────────────────────────────────────┐   │
│  │         Admin Dashboard (React App)                  │   │
│  │         - SPA (Single Page Application)               │   │
│  │         - Port: 3000 (dev) / 80 (prod)               │   │
│  └───────────────────┬─────────────────────────────────┘   │
└───────────────────────┼─────────────────────────────────────┘
                         │ HTTP/HTTPS
                         │ API Calls
                         │ (REST API)
                         ▼
┌─────────────────────────────────────────────────────────────┐
│              IAM API (Go + Gin)                             │
│              - REST API Server                              │
│              - Port: 8080 (internal)                        │
│              - Handles: Auth, Users, Roles, Permissions    │
└───────────────────┬───────────────────────────────────────┘
                     │
                     ▼
         ┌───────────────────────────┐
         │   PostgreSQL Database      │
         │   Redis Cache              │
         │   ORY Hydra                │
         └───────────────────────────┘
```

## 🔄 Communication Flow

### 1. Authentication Flow

```
Admin Dashboard                    IAM API
     │                                 │
     │  POST /api/v1/auth/login        │
     │  Headers: X-Tenant-ID           │
     │  Body: {username, password}     │
     ├────────────────────────────────>│
     │                                 │
     │  Response: {access_token, ...} │
     │<────────────────────────────────┤
     │                                 │
     │  Store token in localStorage    │
     │  or httpOnly cookie             │
     │                                 │
```

### 2. API Request Flow

```
Admin Dashboard                    IAM API
     │                                 │
     │  GET /api/v1/users              │
     │  Headers:                       │
     │    Authorization: Bearer <token> │
     │    X-Tenant-ID: <tenant-id>     │
     ├────────────────────────────────>│
     │                                 │
     │  Validate token                 │
     │  Check permissions              │
     │  Query database                 │
     │                                 │
     │  Response: {users: [...]}       │
     │<────────────────────────────────┤
     │                                 │
     │  Update UI with data            │
     │                                 │
```

## 📍 Deployment Scenarios

### Scenario 1: Local Development

#### Architecture
```
┌─────────────────────────────────────────────────────┐
│              Local Machine                          │
│                                                      │
│  ┌──────────────────┐    ┌──────────────────┐      │
│  │  Admin Dashboard │    │    IAM API        │      │
│  │  (Vite Dev)      │    │  (Go Server)      │      │
│  │  Port: 3000      │    │  Port: 8080       │      │
│  │  http://localhost│    │  http://localhost │      │
│  └────────┬─────────┘    └────────┬─────────┘     │
│           │                         │                │
│           │  API Calls              │                │
│           │  http://localhost:8080  │                │
│           └─────────┬──────────────┘                │
│                     │                                │
│           ┌─────────▼──────────┐                    │
│           │  PostgreSQL:5433   │                    │
│           │  Redis:6379         │                    │
│           └─────────────────────┘                    │
└─────────────────────────────────────────────────────┘
```

#### Configuration

**Admin Dashboard** (`.env`):
```bash
VITE_API_BASE_URL=http://localhost:8080
```

**IAM API** (`config/config.dev.yaml`):
```yaml
server:
  port: 8080
  host: "0.0.0.0"

database:
  host: "localhost"
  port: 5433
  name: "iam"
  user: "your_user"
  password: "${DATABASE_PASSWORD}"
```

**CORS Configuration**:
- Currently allows all origins (`*`) - suitable for development
- Frontend can make requests from `http://localhost:3000` to `http://localhost:8080`

#### Running Locally

**Terminal 1: Backend API**
```bash
# Set environment variables
export DATABASE_HOST="localhost"
export DATABASE_PORT="5433"
export DATABASE_NAME="iam"
export DATABASE_USER="your_user"
export DATABASE_PASSWORD="your_password"

# Start API
go run cmd/server/main.go
# API available at http://localhost:8080
```

**Terminal 2: Admin Dashboard**
```bash
cd frontend/admin-dashboard
npm install
npm run dev
# Dashboard available at http://localhost:3000
```

**Terminal 3: E2E Testing App** (optional)
```bash
cd frontend/e2e-test-app
npm install
npm run dev
# App available at http://localhost:3001
```

#### How They Connect

1. **Admin Dashboard** runs on `http://localhost:3000`
2. **IAM API** runs on `http://localhost:8080`
3. Dashboard makes API calls to `http://localhost:8080/api/v1/*`
4. CORS middleware allows cross-origin requests
5. Browser handles CORS preflight requests automatically

---

### Scenario 2: Kubernetes Deployment

#### Architecture
```
┌─────────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                        │
│                                                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              Ingress Controller                      │    │
│  │  (nginx-ingress)                                      │    │
│  │  - Routes: /admin → Admin Dashboard                  │    │
│  │  - Routes: /api → IAM API                            │    │
│  └──────────────┬──────────────────┬───────────────────┘    │
│                 │                  │                         │
│    ┌────────────▼──────┐  ┌───────▼──────────┐            │
│    │  Admin Dashboard  │  │    IAM API        │            │
│    │  Service          │  │    Service        │            │
│    │  (ClusterIP)      │  │    (ClusterIP)   │            │
│    │  Port: 80         │  │    Port: 80      │            │
│    └────────────┬──────┘  └───────┬──────────┘            │
│                 │                  │                         │
│    ┌────────────▼──────┐  ┌───────▼──────────┐            │
│    │  Admin Dashboard   │  │    IAM API        │            │
│    │  Pods (3 replicas) │  │    Pods (3 replicas)          │
│    │  - React SPA       │  │    - Go Server    │            │
│    │  - Nginx serving   │  │    - Port: 8080   │            │
│    └────────────────────┘  └───────┬──────────┘            │
│                                    │                         │
│                          ┌─────────▼──────────┐             │
│                          │  PostgreSQL       │             │
│                          │  Redis             │             │
│                          │  (StatefulSets)   │             │
│                          └───────────────────┘             │
└─────────────────────────────────────────────────────────────┘
```

#### Configuration

**Ingress** (`k8s/ingress.yaml`):
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: nuage-identity-ingress
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /$1
spec:
  rules:
  - host: iam.example.com
    http:
      paths:
      # Admin Dashboard
      - path: /admin(/|$)(.*)
        pathType: Prefix
        backend:
          service:
            name: admin-dashboard
            port:
              number: 80
      # IAM API
      - path: /api(/|$)(.*)
        pathType: Prefix
        backend:
          service:
            name: iam-api
            port:
              number: 80
```

**Admin Dashboard Service**:
```yaml
apiVersion: v1
kind: Service
metadata:
  name: admin-dashboard
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: 80
  selector:
    app: admin-dashboard
```

**IAM API Service**:
```yaml
apiVersion: v1
kind: Service
metadata:
  name: iam-api
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: iam-api
```

**Admin Dashboard Config** (`.env.production`):
```bash
VITE_API_BASE_URL=https://iam.example.com/api
```

**IAM API Config** (via ConfigMap):
```yaml
server:
  port: 8080
  host: "0.0.0.0"

database:
  host: "postgres-iam"  # Kubernetes service name
  port: 5432
  name: "iam_db"
  user: "iam_user"
  password: "${DATABASE_PASSWORD}"  # From Secret
```

**CORS Configuration** (should be updated for production):
```go
// Should allow specific origins, not "*"
c.Writer.Header().Set("Access-Control-Allow-Origin", "https://iam.example.com")
```

#### How They Connect

1. **User accesses**: `https://iam.example.com/admin`
2. **Ingress routes** to Admin Dashboard service
3. **Dashboard loads** React app (served by Nginx)
4. **Dashboard makes API calls** to `https://iam.example.com/api/v1/*`
5. **Ingress routes** API calls to IAM API service
6. **IAM API processes** requests and returns responses
7. **Same domain** = No CORS issues (or CORS configured for same domain)

---

### Scenario 3: Cloud Deployment (AWS/GCP/Azure)

#### Architecture Option A: Separate Domains

```
┌─────────────────────────────────────────────────────────────┐
│                    Cloud Provider                            │
│                                                               │
│  ┌──────────────────────┐    ┌──────────────────────┐      │
│  │  Admin Dashboard     │    │    IAM API            │      │
│  │  (CloudFront/S3)     │    │  (ECS/EKS/App Engine) │      │
│  │  admin.iam.com       │    │  api.iam.com          │      │
│  └──────────┬───────────┘    └──────────┬───────────┘      │
│             │                            │                   │
│             │  API Calls                 │                   │
│             │  https://api.iam.com       │                   │
│             └──────────┬─────────────────┘                   │
│                        │                                     │
│              ┌─────────▼──────────┐                          │
│              │  RDS/Cloud SQL    │                          │
│              │  ElastiCache      │                          │
│              └───────────────────┘                          │
└─────────────────────────────────────────────────────────────┘
```

#### Configuration

**Admin Dashboard** (`.env.production`):
```bash
VITE_API_BASE_URL=https://api.iam.com
```

**IAM API CORS** (must allow dashboard domain):
```go
// Allow specific origin
allowedOrigins := []string{
    "https://admin.iam.com",
    "https://iam.example.com",
}
origin := c.Request.Header.Get("Origin")
if contains(allowedOrigins, origin) {
    c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
}
```

#### Architecture Option B: Single Domain with Reverse Proxy

```
┌─────────────────────────────────────────────────────────────┐
│                    Cloud Provider                            │
│                                                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │         Application Load Balancer / API Gateway      │    │
│  │         iam.example.com                              │    │
│  └──────────────┬──────────────────┬───────────────────┘    │
│                 │                  │                         │
│    ┌────────────▼──────┐  ┌───────▼──────────┐            │
│    │  Admin Dashboard  │  │    IAM API        │            │
│    │  (S3/CloudFront)  │  │  (ECS/EKS)        │            │
│    │  /admin/*         │  │  /api/*           │            │
│    └───────────────────┘  └───────────────────┘            │
└─────────────────────────────────────────────────────────────┘
```

**Configuration**: Similar to Kubernetes, using ALB/API Gateway routing

---

## 🔐 Security Considerations

### CORS Configuration

**Development** (Current - allows all):
```go
c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
```

**Production** (Should be restricted):
```go
// Option 1: Single origin
c.Writer.Header().Set("Access-Control-Allow-Origin", "https://admin.iam.com")

// Option 2: Multiple allowed origins
allowedOrigins := []string{
    "https://admin.iam.com",
    "https://iam.example.com",
}
origin := c.Request.Header.Get("Origin")
if contains(allowedOrigins, origin) {
    c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
}
```

### Authentication Token Storage

**Option 1: localStorage** (Current approach)
- ✅ Simple to implement
- ⚠️ Vulnerable to XSS attacks
- ✅ Works across tabs

**Option 2: httpOnly Cookies** (Recommended for production)
- ✅ More secure (not accessible via JavaScript)
- ✅ Automatically sent with requests
- ⚠️ Requires CSRF protection

**Option 3: Session Storage**
- ✅ Cleared when tab closes
- ⚠️ Still vulnerable to XSS

### HTTPS in Production

- ✅ Always use HTTPS in production
- ✅ Use TLS certificates (Let's Encrypt, AWS ACM, etc.)
- ✅ Redirect HTTP to HTTPS
- ✅ Use secure cookies (Secure, HttpOnly, SameSite)

## 📊 API Communication Patterns

### 1. RESTful API Calls

```typescript
// Admin Dashboard API Client
const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add auth token to requests
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  const tenantId = localStorage.getItem('tenant_id');
  if (tenantId) {
    config.headers['X-Tenant-ID'] = tenantId;
  }
  return config;
});

// Handle token refresh
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      // Refresh token or redirect to login
      await refreshToken();
    }
    return Promise.reject(error);
  }
);
```

### 2. Error Handling

```typescript
try {
  const response = await apiClient.get('/api/v1/users');
  setUsers(response.data);
} catch (error) {
  if (error.response?.status === 401) {
    // Unauthorized - redirect to login
    router.push('/login');
  } else if (error.response?.status === 403) {
    // Forbidden - show error message
    showError('You do not have permission to perform this action');
  } else if (error.response?.status === 429) {
    // Rate limited
    showError('Too many requests. Please try again later.');
  } else {
    // Other errors
    showError('An error occurred. Please try again.');
  }
}
```

## 🚀 Deployment Checklist

### Local Development
- [x] Backend API running on port 8080
- [x] Frontend dev server on port 3000
- [x] CORS configured (allows all origins)
- [x] Environment variables set
- [x] Database connection configured

### Kubernetes
- [ ] Ingress configured with routing rules
- [ ] Services created for both apps
- [ ] ConfigMaps for configuration
- [ ] Secrets for sensitive data
- [ ] CORS updated for production domains
- [ ] Health checks configured
- [ ] Resource limits set

### Cloud Deployment
- [ ] Domain names configured
- [ ] SSL/TLS certificates
- [ ] Load balancer/API Gateway configured
- [ ] CORS updated for production domains
- [ ] Environment variables set
- [ ] Database connection configured
- [ ] Monitoring and logging set up

## 📝 Configuration Examples

### Environment Variables

**Local Development**:
```bash
# Backend
export DATABASE_HOST="localhost"
export DATABASE_PORT="5433"
export DATABASE_NAME="iam"
export DATABASE_USER="your_user"
export DATABASE_PASSWORD="your_password"

# Frontend
VITE_API_BASE_URL=http://localhost:8080
```

**Kubernetes** (ConfigMap):
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: iam-api-config
data:
  config.yaml: |
    server:
      port: 8080
    database:
      host: postgres-iam
      port: 5432
      name: iam_db
```

**Cloud** (Environment Variables):
```bash
# ECS Task Definition / App Engine app.yaml
DATABASE_HOST=postgres-instance.region.rds.amazonaws.com
DATABASE_PORT=5432
DATABASE_NAME=iam_db
DATABASE_USER=iam_user
DATABASE_PASSWORD=${DATABASE_PASSWORD}  # From Secrets Manager
```

## 🔍 Troubleshooting

### CORS Issues
- **Problem**: Browser blocks API requests
- **Solution**: Check CORS headers, verify allowed origins

### Connection Refused
- **Problem**: Frontend can't connect to backend
- **Solution**: Verify API URL, check service is running, check network/firewall

### Authentication Issues
- **Problem**: Tokens not working
- **Solution**: Check token storage, verify token format, check expiration

### 404 Errors
- **Problem**: API endpoints not found
- **Solution**: Verify routing configuration, check ingress rules, verify service endpoints

## 📚 Related Documentation

- [Frontend Implementation Plan](../planning/frontend-implementation-plan.md)
- [Deployment Guide](../deployment/kubernetes.md)
- [API Documentation](../api/README.md)
- [Configuration Guide](../technical/configuration.md)

---

**Last Updated**: 2024  
**Status**: Ready for Implementation

