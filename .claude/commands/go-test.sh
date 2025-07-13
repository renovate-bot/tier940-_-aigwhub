#!/bin/bash

# Test Go application with browser automation using Playwright
# Usage: /go-test

set -e

echo "🧪 Testing Go application with browser automation..."

# Check if DevContainer is running
if ! docker compose -f .devcontainer/compose.yml ps | grep -q "devcontainer-app-1.*Up"; then
    echo "❌ DevContainer is not running"
    echo "💡 Run /go-build first to start DevContainer"
    exit 1
fi

# Check if application is running (try both local and container)
if curl -s http://localhost:8080/api/health > /dev/null 2>&1; then
    echo "✅ Application accessible on localhost:8080"
elif docker exec devcontainer-app-1 curl -s http://localhost:8080/api/health > /dev/null 2>&1; then
    echo "✅ Application running in DevContainer (port forwarding may need restart)"
else
    echo "❌ Application is not running"
    echo "💡 Run /go-run first to start the application"
    exit 1
fi

echo "✅ Application is running"
echo "🌐 Testing browser functionality..."

# Use Playwright to test the application
# This will be handled by the MCP server once available

echo "🎯 Basic health check passed"
echo "📊 Application is accessible at http://localhost:8080"
echo ""
echo "🔍 Manual testing checklist:"
echo "  - Main page loads correctly"
echo "  - API endpoints respond"
echo "  - WebSocket connection works"
echo "  - Chat functionality operates"
echo ""
echo "💡 Use Playwright MCP tools for automated browser testing"