# テストスイート

このディレクトリには AI Gateway Hub のテストが含まれています。

## 📁 構造

```
test/
├── unit/                   # ユニットテスト
│   ├── utils_paths_test.go    # パス管理のテスト
│   ├── utils_common_test.go   # 共通ユーティリティのテスト
│   ├── config_test.go         # 設定管理のテスト
│   └── database_test.go       # データベースのテスト
├── integration/            # インテグレーションテスト
│   └── provider_test.go       # プロバイダーのテスト
├── e2e/                   # E2Eテスト
│   └── api_test.go           # API全体のテスト
├── Makefile               # テスト実行用Makefile
└── README.md              # このファイル
```

## 🚀 テスト実行方法

### スクリプト経由

```bash
# すべてのテストを実行
./scripts/go-test.sh

# ユニットテストのみ
./scripts/go-test.sh unit

# インテグレーションテストのみ
./scripts/go-test.sh integration

# E2Eテストのみ
./scripts/go-test.sh e2e

# カバレッジ付きでテスト実行
./scripts/go-test.sh coverage

# テストアーティファクトの削除
./scripts/go-test.sh clean
```

### Makefile経由

```bash
cd test/

# すべてのテストを実行
make

# 個別実行
make unit
make integration
make e2e
make coverage
make clean

# ヘルプ表示
make help
```

### 直接実行

```bash
# ユニットテスト
go test -v ./test/unit/...

# インテグレーションテスト
go test -v ./test/integration/...

# E2Eテスト
go test -v ./test/e2e/...

# カバレッジ付き
go test -v -coverprofile=coverage.out ./test/unit/... ./test/integration/...
go tool cover -html=coverage.out -o coverage.html
```

## 📋 テストカテゴリ

### ユニットテスト (`unit/`)

個別のコンポーネントやユーティリティ関数をテストします。

- **utils_paths_test.go**: パス管理機能
  - ディレクトリ作成
  - パス解決
  - ファイル用ディレクトリ作成

- **utils_common_test.go**: 共通ユーティリティ
  - 環境変数読み込み
  - JSON操作
  - エラーハンドリング
  - ファイル操作

- **config_test.go**: 設定管理
  - デフォルト値の確認
  - 環境変数からの読み込み
  - 不正な値の処理

- **database_test.go**: データベース操作
  - SQLite初期化
  - テーブル作成
  - Redis接続

### インテグレーションテスト (`integration/`)

複数のコンポーネント間の連携をテストします。

- **provider_test.go**: プロバイダー機能
  - Claude プロバイダーの動作
  - ログファイル作成
  - プロバイダーレジストリ

### E2Eテスト (`e2e/`)

アプリケーション全体の動作をテストします。

- **api_test.go**: API全体のテスト
  - HTTP APIエンドポイント
  - チャット作成・削除
  - エラーハンドリング
  - CORS設定

## 🔧 テスト環境

### 前提条件

- DevContainer が起動していること
- Redis サービスが利用可能であること（オプション）
- Claude CLI がインストールされていること（一部テスト用）

### テスト用データベース

テストは一時的なSQLiteデータベースを使用し、テスト完了後に自動的に削除されます。

### モック・スタブ

- Claude CLIは実際のAPIを呼び出さず、`echo`コマンドなどで代替
- Redis接続エラーは無視され、テストが継続される
- 一時ディレクトリを使用してファイルシステムの汚染を防止

## 📊 カバレッジ

カバレッジレポートは `coverage.html` として生成されます：

```bash
./scripts/go-test.sh coverage
open coverage.html  # ブラウザで開く
```

## 💡 テスト作成ガイドライン

### ユニットテスト

1. 一時ディレクトリを使用
2. テスト間の独立性を保つ
3. 外部依存を最小限に抑制
4. エラーケースも含めてテスト

### インテグレーションテスト

1. 実際のファイルシステム操作を含む
2. 複数コンポーネントの協調をテスト
3. 設定可能な外部依存を使用

### E2Eテスト

1. 完全なHTTPサーバーを起動
2. 実際のAPIリクエストを送信
3. レスポンス形式と内容を検証
4. エラーハンドリングを確認

## 🐛 トラブルシューティング

### テスト失敗時の対処

1. **DevContainer未起動**: `./scripts/go-build.sh` を実行
2. **権限エラー**: スクリプトが権限修正を自動実行
3. **Redis接続エラー**: 一部テストはRedis無しでも動作
4. **Claude CLI認証エラー**: モックを使用するため影響なし

### 個別テスト実行

```bash
# 特定のテスト関数のみ実行
go test -v -run TestPathManager ./test/unit/

# 特定のテストファイルのみ実行
go test -v ./test/unit/utils_paths_test.go
```