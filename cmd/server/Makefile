# Makefile for clauded-port-forward
# Docker build and deployment

# Variables
IMAGE_NAME=friddlecopper/clauded-port-forward
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Docker variables
DOCKER_TAG?=latest

# Default target
.PHONY: all
all: clean build

# Build Go binary
.PHONY: build-go
build-go:
	@echo "Building Go binary..."
	@mkdir -p dist
	go build -o dist/clauded-port-forward ./cmd/server
	@echo "Go binary built: dist/clauded-port-forward"

# Build Docker image (Native)
.PHONY: build
build:
	@echo "Building ${IMAGE_NAME} Docker image (Native)..."
	docker build -t ${IMAGE_NAME}:${DOCKER_TAG} .
	@echo "Docker image built: ${IMAGE_NAME}:${DOCKER_TAG}"

# Build Docker image (AMD64/x86)
.PHONY: build-amd64
build-amd64:
	@echo "Building ${IMAGE_NAME} Docker image (linux/amd64)..."
	docker build --platform linux/amd64 -t ${IMAGE_NAME}:${DOCKER_TAG} .
	@echo "Docker image built: ${IMAGE_NAME}:${DOCKER_TAG} (amd64)"

# Build Docker image (ARM64)
.PHONY: build-arm64
build-arm64:
	@echo "Building ${IMAGE_NAME} Docker image (linux/arm64)..."
	docker build --platform linux/arm64 -t ${IMAGE_NAME}:arm64 .
	@echo "Docker image built: ${IMAGE_NAME}:arm64"

# Build multi-arch image (AMD64 & ARM64)
.PHONY: build-multiarch
build-multiarch:
	@echo "Building ${IMAGE_NAME} multi-arch Docker image (linux/amd64 & linux/arm64)..."
	docker buildx build --platform linux/amd64,linux/arm64 -t ${IMAGE_NAME}:latest .
	@echo "Multi-arch Docker image built: ${IMAGE_NAME}:latest"

# Build and push multi-arch image
.PHONY: push-multiarch
push-multiarch:
	@echo "Building and pushing ${IMAGE_NAME} multi-arch Docker image..."
	docker buildx build --platform linux/amd64,linux/arm64 -t ${IMAGE_NAME}:latest --push .
	@echo "Multi-arch image built and pushed: ${IMAGE_NAME}:latest"

# Build and push Docker image
.PHONY: push
push: build
	@echo "Pushing Docker image to registry..."
	docker push ${IMAGE_NAME}:${DOCKER_TAG}
	@echo "Image pushed: ${IMAGE_NAME}:${DOCKER_TAG}"

# Push AMD64 specifically
.PHONY: push-amd64
push-amd64: build-amd64
	@echo "Pushing ${IMAGE_NAME}:${DOCKER_TAG} (amd64)..."
	docker push ${IMAGE_NAME}:${DOCKER_TAG}
	@echo "Image pushed"

# Build with version tag
.PHONY: build-version
build-version:
	@echo "Building ${IMAGE_NAME}:${VERSION}..."
	docker build -t ${IMAGE_NAME}:${VERSION} .
	docker tag ${IMAGE_NAME}:${VERSION} ${IMAGE_NAME}:latest
	@echo "Docker image built: ${IMAGE_NAME}:${VERSION}"

# Push with version tag
.PHONY: push-version
push-version: build-version
	docker push ${IMAGE_NAME}:${VERSION}
	docker push ${IMAGE_NAME}:latest

# Run with docker-compose
.PHONY: up
up:
	docker-compose up -d

# Stop docker-compose
.PHONY: down
down:
	docker-compose down

# View logs
.PHONY: logs
logs:
	docker-compose logs -f

# Clean
.PHONY: clean
clean:
	@echo "Cleaning build files..."
	@rm -rf dist
	@docker rmi ${IMAGE_NAME}:${DOCKER_TAG} 2>/dev/null || true
	@echo "Clean completed"

# Help
.PHONY: help
help:
	@echo "Available Make targets:"
	@echo "  build           - Build Docker image (native)"
	@echo "  build-amd64     - Build Docker image (linux/amd64)"
	@echo "  build-arm64     - Build Docker image (linux/arm64)"
	@echo "  build-multiarch - Build multi-arch Docker image (amd64 & arm64)"
	@echo "  push            - Build and push Docker image (native)"
	@echo "  push-amd64      - Build and push (amd64)"
	@echo "  push-multiarch  - Build and push multi-arch image"
	@echo "  build-go        - Build Go binary only"
	@echo "  build-version   - Build with version tag"
	@echo "  push-version    - Push with version tag"
	@echo "  up              - Start with docker-compose"
	@echo "  down            - Stop docker-compose"
	@echo "  logs            - View docker-compose logs"
	@echo "  clean           - Clean build files"
