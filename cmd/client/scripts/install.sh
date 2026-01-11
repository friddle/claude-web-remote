#!/bin/bash

# Claude Code Auto Installation Script
# Supported: macOS, Debian, Ubuntu, Alpine

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect operating system
detect_os() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    elif [[ -f /etc/os-release ]]; then
        . /etc/os-release
        case "$ID" in
            debian)
                echo "debian"
                ;;
            ubuntu)
                echo "ubuntu"
                ;;
            alpine)
                echo "alpine"
                ;;
            *)
                echo "unknown"
                ;;
        esac
    else
        echo "unknown"
    fi
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Configure npm to use Aliyun mirror (for China network environment)
configure_npm_mirror() {
    log_info "Configuring npm registry..."

    # Ensure npm is in PATH
    export PATH="/usr/bin:/usr/local/bin:/usr/local/sbin:$PATH"
    hash -r 2>/dev/null || true

    if ! command -v npm >/dev/null 2>&1; then
        log_warn "npm not found, skipping npm registry configuration"
        return 0
    fi

    npm config set registry https://registry.npmmirror.com
    log_info "npm registry set to Aliyun mirror"
}

# Install Node.js
install_nodejs() {
    local os=$1
    local use_mirror=$2

    log_info "Checking Node.js installation status..."

    if command_exists node; then
        local node_version=$(node --version)
        log_info "Node.js is already installed: $node_version"
        return 0
    fi

    log_warn "Node.js is not installed, starting installation..."

    case $os in
        macos)
            if ! command_exists brew; then
                log_error "Homebrew is not installed, please install Homebrew first"
                log_info "Install command: /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
                return 1
            fi
            brew install node
            ;;
        debian|ubuntu)
            sudo apt-get update
            sudo apt-get install -y ca-certificates curl gnupg
            sudo mkdir -p /etc/apt/keyrings

            # Try Aliyun mirror first if enabled
            if [ "$use_mirror" = "true" ]; then
                log_info "Using Aliyun mirror for Node.js installation..."
                # Use Aliyun Node.js mirror
                curl -fsSL https://mirrors.aliyun.com/nodesource/setup_20.x | sudo -E bash -
            else
                # Use NodeSource repository to install Node.js 20.x
                curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
            fi

            sudo apt-get install -y nodejs npm
            ;;
        alpine)
            sudo apk update
            # Use Aliyun mirror for apk if enabled
            if [ "$use_mirror" = "true" ]; then
                log_info "Configuring Aliyun mirror for apk..."
                sudo sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
            fi
            sudo apk add --no-cache nodejs npm
            ;;
        *)
            log_error "Unsupported operating system: $os"
            return 1
            ;;
    esac

    log_info "Node.js installation completed: $(node --version)"
}

# Install claude-code
install_claude_code() {
    local use_mirror=$1

    log_info "Checking claude-code installation status..."

    if command_exists claude-code; then
        local claude_version=$(claude-code --version 2>/dev/null || echo "unknown")
        log_info "claude-code is already installed: $claude_version"
        return 0
    fi

    log_warn "claude-code is not installed, starting installation..."

    # Configure npm to use user directory if not root
    if [ "$EUID" -ne 0 ]; then
        log_info "Not running as root, setting npm prefix to user directory..."
        npm config set prefix "$HOME/.local"
        export PATH="$HOME/.local/bin:$PATH"
    fi

    # Configure npm mirror if enabled
    if [ "$use_mirror" = "true" ]; then
        configure_npm_mirror
    fi

    # Ensure npm is visible in PATH after installation
    export PATH="/usr/bin:/usr/local/bin:/usr/local/sbin:$PATH"
    hash -r 2>/dev/null || true

    # Verify npm is available
    if ! command -v npm >/dev/null 2>&1; then
        log_error "npm command not found after Node.js installation"
        log_error "Node.js path: $(which node 2>/dev/null || echo 'not found')"
        log_error "Please ensure npm is properly installed with Node.js"
        return 1
    fi

    # Install using npm globally
    log_info "Installing claude-code via npm..."
    log_info "npm location: $(which npm)"
    npm install -g @anthropic-ai/claude-code

    if command_exists claude-code; then
        log_info "claude-code installation successful: $(claude-code --version 2>/dev/null || echo "installed")"
        return 0
    else
        log_error "claude-code installation failed"
        return 1
    fi
}

# Detect if in China network environment
detect_china_network() {
    # Check environment variable
    if [ "$USE_ALIYUN_MIRROR" = "true" ]; then
        echo "true"
        return
    fi

    # Try to detect by checking timezone
    if [ "$(cat /etc/timezone 2>/dev/null || echo $TZ)" = "Asia/Shanghai" ]; then
        echo "true"
        return
    fi

    # Try to ping a Chinese site to detect network environment
    if command_exists ping; then
        if ping -c 1 -W 1 mirrors.aliyun.com >/dev/null 2>&1; then
            echo "true"
            return
        fi
    fi

    echo "false"
}

# Main installation flow
main() {
    log_info "Starting Claude Code installation..."
    log_info "Detecting operating system..."

    # Check for proxy environment variables
    if [ -n "$http_proxy" ] || [ -n "$https_proxy" ]; then
        log_info "Proxy detected: http_proxy=$http_proxy https_proxy=$https_proxy"
        export http_proxy https_proxy
    fi

    local os=$(detect_os)
    log_info "Detected OS: $os"

    case $os in
        macos|debian|ubuntu|alpine)
            ;;
        *)
            log_error "Unsupported operating system: $os"
            log_error "Please install claude-code manually: https://claude.com/claude-code"
            exit 1
            ;;
    esac

    # Detect network environment
    local use_mirror=$(detect_china_network)
    if [ "$use_mirror" = "true" ]; then
        log_info "China network environment detected, will use Aliyun mirrors"
    else
        log_info "Using official mirrors (set USE_ALIYUN_MIRROR=true to use Aliyun mirrors)"
    fi

    # Install Node.js
    if ! install_nodejs $os $use_mirror; then
        log_error "Node.js installation failed"
        exit 1
    fi

    # Install claude-code
    if ! install_claude_code $use_mirror; then
        log_error "claude-code installation failed"
        exit 1
    fi

    log_info "✅ Claude Code installation completed!"
    log_info "You can now use 'claude-code' command to start Claude Code"
}

# Run main flow
main "$@"
