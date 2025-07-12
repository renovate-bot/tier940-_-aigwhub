FROM mcr.microsoft.com/devcontainers/go:1-1.23-bullseye

# Install additional OS packages
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
    git \
    make \
    curl \
    ca-certificates \
    && apt-get clean -y && rm -rf /var/lib/apt/lists/*

# Install Node.js (for Claude CLI)
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs \
    && npm install -g npm@latest

# Create workspace directory
WORKDIR /workspace

# Switch back to non-root user
USER vscode