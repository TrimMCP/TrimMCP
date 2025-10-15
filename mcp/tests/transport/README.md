# Multi-Transport Test Suite

This test suite validates that SlimMCP produces **identical responses** regardless of the transport layer used (stdio, HTTP, HTTPS, SSE).

## Purpose

SlimMCP is a transparent proxy that should work seamlessly with any MCP transport protocol. This test harness ensures:

1. **Protocol Compatibility**: All transport types produce equivalent responses
2. **Response Consistency**: Filtering/optimization doesn't introduce transport-specific differences
3. **Reliability**: Each transport can start, execute requests, and stop cleanly
4. **Performance**: Response times are tracked across transports for benchmarking

## Architecture

```
tests/transport/
├── harness/                    # Transport adapters
│   ├── common.go              # Shared interfaces and test scenarios
│   ├── stdio.go               # Stdio transport (native MCP)
│   ├── http.go                # HTTP/HTTPS transport
│   └── sse.go                 # Server-Sent Events transport
│
├── fixtures/                  # Test servers and utilities
│   ├── github-mcp-bridge.js   # HTTP bridge for GitHub MCP
│   ├── start-github-bridge.sh # Launch HTTP bridge
│   └── start-github-mcp.sh    # Launch stdio MCP server
│
├── transport_test.go          # Main test suite
└── README.md                  # This file
```

## How It Works

### 1. Transport Interface

All transports implement a common interface:

```go
type Transport interface {
    Name() string
    Start(ctx context.Context) error
    Stop() error
    Execute(ctx context.Context, request json.RawMessage) (json.RawMessage, error)
    IsHealthy(ctx context.Context) bool
}
```

### 2. Test Scenarios

Scenarios are transport-agnostic test cases:

```go
type TestScenario struct {
    Name        string
    Description string
    Request     json.RawMessage
    Timeout     time.Duration
    ValidateResponse func(t *testing.T, response json.RawMessage) error
}
```

### 3. Response Normalization

Before comparing responses, we normalize them to remove transport-specific metadata:

- Timing information
- Transport-specific headers
- Server version strings

### 4. Equivalence Testing

For each scenario:
1. Execute the same request on all transports
2. Normalize all responses
3. Assert that all normalized responses are identical

## Running Tests

### Prerequisites

```bash
# Install dependencies
npm install -g @modelcontextprotocol/server-github

# Set GitHub token
export GITHUB_PERSONAL_ACCESS_TOKEN="your-token"
# or
export GITHUB_TOKEN="your-token"
```

### Run All Transport Tests

```bash
cd tests/transport
go test -v
```

### Run Specific Test

```bash
go test -v -run TestIdenticalResponsesAcrossTransports
```

### Run with Coverage

```bash
go test -v -cover
```

## Supported Transports

### Stdio (Native MCP)
- **Status**: ✅ Implemented
- **Server**: `@modelcontextprotocol/server-github`
- **Protocol**: JSON-RPC over stdin/stdout
- **Use Case**: Claude Desktop integration

### HTTP
- **Status**: ✅ Implemented
- **Server**: `github-mcp-bridge.js` (wrapper around stdio server)
- **Protocol**: JSON-RPC over HTTP POST
- **Use Case**: Web-based integrations

### HTTPS
- **Status**: ✅ Implemented
- **Server**: Same as HTTP with TLS
- **Protocol**: JSON-RPC over HTTPS POST
- **Use Case**: Production deployments

### SSE (Server-Sent Events)
- **Status**: ✅ Implemented (adapter ready, needs backend)
- **Protocol**: Requests via HTTP POST, responses via SSE stream
- **Use Case**: Real-time streaming applications

## Adding New Scenarios

To add a new test scenario:

```go
// In transport_test.go
scenarios = append(scenarios, harness.TestScenario{
    Name:        "my_new_test",
    Description: "Test description",
    Request: harness.mustMarshal(harness.MCPRequest{
        JSONRPC: "2.0",
        ID:      123,
        Method:  "tools/call",
        Params:  harness.mustMarshal(map[string]interface{}{
            "name": "tool_name",
            "arguments": map[string]interface{}{
                "param": "value",
            },
        }),
    }),
    Timeout: 10 * time.Second,
    ValidateResponse: func(t *testing.T, response json.RawMessage) error {
        // Custom validation logic
        return nil
    },
})
```

## Adding New Transports

To add support for a new transport:

1. **Create adapter**: Implement `harness.Transport` interface in `harness/`
2. **Add to test**: Instantiate transport in `transport_test.go`
3. **Add fixture**: If needed, add server startup script to `fixtures/`

Example:

```go
// harness/websocket.go
type WebSocketTransport struct {
    // ...
}

func (w *WebSocketTransport) Start(ctx context.Context) error { ... }
func (w *WebSocketTransport) Stop() error { ... }
func (w *WebSocketTransport) Execute(ctx context.Context, request json.RawMessage) (json.RawMessage, error) { ... }
func (w *WebSocketTransport) IsHealthy(ctx context.Context) bool { ... }

// transport_test.go
transports = append(transports, harness.NewWebSocketTransport(...))
```

## Fixtures

### github-mcp-bridge.js

HTTP bridge that wraps the GitHub MCP stdio server:

```bash
# Start the bridge
cd fixtures/
./start-github-bridge.sh

# Now available at http://localhost:3100
curl -X POST http://localhost:3100 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{}}}'
```

### start-github-mcp.sh

Direct stdio server launcher:

```bash
# Start stdio server
cd fixtures/
./start-github-mcp.sh
```

## Debugging

### Enable verbose output

```bash
go test -v -run TestIdenticalResponsesAcrossTransports 2>&1 | tee test.log
```

### Check transport health

```bash
# HTTP bridge
curl -X POST http://localhost:3100 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{}}}'

# Stdio (manual)
cd fixtures/
./start-github-mcp.sh
# Type JSON-RPC request into stdin
```

### Common Issues

**"GITHUB_PERSONAL_ACCESS_TOKEN not set"**
- Export the environment variable before running tests

**"Failed to start http-github-bridge"**
- The HTTP bridge might not be running
- Tests will automatically skip unavailable transports

**"Request timeout"**
- Increase timeout in test scenario
- Check if backend server is responding

## Performance Benchmarking

To compare transport performance:

```bash
go test -bench=. -benchmem
```

This will show:
- Latency per transport
- Memory allocations
- Throughput comparisons

## CI Integration

Example GitHub Actions workflow:

```yaml
- name: Run Transport Tests
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  run: |
    cd mcp/tests/transport
    go test -v -cover
```

## Future Enhancements

- [ ] WebSocket transport adapter
- [ ] gRPC transport adapter (if MCP adds support)
- [ ] Parallel execution benchmarks
- [ ] Load testing across transports
- [ ] Fuzz testing for request/response handling
- [ ] Response caching validation across transports

## References

- [MCP Protocol Specification](https://spec.modelcontextprotocol.io/)
- [SlimMCP Architecture](../../docs/architecture.md)
- [JSON-RPC 2.0 Specification](https://www.jsonrpc.org/specification)
