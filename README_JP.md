# AI Gateway Hub

複数のAI CLIツール（Claude Code、Gemini CLI等）をブラウザで利用できるモダンなWebインターフェース。

> **注意**: すべての静的リソース（HTMLテンプレート、国際化ファイル）は実行ファイルに埋め込まれており、単体実行が可能です。

## ✨ 特徴

- 🌐 **ブラウザアクセス**: どのデバイスからでもWebブラウザで複数のAI CLIツールを利用
- 🚀 **インストール不要**: ターミナル設定をスキップして即座にAIツールにアクセス  
- 👥 **チーム協力**: セッションやコーディング議論を簡単に共有
- 🔧 **マルチプロバイダー対応**: Claude Code、Gemini CLI、将来のAIツールを一つのインターフェースで
- 🔄 **セッション永続化**: Redisベースでコーディングコンテキストを絶対に失わない
- 💬 **リアルタイムコーディング**: WebSocketによる即座のAI応答

## 🚀 クイックスタート

### 前提条件

1. **Claude CLI**: インストールと認証
   ```bash
   # Claude CLIをインストール
   npm install -g @anthropic-ai/claude-code
   
   # 認証（APIキーが必要）
   claude auth
   ```

2. **Redisサーバー**: セッション管理用
   ```bash
   # Docker使用（推奨）
   docker run -d -p 6379:6379 --name redis redis:7.2-alpine
   
   # またはローカルインストール（Ubuntu/Debian）
   sudo apt install redis-server
   sudo systemctl start redis-server
   
   # またはmacOSでHomebrew
   brew install redis
   brew services start redis
   ```

### インストール

#### 方法1: ビルド済みバイナリをダウンロード（推奨）

1. [リリースページ](https://github.com/tier940/aigwhub/releases)に移動
2. プラットフォームに適したバイナリをダウンロード：
   - **Linux x64**: `ai-gateway-hub-v*-linux-amd64.tar.gz`
   - **Linux ARM64**: `ai-gateway-hub-v*-linux-arm64.tar.gz`
   - **macOS x64**: `ai-gateway-hub-v*-darwin-amd64.tar.gz`
   - **macOS ARM64**: `ai-gateway-hub-v*-darwin-arm64.tar.gz`
   - **Windows x64**: `ai-gateway-hub-v*-windows-amd64.zip`

3. 展開して実行：
```bash
# Linux/macOS
tar -xzf ai-gateway-hub-v*-*.tar.gz
chmod +x ai-gateway-hub
./ai-gateway-hub

# Windows
# zipファイルを展開してai-gateway-hub.exeを実行
```

#### 方法2: ソースからビルド（開発用）

```bash
# runディレクトリに移動
cd run/

# .env.exampleから.envを作成
cp ../.env.example .env

# 必要に応じて.envを編集
# 特にREDIS_ADDRを適切に設定

# Redisを起動（Dockerを使用する場合）
docker run -p 6379:6379 redis:7.2-alpine

# アプリケーションを起動
./ai-gateway-hub

# ブラウザでアクセス
open http://localhost:8080
```

### 開発環境（DevContainer）

開発やビルドを行う場合：

```bash
# VS Code でプロジェクトを開く
code .

# DevContainer で再開（初回は自動でセットアップ）
# Ctrl+Shift+P → "Dev Containers: Reopen in Container"

# DevContainer内でビルド
./scripts/build-in-container.sh

# ビルド後、runディレクトリが作成される
# run/
# ├── ai-gateway-hub    # 実行ファイル（埋め込みリソース付き）
# └── .env             # 環境設定ファイル

# DevContainer内で実行
./scripts/run-in-container.sh
```

### ビルド成果物について

DevContainerでビルドすると、以下が生成されます：

- `ai-gateway-hub`: すべてのリソースが埋め込まれた単体実行ファイル
- `run/`: 配布用ディレクトリ
  - HTMLテンプレート、国際化ファイルはすべて実行ファイルに埋め込み済み
  - 外部依存はRedisのみ
  - `.env`ファイルで設定をカスタマイズ可能

## 🔒 セキュリティ注意事項

⚠️ **重要**: 本アプリケーションはAI CLIコマンドを直接実行するため、セキュリティリスクが存在します。

- **Docker環境での使用を強く推奨**
- 本番環境での使用は自己責任で行ってください
- 適切なネットワーク制限とアクセス制御を実装することを推奨

## 📚 ドキュメント

- **[CLAUDE.md](./CLAUDE.md)** - 開発者向け詳細技術仕様
- **API エンドポイント**: `/api/health` でヘルスチェック
- **WebSocket**: `/ws` でリアルタイム通信

## 🤝 コントリビューション

1. リポジトリをフォーク
2. フィーチャーブランチを作成（`git checkout -b feature/amazing-feature`）
3. 変更をコミット（`git commit -m 'Add amazing feature'`）
4. プルリクエストを作成

## 📄 ライセンス

MIT License

---

**⚠️ 免責事項**: 本アプリケーションはAI CLIを直接実行するため、セキュリティリスクが存在する可能性があります。Docker環境での使用を強く推奨し、本番環境での使用は自己責任で行ってください。