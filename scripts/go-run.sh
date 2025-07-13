#!/bin/bash

# Run Go application in DevContainer
# Usage: ./scripts/go-run.sh

set -e

echo "🚀 Starting Go application in DevContainer..."

# Check DevContainer services
echo "📦 Checking DevContainer services..."
cd .devcontainer
if ! docker compose ps | grep -q "devcontainer-app-1.*Up"; then
    echo "❌ DevContainer is not running"
    echo "💡 Run ./scripts/go-build.sh first to start services"
    exit 1
fi

# Check and start Redis service
echo "📡 Checking Redis service..."
if ! docker compose ps | grep -q "devcontainer-redis-1.*Up"; then
    echo "🔄 Starting Redis service..."
    docker compose up -d redis
    sleep 2
fi
cd ..

# Check executable exists
if [ ! -f "run/ai-gateway-hub" ]; then
    echo "❌ Executable not found"
    echo "💡 Run ./scripts/go-build.sh first to build the application"
    exit 1
fi

# Check .env file
if [ ! -f "run/.env" ]; then
    echo "📝 Creating .env file..."
    cp .env.example run/.env
fi

echo "🎯 Application starting on http://localhost:8080"
echo "🛑 Press Ctrl+C to stop"
echo ""

# Run application in DevContainer (interactive mode)
# Set Redis connection to DevContainer's redis service
docker exec -it -w /workspace/run -e REDIS_ADDR=redis:6379 devcontainer-app-1 ./ai-gateway-hub