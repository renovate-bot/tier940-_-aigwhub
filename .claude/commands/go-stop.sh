#!/bin/bash

# Stop running Go application in DevContainer
# Usage: /go-stop

set -e

echo "🛑 Stopping running Go application in DevContainer..."

# Check DevContainer services
if ! docker compose -f .devcontainer/compose.yml ps | grep -q "devcontainer-app-1.*Up"; then
    echo "❌ DevContainer is not running"
    exit 1
fi

# Search for ai-gateway-hub processes in DevContainer
echo "🔍 Searching for ai-gateway-hub processes..."
PIDS=$(docker exec devcontainer-app-1 pgrep -f "ai-gateway-hub" 2>/dev/null || true)

if [ -z "$PIDS" ]; then
    echo "ℹ️ No running ai-gateway-hub processes found"
    echo "✅ Application may already be stopped"
else
    echo "📍 Found processes: $PIDS"
    echo "🔄 Executing graceful shutdown..."
    
    # Graceful shutdown (SIGTERM)
    for PID in $PIDS; do
        echo "  - Stopping process $PID..."
        docker exec devcontainer-app-1 kill -TERM "$PID" 2>/dev/null || true
    done
    
    # Wait and check if processes terminated
    sleep 3
    
    # Force kill if still running
    REMAINING_PIDS=$(docker exec devcontainer-app-1 pgrep -f "ai-gateway-hub" 2>/dev/null || true)
    if [ -n "$REMAINING_PIDS" ]; then
        echo "⚠️ Graceful shutdown incomplete. Force killing remaining processes..."
        for PID in $REMAINING_PIDS; do
            echo "  - Force killing process $PID..."
            docker exec devcontainer-app-1 kill -KILL "$PID" 2>/dev/null || true
        done
    fi
    
    echo "✅ ai-gateway-hub processes stopped"
fi

# Check port 8080 usage
echo "🔍 Checking port 8080 usage..."
PORT_CHECK=$(docker exec devcontainer-app-1 netstat -tulpn 2>/dev/null | grep ":8080 " || true)
if [ -z "$PORT_CHECK" ]; then
    echo "✅ Port 8080 is free"
else
    echo "⚠️ Port 8080 still in use:"
    echo "$PORT_CHECK"
fi

echo ""
echo "🎯 Next steps:"
echo "  - Use /go-build to rebuild"
echo "  - Use /go-run to restart"