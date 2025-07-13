#!/bin/bash

# Build Go application in DevContainer
# Usage: ./scripts/go-build.sh

set -e

echo "🔨 Building Go application in DevContainer..."

# Start DevContainer services if needed
echo "📦 Starting DevContainer services..."
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
docker exec -w /workspace devcontainer-app-1 go build -o ./run/ai-gateway-hub .

echo "✅ Build completed successfully!"
echo "📍 Executable: run/ai-gateway-hub"
echo "📍 Configuration: run/.env"
echo ""
echo "Next steps:"
echo "  - Use ./scripts/go-run.sh to start the application"
echo "  - Edit run/.env to customize configuration"
