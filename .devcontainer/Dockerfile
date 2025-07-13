FROM mcr.microsoft.com/devcontainers/go:1-1.23-bullseye

# Install additional OS packages and SQLite dependencies
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
    git \
    make \
    curl \
    ca-certificates \
    build-essential \
    sqlite3 \
    libsqlite3-dev \
    && apt-get clean -y && rm -rf /var/lib/apt/lists/*

# Install Node.js (for Claude CLI)
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y nodejs \
    && npm install -g npm@latest \
    && npm install -g @anthropic-ai/claude-code

# Install Go tools
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Create workspace directory
WORKDIR /workspace

# Switch back to non-root user
USER vscode