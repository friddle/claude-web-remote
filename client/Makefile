# Makefile for gotty-piko-client
SHELL=/bin/bash
# 支持多平台编译

# 变量定义
BINARY_NAME=gottyp
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go 相关变量
GO=go
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
CGO_ENABLED?=0

# Android SDK 相关变量
ANDROID_SDK?=$(HOME)/sdk/Android/Sdk
ANDROID_NDK=$(ANDROID_SDK)/ndk/default

# 编译参数
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT} -s -w"

# 输出目录
DIST_DIR=dist
BUILD_DIR=dist

# 支持的平台
PLATFORMS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

# 默认目标
.PHONY: all
all: build

# 构建当前平台
.PHONY: build
build:
	@echo "构建 ${BINARY_NAME} for ${GOOS}/${GOARCH}..."
	@mkdir -p ${BUILD_DIR}
	CGO_ENABLED=${CGO_ENABLED} ${GO} build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME} ./main.go
	@echo "构建完成: ${BUILD_DIR}/${BINARY_NAME}"

# 构建所有平台
.PHONY: build-all
build-all: 
	@echo "构建所有平台的 ${BINARY_NAME}..."
	@mkdir -p ${DIST_DIR}
	@for platform in ${PLATFORMS}; do \
		IFS='/' read -r os arch <<< "$$platform"; \
		echo "构建 $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 ${GO} build ${LDFLAGS} -o ${DIST_DIR}/${BINARY_NAME}-$$os-$$arch ./main.go; \
		if [ "$$os" = "windows" ]; then \
			mv ${DIST_DIR}/${BINARY_NAME}-$$os-$$arch ${DIST_DIR}/${BINARY_NAME}-$$os-$$arch.exe; \
		fi; \
	done
	@echo "所有平台构建完成，输出目录: ${DIST_DIR}"

# 构建特定平台
.PHONY: build-linux
build-linux:
	@echo "构建 Linux 版本..."
	@mkdir -p ${DIST_DIR}
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 ${GO} build ${LDFLAGS} -o ${DIST_DIR}/${BINARY_NAME}-linux-amd64 ./main.go
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 ${GO} build ${LDFLAGS} -o ${DIST_DIR}/${BINARY_NAME}-linux-arm64 ./main.go
	@echo "Linux 版本构建完成"

.PHONY: build-darwin
build-darwin:
	@echo "构建 macOS 版本..."
	@mkdir -p ${DIST_DIR}
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 ${GO} build ${LDFLAGS} -o ${DIST_DIR}/${BINARY_NAME}-darwin-amd64 ./main.go
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 ${GO} build ${LDFLAGS} -o ${DIST_DIR}/${BINARY_NAME}-darwin-arm64 ./main.go
	@echo "macOS 版本构建完成"

# Android 编译目标
.PHONY: build-android
build-android:
	@echo "构建所有 Android API 版本 (28-35)..."
	@mkdir -p ${DIST_DIR}
	@if [ ! -d "$(ANDROID_NDK)" ]; then \
		echo "错误: Android NDK 未找到，请检查 ANDROID_SDK 环境变量"; \
		echo "当前 ANDROID_SDK: $(ANDROID_SDK)"; \
		exit 1; \
	fi
	@for api in 28 29 30 31 32 33 34 35; do \
		echo "构建 Android API $$api 版本..."; \
		GOOS=android GOARCH=arm64 CGO_ENABLED=1 \
		CC=$(ANDROID_NDK)/toolchains/llvm/prebuilt/darwin-x86_64/bin/aarch64-linux-android$$api-clang \
		CXX=$(ANDROID_NDK)/toolchains/llvm/prebuilt/darwin-x86_64/bin/aarch64-linux-android$$api-clang++ \
		${GO} build ${LDFLAGS} -o ${DIST_DIR}/${BINARY_NAME}-android-arm64-api$$api ./main.go; \
	done
	@echo "所有 Android API 版本构建完成，输出目录: ${DIST_DIR}"

# 帮助
.PHONY: help
help:
	@echo "可用的 Make 目标:"
	@echo "  build        - 构建当前平台版本"
	@echo "  build-all    - 构建所有平台版本"
	@echo "  build-linux  - 构建 Linux 版本"
	@echo "  build-darwin - 构建 macOS 版本"
	@echo "  build-android- 构建 Android API ${ANDROID_API} 版本"
	@echo "  build-android-all - 构建所有 Android API 版本 (28-35)"
	@echo "  help         - 显示此帮助信息"
	@echo ""
	@echo "环境变量:"
	@echo "  ANDROID_SDK  - Android SDK 路径 (默认: ~/sdk/Android)"
	@echo "  ANDROID_API  - Android API 版本 (默认: 23)" 