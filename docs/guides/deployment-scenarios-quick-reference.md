# Deployment Scenarios - Quick Reference

Quick reference for how Admin Dashboard and IAM API work together in different environments.

## 🏠 Local Development

### Setup
```bash
# Terminal 1: Backend API
export DATABASE_PORT=5433
go run cmd/server/main.go
# → http://localhost:8080

# Terminal 2: Admin Dashboard
cd frontend/admin-dashboard
npm run dev
# → http://localhost:3000
```

### Configuration
- **Frontend API URL**: `http://localhost:8080`
- **CORS**: Allows all origins (`*`)
- **Connection**: Direct HTTP calls from browser

### How It Works
```
Browser (localhost:3000) → API (localhost:8080)
```

---

## ☸️ Kubernetes

### Architecture
```
Ingress (iam.example.com)
  ├── /admin → Admin Dashboard Service
  └── /api → IAM API Service
```

### Configuration
- **Frontend API URL**: `https://iam.example.com/api`
- **CORS**: Same domain (no CORS needed) OR configured origins
- **Connection**: Via Ingress routing

### How It Works
```
User → Ingress → Admin Dashboard (serves React app)
                ↓
User's Browser → Ingress → IAM API (processes requests)
```

---

## ☁️ Cloud (AWS/GCP/Azure)

### Option A: Separate Domains
```
admin.iam.com → Admin Dashboard (S3/CloudFront)
api.iam.com → IAM API (ECS/EKS/App Engine)
```

### Option B: Single Domain
```
iam.example.com/admin → Admin Dashboard
iam.example.com/api → IAM API
```

### Configuration
- **Frontend API URL**: `https://api.iam.com` or `https://iam.example.com/api`
- **CORS**: Must allow dashboard domain(s)
- **Connection**: Via Load Balancer/API Gateway

---

## 🔑 Key Points

1. **Frontend is separate** - React SPA, independent deployment
2. **Backend is API** - REST API, handles all business logic
3. **Communication** - HTTP/HTTPS API calls
4. **Authentication** - Tokens (JWT) in headers
5. **CORS** - Configured based on deployment pattern

---

## 📋 Configuration Matrix

| Environment | Frontend URL | API URL | CORS Setting |
|------------|--------------|---------|--------------|
| Local Dev | localhost:3000 | localhost:8080 | `*` (all) |
| K8s Same Domain | iam.com/admin | iam.com/api | Same domain (no CORS) |
| K8s Separate | admin.iam.com | api.iam.com | Specific origins |
| Cloud Separate | admin.iam.com | api.iam.com | Specific origins |
| Cloud Single | iam.com/admin | iam.com/api | Same domain (no CORS) |

---

## 🚀 Quick Start Commands

### Local
```bash
# Backend
go run cmd/server/main.go

# Frontend
cd frontend/admin-dashboard && npm run dev
```

### Kubernetes
```bash
# Deploy
kubectl apply -f k8s/

# Or with Helm
helm install nuage-identity ./helm/nuage-identity
```

### Cloud
```bash
# Deploy backend (example: AWS ECS)
aws ecs update-service --cluster iam-cluster --service iam-api

# Deploy frontend (example: S3 + CloudFront)
aws s3 sync frontend/admin-dashboard/dist s3://iam-admin-bucket
```

---

**See full documentation**: [Frontend-Backend Integration](../architecture/frontend-backend-integration.md)

