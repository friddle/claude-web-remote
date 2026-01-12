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

# Check if sudo is available
can_use_sudo() {
    # If already root, no need for sudo
    if [ "$EUID" -eq 0 ]; then
        return 0
    fi

    # Try sudo with a simple command (no password)
    if sudo -n true 2>/dev/null; then
        return 0
    fi

    # Try sudo with password (will prompt user)
    if sudo -v >/dev/null 2>&1; then
        return 0
    fi

    return 1
}

# Set SUDO_PREFIX based on whether we can use sudo
SUDO_PREFIX=""
USE_SUDO="false"

setup_sudo() {
    if [ "$EUID" -eq 0 ]; then
        USE_SUDO="true"
        SUDO_PREFIX=""
        log_info "Running as root"
    elif can_use_sudo; then
        USE_SUDO="true"
        SUDO_PREFIX="sudo"
        log_info "sudo is available, will use it for installation"
    else
        USE_SUDO="false"
        SUDO_PREFIX=""
        log_info "sudo not available, will install to user directory"
    fi
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

# Get npm bin directory
get_npm_bin_dir() {
    # Try to get npm bin directory, fallback to common locations
    local npm_bin=$(npm bin -g 2>/dev/null)
    if [ -n "$npm_bin" ]; then
        echo "$npm_bin"
        return
    fi

    # Fallback to common npm bin directories
    for dir in "/opt/homebrew/bin" "/usr/local/bin" "$HOME/.local/bin"; do
        if [ -d "$dir" ]; then
            echo "$dir"
            return
        fi
    done

    # Default fallback
    echo "/usr/local/bin"
}

# Setup PATH in shell profile
setup_path() {
    local path_entry="$1"

    # Check if PATH entry already exists in common shell configs
    for profile in "$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile"; do
        if [ -f "$profile" ]; then
            if grep -q "PATH.*$path_entry" "$profile" 2>/dev/null; then
                return 0  # Already exists
            fi
        fi
    done

    # Add to .profile (for login shells)
    if [ -f "$HOME/.profile" ]; then
        echo "export PATH=\"\$PATH:$path_entry\"" >> "$HOME/.profile"
    fi

    # Add to .bashrc (for bash shells)
    if [ -f "$HOME/.bashrc" ]; then
        echo "export PATH=\"\$PATH:$path_entry\"" >> "$HOME/.bashrc"
    fi

    # Add to .zshrc (for zsh shells)
    if [ -f "$HOME/.zshrc" ]; then
        echo "export PATH=\"\$PATH:$path_entry\"" >> "$HOME/.zshrc"
    fi

    log_info "Added $path_entry to PATH in shell profile"
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
            $SUDO_PREFIX apt-get update
            $SUDO_PREFIX apt-get install -y ca-certificates curl gnupg
            $SUDO_PREFIX mkdir -p /etc/apt/keyrings

            # Try Aliyun mirror first if enabled
            if [ "$use_mirror" = "true" ]; then
                log_info "Using Aliyun mirror for Node.js installation..."
                # Use Aliyun Node.js mirror
                curl -fsSL https://mirrors.aliyun.com/nodesource/setup_20.x | $SUDO_PREFIX -E bash -
            else
                # Use NodeSource repository to install Node.js 20.x
                curl -fsSL https://deb.nodesource.com/setup_20.x | $SUDO_PREFIX -E bash -
            fi

            $SUDO_PREFIX apt-get install -y nodejs npm

            # Ensure common Node.js paths are in PATH
            export PATH="/usr/bin:/usr/local/bin:/usr/local/sbin:$PATH"
            hash -r 2>/dev/null || true

            # Persist PATH to shell profiles for future sessions (only for user installation)
            if [ "$USE_SUDO" = "false" ]; then
                setup_path "/usr/local/bin"
                setup_path "/usr/bin"
            fi

            # Verify npm is available after installation
            if ! command -v npm >/dev/null 2>&1; then
                log_error "npm was not installed properly"
                log_error "node: $(which node 2>/dev/null || echo 'not found')"
                log_error "npm: $(which npm 2>/dev/null || echo 'not found')"
                return 1
            fi
            ;;
        alpine)
            $SUDO_PREFIX apk update
            # Use Aliyun mirror for apk if enabled
            if [ "$use_mirror" = "true" ]; then
                log_info "Configuring Aliyun mirror for apk..."
                $SUDO_PREFIX sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
            fi
            $SUDO_PREFIX apk add --no-cache nodejs npm
            ;;
        *)
            log_error "Unsupported operating system: $os"
            return 1
            ;;
    esac

    log_info "Node.js installation completed: $(node --version)"
}

# Install claude
install_claude_code() {
    local use_mirror=$1

    log_info "Checking claude installation status..."

    # Check for claude in common paths
    if command_exists claude; then
        local claude_version=$(claude --version 2>/dev/null || echo "unknown")
        log_info "claude is already installed: $claude_version"
        return 0
    fi

    # Also check global npm paths when using sudo
    if [ "$USE_SUDO" = "true" ]; then
        export PATH="/usr/local/bin:/usr/bin:$PATH"
        hash -r 2>/dev/null || true
        if command_exists claude; then
            local claude_version=$(claude --version 2>/dev/null || echo "unknown")
            log_info "claude is already installed: $claude_version"
            return 0
        fi
    fi

    log_warn "claude is not installed, starting installation..."

    # Configure npm to use user directory if not using sudo
    if [ "$USE_SUDO" = "false" ]; then
        log_info "No sudo available, setting npm prefix to user directory..."
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
    log_info "Installing claude via npm..."
    log_info "npm location: $(which npm)"
    $SUDO_PREFIX npm install -g @anthropic-ai/claude-code

    # Create symlink from claude-code to claude if it doesn't exist
    if [ "$USE_SUDO" = "true" ]; then
        # Get npm bin directory and add to PATH
        NPM_BIN_DIR=$(get_npm_bin_dir)
        log_info "npm bin directory: $NPM_BIN_DIR"
        export PATH="$NPM_BIN_DIR:$PATH"
    else
        export PATH="$HOME/.local/bin:$PATH"
    fi
    hash -r 2>/dev/null || true

    # Try to find where claude-code was installed and create symlink
    if command_exists claude-code; then
        CLAUDE_CODE_PATH=$(which claude-code)
        CLAUDE_DIR=$(dirname "$CLAUDE_CODE_PATH")
        if [ ! -e "$CLAUDE_DIR/claude" ]; then
            ln -s "$CLAUDE_CODE_PATH" "$CLAUDE_DIR/claude"
            log_info "Created symlink from claude-code to claude"
        fi
    fi

    if command_exists claude; then
        log_info "claude installation successful: $(claude --version 2>/dev/null || echo "installed")"
        return 0
    else
        log_error "claude installation failed"
        log_error "PATH: $PATH"
        log_error "which claude: $(which claude 2>/dev/null || echo 'not found')"
        log_error "ls /usr/local/bin/claude: $(ls -la /usr/local/bin/claude 2>/dev/null || echo 'not found')"
        log_error "ls /opt/homebrew/bin/claude: $(ls -la /opt/homebrew/bin/claude 2>/dev/null || echo 'not found')"
        log_error "npm config get prefix: $(npm config get prefix 2>/dev/null || echo 'not found')"
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

    # Setup sudo first
    setup_sudo

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
        log_error "claude installation failed"
        exit 1
    fi

    log_info "✅ Claude Code installation completed!"
    log_info "You can now use 'claude' command to start Claude Code"
}

# Run main flow
main "$@"
