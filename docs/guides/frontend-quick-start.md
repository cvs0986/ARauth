# Frontend Quick Start Guide

This guide will help you set up and start developing the frontend applications for Nuage Identity.

## Prerequisites

- Node.js 18+ and npm/yarn/pnpm
- Backend API running on `http://localhost:8080`
- PostgreSQL running on port `5433`
- Redis running (optional but recommended)

## Quick Setup

### 1. Install Node.js (if not installed)

```bash
# Using nvm (recommended)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
nvm install 20
nvm use 20

# Or download from nodejs.org
```

### 2. Initialize Frontend Projects

```bash
# Navigate to project root
cd /home/eshwar/Documents/Veer/nuage-indentity

# Run setup script (to be created)
bash scripts/setup-frontend.sh
```

### 3. Start Development Servers

```bash
# Terminal 1: Admin Dashboard
cd frontend/admin-dashboard
npm install
npm run dev
# Runs on http://localhost:3000

# Terminal 2: E2E Testing App
cd frontend/e2e-test-app
npm install
npm run dev
# Runs on http://localhost:3001
```

### 4. Verify Setup

1. **Check Admin Dashboard**: Open http://localhost:3000
2. **Check E2E App**: Open http://localhost:3001
3. **Check API**: `curl http://localhost:8080/health`

## Development Workflow

### Daily Development

1. **Start Backend** (if not running):
   ```bash
   go run cmd/server/main.go
   ```

2. **Start Frontend**:
   ```bash
   # Admin Dashboard
   cd frontend/admin-dashboard && npm run dev
   
   # E2E Testing App
   cd frontend/e2e-test-app && npm run dev
   ```

3. **Make Changes**: Files auto-reload with hot module replacement

4. **Run Tests**:
   ```bash
   npm test              # Unit tests
   npm run test:e2e      # E2E tests
   ```

### API Integration

The frontend apps connect to the backend API at:
- **Development**: `http://localhost:8080`
- **Production**: Set via `VITE_API_BASE_URL` environment variable

### Environment Variables

Create `.env` files in each frontend project:

```bash
# frontend/admin-dashboard/.env
VITE_API_BASE_URL=http://localhost:8080
VITE_APP_NAME=Nuage Identity Admin

# frontend/e2e-test-app/.env
VITE_API_BASE_URL=http://localhost:8080
VITE_APP_NAME=Nuage Identity Test App
```

## Testing Scenarios

### Scenario 1: Complete User Journey

1. **Create Tenant** (Admin Dashboard)
   - Navigate to Tenants
   - Click "Create Tenant"
   - Fill form: Name, Domain
   - Submit

2. **Create User** (Admin Dashboard)
   - Navigate to Users
   - Click "Create User"
   - Fill form: Username, Email, Password, Tenant
   - Assign roles
   - Submit

3. **Login** (E2E Testing App)
   - Navigate to Login page
   - Enter credentials
   - Select tenant
   - Submit

4. **MFA Setup** (E2E Testing App)
   - Navigate to Profile → Security
   - Click "Enable MFA"
   - Scan QR code or enter secret
   - Verify with TOTP code
   - Save recovery codes

5. **Test Permissions** (E2E Testing App)
   - Navigate to Roles & Permissions
   - View assigned roles
   - Test permission-based features

### Scenario 2: Admin Management

1. **Manage Roles** (Admin Dashboard)
   - Create role
   - Assign permissions
   - View role details

2. **Manage Permissions** (Admin Dashboard)
   - Create permission
   - View permission details
   - Check role assignments

3. **User Management** (Admin Dashboard)
   - List users
   - Filter by tenant
   - Edit user
   - Assign roles
   - View user permissions

### Scenario 3: Security Testing

1. **Rate Limiting**
   - Attempt multiple failed logins
   - Verify rate limit response

2. **MFA Challenge**
   - Login with MFA-enabled account
   - Complete MFA challenge
   - Verify access

3. **Permission Testing**
   - Login as user with limited permissions
   - Attempt unauthorized actions
   - Verify 403 responses

## Troubleshooting

### API Connection Issues

**Problem**: Frontend can't connect to backend

**Solutions**:
- Verify backend is running: `curl http://localhost:8080/health`
- Check CORS settings in backend
- Verify `VITE_API_BASE_URL` in `.env`
- Check browser console for errors

### Authentication Issues

**Problem**: Login fails or tokens not working

**Solutions**:
- Check tenant ID is set in headers
- Verify token storage (localStorage/cookies)
- Check token expiration
- Verify Hydra is running (if using OAuth2)

### Build Issues

**Problem**: Build fails or errors

**Solutions**:
- Clear node_modules: `rm -rf node_modules && npm install`
- Clear build cache: `rm -rf dist .vite`
- Check Node.js version: `node --version` (should be 18+)
- Check for TypeScript errors: `npm run type-check`

## Next Steps

1. Read the [Full Implementation Plan](../planning/frontend-implementation-plan.md)
2. Start with Phase 1: Foundation & Setup
3. Follow the phase-by-phase implementation guide
4. Refer to API documentation for endpoint details

## Resources

- [React Documentation](https://react.dev/)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)
- [Vite Guide](https://vitejs.dev/guide/)
- [API Documentation](../api/README.md)

