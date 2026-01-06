#!/bin/bash

# Nuage Identity - Development Tools Installation Script
# Installs required development tools for Go development

set -e

echo "========================================="
echo "Nuage Identity - Development Tools Setup"
echo "========================================="
echo ""

# Ensure Go is in PATH
export PATH=$PATH:/home/eshwar/go-install/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

# Check if Go is available
if ! command -v go &> /dev/null; then
    echo "ERROR: Go is not installed or not in PATH"
    echo "Please run scripts/setup-go.sh first"
    exit 1
fi

echo "Go version: $(go version)"
echo ""

# Install Go tools
echo "Installing Go development tools..."

# golangci-lint
echo "Installing golangci-lint..."
if ! command -v golangci-lint &> /dev/null; then
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
    echo "✓ golangci-lint installed"
else
    echo "✓ golangci-lint already installed"
fi

# goimports
echo "Installing goimports..."
go install golang.org/x/tools/cmd/goimports@latest
echo "✓ goimports installed"

# golang-migrate
echo "Installing golang-migrate..."
if ! command -v migrate &> /dev/null; then
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
    echo "✓ golang-migrate installed"
else
    echo "✓ golang-migrate already installed"
fi

# mockgen
echo "Installing mockgen..."
go install github.com/golang/mock/mockgen@latest
echo "✓ mockgen installed"

# Check Docker
echo ""
echo "Checking Docker..."
if command -v docker &> /dev/null; then
    echo "✓ Docker is installed: $(docker --version)"
else
    echo "⚠ Docker is not installed. Please install Docker for local development."
    echo "  Fedora: sudo dnf install docker docker-compose"
fi

# Check Docker Compose
if command -v docker-compose &> /dev/null || docker compose version &> /dev/null; then
    echo "✓ Docker Compose is available"
else
    echo "⚠ Docker Compose is not installed."
fi

# Check Git
echo ""
echo "Checking Git..."
if command -v git &> /dev/null; then
    echo "✓ Git is installed: $(git --version)"
else
    echo "⚠ Git is not installed. Please install Git."
    echo "  Fedora: sudo dnf install git"
fi

echo ""
echo "========================================="
echo "Development tools setup complete!"
echo "========================================="
echo ""
echo "Installed tools:"
echo "  - golangci-lint: $(golangci-lint version 2>/dev/null || echo 'installed')"
echo "  - goimports: installed"
echo "  - golang-migrate: $(migrate -version 2>/dev/null | head -1 || echo 'installed')"
echo "  - mockgen: installed"
echo ""
echo "Note: If tools are not found, ensure GOPATH/bin is in your PATH:"
echo "  export PATH=\$PATH:\$GOPATH/bin"

