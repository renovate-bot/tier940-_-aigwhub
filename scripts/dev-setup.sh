#!/bin/bash
set -e

echo "🚀 Setting up AI Gateway Hub development environment..."

# Create necessary directories
echo "📁 Creating directories..."
mkdir -p data logs

# Copy environment file if it doesn't exist
if [ ! -f .env ]; then
    echo "📋 Creating .env file from example..."
    if [ -f .env.example ]; then
        cp .env.example .env
    else
        echo "⚠️  Warning: .env.example not found"
    fi
fi

# Install Go dependencies
echo "📦 Installing Go dependencies..."
go mod download

# Install development tools
echo "🔧 Installing development tools..."
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install Claude CLI
echo "🤖 Installing Claude CLI..."
npm install -g @anthropic-ai/claude

echo "✅ Development environment setup complete!"
echo ""
echo "To start the application, run:"
echo "  make dev"
echo ""
echo "Or directly:"
echo "  go run main.go"
