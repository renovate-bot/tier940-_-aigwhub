#!/bin/bash

# Run AI Gateway Hub in DevContainer
# Usage: ./scripts/run-in-container.sh (executed inside DevContainer)

set -e

echo "=== Starting AI Gateway Hub in DevContainer ==="

# 作業ディレクトリに移動
cd /workspace

# Redisサービスを確認・起動
if ! docker ps | grep -q redis; then
    echo "Starting Redis service..."
    cd .devcontainer
    docker compose up -d redis
    cd ..
    sleep 2
fi

# アプリケーションを実行
echo "Starting AI Gateway Hub on http://localhost:8080..."
echo "Press Ctrl+C to stop"
export REDIS_ADDR=redis:6379
./ai-gateway-hub