#!/bin/bash
# Debug script for running opencode

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Find opencode binary - try multiple locations
OPENCODE=""
if [ -f "$SCRIPT_DIR/opencode" ]; then
    OPENCODE="$SCRIPT_DIR/opencode"
elif [ -f "/usr/local/bin/opencode" ]; then
    OPENCODE="/usr/local/bin/opencode"
elif [ -f "$HOME/.local/bin/opencode" ]; then
    OPENCODE="$HOME/.local/bin/opencode"
else
    echo "Error: opencode binary not found"
    echo "Searched in:"
    echo "  - $SCRIPT_DIR/opencode"
    echo "  - /usr/local/bin/opencode"
    echo "  - $HOME/.local/bin/opencode"
    exit 1
fi

echo "🚀 Starting opencode for debugging..."
echo "Binary: $OPENCODE"
echo "Arguments: $@"
echo ""

# Run opencode with all provided arguments
exec "$OPENCODE" "$@"
