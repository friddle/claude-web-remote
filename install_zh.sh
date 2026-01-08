#!/bin/bash

# Clauded 安装脚本（中文版）
# 此脚本根据您的操作系统和架构下载并安装 clauded

set -e

VERSION="${VERSION:-v0.1}"
REPO="friddle/claude-web-remote"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
# 使用 GitHub 加速代理
DOWNLOAD_URL="https://ghfast.top/https://github.com/${REPO}/releases/download/${VERSION}"

# 输出颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # 无颜色

# 打印函数
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
    echo -e "${BLUE}  Clauded 安装脚本${NC}"
    echo -e "${BLUE}  版本: ${VERSION}${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo ""
}

# 检测操作系统和架构
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
            print_error "不支持的操作系统: $OS"
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
            print_error "不支持的架构: $ARCH"
            exit 1
            ;;
    esac

    PLATFORM="${OS}-${ARCH}"
    print_success "检测到平台: $PLATFORM"
}

# 检查前置条件
check_prerequisites() {
    print_info "检查前置条件..."

    # 检查 curl 或 wget
    if command -v curl >/dev/null 2>&1; then
        DOWNLOAD_CMD="curl -fsSL"
    elif command -v wget >/dev/null 2>&1; then
        DOWNLOAD_CMD="wget -qO-"
    else
        print_error "未安装 curl 或 wget，请先安装其中之一"
        exit 1
    fi

    # 检查 INSTALL_DIR 是否可写
    if [ ! -w "$INSTALL_DIR" ]; then
        print_warning "安装目录 $INSTALL_DIR 不可写"
        print_info "改为安装到 $HOME/.local/bin..."
        INSTALL_DIR="$HOME/.local/bin"
        mkdir -p "$INSTALL_DIR"
    fi

    print_success "前置条件检查通过"
}

# 下载二进制文件
download_binary() {
    print_info "正在下载 clauded ${VERSION} (${PLATFORM})..." >&2

    BINARY_NAME="clauded-${PLATFORM}"
    DOWNLOAD_URL="${DOWNLOAD_URL}/${BINARY_NAME}"

    TEMP_DIR=$(mktemp -d)
    BINARY_PATH="${TEMP_DIR}/clauded"

    if ! $DOWNLOAD_CMD "${DOWNLOAD_URL}" -o "$BINARY_PATH"; then
        print_error "下载失败，无法从 ${DOWNLOAD_URL} 下载" >&2
        rm -rf "$TEMP_DIR"
        exit 1
    fi

    print_success "已下载到 $BINARY_PATH" >&2
    # 仅将 TEMP_DIR 输出到 stdout
    echo "$TEMP_DIR"
}

# 安装二进制文件
install_binary() {
    local TEMP_DIR=$1
    local BINARY_PATH="${TEMP_DIR}/clauded"

    print_info "正在安装 clauded 到 $INSTALL_DIR..."

    # 设置可执行权限
    chmod +x "$BINARY_PATH"

    # 安装二进制文件
    if ! mv "$BINARY_PATH" "$INSTALL_DIR/clauded"; then
        print_error "安装失败，无法移动到 $INSTALL_DIR"
        rm -rf "$TEMP_DIR"
        exit 1
    fi

    # 清理临时目录
    rm -rf "$TEMP_DIR"

    print_success "已安装 clauded 到 $INSTALL_DIR/clauded"
}

# 验证安装
verify_installation() {
    print_info "正在验证安装..."

    if command -v clauded >/dev/null 2>&1; then
        INSTALLED_VERSION=$(clauded --version 2>/dev/null || echo "unknown")
        print_success "Clauded 安装成功!"
        echo ""
        echo -e "  ${GREEN}二进制文件:${NC} $(which clauded)"
        echo -e "  ${GREEN}版本:${NC} ${INSTALLED_VERSION}"
        echo ""
    else
        print_error "在 PATH 中找不到 clauded"
        print_info "请将 $INSTALL_DIR 添加到您的 PATH:"
        echo ""
        echo "  export PATH=\"
$PATH:$INSTALL_DIR\""
        echo ""
        exit 1
    fi
}

# 启动 clauded（使用提供的参数）
start_clauded() {
    print_info "正在启动 clauded..."

    # 解析命令行参数
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
                echo "未知选项: $1"
                exit 1
                ;;
        esac
    done

    # 如果未提供必需参数，则提示用户输入
    if [ -z "$SESSION_NAME" ]; then
        read -p "请输入会话名称: " SESSION_NAME
    fi

    if [ -z "$REMOTE_SERVER" ]; then
        read -p "请输入远程服务器 (例如: localhost:8022 或 your-domain.com): " REMOTE_SERVER
    fi

    # 构建命令
    CMD="clauded --session $SESSION_NAME --remote $REMOTE_SERVER"

    if [ -n "$PASSWORD" ]; then
        CMD="$CMD --password $PASSWORD"
    fi

    echo ""
    print_info "使用以下配置启动 clauded:"
    echo "  会话: $SESSION_NAME"
    echo "  服务器: $REMOTE_SERVER"
    echo ""

    # 执行 clauded
    exec $CMD
}

# 主安装流程
main() {
    print_header

    # 请求确认
    read -p "是否继续安装? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_warning "安装已取消"
        exit 0
    fi

    # 安装步骤
    detect_platform
    check_prerequisites
    TEMP_DIR=$(download_binary)
    install_binary "$TEMP_DIR"
    verify_installation

    echo ""
    print_success "安装完成!"
    echo "提示: 如果未发现 Claude Code，clauded 将会自动进行安装。"
    echo ""

    # 询问用户是否立即启动 clauded
    read -p "是否立即启动 clauded? (y/N) " -n 1 -r
    echo ""
    echo ""

    if [[ $REPLY =~ ^[Yy]$ ]]; then
        start_clauded "$@"
    else
        echo "您可以稍后使用以下命令启动 clauded:"
        echo ""
        echo -e "  ${GREEN}clauded --session <名称> --remote <服务器>${NC}"
        echo ""
        echo "示例:"
        echo -e "  ${GREEN}clauded --session mysession --remote localhost:8022${NC}"
        echo ""
    fi
}

# 使用所有参数运行主函数
main "$@"