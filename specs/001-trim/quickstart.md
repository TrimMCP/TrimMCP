# Developer Quickstart Guide

**Feature**: MCP Response Optimizer Proxy (SlimMCP)
**Branch**: `001-trim`
**Date**: 2025-10-13

## Overview

This guide helps developers get started with SlimMCP development. Follow these steps to set up your environment, run the proxy locally, run tests, and understand the codebase structure.

---

## Prerequisites

### Required Software

- **Go 1.21+**: [Download](https://go.dev/dl/)
- **Git**: For version control
- **Make**: For build automation (optional but recommended)
- **Docker**: For containerized testing (optional)

### Verify Installation

```bash
go version  # Should show Go 1.21 or higher
git --version
make --version
docker --version  # Optional
```

---

## Development Environment Setup

**Note for Core Team**: If you're working in the development monorepo, clone `IdeaCraftersHQ/oss-mcp-slim` and work in the `mcp/` directory. All paths below assume you're in the `mcp/` directory context.

### 1. Clone the Repository

**Production Repository** (External contributors):
```bash
git clone https://github.com/slimmcp/slimmcp.git
cd slimmcp
```

**Development Monorepo** (Core team):
```bash
git clone https://github.com/IdeaCraftersHQ/oss-mcp-slim.git
cd oss-mcp-slim/mcp
```

### 2. Install Dependencies

```bash
# Download Go module dependencies
go mod download

# Install development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### 3. Verify Setup

```bash
# Run tests to verify everything works
go test ./...

# Should see output like:
# ok      github.com/slimmcp/slimmcp/pkg/proxy     0.234s
# ok      github.com/slimmcp/slimmcp/pkg/filter    0.156s
# ...
```

---

## Running the Proxy Locally

### Quick Start

```bash
# Create a basic configuration file
cat > slim.yaml <<EOF
proxy:
  listenAddress: "localhost:8080"
  strategy: smart
  servers:
    mock:
      url: "http://localhost:3000"

monitoring:
  enabled: true
  metricsPort: 9090
  logLevel: debug
EOF

# Build and run
go build -o slim cmd/slimmcp/main.go
./slim start --config slim.yaml
```

### Using Make (Recommended)

```bash
# Build
make build

# Run with default config
make run

# Run with custom config
make run CONFIG=examples/dev.yaml

# Run tests
make test

# Run linter
make lint
```

### Start a Mock MCP Server

For testing, you can start a mock MCP server:

```bash
# In a separate terminal
go run internal/testutil/mock_server.go --port 3000
```

Or use the provided test server:

```bash
make test-server
```

### Verify It's Working

```bash
# Health check
curl http://localhost:8080/health

# Metrics
curl http://localhost:9090/metrics

# Make a test request
curl -X POST http://localhost:8080/mock \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "test-tool",
      "arguments": {"id": "123"}
    },
    "id": 1
  }'
```

---

## Running Tests

### All Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Specific Test Types

```bash
# Unit tests only (fast)
go test ./pkg/...

# Integration tests
go test ./tests/integration/...

# Contract tests
go test ./tests/contract/...

# Acceptance tests (user stories)
go test ./tests/acceptance/...

# Benchmarks
go test -bench=. ./pkg/filter/...
```

### Test a Specific Package

```bash
# Test the filter package
go test ./pkg/filter/

# Test with race detector
go test -race ./pkg/filter/

# Run specific test function
go test ./pkg/filter/ -run TestFilterResponse

# Run tests matching pattern
go test ./... -run Filter
```

---

## Project Structure Overview

```
slimmcp/
├── cmd/slimmcp/         # Main application entry point
├── pkg/                 # Public, reusable packages
│   ├── proxy/          # HTTP server and routing
│   ├── analyzer/       # Request analysis and field inference
│   ├── filter/         # Response filtering engine
│   ├── schema/         # Tool schema management
│   ├── capability/     # Server capability detection
│   ├── client/         # MCP client for downstream servers
│   ├── metrics/        # Prometheus metrics collection
│   ├── config/         # Configuration loading and validation
│   └── types/          # Shared data types
├── internal/           # Private packages
│   └── testutil/       # Test utilities and mocks
├── tests/              # Test suites
│   ├── integration/    # Integration tests
│   ├── contract/       # Protocol compliance tests
│   └── acceptance/     # User story tests
└── examples/           # Example configurations
```

---

## Adding a New Filtering Strategy

### 1. Define Strategy Interface

Strategies implement the `Analyzer` interface in `pkg/analyzer/analyzer.go`:

```go
type Analyzer interface {
    InferFields(req *types.MCPRequest) ([]string, error)
}
```

### 2. Create Strategy File

Create `pkg/analyzer/mystrategies.go`:

```go
package analyzer

import "github.com/slimmcp/slimmcp/pkg/types"

type MyStrategy struct {
    config *types.Config
}

func NewMyStrategy(cfg *types.Config) *MyStrategy {
    return &MyStrategy{config: cfg}
}

func (s *MyStrategy) InferFields(req *types.MCPRequest) ([]string, error) {
    // Your inference logic here
    fields := []string{"id", "type", "url"}

    // Add more fields based on your logic

    return fields, nil
}
```

### 3. Register Strategy

Update `pkg/analyzer/factory.go`:

```go
func NewAnalyzer(strategy string, cfg *types.Config) (Analyzer, error) {
    switch strategy {
    case "smart":
        return NewSmartAnalyzer(cfg), nil
    case "minimal":
        return NewMinimalAnalyzer(cfg), nil
    case "mystrategy":  // Add this
        return NewMyStrategy(cfg), nil
    default:
        return nil, fmt.Errorf("unknown strategy: %s", strategy)
    }
}
```

### 4. Write Tests

Create `pkg/analyzer/mystrategy_test.go`:

```go
func TestMyStrategy(t *testing.T) {
    tests := []struct {
        name           string
        request        *types.MCPRequest
        expectedFields []string
    }{
        {
            name: "basic request",
            request: &types.MCPRequest{
                Method: "tools/call",
                Params: &types.RequestParams{
                    Name: "test-tool",
                    Arguments: map[string]interface{}{
                        "id": "123",
                    },
                },
            },
            expectedFields: []string{"id", "type", "url"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            strategy := NewMyStrategy(nil)
            fields, err := strategy.InferFields(tt.request)

            assert.NoError(t, err)
            assert.ElementsMatch(t, tt.expectedFields, fields)
        })
    }
}
```

### 5. Update Configuration

Add strategy to configuration validation in `pkg/config/validator.go`:

```go
var validStrategies = []string{
    "smart",
    "minimal",
    "explicit",
    "passthrough",
    "mystrategy",  // Add this
}
```

---

## Debugging Request/Response Flow

### Enable Debug Logging

```yaml
# slim.yaml
monitoring:
  logLevel: debug
```

### Add Debug Logs

```go
import "log/slog"

func (s *Server) HandleRequest(w http.ResponseWriter, r *http.Request) {
    slog.Debug("Received request",
        "method", r.Method,
        "path", r.URL.Path,
        "remote", r.RemoteAddr,
    )

    // ... processing ...

    slog.Debug("Sending response",
        "status", statusCode,
        "size", responseSize,
    )
}
```

### Use Request Tracing

The proxy automatically adds request IDs to context:

```go
func (s *Server) HandleRequest(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    requestID := ctx.Value("request_id").(string)

    slog.Info("Processing request", "request_id", requestID)
}
```

### Monitor Metrics

```bash
# Watch metrics in real-time
watch -n 1 curl -s http://localhost:9090/metrics | grep mcp_optimizer
```

---

## Common Development Tasks

### Run Proxy with Test Data

```bash
# Start mock server with test responses
make test-server

# In another terminal, start proxy
make run

# Send test requests
curl -X POST http://localhost:8080/mock \
  -H "Content-Type: application/json" \
  -d @tests/fixtures/notion-fetch-request.json
```

### Profile Performance

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. ./pkg/filter/
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=. ./pkg/filter/
go tool pprof mem.prof
```

### Check for Race Conditions

```bash
# Run with race detector
go test -race ./...

# Run specific test with race detector
go test -race ./pkg/filter/ -run TestConcurrentFiltering
```

### Update Dependencies

```bash
# Update all dependencies
go get -u ./...

# Update specific dependency
go get -u github.com/prometheus/client_golang@latest

# Tidy up
go mod tidy
```

---

## Common Troubleshooting

### Port Already in Use

```bash
# Find process using port 8080
lsof -i :8080

# Kill process
kill -9 <PID>

# Or use a different port in config
proxy:
  listenAddress: "localhost:8081"
```

### Configuration Not Loading

```bash
# Verify YAML syntax
cat slim.yaml | yq eval

# Check file permissions
ls -la slim.yaml

# Use absolute path
./slim start --config /absolute/path/to/slim.yaml
```

### Tests Failing

```bash
# Clean test cache
go clean -testcache

# Re-download dependencies
go mod download

# Verify Go version
go version  # Must be 1.21+
```

### Build Fails

```bash
# Clean build cache
go clean -cache

# Rebuild from scratch
go build -a -o slim cmd/slimmcp/main.go
```

---

## Next Steps

1. **Read the Architecture**: `docs/architecture.md`
2. **Review the Spec**: `specs/001-trim/spec.md`
3. **Check the Plan**: `specs/001-trim/plan.md`
4. **Explore Data Model**: `specs/001-trim/data-model.md`
5. **Read Research Decisions**: `specs/001-trim/research.md`

## Getting Help

- **Documentation**: `docs/` directory
- **Examples**: `examples/` directory
- **Tests**: `tests/` directory (great learning resource)
- **Issues**: Open an issue on GitHub
- **Discord**: Join our Discord community

---

**Happy coding! 🚀**
