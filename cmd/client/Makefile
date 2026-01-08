# Makefile for clauded client
SHELL=/bin/bash

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

# 支持的平台
PLATFORMS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

# 默认目标
.PHONY: all
all: build

# 构建当前平台
.PHONY: build
build:
	@echo "🔨 构建 ${BINARY_NAME} for ${GOOS}/${GOARCH}..."
	@mkdir -p ${OUTPUT_DIR}
	CGO_ENABLED=${CGO_ENABLED} ${GO} build ${LDFLAGS} -o ${OUTPUT_DIR}/${BINARY_NAME}-${GOOS}-${GOARCH} ./main.go
	@echo "✅ 构建完成: ${OUTPUT_DIR}/${BINARY_NAME}-${GOOS}-${GOARCH}"

# 构建所有平台
.PHONY: build-all
build-all:
	@echo "🔨 构建所有平台的 ${BINARY_NAME}..."
	@mkdir -p ${OUTPUT_DIR}
	@for platform in ${PLATFORMS}; do \
		IFS='/' read -r os arch <<< "$$platform"; \
		echo "  📦 构建 $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 ${GO} build ${LDFLAGS} -o ${OUTPUT_DIR}/${BINARY_NAME}-$$os-$$arch ./main.go; \
	done
	@echo "✅ 所有平台构建完成"
	@ls -lh ${OUTPUT_DIR}

# 清理输出目录
.PHONY: clean
clean:
	@echo "🧹 清理输出目录..."
	@rm -rf ${OUTPUT_DIR}
	@echo "✅ 清理完成"

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

# 帮助
.PHONY: help
help:
	@echo "📖 可用的 Make 目标:"
	@echo ""
	@echo "  make build         - 构建当前平台版本"
	@echo "  make build-all     - 构建所有支持平台版本"
	@echo "  make clean         - 清理输出目录"
	@echo "  make info          - 显示构建信息"