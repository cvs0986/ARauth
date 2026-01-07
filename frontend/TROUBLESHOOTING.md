# Frontend Troubleshooting Guide

## Common Issues and Solutions

### Issue 1: Tailwind CSS v4 Error

**Error**: `It looks like you're trying to use tailwindcss directly as a PostCSS plugin`

**Solution**: 
- Use Tailwind CSS v3.4 instead of v4
- PostCSS config works correctly with v3.4

```bash
npm uninstall tailwindcss
npm install -D tailwindcss@^3.4.0
```

### Issue 2: Missing Dependencies

**Error**: `Failed to resolve import "class-variance-authority"` or similar

**Solution**: Install missing dependencies

```bash
# Admin Dashboard
cd frontend/admin-dashboard
npm install class-variance-authority lucide-react @radix-ui/react-slot @radix-ui/react-label @radix-ui/react-dialog

# E2E Test App
cd frontend/e2e-test-app
npm install axios class-variance-authority lucide-react @radix-ui/react-slot @radix-ui/react-label @radix-ui/react-dialog
```

### Issue 3: Axios Not Found in E2E App

**Error**: `Failed to resolve import "axios" from shared/utils/api-client.ts`

**Solution**: Install axios in e2e-test-app

```bash
cd frontend/e2e-test-app
npm install axios
```

### Issue 4: Port Already in Use

**Error**: `Port 5173 is in use`

**Solution**: 
- Stop the other process using the port
- Or use a different port: `npm run dev -- --port 5174`

## Quick Fix Commands

### For Admin Dashboard
```bash
cd frontend/admin-dashboard
npm install
npm run dev
```

### For E2E Test App
```bash
cd frontend/e2e-test-app
npm install
npm run dev
```

## Required Dependencies

### Both Apps Need
- `axios` - HTTP client
- `react-router-dom` - Routing
- `@tanstack/react-query` - Data fetching
- `zustand` - State management
- `react-hook-form` - Form handling
- `zod` - Validation
- `tailwindcss@^3.4.0` - Styling
- `class-variance-authority` - Component variants
- `lucide-react` - Icons
- `@radix-ui/*` - UI primitives

### shadcn/ui Dependencies
- `clsx` - Class name utility
- `tailwind-merge` - Tailwind class merging

---

**Last Updated**: 2024-01-07

