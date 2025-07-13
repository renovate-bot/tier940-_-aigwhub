# Multi-stage build for production
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o ai-gateway-hub ./main.go

# Production image
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    sqlite \
    redis \
    curl \
    nodejs \
    npm \
    && npm install -g @anthropic-ai/claude-code \
    && addgroup -g 1001 -S appgroup \
    && adduser -u 1001 -S appuser -G appgroup

# Create necessary directories
RUN mkdir -p /app/data /app/logs /app/locales /app/web && \
    chown -R appuser:appgroup /app

# Copy binary from builder
COPY --from=builder /app/ai-gateway-hub /app/
COPY --from=builder /app/web /app/web/
COPY --from=builder /app/locales /app/locales/
COPY --from=builder /app/.env.example /app/

# Set permissions
RUN chmod +x /app/ai-gateway-hub && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Set working directory
WORKDIR /app

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/api/health || exit 1

# Default environment variables
ENV PORT=8080 \
    SQLITE_DB_FILE=./data/ai_gateway.db \
    REDIS_ADDR=redis:6379 \
    LOG_DIR=./logs \
    LOG_LEVEL=info \
    CLAUDE_CLI_PATH=/usr/bin/claude

# Run the application
CMD ["./ai-gateway-hub"]