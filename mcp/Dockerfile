# SlimMCP Docker Image
# Multi-stage build for minimal image size

# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build binary with optimizations
# -ldflags explanation:
#   -s: strip symbol table
#   -w: strip DWARF debug info
#   -X: set version variable
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -trimpath \
    -o slim \
    ./cmd/slimmcp

# Verify binary works
RUN ./slim --version

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    && addgroup -g 1000 slimmcp \
    && adduser -D -u 1000 -G slimmcp slimmcp

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/slim /usr/local/bin/slim

# Copy example configurations for reference
COPY --from=builder /build/examples /app/examples

# Copy CHANGELOG and README
COPY --from=builder /build/CHANGELOG.md /app/
COPY --from=builder /build/README.md /app/

# Set ownership
RUN chown -R slimmcp:slimmcp /app

# Switch to non-root user
USER slimmcp

# Expose ports
# 8080: Proxy server
# 9090: Metrics server
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Default command
ENTRYPOINT ["/usr/local/bin/slim"]
CMD ["start", "--config", "/config/slim.yaml"]

# Labels
LABEL org.opencontainers.image.title="SlimMCP"
LABEL org.opencontainers.image.description="Transparent MCP proxy for 60-85% token reduction"
LABEL org.opencontainers.image.vendor="IdeaCraftersHQ"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/slimmcp/slimmcp"
LABEL org.opencontainers.image.documentation="https://docs.slimmcp.com"
