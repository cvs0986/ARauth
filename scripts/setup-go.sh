#!/bin/bash

# Nuage Identity - Go Installation Script for Fedora Linux
# This script installs Go 1.21+ on Fedora Linux

set -e

echo "========================================="
echo "Nuage Identity - Go Installation"
echo "========================================="
echo ""

# Check if Go is already installed
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    echo "Go is already installed: version $GO_VERSION"
    
    # Check if version is 1.21 or higher
    MAJOR=$(echo $GO_VERSION | cut -d. -f1)
    MINOR=$(echo $GO_VERSION | cut -d. -f2)
    
    if [ "$MAJOR" -gt 1 ] || ([ "$MAJOR" -eq 1 ] && [ "$MINOR" -ge 21 ]); then
        echo "Go version $GO_VERSION meets requirements (>= 1.21)"
        exit 0
    else
        echo "Go version $GO_VERSION is too old. Need >= 1.21"
    fi
fi

echo "Installing Go 1.21+..."

# Try using dnf first (may require sudo)
if command -v dnf &> /dev/null; then
    echo "Attempting to install Go using dnf (may require sudo password)..."
    if sudo dnf install -y golang 2>/dev/null; then
        if command -v go &> /dev/null; then
            GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
            echo "Go installed successfully via dnf: version $GO_VERSION"
            exit 0
        fi
    else
        echo "dnf installation requires sudo. Trying manual user-space installation..."
    fi
fi

# Manual installation to user directory (no sudo required)
echo "Installing Go to user directory (~/go-install)..."
GO_INSTALL_DIR="$HOME/go-install"
mkdir -p "$GO_INSTALL_DIR"

# Set Go version
GO_VERSION="1.21.5"
GO_ARCH="linux-amd64"
GO_TAR="go${GO_VERSION}.${GO_ARCH}.tar.gz"
GO_URL="https://go.dev/dl/${GO_TAR}"

# Create temporary directory
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

echo "Downloading Go ${GO_VERSION}..."
if command -v wget &> /dev/null; then
    wget -q "$GO_URL" || { echo "Failed to download Go"; exit 1; }
elif command -v curl &> /dev/null; then
    curl -L -o "$GO_TAR" "$GO_URL" || { echo "Failed to download Go"; exit 1; }
else
    echo "ERROR: Neither wget nor curl is available"
    exit 1
fi

echo "Extracting Go..."
rm -rf "$GO_INSTALL_DIR/go"
tar -C "$GO_INSTALL_DIR" -xzf "$GO_TAR"

# Cleanup
cd -
rm -rf "$TEMP_DIR"

# Add Go to PATH in bashrc
if ! grep -q "$GO_INSTALL_DIR/go/bin" ~/.bashrc; then
    echo "" >> ~/.bashrc
    echo "# Go installation" >> ~/.bashrc
    echo "export PATH=\$PATH:$GO_INSTALL_DIR/go/bin" >> ~/.bashrc
    echo "export GOPATH=\$HOME/go" >> ~/.bashrc
    echo "export PATH=\$PATH:\$GOPATH/bin" >> ~/.bashrc
fi

# Add to current session
export PATH=$PATH:$GO_INSTALL_DIR/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

# Verify installation
if command -v go &> /dev/null; then
    GO_VERSION=$(go version)
    echo ""
    echo "========================================="
    echo "Go installed successfully!"
    echo "Version: $GO_VERSION"
    echo "========================================="
    echo ""
    echo "Go binary location: $(which go)"
    echo "GOROOT: ${GOROOT:-$GO_INSTALL_DIR/go}"
    echo "GOPATH: ${GOPATH:-$HOME/go}"
    echo ""
    echo "Note: If you just installed Go, you may need to:"
    echo "  1. Restart your terminal, or"
    echo "  2. Run: source ~/.bashrc"
    echo "  3. Or run: export PATH=\$PATH:$GO_INSTALL_DIR/go/bin"
    echo ""
else
    echo "ERROR: Go installation failed!"
    exit 1
fi
