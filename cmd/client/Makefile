# Makefile for gotty-piko-client
SHELL=/bin/bash
# 支持 macOS 和 Linux 平台编译

# 变量定义
BINARY_NAME=clauded
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go 相关变量
GO=go
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
CGO_ENABLED?=0

# 编译参数
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT} -s -w"

# 输出目录
OUTPUT_DIR=output

# 支持的平台 (仅 macOS 和 Linux)
PLATFORMS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

# 默认目标
.PHONY: all
all: build

# 构建当前平台
.PHONY: build
build:
	@echo "🔨 构建 ${BINARY_NAME} for ${GOOS}/${GOARCH}..."
	@mkdir -p ${OUTPUT_DIR}
	CGO_ENABLED=${CGO_ENABLED} ${GO} build ${LDFLAGS} -o ${OUTPUT_DIR}/${BINARY_NAME} ./main.go
	@echo "✅ 构建完成: ${OUTPUT_DIR}/${BINARY_NAME}"

# 清理输出目录
.PHONY: clean
clean:
	@echo "🧹 清理输出目录..."
	@rm -rf ${OUTPUT_DIR}
	@echo "✅ 清理完成"

# 构建所有平台并打包
.PHONY: build-all
build-all:
	@echo "🔨 构建所有平台的 ${BINARY_NAME} (仅 macOS 和 Linux)..."
	@mkdir -p ${OUTPUT_DIR}
	@for platform in ${PLATFORMS}; do \
		IFS='/' read -r os arch <<< "$$platform"; \
		echo "  📦 构建 $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 ${GO} build ${LDFLAGS} -o ${OUTPUT_DIR}/${BINARY_NAME}-$$os-$$arch ./main.go; \
	done
	@echo ""
	@echo "📦 打包所有平台..."
	@for platform in ${PLATFORMS}; do \
		IFS='/' read -r os arch <<< "$$platform"; \
		binary_name="${BINARY_NAME}-$$os-$$arch"; \
		archive_name="${BINARY_NAME}-${VERSION}-$$os-$$arch.tar.gz"; \
		echo "  📦 打包 $$archive_name..."; \
		tar -czf ${OUTPUT_DIR}/$$archive_name -C ${OUTPUT_DIR} $$binary_name; \
	done
	@echo ""
	@echo "✅ 所有平台构建并打包完成"
	@echo "📁 输出目录: ${OUTPUT_DIR}/"
	@ls -lh ${OUTPUT_DIR}/*.tar.gz 2>/dev/null || echo "  (没有找到打包文件)"

# 仅打包已构建的二进制文件
.PHONY: package
package:
	@echo "📦 打包已构建的二进制文件..."
	@if [ ! -d "${OUTPUT_DIR}" ]; then \
		echo "❌ 错误: ${OUTPUT_DIR} 目录不存在，请先运行 make build-all"; \
		exit 1; \
	fi
	@for platform in ${PLATFORMS}; do \
		IFS='/' read -r os arch <<< "$$platform"; \
		binary_name="${BINARY_NAME}-$$os-$$arch"; \
		if [ -f "${OUTPUT_DIR}/$$binary_name" ]; then \
			archive_name="${BINARY_NAME}-${VERSION}-$$os-$$arch.tar.gz"; \
			echo "  📦 打包 $$archive_name..."; \
			tar -czf ${OUTPUT_DIR}/$$archive_name -C ${OUTPUT_DIR} $$binary_name; \
		fi \
	done
	@echo ""
	@echo "✅ 打包完成"
	@ls -lh ${OUTPUT_DIR}/*.tar.gz 2>/dev/null || echo "  (没有找到打包文件)"

# 构建特定平台
.PHONY: build-linux
build-linux:
	@echo "🔨 构建 Linux 版本..."
	@mkdir -p ${OUTPUT_DIR}
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 ${GO} build ${LDFLAGS} -o ${OUTPUT_DIR}/${BINARY_NAME}-linux-amd64 ./main.go
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 ${GO} build ${LDFLAGS} -o ${OUTPUT_DIR}/${BINARY_NAME}-linux-arm64 ./main.go
	@echo "✅ Linux 版本构建完成"

.PHONY: build-darwin
build-darwin:
	@echo "🔨 构建 macOS 版本..."
	@mkdir -p ${OUTPUT_DIR}
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 ${GO} build ${LDFLAGS} -o ${OUTPUT_DIR}/${BINARY_NAME}-darwin-amd64 ./main.go
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 ${GO} build ${LDFLAGS} -o ${OUTPUT_DIR}/${BINARY_NAME}-darwin-arm64 ./main.go
	@echo "✅ macOS 版本构建完成"

# 显示构建信息
.PHONY: info
info:
	@echo "📋 构建信息:"
	@echo "  二进制名称: ${BINARY_NAME}"
	@echo "  版本: ${VERSION}"
	@echo "  构建时间: ${BUILD_TIME}"
	@echo "  Git 提交: ${GIT_COMMIT}"
	@echo "  Go 版本: $$(go version)"
	@echo "  当前平台: ${GOOS}/${GOARCH}"
	@echo ""
	@echo "📦 支持的平台:"
	@for platform in ${PLATFORMS}; do \
		echo "  - $$platform"; \
	done

# 帮助
.PHONY: help
help:
	@echo "📖 可用的 Make 目标:"
	@echo ""
	@echo "构建相关:"
	@echo "  make build         - 构建当前平台版本"
	@echo "  make build-all     - 构建所有支持平台版本并打包 (推荐)"
	@echo "  make build-linux   - 仅构建 Linux 版本 (amd64, arm64)"
	@echo "  make build-darwin  - 仅构建 macOS 版本 (amd64, arm64)"
	@echo ""
	@echo "打包相关:"
	@echo "  make package       - 打包已构建的二进制文件为 tar.gz"
	@echo ""
	@echo "其他:"
	@echo "  make clean         - 清理输出目录"
	@echo "  make info          - 显示构建信息"
	@echo "  make help          - 显示此帮助信息"
	@echo ""
	@echo "📦 支持的平台:"
	@echo "  - Linux (amd64, arm64)"
	@echo "  - macOS (amd64, arm64)"
	@echo ""
	@echo "💡 示例:"
	@echo "  make build-all    # 构建所有平台并打包"
	@echo "  make clean && make build-all  # 清理后重新构建"
	@echo ""
	@echo "📁 输出目录: ${OUTPUT_DIR}/"
	@echo "  - 二进制文件: clauded-{os}-{arch}"
	@echo "  - 打包文件: clauded-{version}-{os}-{arch}.tar.gz" 