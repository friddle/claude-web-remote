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
OUTPUT_FILE="output/$BINARY_NAME"
INSTALL_DIR="$HOME/bin"
INSTALL_PATH="$INSTALL_DIR/$BINARY_NAME"

if [ ! -f "$OUTPUT_FILE" ]; then
    echo "❌ Build failed: $OUTPUT_FILE not found."
    exit 1
fi

echo "📂 Ensuring $INSTALL_DIR exists..."
mkdir -p "$INSTALL_DIR"

echo "🚀 Installing $BINARY_NAME to $INSTALL_PATH..."
cp "$OUTPUT_FILE" "$INSTALL_PATH"

echo "✅ Build and install successful!"
echo "🏃 Running $INSTALL_PATH ..."
echo "----------------------------------------"
