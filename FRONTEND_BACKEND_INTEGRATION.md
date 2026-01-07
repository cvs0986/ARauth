# Frontend-Backend Integration Summary

## ✅ Confirmation: How Admin Dashboard & IAM API Work Together

This document confirms how the Admin Dashboard (frontend) and IAM API (backend) work together in different deployment scenarios.

## 🏗️ Architecture Overview

The Admin Dashboard and IAM API are **separate applications** that communicate via HTTP/HTTPS:

- **Admin Dashboard**: React SPA (Single Page Application)
- **IAM API**: Go REST API server
- **Communication**: HTTP API calls (REST)
- **Authentication**: JWT tokens in Authorization header
- **CORS**: Configured to allow cross-origin requests

## 📍 Deployment Scenarios

### 1. Local Development

**Setup**:
- Admin Dashboard: `http://localhost:3000` (Vite dev server)
- IAM API: `http://localhost:8080` (Go server)

**How They Connect**:
```
Browser → Admin Dashboard (localhost:3000)
         ↓ (API calls)
         IAM API (localhost:8080)
         ↓
         PostgreSQL (localhost:5433)
```

**Configuration**:
- Frontend `.env`: `VITE_API_BASE_URL=http://localhost:8080`
- CORS allows all origins (`*`) - suitable for development

**Running**:
```bash
# Terminal 1: Backend
go run cmd/server/main.go

# Terminal 2: Frontend
cd frontend/admin-dashboard && npm run dev
```

---

### 2. Kubernetes Deployment

**Architecture**:
```
User → Ingress (iam.example.com)
       ├── /admin → Admin Dashboard Service → Pods
       └── /api → IAM API Service → Pods
```

**How They Connect**:
- Both apps deployed as separate Kubernetes services
- Ingress routes traffic based on path
- Same domain = No CORS issues (or CORS configured)
- Internal communication via Kubernetes service names

**Configuration**:
- Frontend API URL: `https://iam.example.com/api` (relative or full URL)
- Ingress handles routing
- Services use ClusterIP for internal communication

---

### 3. Cloud Deployment (AWS/GCP/Azure)

**Option A: Separate Domains**
```
admin.iam.com → Admin Dashboard (S3/CloudFront)
api.iam.com → IAM API (ECS/EKS/App Engine)
```

**Option B: Single Domain**
```
iam.example.com/admin → Admin Dashboard
iam.example.com/api → IAM API
```

**How They Connect**:
- Frontend makes API calls to backend domain
- CORS must be configured to allow frontend domain
- Load balancer/API Gateway routes traffic

---

## 🔄 Communication Flow

### 1. User Login
```
Admin Dashboard → POST /api/v1/auth/login
                → IAM API validates credentials
                → Returns JWT token
                → Dashboard stores token
```

### 2. API Request
```
Admin Dashboard → GET /api/v1/users
                → Headers: Authorization: Bearer <token>
                → IAM API validates token
                → Returns user data
                → Dashboard displays data
```

## 🔐 Security & CORS

### Current CORS Configuration
- **Development**: Allows all origins (`*`) ✅
- **Production**: Should be restricted to specific domains ⚠️

### Token Storage
- **Development**: localStorage (simple)
- **Production**: Consider httpOnly cookies (more secure)

### HTTPS
- **Development**: HTTP is fine
- **Production**: Always use HTTPS

## 📋 Key Points

1. ✅ **Separate Applications**: Frontend and backend are independent
2. ✅ **API Communication**: Frontend makes HTTP requests to backend
3. ✅ **Configurable**: API URL configured via environment variables
4. ✅ **CORS Enabled**: Cross-origin requests are supported
5. ✅ **Flexible Deployment**: Can deploy separately or together

## 🚀 Quick Start

### Local Development
```bash
# 1. Start Backend
export DATABASE_PORT=5433
go run cmd/server/main.go

# 2. Start Frontend
cd frontend/admin-dashboard
npm run dev

# 3. Access
# Dashboard: http://localhost:3000
# API: http://localhost:8080
```

### Configuration Files
- **Backend**: `config/config.yaml` or environment variables
- **Frontend**: `.env` file with `VITE_API_BASE_URL`

## 📚 Full Documentation

For detailed information, see:
- **[Frontend-Backend Integration Guide](docs/architecture/frontend-backend-integration.md)** - Complete guide
- **[Deployment Scenarios Quick Reference](docs/guides/deployment-scenarios-quick-reference.md)** - Quick reference
- **[Frontend Implementation Plan](docs/planning/frontend-implementation-plan.md)** - Frontend development plan

## ✅ Summary

**Yes, they work together seamlessly!**

- ✅ Admin Dashboard is a separate React app
- ✅ IAM API is a separate Go service
- ✅ They communicate via HTTP API calls
- ✅ CORS is configured for cross-origin requests
- ✅ Works in local, Kubernetes, and cloud deployments
- ✅ Configuration is flexible and environment-based

**You can deploy them:**
- Together (same domain, different paths)
- Separately (different domains/subdomains)
- Any combination that fits your infrastructure

---

**Last Updated**: 2024  
**Status**: Confirmed and Documented

