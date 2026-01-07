# Clear Vite Cache - Instructions

## Problem
Vite sometimes caches dependencies incorrectly, causing import resolution errors even after installing packages.

## Solution: Clear Vite Cache

### For Admin Dashboard
```bash
cd frontend/admin-dashboard
rm -rf node_modules/.vite
npm run dev
```

### For E2E Test App
```bash
cd frontend/e2e-test-app
rm -rf node_modules/.vite
npm run dev
```

## What This Does
- Removes Vite's pre-bundled dependency cache
- Forces Vite to re-scan and re-bundle dependencies
- Resolves import resolution issues

## When to Use
- After installing new dependencies
- When seeing "Failed to resolve import" errors
- After updating package.json
- When dependencies seem to not be found

## Both Apps Fixed
Both apps now have:
- ✅ axios in optimizeDeps
- ✅ Server port configured
- ✅ All dependencies installed

---

**Last Updated**: 2024-01-08

