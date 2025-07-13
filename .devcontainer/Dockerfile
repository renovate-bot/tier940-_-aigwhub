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
    xz-utils \
    && apt-get clean -y && rm -rf /var/lib/apt/lists/*

# Install Node.js
RUN curl -fsSL https://nodejs.org/dist/v22.11.0/node-v22.11.0-linux-x64.tar.xz -o node.tar.xz \
    && tar -xf node.tar.xz -C /usr/local --strip-components=1 \
    && rm node.tar.xz \
    && npm install -g @anthropic-ai/claude-code

# Install Go tools
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Create workspace directory
WORKDIR /workspace

# Switch back to non-root user
USER vscode
