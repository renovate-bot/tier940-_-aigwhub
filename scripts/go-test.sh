#!/bin/bash

# Test Go application in DevContainer
# Usage: ./scripts/go-test.sh [test-type]

set -e

echo "🧪 Running Go tests in DevContainer..."

# Check DevContainer services
echo "📦 Checking DevContainer services..."
cd .devcontainer
if ! docker compose ps | grep -q "devcontainer-app-1.*Up"; then
    echo "❌ DevContainer is not running"
    echo "💡 Run ./scripts/go-build.sh first to start services"
    exit 1
fi
cd ..

# Fix Go module permissions
echo "🔧 Fixing Go module permissions..."
docker exec -u root devcontainer-app-1 chown -R vscode:vscode /go || true

# Run tests based on argument
TEST_TYPE=${1:-"all"}

case $TEST_TYPE in
    "unit")
        echo "🔬 Running unit tests..."
        docker exec -w /workspace devcontainer-app-1 go test -v ./test/unit/...
        ;;
    "integration")
        echo "🔗 Running integration tests..."
        docker exec -w /workspace devcontainer-app-1 go test -v ./test/integration/...
        ;;
    "e2e")
        echo "🌐 Running E2E tests..."
        docker exec -w /workspace devcontainer-app-1 go test -v ./test/e2e/...
        ;;
    "coverage")
        echo "📊 Running tests with coverage..."
        docker exec -w /workspace devcontainer-app-1 go test -v -coverprofile=coverage.out ./test/unit/... ./test/integration/...
        docker exec -w /workspace devcontainer-app-1 go tool cover -html=coverage.out -o coverage.html
        echo "📈 Coverage report generated: coverage.html"
        ;;
    "all")
        echo "🚀 Running all tests..."
        echo ""
        echo "1️⃣ Unit Tests:"
        docker exec -w /workspace devcontainer-app-1 go test -v ./test/unit/...
        echo ""
        echo "2️⃣ Integration Tests:"
        docker exec -w /workspace devcontainer-app-1 go test -v ./test/integration/...
        echo ""
        echo "3️⃣ E2E Tests:"
        docker exec -w /workspace devcontainer-app-1 go test -v ./test/e2e/...
        ;;
    "clean")
        echo "🧹 Cleaning test artifacts..."
        docker exec -w /workspace devcontainer-app-1 rm -f coverage.out coverage.html
        docker exec -w /workspace devcontainer-app-1 go clean -testcache
        echo "✅ Test artifacts cleaned"
        ;;
    *)
        echo "❌ Invalid test type: $TEST_TYPE"
        echo "Valid options: unit, integration, e2e, coverage, all, clean"
        exit 1
        ;;
esac

echo ""
echo "✅ Tests completed successfully!"
echo ""
echo "📋 Available test commands:"
echo "  ./scripts/go-test.sh unit        - Run unit tests only"
echo "  ./scripts/go-test.sh integration - Run integration tests only"
echo "  ./scripts/go-test.sh e2e         - Run E2E tests only"
echo "  ./scripts/go-test.sh coverage    - Run tests with coverage report"
echo "  ./scripts/go-test.sh all         - Run all tests (default)"
echo "  ./scripts/go-test.sh clean       - Clean test artifacts"