# Quick Start Testing - TL;DR Version

## 🚀 Start Everything

### Terminal 1: Backend
```bash
cd /home/eshwar/Documents/Veer/nuage-indentity
export DATABASE_HOST=127.0.0.1 DATABASE_PORT=5433 DATABASE_USER=dcim_user DATABASE_PASSWORD=dcim_password DATABASE_NAME=iam DATABASE_SSL_MODE=disable JWT_SECRET=test-jwt-secret-key-min-32-characters-long ENCRYPTION_KEY=01234567890123456789012345678901
go run cmd/server/main.go
```

### Terminal 2: Admin Dashboard
```bash
cd frontend/admin-dashboard
npm run dev
```

### Terminal 3: E2E Test App
```bash
cd frontend/e2e-test-app
npm run dev
```

## 📝 Initial Setup (One-Time)

### 1. Create Tenant
```bash
curl -X POST http://localhost:8080/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Company", "domain": "test.local", "status": "active"}'
```
**Save the `id` from response!**

### 2. Create Admin User
```bash
# Replace TENANT_ID with id from step 1
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: TENANT_ID" \
  -d '{
    "username": "admin",
    "email": "admin@test.local",
    "password": "Admin@123456",
    "first_name": "Admin",
    "last_name": "User"
  }'
```

## 🧪 Test

1. **Admin Dashboard**: http://localhost:5173
   - Login: `admin` / `Admin@123456` / `TENANT_ID`
   - Test all CRUD operations

2. **E2E Test App**: http://localhost:5174
   - Register new user
   - Login
   - Test MFA
   - View profile
   - View roles & permissions

## ✅ Verify Backend
```bash
curl http://localhost:8080/health
```

---

**For detailed steps, see**: `TESTING_GUIDE.md`

