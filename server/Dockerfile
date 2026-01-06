# Build stage
FROM docker.linkos.org/library/golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Runtime stage
FROM docker.linkos.org/library/alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS and piko
# Download correct architecture based on TARGETPLATFORM
RUN apk --no-cache add ca-certificates curl && \
    ARCH=$(uname -m) && \
    if [ "$ARCH" = "aarch64" ]; then \
        PIKO_ARCH="linux-arm64"; \
    elif [ "$ARCH" = "x86_64" ]; then \
        PIKO_ARCH="linux-amd64"; \
    else \
        echo "Unsupported architecture: $ARCH" && exit 1; \
    fi && \
    curl -L "https://github.com/andydunstall/piko/releases/latest/download/piko-${PIKO_ARCH}" -o /usr/local/bin/piko && \
    chmod +x /usr/local/bin/piko

# Copy the binary from builder
COPY --from=builder /app/server .

# Expose ports
EXPOSE 8022 8023 8088

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD curl -f http://localhost:8088/health || exit 1

# Run the server
CMD ["./server"]
