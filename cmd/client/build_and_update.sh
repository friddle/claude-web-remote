#!/bin/bash
set -e

# 获取脚本所在目录并进入
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

echo "🧹 Cleaning up old build artifacts..."
make clean

echo "🔨 Building project..."
make build

BINARY_NAME="clauded"
OUTPUT_DIR="output"
INSTALL_DIR="$HOME/bin"
INSTALL_PATH="$INSTALL_DIR/$BINARY_NAME"

# Detect platform binary
GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)
PLATFORM_BINARY="$OUTPUT_DIR/${BINARY_NAME}-${GOOS}-${GOARCH}"

if [ ! -f "$PLATFORM_BINARY" ]; then
    echo "❌ Build failed: $PLATFORM_BINARY not found."
    echo "📂 Available files in output directory:"
    ls -la "$OUTPUT_DIR" 2>/dev/null || echo "  (output directory not found)"
    exit 1
fi

echo "📦 Found platform binary: $PLATFORM_BINARY"

echo "📂 Ensuring $INSTALL_DIR exists..."
mkdir -p "$INSTALL_DIR"

echo "🚀 Installing $BINARY_NAME to $INSTALL_PATH..."
cp "$PLATFORM_BINARY" "$INSTALL_PATH"

echo "✅ Build and install successful!"
echo "🏃 Running $INSTALL_PATH ..."
echo "----------------------------------------"
