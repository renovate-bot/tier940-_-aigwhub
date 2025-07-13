#!/bin/bash

# Release script for AI Gateway Hub
set -e

VERSION=${1:-v0.1.0}

echo "🚀 Creating release $VERSION"

# Validate version format
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "❌ Invalid version format. Use semantic versioning like v0.1.0"
    exit 1
fi

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "❌ Not a git repository"
    exit 1
fi

# Check if working directory is clean
if [[ -n $(git status --porcelain) ]]; then
    echo "❌ Working directory is not clean. Please commit or stash changes."
    git status --short
    exit 1
fi

# Check if tag already exists
if git tag -l | grep -q "^$VERSION$"; then
    echo "❌ Tag $VERSION already exists"
    exit 1
fi

# Fetch latest changes
echo "📡 Fetching latest changes..."
git fetch origin

# Check if current branch is up to date
LOCAL=$(git rev-parse HEAD)
REMOTE=$(git rev-parse origin/$(git branch --show-current))

if [[ $LOCAL != $REMOTE ]]; then
    echo "❌ Local branch is not up to date with remote. Please pull latest changes."
    exit 1
fi

# Run tests
echo "🧪 Running tests..."
if command -v ./scripts/go-test.sh &> /dev/null; then
    ./scripts/go-test.sh unit
else
    echo "⚠️  Test script not found, skipping tests"
fi

# Build to ensure everything compiles
echo "🔨 Building application..."
if command -v ./scripts/go-build.sh &> /dev/null; then
    ./scripts/go-build.sh
else
    go build -o ai-gateway-hub ./main.go
fi

# Create and push tag
echo "🏷️  Creating tag $VERSION..."
git tag -a $VERSION -m "Release $VERSION

🚀 AI Gateway Hub $VERSION

✨ What's New:
- Fixed chat interface and resolved template conflicts
- Implemented unified logging system with industry-standard libraries  
- Enhanced WebSocket communication reliability
- Improved Claude CLI integration with proper authentication

🔧 Technical Improvements:
- Migrated to logrus for standardized logging
- Replaced custom configuration with viper
- Enhanced error handling and debugging
- Improved DevContainer setup

🤖 Generated with Claude Code
"

echo "📤 Pushing tag to origin..."
git push origin $VERSION

echo "✅ Release $VERSION created successfully!"
echo "🔗 GitHub Actions will now build and create the release automatically"
echo "📦 Check the progress at: https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[:/]\([^/]*\/[^/]*\).*/\1/' | sed 's/\.git$//')/actions"

# Wait a moment and check if workflow started
sleep 3
echo "🎯 Workflow should start shortly. You can monitor the release process in GitHub Actions."