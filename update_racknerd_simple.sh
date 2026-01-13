#!/bin/bash
# Simple update script for racknerd server
# Usage: curl -fsSL https://raw.githubusercontent.com/friddle/claude-web-remote/main/update_racknerd_simple.sh | bash

set -e

VERSION="v0.4"
REPO="friddle/claude-web-remote"
DEPLOY_DIR="$HOME/racknerd"
BINARY_NAME="clauded-linux-amd64"

echo "🚀 Updating clauded to $VERSION..."

# Create deploy directory
mkdir -p "$DEPLOY_DIR"
cd "$DEPLOY_DIR"

# Detect architecture
ARCH=$(uname -m)
if [ "$ARCH" = "aarch64" ]; then
    BINARY_NAME="clauded-linux-arm64"
elif [ "$ARCH" != "x86_64" ]; then
    echo "❌ Unsupported architecture: $ARCH"
    exit 1
fi

echo "📥 Downloading $BINARY_NAME..."
curl -fsSL "https://github.com/$REPO/releases/download/$VERSION/$BINARY_NAME" -o clauded
chmod +x clauded

echo "✅ Download complete: $DEPLOY_DIR/clauded"
echo "🔧 To update Docker container, run:"
echo "   cd $DEPLOY_DIR && docker-compose pull && docker-compose up -d"
