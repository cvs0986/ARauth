#!/bin/bash

# Frontend Setup Script for Nuage Identity
# This script initializes the frontend projects (Admin Dashboard and E2E Testing App)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
FRONTEND_DIR="$PROJECT_ROOT/frontend"

echo "🚀 Setting up Nuage Identity Frontend Applications"
echo "=================================================="

# Check Node.js
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js 18+ first."
    echo "   Visit: https://nodejs.org/"
    exit 1
fi

NODE_VERSION=$(node -v | cut -d'v' -f2 | cut -d'.' -f1)
if [ "$NODE_VERSION" -lt 18 ]; then
    echo "❌ Node.js version 18+ is required. Current version: $(node -v)"
    exit 1
fi

echo "✅ Node.js version: $(node -v)"

# Check npm
if ! command -v npm &> /dev/null; then
    echo "❌ npm is not installed."
    exit 1
fi

echo "✅ npm version: $(npm -v)"

# Create frontend directory
mkdir -p "$FRONTEND_DIR"
cd "$FRONTEND_DIR"

echo ""
echo "📦 Creating frontend projects..."
echo ""

# Function to create a React + TypeScript + Vite project
create_project() {
    local project_name=$1
    local project_dir="$FRONTEND_DIR/$project_name"
    
    echo "Creating $project_name..."
    
    if [ -d "$project_dir" ]; then
        echo "⚠️  Directory $project_dir already exists. Skipping..."
        return
    fi
    
    # Create Vite + React + TypeScript project
    npm create vite@latest "$project_name" -- --template react-ts
    
    cd "$project_dir"
    
    # Install dependencies
    echo "Installing dependencies for $project_name..."
    npm install
    
    # Install additional dependencies
    echo "Installing additional packages..."
    npm install axios react-router-dom @tanstack/react-query zustand react-hook-form zod
    npm install -D @types/node
    
    # Install UI library (using shadcn/ui setup - will need manual configuration)
    echo "📝 Note: UI library setup requires manual configuration"
    echo "   Consider using: shadcn-ui, Ant Design, or Material-UI"
    
    cd "$FRONTEND_DIR"
}

# Create projects
create_project "admin-dashboard"
create_project "e2e-test-app"

# Create shared directory structure
echo ""
echo "📁 Creating shared directory structure..."
mkdir -p "$FRONTEND_DIR/shared/api-client"
mkdir -p "$FRONTEND_DIR/shared/types"
mkdir -p "$FRONTEND_DIR/shared/utils"
mkdir -p "$FRONTEND_DIR/shared/constants"

# Create .env.example files
echo ""
echo "📝 Creating environment files..."

cat > "$FRONTEND_DIR/admin-dashboard/.env.example" << EOF
VITE_API_BASE_URL=http://localhost:8080
VITE_APP_NAME=Nuage Identity Admin
VITE_APP_VERSION=1.0.0
EOF

cat > "$FRONTEND_DIR/e2e-test-app/.env.example" << EOF
VITE_API_BASE_URL=http://localhost:8080
VITE_APP_NAME=Nuage Identity Test App
VITE_APP_VERSION=1.0.0
EOF

# Create README for frontend
cat > "$FRONTEND_DIR/README.md" << 'EOF'
# Nuage Identity Frontend Applications

This directory contains the frontend applications for Nuage Identity IAM.

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

1. Install dependencies for each project:
```bash
cd admin-dashboard && npm install
cd ../e2e-test-app && npm install
```

2. Copy environment files:
```bash
cp admin-dashboard/.env.example admin-dashboard/.env
cp e2e-test-app/.env.example e2e-test-app/.env
```

3. Start development servers:
```bash
# Terminal 1: Admin Dashboard
cd admin-dashboard && npm run dev

# Terminal 2: E2E Testing App
cd e2e-test-app && npm run dev
```

## Documentation

- [Implementation Plan](../../docs/planning/frontend-implementation-plan.md)
- [Quick Start Guide](../../docs/guides/frontend-quick-start.md)
- [API Documentation](../../docs/api/README.md)

## Development

### Project Structure

Each project follows a similar structure:
```
src/
├── components/     # Reusable UI components
├── pages/          # Page components
├── hooks/          # Custom React hooks
├── services/       # API service layer
├── store/          # State management
├── types/          # TypeScript types
├── utils/          # Utility functions
└── App.tsx         # Root component
```

### API Client

The API client is generated from the OpenAPI specification:
```bash
# Generate TypeScript client from OpenAPI spec
npx @openapitools/openapi-generator-cli generate \
  -i ../../docs/api/openapi.yaml \
  -g typescript-axios \
  -o shared/api-client
```

### Testing

```bash
# Unit tests
npm test

# E2E tests
npm run test:e2e

# Coverage
npm run test:coverage
```

## Build

```bash
# Build for production
npm run build

# Preview production build
npm run preview
```

## Deployment

See [Deployment Guide](../../docs/deployment/frontend.md) for production deployment instructions.
EOF

echo ""
echo "✅ Frontend setup complete!"
echo ""
echo "Next steps:"
echo "1. Copy .env.example to .env in each project"
echo "2. Install UI component library (e.g., shadcn-ui, Ant Design)"
echo "3. Generate API client from OpenAPI spec"
echo "4. Start development:"
echo "   cd frontend/admin-dashboard && npm run dev"
echo "   cd frontend/e2e-test-app && npm run dev"
echo ""
echo "📚 Documentation:"
echo "   - Implementation Plan: docs/planning/frontend-implementation-plan.md"
echo "   - Quick Start: docs/guides/frontend-quick-start.md"
echo ""

