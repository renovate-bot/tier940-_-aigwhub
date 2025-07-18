# AI Gateway Hub - Development Guide

## 🧠 Triadic Development with Claude CLI and Gemini CLI
- AI Gateway Hub promotes a **triadic development principle** that combines:

- 🧍 **Human (Developer)** – Decision-maker: defines goals, initiates actions
- 🛠️ **Claude CLI** – Executor: handles task breakdown, implementation, file operations
- 🔍 **Gemini CLI** – Advisor: conducts web searches, API/library analysis, technical debugging

- This collaborative workflow aims to **maximize development speed and code quality** by leveraging the strengths of each role.

### 🔧 Claude CLI – Code Execution Engine
- Breaks down high-level instructions into actionable tasks
- Executes code generation, refactoring, and file handling
- Follows user instructions systematically but lacks context awareness

### 📚 Gemini CLI – Technical Research Specialist
- Investigates external documentation, error messages, and library behavior
- Provides current information via search (e.g., Google Search)
- Offers opinions, flags assumptions, and gives micro-level validation

### 🔄 Workflow Example

```plaintext
[You] → "Implement this feature using Redis Streams"
 ↓
[Claude CLI] → Generates initial implementation
 ↓
[Gemini CLI] → Validates API usage, suggests edge case handling
 ↓
[You] → Reviews and finalizes the output
```

## Overview
- AI Gateway Hub is a modern web interface for using multiple AI CLI tools (Claude Code, Gemini CLI, etc.) in a browser. It uses Go's html/template engine for server-side rendering and Alpine.js for lightweight client-side interactions.

## 🔧 Technology Stack
- **Frontend**: Go html/template + Alpine.js + Tailwind CSS
- **Backend**: Go + Gin + Gorilla WebSocket
- **Database**: SQLite + Redis
- **Session Management**: Redis
- **Real-time Communication**: WebSocket (Gorilla WebSocket)
- **Containers**: Docker

### Detailed Versions
- **Go 1.23** - Latest stable
- **Gin v1.9.x** - Stable web framework
- **html/template** - Go standard template engine
- **Alpine.js v3.13** - Lightweight JavaScript, CDN delivered
- **Node.js v22** - LTS version, Claude CLI runtime
- **Tailwind CSS v3.3** - Utility-first CSS, CDN delivered
- **gorilla/websocket v1.5.x** - Long-term proven
- **go-redis v8.11.x** - Long-term stable Redis client
- **SQLite 3.42+** - Embedded database
- **Redis 7.2** - Latest stable
- **Docker CE 24.x** - Enterprise standard
- **Ubuntu 22.04** - Long-term security support

## 📁 Project Structure

```
ai-gateway-hub/
├── README.md                  # User overview
├── CLAUDE.md                  # Developer details (this file)
├── go.mod
├── go.sum
├── main.go                    # Application entry point
├── .devcontainer/             # DevContainer settings
│   ├── devcontainer.json
│   ├── Dockerfile
│   └── compose.yml
├── internal/
│   ├── config/                # Configuration management
│   ├── database/              # Database layer
│   ├── handlers/              # HTTP handlers
│   ├── i18n/                  # Internationalization
│   ├── middleware/            # Middleware
│   ├── providers/             # AI provider implementations
│   ├── services/              # Business logic
│   └── models/                # Data models
├── web/
│   └── templates/             # Go html/template
│   │   ├── layout.html
│   │   ├── index.html
│   │   ├── chat.html
│   │   ├── settings.html
│   │   └── error.html
├── locales/                   # i18n files
│   ├── en/
│   │   └── messages.json
│   └── ja/
│       └── messages.json
├── data/                      # SQLite files
├── logs/                      # Log files
└── scripts/                   # Utility scripts
```

## 📇 Architecture

### System Diagram

```
[Web Browser]
    ↓ HTTP/WebSocket
[Go Web Server (Gin)]
    ↓ html/template + Alpine.js
[WebSocketHub] ←→ [AIProvider Registry]
    ↓                    ↓
[Redis Sessions]    [CLI Execution]
    ↓
[SQLite Metadata]
```

### Components

1. **Go html/template Frontend**
- Server-side HTML rendering
- Lightweight client interactions via Alpine.js
- Tailwind CSS for styling (CDN-based)
- Real-time communication via WebSocket

2. **Go Backend (Pluggable Design)**
- HTTP APIs using Gin
- Real-time WebSocket using Gorilla
- AIProvider abstraction layer (interface-based)
- Redis session management
- SQLite metadata persistence

3. **AIProvider Plugin System**
- Claude CLI Provider (initial implementation)
- Gemini CLI Provider (planned)
- Unified interface
- Pluggable authentication

4. **Data Layer**
- SQLite: metadata + chat history
- Redis: active sessions + WebSocket management
- Logs: full execution history (per provider)

## 🔧 Configuration

### Environment Variables

```bash
# Server Settings
PORT=8080
SQLITE_DB_FILE=./data/ai_gateway.db
REDIS_ADDR=localhost:6379
STATIC_DIR=./web/static
TEMPLATE_DIR=./web/templates

# Logging
LOG_DIR=./logs
LOG_LEVEL=info

# Session Management
MAX_SESSIONS=100
SESSION_TIMEOUT=3600
WEBSOCKET_TIMEOUT=7200

# AI Provider Settings
CLAUDE_CLI_PATH=claude
GEMINI_CLI_PATH=gemini

# Claude CLI Options
CLAUDE_SKIP_PERMISSIONS=false
CLAUDE_EXTRA_ARGS=

# Feature Flags
ENABLE_PROVIDER_AUTO_DISCOVERY=true
ENABLE_HEALTH_CHECKS=true
```

### Claude CLI Options
- The following environment variables allow you to configure Claude CLI behavior:

- **CLAUDE_SKIP_PERMISSIONS**: Set to `true` to enable the `--dangerously-skip-permissions` flag. This skips permission prompts during Claude CLI operations. Default: `false`
- **CLAUDE_EXTRA_ARGS**: Additional arguments to pass to Claude CLI. Examples:
  - `--model claude-3-opus-20240229` - Use a specific model
  - `--max-tokens 8192` - Set maximum token limit
  - `--model claude-3-opus-20240229 --max-tokens 8192` - Multiple arguments

- Example configuration:
```bash
CLAUDE_SKIP_PERMISSIONS=true
CLAUDE_EXTRA_ARGS=--model claude-3-opus-20240229 --max-tokens 8192
```

## 📡 API Endpoints

### HTTP API

```
GET  /                    # Main page
GET  /chat/:id           # Chat page
GET  /api/chats          # List chats
POST /api/chats          # Create chat
DELETE /api/chats/:id    # Delete chat
GET  /api/providers      # List available providers
GET  /api/health         # Health check
```

### WebSocket

```
/ws                      # WebSocket connection
```

### WebSocket Message Format

```json
{
  "type": "ai_prompt|ai_response|session_status|error",
  "data": {
    "chat_id": 123,
    "provider": "claude",
    "content": "message content",
    "timestamp": "2025-07-12T10:30:00Z",
    "stream": true
  }
}
```

## 🌐 Internationalization (i18n)

### Supported Languages
- English (default)
- Japanese

### Language Switching
- Auto-detect via Accept-Language header
- Manual override via `?lang=ja`

### Translation Files
- `locales/en/messages.json`
- `locales/ja/messages.json`

### Local Development

```bash
# Clone the repo
git clone https://github.com/yourusername/ai-gateway-hub.git
cd ai-gateway-hub

# Download Go dependencies
go mod download

# Start Redis (Docker example)
docker run -p 6379:6379 redis:7.2-alpine

# Create data and log directories
mkdir -p ./data ./logs

# Install dev tools
make tools

# Run the app
docker compose up
```

## 🤚 Contribution
1. Fork the repo
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Create a pull request

### Development Guidelines
- Follow Go standard coding conventions
- Use `golangci-lint` for code quality
- Add tests for new features
- Always include both English and Japanese i18n
- Prioritize security in all changes

## 📅 License
- This project is licensed under the MIT License.

## 🙏 Acknowledgements
- [Anthropic](https://anthropic.com) - Claude AI
- [Claude CLI](https://github.com/anthropics/claude-cli)
- [Go](https://golang.org)
- [Redis](https://redis.io)
- [Alpine.js](https://alpinejs.dev)

---

**⚠️ Disclaimer**: This app directly executes AI CLI commands and may pose security risks. It is strongly recommended to use it inside a Docker sandbox. Use in production at your own risk.
