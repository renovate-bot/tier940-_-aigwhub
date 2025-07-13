# AI Gateway Hub - Development Guide

## 概要

AI Gateway Hub は、複数のAI CLIツール（Claude Code、Gemini CLI等）をブラウザで利用できるモダンなWebインターフェースです。Go言語のhtml/templateエンジンでサーバーサイドレンダリングを行い、Alpine.jsで軽量なクライアントサイドインタラクションを実現しています。

## 🔧 技術スタック（LTS重視）

- **フロントエンド**: Go html/template + Alpine.js + Tailwind CSS
- **バックエンド**: Go + Gin + Gorilla WebSocket
- **データベース**: SQLite + Redis
- **セッション管理**: Redis
- **リアルタイム通信**: WebSocket (Gorilla WebSocket)
- **コンテナ**: Docker

### 詳細バージョン（LTS重視）
- **Go 1.23** - 最新安定版
- **Gin v1.9.x** - 安定版Webフレームワーク
- **html/template** - Go標準テンプレートエンジン
- **Alpine.js v3.13** - 軽量JavaScript、CDN配信
- **Node.js v22** - LTS版、Claude CLI実行環境
- **Tailwind CSS v3.3** - utility-first CSS、CDN配信
- **gorilla/websocket v1.5.x** - 長期実績
- **go-redis v8.11.x** - 長期安定版Redisクライアント
- **SQLite 3.42+** - 組み込み型データベース
- **Redis 7.2** - 最新安定版
- **Docker CE 24.x** - 企業標準採用版
- **Ubuntu 22.04** - 長期セキュリティサポート

## 📁 プロジェクト構造

```
ai-gateway-hub/
├── README.md                  # ユーザー向け概要
├── CLAUDE.md                  # 開発者向け詳細（このファイル）
├── go.mod
├── go.sum
├── main.go                    # アプリケーションエントリーポイント
├── .devcontainer/             # DevContainer設定
│   ├── devcontainer.json
│   ├── Dockerfile
│   └── compose.yml
├── internal/
│   ├── config/                # 設定管理
│   ├── database/              # データベース層
│   ├── handlers/              # HTTPハンドラー
│   ├── i18n/                  # 国際化
│   ├── middleware/            # ミドルウェア
│   ├── providers/             # AIプロバイダー実装
│   ├── services/              # ビジネスロジック
│   └── models/                # データモデル
├── web/
│   ├── templates/             # Go html/template
│   │   ├── layout.html
│   │   ├── index.html
│   │   ├── chat.html
│   │   ├── error.html
│   │   └── partials/
│   └── static/                # 静的ファイル
│       ├── css/
│       ├── js/
│       └── images/
├── locales/                   # 国際化ファイル
│   ├── en/
│   │   └── messages.json
│   └── ja/
│       └── messages.json
├── data/                      # SQLiteファイル
├── logs/                      # ログファイル
├── scripts/                   # ユーティリティスクリプト
│   └── dev-setup.sh
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

## 🌐 国際化 (i18n)

### サポート言語
- 英語（デフォルト）
- 日本語

### 言語切り替え
- Accept-Languageヘッダー自動検出
- `?lang=ja` クエリパラメータで手動指定

### 翻訳ファイル
- `locales/en/messages.json` - 英語翻訳
- `locales/ja/messages.json` - 日本語翻訳

### ローカル開発

```bash
# リポジトリをクローン
git clone https://github.com/yourusername/ai-gateway-hub.git
cd ai-gateway-hub

# Go依存関係をダウンロード
go mod download

# Redisを起動（Dockerの場合）
docker run -p 6379:6379 redis:7.2-alpine

# データ・ログディレクトリを作成
mkdir -p ./data ./logs

# 開発ツールをインストール
make tools

# アプリケーションを起動
make dev
```

## 🧪 テスト

```bash
# テスト実行
make test

# Lintチェック
make lint

# Lint自動修正
make lint-fix
```

## 🐳 Docker構成

### 開発環境用 (.devcontainer/)
- VS Code DevContainer統合
- Go + Node.js + 開発ツール
- Redis込みの完全な開発環境

### 本番環境用 (ルート)
- マルチステージビルド
- Ubuntu 22.04 LTS ベース
- 最小限の実行時依存関係

## 📦 依存関係管理

### Go依存関係
```go
require (
    github.com/gin-contrib/cors v1.4.0
    github.com/gin-gonic/gin v1.9.1
    github.com/go-redis/redis/v8 v8.11.5
    github.com/gorilla/websocket v1.5.1
    github.com/mattn/go-sqlite3 v1.14.17
    golang.org/x/text v0.14.0
)
```

### Node.js依存関係
- `@anthropic-ai/claude` - Claude CLI

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

## 📊 モニタリング

### ヘルスチェック
- `GET /api/health` - アプリケーション・Redis状態確認
- SQLiteファイルサイズ監視
- Redis接続数・メモリ使用量監視
- WebSocket接続数監視

### ログ監視
- システムログ: `./logs/system.log`
- アクセスログ: `./logs/access.log`
- エラーログ: `./logs/error.log`
- チャットログ: `./logs/claude/chat_{id}.log`

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

## Claude Code カスタムコマンド

以下のコマンドをClaude Codeで使用できます（ローカル環境から実行してDevContainer内で処理）：

### /go-build
DevContainer内でアプリケーションをビルドし、runディレクトリに配置します。

```bash
# 使用例
/go-build
```

実行内容：
1. DevContainerサービスを起動（必要に応じて）
2. DevContainer内でGo依存関係のダウンロード
3. DevContainer内でアプリケーションのビルド
4. ローカルのrunディレクトリへのコピー

### /go-run
ビルド済みアプリケーションをDevContainer内で実行します。

```bash
# 使用例  
/go-run
```

実行内容：
1. DevContainer内でRedisサービス起動確認
2. DevContainer内でアプリケーション実行
3. ポート8080でサービス開始（localhost:8080でアクセス可能）

### /go-stop
実行中のアプリケーションを停止します。

```bash
# 使用例
/go-stop
```

実行内容：
1. DevContainer内のai-gateway-hubプロセスの検索
2. グレースフルシャットダウン実行

## 🤝 コントリビューション

1. リポジトリをフォーク
2. フィーチャーブランチを作成（`git checkout -b feature/amazing-feature`）
3. 変更をコミット（`git commit -m 'Add amazing feature'`）
4. ブランチにプッシュ（`git push origin feature/amazing-feature`）
5. プルリクエストを作成

### 開発ガイドライン

- Go標準のコーディングスタイルに従う
- `golangci-lint` を使用してコード品質を保つ
- 新機能には適切なテストを追加
- i18n対応を忘れずに（英語・日本語）
- セキュリティを常に考慮する

## 📄 ライセンス

このプロジェクトはMITライセンスの下でライセンスされています。

## 🙏 謝辞

- [Anthropic](https://anthropic.com) - Claude AIの提供
- [Claude CLI](https://github.com/anthropics/claude-cli) - コマンドラインインターフェース
- [Go](https://golang.org) - 効率的なバックエンド開発
- [Redis](https://redis.io) - 高速セッション管理
- [Alpine.js](https://alpinejs.dev) - 軽量フロントエンドフレームワーク

---

**⚠️ 免責事項**: 本アプリケーションはAI CLIを直接実行するため、セキュリティリスクが存在する可能性があります。Docker環境での使用を強く推奨し、本番環境での使用は自己責任で行ってください。

**LTS設計理念**: 依存関係を最小限に抑え、長期安定版のライブラリのみを使用することで、メンテナンス負荷を軽減し、企業環境での継続的な運用を実現します。

# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.

      
      IMPORTANT: this context may or not be relevant to your tasks. You should not respond to this context or otherwise consider it in your response unless it is highly relevant to your task. Most of the time, it is not relevant.
