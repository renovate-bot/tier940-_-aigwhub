#!/bin/bash

# Build Go application in DevContainer
# Usage: /go-build

set -e

echo "🔨 Building Go application in DevContainer..."

# Start DevContainer services if needed
echo "📦 Checking DevContainer services..."
cd .devcontainer
if ! docker compose ps | grep -q "devcontainer-app-1.*Up"; then
    echo "🚀 Starting DevContainer services..."
    docker compose up -d
    # Wait for services to start
    sleep 3
fi
cd ..

# Fix Go module permissions
echo "🔧 Fixing Go module permissions..."
docker exec -u root devcontainer-app-1 chown -R vscode:vscode /go || true

# Download Go dependencies in DevContainer
echo "📥 Downloading Go dependencies..."
docker exec -w /workspace devcontainer-app-1 go mod download

# Build application in DevContainer
echo "🏗️ Building application..."
docker exec -w /workspace devcontainer-app-1 go build -o ai-gateway-hub .

# Clean and prepare run directory
echo "🗑️ Cleaning previous build artifacts..."
rm -rf run

echo "📁 Creating run directory structure..."
mkdir -p run/data run/logs

# Copy build results to run directory
echo "📋 Copying build artifacts..."
cp ai-gateway-hub run/
cp .env.example run/.env

echo "✅ Build completed successfully!"
echo "📍 Executable: run/ai-gateway-hub"
echo "📍 Configuration: run/.env"
echo ""
echo "Next steps:"
echo "  - Use /go-run to start the application"
echo "  - Edit run/.env to customize configuration"