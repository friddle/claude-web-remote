#!/bin/bash

# Clauded Installation Script
# This script downloads and installs clauded based on your OS and architecture

set -e

VERSION="${VERSION:-v0.1}"
REPO="friddle/clauded"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print functions
print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_header() {
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  Clauded Installation Script${NC}"
    echo -e "${BLUE}  Version: ${VERSION}${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo ""
}

# Detect OS and architecture
detect_platform() {
    OS="$(uname -s)"
    ARCH="$(uname -m)"

    case "$OS" in
        Linux*)
            OS="linux"
            ;;
        Darwin*)
            OS="darwin"
            ;;
        *)
            print_error "Unsupported operating system: $OS"
            exit 1
            ;;
    esac

    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    PLATFORM="${OS}-${ARCH}"
    print_success "Detected platform: $PLATFORM"
}

# Check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."

    # Check for curl or wget
    if command -v curl >/dev/null 2>&1; then
        DOWNLOAD_CMD="curl -fsSL"
    elif command -v wget >/dev/null 2>&1; then
        DOWNLOAD_CMD="wget -qO-"
    else
        print_error "Neither curl nor wget is installed. Please install one of them."
        exit 1
    fi

    # Check if INSTALL_DIR is writable
    if [ ! -w "$INSTALL_DIR" ]; then
        print_warning "Install directory $INSTALL_DIR is not writable."
        print_info "Installing to $HOME/.local/bin instead..."
        INSTALL_DIR="$HOME/.local/bin"
        mkdir -p "$INSTALL_DIR"
    fi

    print_success "Prerequisites check passed"
}

# Display environment setup reminder
show_env_reminder() {
    echo ""
    echo -e "${YELLOW}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${YELLOW}  IMPORTANT: Claude Code Environment Setup${NC}"
    echo -e "${YELLOW}═══════════════════════════════════════════════════════════════${NC}"
    echo ""
    echo -e "${RED}Before starting clauded, make sure you have:${NC}"
    echo ""
    echo "1. Installed Claude Code CLI:"
    echo "   npm install -g @anthropic-ai/claude-code"
    echo ""
    echo "2. Configured your Anthropic API key:"
    echo "   export ANTHROPIC_API_KEY='your-api-key-here'"
    echo ""
    echo "3. OR authenticated with Claude Code:"
    echo "   claude auth login"
    echo ""
    echo "4. For custom API endpoints, configure:"
    echo "   export ANTHROPIC_BASE_URL='https://your-endpoint.com'"
    echo "   export ANTHROPIC_AUTH_TOKEN='your-token'"
    echo ""
    echo -e "${YELLOW}═══════════════════════════════════════════════════════════════${NC}"
    echo ""
}

# Download binary
download_binary() {
    print_info "Downloading clauded ${VERSION} for ${PLATFORM}..." >&2

    BINARY_NAME="clauded-${PLATFORM}"
    DOWNLOAD_URL="${DOWNLOAD_URL}/${BINARY_NAME}"

    TEMP_DIR=$(mktemp -d)
    BINARY_PATH="${TEMP_DIR}/clauded"

    if ! $DOWNLOAD_CMD "${DOWNLOAD_URL}" -o "$BINARY_PATH"; then
        print_error "Failed to download clauded from ${DOWNLOAD_URL}" >&2
        rm -rf "$TEMP_DIR"
        exit 1
    fi

    print_success "Downloaded to $BINARY_PATH" >&2
    # Only output TEMP_DIR to stdout
    echo "$TEMP_DIR"
}

# Install binary
install_binary() {
    local TEMP_DIR=$1
    local BINARY_PATH="${TEMP_DIR}/clauded"

    print_info "Installing clauded to $INSTALL_DIR..."

    # Make binary executable
    chmod +x "$BINARY_PATH"

    # Install binary
    if ! mv "$BINARY_PATH" "$INSTALL_DIR/clauded"; then
        print_error "Failed to install clauded to $INSTALL_DIR"
        rm -rf "$TEMP_DIR"
        exit 1
    fi

    # Clean up temp directory
    rm -rf "$TEMP_DIR"

    print_success "Installed clauded to $INSTALL_DIR/clauded"
}

# Verify installation
verify_installation() {
    print_info "Verifying installation..."

    if command -v clauded >/dev/null 2>&1; then
        INSTALLED_VERSION=$(clauded --version 2>/dev/null || echo "unknown")
        print_success "Clauded installed successfully!"
        echo ""
        echo -e "  ${GREEN}Binary:${NC} $(which clauded)"
        echo -e "  ${GREEN}Version:${NC} ${INSTALLED_VERSION}"
        echo ""
    else
        print_error "Clauded binary not found in PATH"
        print_info "Add $INSTALL_DIR to your PATH:"
        echo ""
        echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
        echo ""
        exit 1
    fi
}

# Start clauded with provided arguments
start_clauded() {
    print_info "Starting clauded..."

    # Parse command line arguments
    SESSION_NAME=""
    REMOTE_SERVER=""
    PASSWORD=""

    while [[ $# -gt 0 ]]; do
        case $1 in
            --session)
                SESSION_NAME="$2"
                shift 2
                ;;
            --remote)
                REMOTE_SERVER="$2"
                shift 2
                ;;
            --password)
                PASSWORD="$2"
                shift 2
                ;;
            *)
                echo "Unknown option: $1"
                exit 1
                ;;
        esac
    done

    # Prompt for required arguments if not provided
    if [ -z "$SESSION_NAME" ]; then
        read -p "Enter session name: " SESSION_NAME
    fi

    if [ -z "$REMOTE_SERVER" ]; then
        read -p "Enter remote server (e.g., localhost:8022 or your-domain.com): " REMOTE_SERVER
    fi

    # Build command
    CMD="clauded --session $SESSION_NAME --remote $REMOTE_SERVER"

    if [ -n "$PASSWORD" ]; then
        CMD="$CMD --password $PASSWORD"
    fi

    echo ""
    print_info "Starting clauded with:"
    echo "  Session: $SESSION_NAME"
    echo "  Remote: $REMOTE_SERVER"
    echo ""

    # Execute clauded
    exec $CMD
}

# Main installation flow
main() {
    print_header

    # Show environment reminder first
    show_env_reminder

    # Ask for confirmation
    read -p "Continue with installation? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_warning "Installation cancelled"
        exit 0
    fi

    # Installation steps
    detect_platform
    check_prerequisites
    TEMP_DIR=$(download_binary)
    install_binary "$TEMP_DIR"
    verify_installation

    echo ""
    print_success "Installation complete!"
    echo ""

    # Ask if user wants to start clauded now
    read -p "Start clauded now? (y/N) " -n 1 -r
    echo ""
    echo ""

    if [[ $REPLY =~ ^[Yy]$ ]]; then
        start_clauded "$@"
    else
        echo "You can start clauded later with:"
        echo ""
        echo -e "  ${GREEN}clauded --session <name> --remote <server>${NC}"
        echo ""
        echo "Example:"
        echo -e "  ${GREEN}clauded --session mysession --remote localhost:8022${NC}"
        echo ""
    fi
}

# Run main function with all arguments
main "$@"
