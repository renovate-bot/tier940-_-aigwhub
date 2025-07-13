#!/bin/bash

# このスクリプトはDevContainer内で実行されることを前提としています

set -e

echo "=== Building AI Gateway Hub in DevContainer ==="

# 作業ディレクトリに移動
cd /workspace

# Go依存関係をダウンロード
echo "Downloading Go dependencies..."
go mod download

# ビルド
echo "Building application..."
go build -o ai-gateway-hub .

# 実行用ディレクトリを作成
echo "Setting up run directory..."
rm -rf run
mkdir -p run/data run/logs

# 実行ファイルと設定ファイルをコピー
cp ai-gateway-hub run/
cp .env.example run/.env

echo "=== Build completed ==="
echo "The standalone executable is in the 'run' directory"
echo "To run the application:"
echo "  1. Start Redis: docker compose -f .devcontainer/compose.yml up -d redis"
echo "  2. cd run && ./ai-gateway-hub"