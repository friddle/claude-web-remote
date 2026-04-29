#!/bin/bash
# Debug script for running clauded with clauded.tools.yicoson.cn

# Default configuration
HOST="clauded.tools.yicoson.cn"
ANTHROPIC_BASE_URL="https://open.bigmodel.cn/api/anthropic"
ANTHROPIC_AUTH_TOKEN="34d9f2b58ab545448bf7e29948b66d80.HIRfPtswrGiCndmm"
FLAGS="--allow-dangerously-skip-permissions --dangerously-skip-permissions"
AUTO_EXIT=false

# Fixed password
PASSWORD="sybran_20250708"

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "🔧 Debug mode: Cleaning up..."

# Kill existing clauded processes
PIDS=$(pgrep -f "[c]lauded" 2>/dev/null || true)
if [ -n "$PIDS" ]; then
    echo "🛑 Killing existing clauded processes: $PIDS"
    pkill -9 -f "[c]lauded" 2>/dev/null || true
    sleep 1
fi

# Kill existing tmux sessions created by clauded
SESSIONS=$(tmux list-sessions -F "#{session_name}" 2>/dev/null | grep -E "^[a-z]{5}$" || true)
if [ -n "$SESSIONS" ]; then
    echo "🛑 Killing tmux sessions: $SESSIONS"
    echo "$SESSIONS" | xargs -I {} tmux kill-session -t {} 2>/dev/null || true
fi

# Find clauded binary
if [ -f "$SCRIPT_DIR/cmd/client/output/clauded-linux-amd64" ]; then
    CLAUDED="$SCRIPT_DIR/cmd/client/output/clauded-linux-amd64"
elif [ -f "/usr/local/bin/clauded" ]; then
    CLAUDED="/usr/local/bin/clauded"
else
    echo "Error: clauded binary not found"
    echo "Please build first: cd cmd/client && make build"
    exit 1
fi

echo "🚀 Starting clauded in debug mode..."
echo "Host: $HOST"
echo "Binary: $CLAUDED"
echo "Password: $PASSWORD"
echo ""

# Run clauded with debug configuration
exec "$CLAUDED" \
    --remote "https://$HOST" \
    --auth-name=friddle \
    --password "$PASSWORD" \
    --env "ANTHROPIC_BASE_URL=$ANTHROPIC_BASE_URL" \
    --env "ANTHROPIC_AUTH_TOKEN=$ANTHROPIC_AUTH_TOKEN" \
    --flags="$FLAGS" \
    --auto-exit=$AUTO_EXIT \
    --daemon=false


