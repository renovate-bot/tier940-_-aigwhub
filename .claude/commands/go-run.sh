#!/bin/bash

# Run Go application in DevContainer
# Usage: /go-run

set -e

echo "🚀 Running Go application in DevContainer..."

# Check DevContainer services
echo "📦 Checking DevContainer services..."
cd .devcontainer
if ! docker compose ps | grep -q "devcontainer-app-1.*Up"; then
    echo "❌ DevContainer is not running"
    echo "💡 Run /go-build first to start services"
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
    echo "💡 Run /go-build first to build the application"
    exit 1
fi

# Check .env file
if [ ! -f "run/.env" ]; then
    echo "📝 Creating .env file..."
    cp .env.example run/.env
fi

echo "🎯 Starting application..."
echo "📍 URL: http://localhost:8080"
echo "🛑 Stop with /go-stop or press Ctrl+C"
echo ""

# Run application in DevContainer as daemon
# Set Redis connection to DevContainer's redis service
docker exec -d -w /workspace/run -e REDIS_ADDR=redis:6379 devcontainer-app-1 ./ai-gateway-hub

# Wait a bit and check if it started
sleep 2
if docker exec devcontainer-app-1 pgrep -x ai-gateway-hub > /dev/null; then
    echo "✅ Application started successfully as daemon"
    echo "📍 Access at: http://localhost:8080"
    echo "📊 View logs: docker exec devcontainer-app-1 tail -f /workspace/logs/system.log"
else
    echo "❌ Failed to start application"
    echo "💡 Check logs: docker exec devcontainer-app-1 cat /workspace/logs/error.log"
    exit 1
fi