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
	@echo "构建 ${BINARY_NAME} for ${GOOS}/${GOARCH}..."
	@mkdir -p ${OUTPUT_DIR}
	CGO_ENABLED=${CGO_ENABLED} ${GO} build ${LDFLAGS} -o ${OUTPUT_DIR}/${BINARY_NAME} ./main.go
	@echo "构建完成: ${OUTPUT_DIR}/${BINARY_NAME}"

# 构建所有平台
.PHONY: build-all
build-all:
	@echo "构建所有平台的 ${BINARY_NAME} (仅 macOS 和 Linux)..."
	@mkdir -p ${OUTPUT_DIR}
	@for platform in ${PLATFORMS}; do \
		IFS='/' read -r os arch <<< "$$platform"; \
		echo "构建 $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 ${GO} build ${LDFLAGS} -o ${OUTPUT_DIR}/${BINARY_NAME}-$$os-$$arch ./main.go; \
	done
	@echo "所有平台构建完成，输出目录: ${OUTPUT_DIR}"

# 构建特定平台
.PHONY: build-linux
build-linux:
	@echo "构建 Linux 版本..."
	@mkdir -p ${OUTPUT_DIR}
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 ${GO} build ${LDFLAGS} -o ${OUTPUT_DIR}/${BINARY_NAME}-linux-amd64 ./main.go
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 ${GO} build ${LDFLAGS} -o ${OUTPUT_DIR}/${BINARY_NAME}-linux-arm64 ./main.go
	@echo "Linux 版本构建完成"

.PHONY: build-darwin
build-darwin:
	@echo "构建 macOS 版本..."
	@mkdir -p ${OUTPUT_DIR}
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 ${GO} build ${LDFLAGS} -o ${OUTPUT_DIR}/${BINARY_NAME}-darwin-amd64 ./main.go
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 ${GO} build ${LDFLAGS} -o ${OUTPUT_DIR}/${BINARY_NAME}-darwin-arm64 ./main.go
	@echo "macOS 版本构建完成"

# 帮助
.PHONY: help
help:
	@echo "可用的 Make 目标:"
	@echo "  make         - 构建当前平台版本"
	@echo "  make build   - 构建当前平台版本"
	@echo "  make build-all - 构建所有支持平台版本 (仅 macOS 和 Linux)"
	@echo "  make build-linux - 构建 Linux 版本"
	@echo "  make build-darwin - 构建 macOS 版本"
	@echo "  make help    - 显示此帮助信息"
	@echo ""
	@echo "支持的平台: macOS (amd64, arm64), Linux (amd64, arm64)" 