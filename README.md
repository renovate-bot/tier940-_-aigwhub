# AI Gateway Hub

複数のAI CLIツール（Claude Code、Gemini CLI等）をブラウザで利用できるモダンなWebインターフェース。Go言語のhtml/templateエンジンでサーバーサイドレンダリングを行い、Alpine.jsで軽量なクライアントサイドインタラクションを実現。

## ✨ AI Gateway Hubの利点

- 🌐 **ブラウザアクセス**: どのデバイスからでもWebブラウザで複数のAI CLIツールを利用
- 🚀 **インストール不要**: ターミナル設定をスキップして即座にAIツールにアクセス
- 👥 **チーム協力**: セッションやコーディング議論を簡単に共有
- 🔧 **マルチプロバイダー対応**: Claude Code、Gemini CLI、将来のAIツールを一つのインターフェースで
- 🔄 **セッション永続化**: Redisベースでコーディングコンテキストを絶対に失わない
- 📝 **完全履歴**: 全ての開発セッションを保存・エクスポート可能

## 🎯 最適な利用シーン

- **開発者**: ターミナルの複雑さなしにClaude Codeを使いたい
- **チーム**: AIコーディング支援への共有アクセスが必要
- **リモート開発**: ブラウザアクセスが推奨される環境
- **学習・教育**: AI支援によるコーディング学習
- **コードレビュー**: AI駆動のインサイト付きレビュー

## ✨ 特徴

- 🚀 **高速開発**: 瞬時の開発フィードバック
- 💬 **リアルタイムコーディング**: WebSocketによる即座のAI応答
- 🛡️ **セッション永続化**: Redisベースのクラッシュ耐性セッション管理
- 📁 **ファイル管理**: Webインターフェースでプロジェクトファイルのアップロード、編集、管理
- 🎨 **モダンUI**: コーディング最適化されたダーク/ライトテーマ対応レスポンシブデザイン
- 🔄 **自動再接続**: 堅牢な接続処理と自動復旧機能

## 🔧 技術スタック（LTS重視）

- **フロントエンド**: Go html/template + Alpine.js + Tailwind CSS
- **バックエンド**: Go + Gin + Gorilla WebSocket
- **データベース**: SQLite + Redis
- **セッション管理**: Redis
- **リアルタイム通信**: WebSocket (Gorilla WebSocket)
- **コンテナ**: Docker

### 詳細バージョン（LTS重視）
- **Go 1.23** - 最新安定版（2025年7月現在）、Go 1.24も利用可能
- **Gin v1.9.x** - 安定版Webフレームワーク、破壊的変更が少ない
- **html/template** - Go標準テンプレートエンジン
- **Alpine.js v3.13** - 軽量JavaScript、CDN配信
- **Tailwind CSS v3.3** - utility-first CSS、CDN配信
- **gorilla/websocket v1.5.x** - 長期実績、メンテナンス継続中
- **go-redis v8.11.x** - 長期安定版Redisクライアント
- **SQLite 3.42+** - 組み込み型データベース、運用コスト最小
- **Redis 7.2** - 最新安定版（2025年7月現在）
- **Docker CE 24.x** - 企業標準採用版
- **Alpine Linux 3.18** - 長期セキュリティサポート

## 🚀 クイックスタート

### 前提条件

- Claude CLIがインストール済みで認証完了（`claude auth`）
- Go 1.23+
- Redis 7.2+（Dockerでも可）
- Docker（本番デプロイ用）

### ローカル開発

```bash
# リポジトリをクローン
git clone https://github.com/yourusername/ai-gateway-hub.git
cd ai-gateway-hub

# Go依存関係をダウンロード
go mod download

# Redisを起動（Dockerの場合）
docker run -d -p 6379:6379 redis:7.2-alpine

# データ・ログディレクトリを作成
mkdir -p ./data ./logs
sudo mkdir -p /var/log/ai-gateway/claude /var/log/ai-gateway/gemini
sudo chown -R $USER:$USER /var/log/ai-gateway

# アプリケーションを起動
go run main.go
```

ブラウザで `http://localhost:8080` にアクセス

### Dockerデプロイ

```bash
# 先にClaude CLIを認証
claude auth

# ログディレクトリを作成
mkdir -p ./logs ./data
chmod 755 ./logs ./data

# Docker Composeでビルド・実行
compose up -d

# ヘルスチェック
curl http://localhost:8080/api/health
```

## 📁 プロジェクト構造

```
ai-gateway-hub/
├── README.md
├── go.mod
├── go.sum
├── main.go                     # アプリケーションエントリーポイント
├── compose.yml
├── Dockerfile
├── internal/
│   ├── config/                 # 設定管理
│   ├── database/              # データベース層
│   ├── handlers/              # HTTPハンドラー
│   ├── providers/             # AIプロバイダー実装
│   ├── services/              # ビジネスロジック
│   └── models/                # データモデル
├── web/
│   ├── templates/             # Go html/template
│   │   ├── layout.html
│   │   ├── index.html
│   │   ├── chat.html
│   │   └── partials/
│   └── static/                # 静的ファイル
│       ├── css/
│       ├── js/
│       └── images/
├── data/                      # SQLiteファイル
├── logs/                      # ログファイル
└── docs/                      # ドキュメント
```

## 🏗️ アーキテクチャ

### システム構成
```
[Webブラウザ]
    ↓ HTTP/WebSocket
[Go Webサーバー (Gin)]
    ↓ html/template + Alpine.js
[WebSocketHub] ←→ [AIProvider Registry]
    ↓                    ↓
[Redis Sessions]    [Claude CLI実行]
    ↓
[SQLite Metadata]
```

### コンポーネント
1. **Go html/templateフロントエンド**
   - サーバーサイドHTMLレンダリング
   - Alpine.jsで軽量クライアントサイドインタラクション
   - Tailwind CSSでスタイリング（CDN配信）
   - WebSocketでリアルタイム通信

2. **Goバックエンド（プラガブル設計）**
   - Gin WebフレームワークでHTTP API
   - Gorilla WebSocketでリアルタイム通信
   - AIProvider抽象化層（Interface-based）
   - Redis セッション管理
   - SQLite メタデータ永続化

3. **AIProvider Plugin System**
   - Claude CLI Provider（初期実装）
   - Gemini CLI Provider（将来実装）
   - 共通インターフェース準拠
   - プラガブル認証システム

4. **データ層**
   - SQLite: メタデータ + チャット履歴
   - Redis: アクティブセッション + WebSocket管理
   - ログファイル: 完全な実行履歴（Provider別）

## 🔧 設定

### 環境変数

```bash
# サーバー設定
PORT=8080
SQLITE_DB_PATH=./data/ai_gateway.db
REDIS_ADDR=localhost:6379
STATIC_DIR=./web/static
TEMPLATE_DIR=./web/templates

# ログ設定
LOG_DIR=./logs
LOG_LEVEL=info

# セッション管理
MAX_SESSIONS=100
SESSION_TIMEOUT=3600
WEBSOCKET_TIMEOUT=7200

# AIプロバイダー設定
CLAUDE_CLI_PATH=claude
GEMINI_CLI_PATH=gemini

# 機能フラグ
ENABLE_PROVIDER_AUTO_DISCOVERY=true
ENABLE_HEALTH_CHECKS=true
```

## 📡 API エンドポイント

### HTTP API
```
GET  /                    # メインページ
GET  /chat/:id           # チャットページ
GET  /api/chats          # チャット一覧
POST /api/chats          # 新規チャット作成
DELETE /api/chats/:id    # チャット削除
GET  /api/providers      # 利用可能プロバイダー一覧
GET  /api/health         # ヘルスチェック
```

### WebSocket
```
/ws                      # WebSocket接続
```

### WebSocketメッセージフォーマット
```json
{
  "type": "ai_prompt|ai_response|session_status|error",
  "data": {
    "chat_id": 123,
    "provider": "claude",
    "content": "メッセージ内容",
    "timestamp": "2025-07-12T10:30:00Z",
    "stream": true
  }
}
```

## 🔒 セキュリティ考慮事項

⚠️ **重要なセキュリティ通知**

本アプリケーションはAI CLIコマンドを直接実行するため、セキュリティリスクが存在します。

- 適切なセキュリティ対策なしに**本番環境での使用は推奨されません**
- **Docker隔離を強く推奨**
- 適切な入力検証とレート制限の実装が必要

### 推奨セキュリティ対策

1. **Docker環境実行**
   - 制限された権限で隔離環境で実行
   - ネットワークアクセス制限
   - ボリュームマウント最小化

2. **入力検証**
   - プロンプト長制限
   - 特殊文字のエスケープ
   - コマンドインジェクション防止

3. **リソース制限**
   - AI CLI実行タイムアウト設定
   - 同時セッション数制限
   - Redis TTL設定

## 📦 依存関係（go.mod）

```go
module ai-gateway-hub

go 1.23

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/gin-contrib/cors v1.4.0
    github.com/go-redis/redis/v8 v8.11.5
    github.com/gorilla/websocket v1.5.1
    github.com/mattn/go-sqlite3 v1.14.17
)
```

## 🐳 Docker構成

### Dockerfile
```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:3.18
RUN apk add --no-cache ca-certificates nodejs npm
RUN npm install -g @anthropic-ai/claude-cli
COPY --from=builder /app/main /app/
COPY --from=builder /app/web /app/web/
WORKDIR /app
EXPOSE 8080
CMD ["./main"]
```

### compose.yml
```yaml
services:
  app:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - redis
    volumes:
      - ./data:/app/data
      - ./logs:/app/logs
      - ./claude-config:/root/.claude
    environment:
      REDIS_ADDR: redis:6379

  redis:
    image: redis:7.2-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

volumes:
  redis_data:
```

## 🚀 デプロイメント

### ローカル開発
```bash
go mod download
docker run -d -p 6379:6379 redis:7.2-alpine
mkdir -p ./data ./logs
go run main.go
```

### 本番デプロイ
```bash
claude auth
mkdir -p ./data ./logs
chmod 755 ./data ./logs
compose up -d
curl http://localhost:8080/api/health
```

## 📊 モニタリング

### ヘルスチェック
- `GET /api/health` - アプリケーション・Redis状態確認
- SQLiteファイルサイズ監視
- Redis接続数・メモリ使用量監視
- WebSocket接続数監視

### ログ監視
- システムログ: `/var/log/ai-gateway/system.log`
- アクセスログ: `/var/log/ai-gateway/access.log`
- エラーログ: `/var/log/ai-gateway/error.log`
- チャットログ: `/var/log/ai-gateway/{provider}/chat_{id}.log`

## 🛣️ 今後の拡張予定

### 機能拡張
- [ ] ファイルアップロード対応
- [ ] マルチユーザー認証
- [ ] チャット共有機能
- [ ] ログエクスポート機能
- [ ] セッション復旧機能
- [ ] 全文検索機能

### 技術改善
- [ ] Gemini CLI Provider実装
- [ ] Redis Cluster対応
- [ ] Prometheus メトリクス
- [ ] ログローテーション
- [ ] レート制限実装
- [ ] WebAssembly Plugin対応

## 🤝 コントリビューション

1. リポジトリをフォーク
2. フィーチャーブランチを作成（`git checkout -b feature/amazing-feature`）
3. 変更をコミット（`git commit -m 'Add amazing feature'`）
4. ブランチにプッシュ（`git push origin feature/amazing-feature`）
5. プルリクエストを作成

## 📄 ライセンス

このプロジェクトはMITライセンスの下でライセンスされています。

## 🙏 謝辞

- [Anthropic](https://anthropic.com) - Claude AIの提供
- [Claude CLI](https://github.com/anthropics/claude-cli) - コマンドラインインターフェース
- [Go](https://golang.org) - 効率的なバックエンド開発
- [Redis](https://redis.io) - 高速セッション管理
- [Alpine.js](https://alpinejs.dev) - 軽量フロントエンドフレームワーク

## 📞 サポート

- 📚 [ドキュメント](docs/)
- 🐛 [バグ報告](https://github.com/yourusername/ai-gateway-hub/issues)
- 💡 [機能リクエスト](https://github.com/yourusername/ai-gateway-hub/issues)
- 💬 [ディスカッション](https://github.com/yourusername/ai-gateway-hub/discussions)

---

**⚠️ 免責事項**: 本アプリケーションはAI CLIを直接実行するため、セキュリティリスクが存在する可能性があります。Docker環境での使用を強く推奨し、本番環境での使用は自己責任で行ってください。

**LTS設計理念**: 依存関係を最小限に抑え、長期安定版のライブラリのみを使用することで、メンテナンス負荷を軽減し、企業環境での継続的な運用を実現します。
