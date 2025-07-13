#!/bin/bash

# このスクリプトはDevContainer内で実行されることを前提としています

set -e

echo "=== Running AI Gateway Hub in DevContainer ==="

# 作業ディレクトリに移動
cd /workspace

# Redisが起動しているか確認
if ! docker ps | grep -q redis; then
    echo "Starting Redis..."
    cd .devcontainer
    docker compose up -d redis
    cd ..
    sleep 2
fi

# アプリケーションを実行
echo "Starting AI Gateway Hub..."
export REDIS_ADDR=redis:6379
./ai-gateway-hub