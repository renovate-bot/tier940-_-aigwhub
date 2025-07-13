# AI Gateway Hub

Modern web interface for using multiple AI CLI tools (Claude Code, Gemini CLI, etc.) in a browser.

> **Note**: All static resources (HTML templates, internationalization files) are embedded in the executable for standalone deployment.

## ✨ Features

- 🌐 **Browser Access**: Use multiple AI CLI tools from any device via web browser
- 🚀 **Zero Installation**: Skip terminal setup and access AI tools instantly
- 👥 **Team Collaboration**: Easily share sessions and coding discussions
- 🔧 **Multi-Provider Support**: Claude Code, Gemini CLI, and future AI tools in one interface
- 🔄 **Session Persistence**: Redis-based to never lose coding context
- 💬 **Real-time Coding**: Instant AI responses via WebSocket

## 🚀 Quick Start

### Prerequisites

- VS Code with Dev Containers extension
- Docker & Docker Compose
- Claude CLI installed and authenticated (`claude auth`)

### Using Standalone Binary

**⚠️ Important**: All operations should be performed in the `run/` directory.

```bash
# Navigate to run directory (required)
cd run/

# Create .env from example (if not exists)
cp ../.env.example .env

# Edit .env as needed
# Especially REDIS_ADDR configuration

# Start Redis (using Docker)
docker run -p 6379:6379 redis:7.2-alpine

# Start application (must be run from run/ directory)
./ai-gateway-hub

# Access in browser
open http://localhost:8080
```

> **Note**: The application creates `data/` and `logs/` directories relative to its execution path. Always run from the `run/` directory to keep all runtime files contained.

### Development Environment (DevContainer)

For development and building:

```bash
# Open project in VS Code
code .

# Reopen in DevContainer (auto-setup on first run)
# Ctrl+Shift+P → "Dev Containers: Reopen in Container"

# Available Claude Code custom commands:
/go-build    # Build application in DevContainer
/go-run      # Run application in DevContainer
/go-stop     # Stop running application
/go-test     # Test application with browser automation

# Generated files in run/ directory:
# run/
# ├── ai-gateway-hub    # Standalone executable with embedded resources
# ├── .env             # Environment configuration (copy from .env.example)
# ├── data/            # SQLite database files (created at runtime)
# └── logs/            # Application logs (created at runtime)
```

### Build Artifacts

Building in DevContainer generates files in the `run/` directory:

- `ai-gateway-hub`: Standalone executable with embedded resources
- `data/`: SQLite database files (created at runtime)
- `logs/`: Application logs (created at runtime)
- `.env`: Configuration file (copy from `.env.example`)

**Important**: All runtime files are contained in `run/` directory. Always execute the application from this directory to ensure proper file organization.

## 🔒 Security Notice

⚠️ **Important**: This application executes AI CLI commands directly, which may pose security risks.

- **Docker environment strongly recommended**
- Production use at your own risk
- Implement proper network restrictions and access controls

## 📚 Documentation

- **[CLAUDE.md](./CLAUDE.md)** - Detailed technical specifications for developers  
- **[README_JP.md](./README_JP.md)** - Japanese version
- **API Endpoints**: `/api/health` for health checks
- **WebSocket**: `/ws` for real-time communication

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Create a pull request

## 📄 License

MIT License

---

**⚠️ Disclaimer**: This application executes AI CLI directly and may pose security risks. Docker environment usage is strongly recommended. Production use is at your own risk.
