#!/bin/bash
# Deployment script for racknerd server
# This script pulls the latest code and restarts the Docker container

set -e

# Configuration
REPO_URL="https://github.com/friddle/claude-web-remote.git"
DOCKER_IMAGE="friddlecopper/clauded-port-forward:latest"
CONTAINER_NAME="clauded-server"
DEPLOY_DIR="$HOME/racknerd/clauded"

echo "🚀 Starting deployment..."
echo "📍 Deploy directory: $DEPLOY_DIR"

# Create deploy directory if it doesn't exist
mkdir -p "$DEPLOY_DIR"
cd "$DEPLOY_DIR"

# Clone or pull repository
if [ ! -d ".git" ]; then
    echo "📥 Cloning repository..."
    git clone "$REPO_URL" .
else
    echo "📥 Pulling latest changes..."
    git pull origin main
fi

# Check if docker-compose.yaml exists
if [ ! -f "cmd/server/docker-compose.yaml" ]; then
    echo "❌ docker-compose.yaml not found"
    exit 1
fi

cd cmd/server

# Stop existing container
echo "🛑 Stopping existing container..."
if docker ps -q -f name="$CONTAINER_NAME" | grep -q .; then
    docker stop "$CONTAINER_NAME" || true
fi

# Remove old container
echo "🗑️  Removing old container..."
if docker ps -aq -f name="$CONTAINER_NAME" | grep -q .; then
    docker rm "$CONTAINER_NAME" || true
fi

# Pull latest image
echo "📦 Pulling latest Docker image..."
if ! docker pull "$DOCKER_IMAGE"; then
    echo "⚠️  Failed to pull image, building locally instead..."
    docker-compose build
fi

# Start new container
echo "▶️  Starting new container..."
docker-compose up -d

# Wait for container to be healthy
echo "⏳ Waiting for container to be ready..."
sleep 5

# Show logs
echo "📋 Container logs:"
docker-compose logs --tail=20

# Show container status
echo "✅ Deployment complete!"
echo "📊 Container status:"
docker ps -f name="$CONTAINER_NAME"

echo ""
echo "🌐 Access URLs:"
echo "   Health check: http://localhost:8088/health"
echo "   API: http://localhost:8088/api/v1/notifications/stream?session_id=test"
