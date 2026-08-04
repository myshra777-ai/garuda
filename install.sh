#!/bin/bash
set -e

echo "🛡️ Installing Garuda Runtime Engine..."

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
else
    echo "⚠️ Unknown architecture: $ARCH. Falling back to building from source."
    ARCH="source"
fi

# Check for Docker
if ! command -v docker &> /dev/null; then
    echo "⚠️ Docker not found. Garuda requires Docker to run PostgreSQL and Redis."
    echo "   Install Docker from: https://docs.docker.com/get-docker/"
fi

# Check for Go (optional, only needed for building from source)
if ! command -v go &> /dev/null; then
    echo "⚠️ Go not found. Will use pre-built binary if available."
fi

# Download or build
if [ "$ARCH" != "source" ] && command -v curl &> /dev/null; then
    BINARY_URL="https://github.com/myshra777-ai/garuda/releases/latest/download/garuda-${OS}-${ARCH}"
    echo "📥 Downloading binary for ${OS}/${ARCH}..."
    if curl -fsSL -o /usr/local/bin/garuda "$BINARY_URL"; then
        chmod +x /usr/local/bin/garuda
        echo "✅ Binary installed successfully."
    else
        echo "⚠️ Binary download failed. Building from source..."
        go build -o /usr/local/bin/garuda ./cmd/garuda
    fi
else
    echo "🔨 Building Garuda from source..."
    go build -o /usr/local/bin/garuda ./cmd/garuda
fi

echo ""
echo "✅ Garuda successfully installed!"
echo "👉 Run 'garuda up' to start your local runtime and Mission Control."
echo ""
echo "📖 Documentation: https://github.com/myshra777-ai/garuda"