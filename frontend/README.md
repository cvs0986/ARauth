# ARauth Identity Frontend Applications

This directory contains the frontend applications for ARauth Identity IAM.

## Projects

- **admin-dashboard**: Management UI for system administrators
- **e2e-test-app**: Complete frontend application for end-to-end testing
- **shared**: Shared code and utilities between applications

## Quick Start

### Prerequisites
- Node.js 18+
- npm or yarn
- Backend API running on http://localhost:8080

### Setup

1. **Install dependencies for each project**:
```bash
cd admin-dashboard && npm install
cd ../e2e-test-app && npm install
```

2. **Copy environment files**:
```bash
cp admin-dashboard/.env.example admin-dashboard/.env
cp e2e-test-app/.env.example e2e-test-app/.env
```

3. **Start development servers**:
```bash
# Terminal 1: Admin Dashboard
cd admin-dashboard && npm run dev
# Runs on http://localhost:5173

# Terminal 2: E2E Testing App
cd e2e-test-app && npm run dev
# Runs on http://localhost:5174
```

## Project Structure

Each project follows a similar structure:
```
src/
├── components/     # Reusable UI components
├── pages/          # Page components
├── hooks/          # Custom React hooks
├── services/       # API service layer
├── store/          # State management (Zustand)
├── types/          # TypeScript types
├── utils/          # Utility functions
└── App.tsx         # Root component
```

## Shared Code

The `shared/` directory contains code shared between applications:
- `constants/` - API endpoints and constants
- `types/` - Shared TypeScript types
- `utils/` - Shared utilities (API client, etc.)

## Development

### Running Tests
```bash
# Unit tests (when implemented)
npm test

# E2E tests (when implemented)
npm run test:e2e
```

### Building
```bash
# Build for production
npm run build

# Preview production build
npm run preview
```

## Documentation

- [Frontend Implementation Plan](../docs/planning/frontend-implementation-plan.md)
- [Frontend Quick Start](../docs/guides/frontend-quick-start.md)
- [API Documentation](../docs/api/README.md)

## Status

✅ Phase 1: Foundation & Setup - In Progress
- [x] React projects initialized
- [x] Dependencies installed
- [x] Tailwind CSS configured
- [x] Base structure created
- [x] Authentication store setup
- [x] API client configured
- [ ] UI components library
- [ ] Base layout components

---

**Last Updated**: 2024

