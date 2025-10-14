# SlimMCP Technical Stack Documentation

## Overview

SlimMCP is a high-performance, transparent middleware designed to optimize Model Context Protocol (MCP) response sizes. By intelligently filtering and reducing API responses, SlimMCP achieves 60-85% size reduction while maintaining full protocol compatibility.

## 1. Core Technologies

### Programming Language & Runtime
- **Language**: Go (Golang)
  - Version: Go 1.25+ (latest stable release)
  - Chosen for:
    - High performance
    - Excellent concurrency support
    - Compiled language with low overhead
    - Strong typing and memory safety
    - Built-in tooling and standard library

### Key Language Features Utilized
- Goroutines and channels for concurrent processing
- Interfaces for flexible design
- Compile-time type checking
- Efficient memory management
- Zero-cost abstractions

## 2. HTTP & Networking

### Web Server
- **Standard Library**: `net/http`
  - Built-in HTTP server implementation
  - High-performance request handling
  - Lightweight and efficient

### Routing
- **Library**: `github.com/gorilla/mux`
  - Flexible routing capabilities
  - URL path variables
  - Middleware support
  - Performance-optimized route matching

### Client Networking
- Standard library `net/http` for client requests
- Configurable connection pooling
- Timeout and retry mechanisms
- TLS support for secure communications

## 3. Middleware Stack

### CORS (Cross-Origin Resource Sharing)
- **Library**: `github.com/rs/cors`
  - Configurable CORS policy management
  - Wildcard and specific origin support
  - Credential handling
  - Preflight request optimization

**Configuration Options**:
- Allowed origins (with pattern matching)
- Allowed methods (GET, POST, OPTIONS, etc.)
- Allowed headers (including custom MCP headers)
- Exposed headers for client access
- Max age for preflight caching

### Rate Limiting
- **Library**: `golang.org/x/time/rate`
  - Token bucket algorithm implementation
  - Per-client rate limiting
  - Burst capacity management
  - Configurable rate windows

**Features**:
- IP-based rate limiting
- Authenticated user rate limiting
- Endpoint-specific limits
- Graceful degradation
- Rate limit headers (X-RateLimit-*)
- 429 Too Many Requests responses

**Configuration**:
- Requests per second/minute/hour
- Burst size
- Sliding window implementation
- Redis-backed distributed rate limiting (optional)

### Security Headers
- **Library**: Custom middleware using `net/http`
  - Automatic security header injection
  - Configurable policy enforcement
  - CSP (Content Security Policy) support

**Headers Implemented**:
```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains
Content-Security-Policy: default-src 'self'
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: geolocation=(), microphone=(), camera=()
```

### Request/Response Middleware Chain
- **Execution Order**:
  1. Security headers injection
  2. CORS preflight handling
  3. Rate limiting check
  4. Authentication/Authorization (pass-through)
  5. Request logging and tracing
  6. Body size limiting
  7. Request timeout enforcement
  8. Compression (gzip/deflate)
  9. MCP protocol handling
  10. Response filtering and optimization
  11. Performance metrics collection
  12. Error handling and recovery

### Additional Middleware Components

#### Compression
- **Library**: Standard library `compress/gzip`, `compress/flate`
  - Automatic response compression
  - Content negotiation (Accept-Encoding)
  - Configurable compression levels
  - Size threshold for compression

#### Request ID & Correlation
- **Library**: Custom middleware
  - Unique request ID generation
  - X-Request-ID header injection
  - Distributed tracing correlation
  - Log correlation support

#### Body Size Limiting
- Built-in `http.MaxBytesReader`
  - Configurable max request body size
  - Protection against memory exhaustion
  - Early termination of oversized requests

#### Panic Recovery
- Custom recovery middleware
  - Graceful panic handling
  - Stack trace logging
  - 500 Internal Server Error responses
  - Prevents server crashes

### Middleware Configuration Example
```yaml
middleware:
  cors:
    enabled: true
    allowed_origins:
      - "*"
    allowed_methods:
      - GET
      - POST
      - OPTIONS
    allowed_headers:
      - Content-Type
      - Authorization
    max_age: 3600

  rate_limiting:
    enabled: true
    requests_per_minute: 60
    burst_size: 10
    storage: memory  # or redis

  security_headers:
    enabled: true
    hsts_max_age: 31536000
    csp: "default-src 'self'"

  compression:
    enabled: true
    level: 6  # 1-9
    min_size: 1024  # bytes

  body_limit:
    max_size: 10485760  # 10MB
```

## 4. Protocol Implementation

### MCP SDK
- **Library**: `github.com/mark3labs/mcp-go`
  - Native Go implementation of MCP protocol
  - Full JSON-RPC 2.0 support
  - Type-safe method invocations
  - Automatic schema management

### JSON Handling
- **Standard Library**: `encoding/json`
  - High-performance JSON encoding/decoding
  - Reflection-based marshaling/unmarshaling
  - Customizable struct tag handling

### Protocol Extensions
- Custom field selection (_meta.fieldSelection)
- Transparent request/response transformation
- Metadata injection
- Performance tracking

## 5. Data Processing

### JSON Filtering
- Custom filtering algorithm
- Preserve structural integrity
- O(n) time complexity
- Memory-efficient traversal
- Support for nested and complex schemas

### Schema Management
- In-memory schema representation
- Dependency graph for field relationships
- Validation of filtering operations
- Caching of schema metadata

## 6. Monitoring & Observability

### Metrics
- **Library**: `github.com/prometheus/client_golang`
  - Prometheus metrics integration
  - Custom counters and histograms
  - Performance and filtering statistics tracking
  - Exportable metrics endpoint

### Logging
- Standard library `log` package
- Configurable log levels
- JSON-structured logging
- Performance-aware logging

### Distributed Tracing
- OpenTelemetry compatibility
- Trace context propagation
- Request lifecycle tracking

## 7. Configuration Management

### Configuration Formats
- **Library**: `gopkg.in/yaml.v3`
  - YAML configuration parsing
  - Type-safe configuration loading
  - Environment variable interpolation

### Configuration Features
- Hot reload support
- Multiple configuration sources
- Validation of configuration parameters
- Default and environment-specific configurations

## 8. Deployment & Distribution

### Containerization
- Docker support
- Multi-stage builds
- Minimal Alpine-based images
- Optimized layer caching

### Distribution
- Single statically linked binary
- Cross-platform compilation (Linux, macOS, Windows)
- Package managers (Homebrew, Apt)
- Release binaries with checksum verification

## 9. Development Tools

### Testing
- Standard library `testing` package
- Table-driven tests
- Benchmarking support
- Integration and unit test coverage

### Code Quality
- `golangci-lint` for static analysis
- `gofmt` for consistent formatting
- `go vet` for catching common errors

### Dependency Management
- `go mod` for dependency tracking
- Vendoring support
- Reproducible builds

## 10. Performance Optimization

### Concurrency
- Goroutine-based parallel processing
- Buffered channels for communication
- Sync primitives (`sync.WaitGroup`, `sync.Mutex`)
- Context-based cancellation

### Memory Management
- Stack allocation preference
- Escape analysis
- Minimal heap allocations
- Efficient garbage collection

### Caching
- In-memory LRU caches
- Thread-safe cache implementations
- Configurable cache sizes
- Adaptive cache strategies

## 11. Security

### Network Security
- TLS 1.3 support
- Secure default configurations
- Certificate management
- HTTPS enforcement

### Authentication
- Pass-through authentication
- Optional JWT/OAuth integration
- Configurable authentication providers

### Data Privacy
- Minimal data retention
- Configurable PII filtering
- Encryption of sensitive metadata

## 12. Integration

### MCP Ecosystem
- Full compatibility with existing MCP clients
- Transparent middleware design
- No client-side changes required
- Support for various MCP servers

## 13. Dependencies

### Core Dependencies
```go
module github.com/yourdomain/slimmcp

go 1.21

require (
    github.com/gorilla/mux v1.8.1
    github.com/prometheus/client_golang v1.23.2
    github.com/mark3labs/mcp-go v0.5.2
    gopkg.in/yaml.v3 v3.0.1
    github.com/rs/cors v1.11.0
    golang.org/x/time v0.5.0
)
```

## 14. Alternative Technologies Considered

| Technology   | Pros                     | Cons                      | Chosen Technology |
|--------------|--------------------------|---------------------------|-------------------|
| TypeScript   | Rich ecosystem           | Higher runtime overhead   | Go                |
| Python       | Quick prototyping        | Performance limitations   | Go                |
| Rust         | Extreme performance      | Steeper learning curve    | Go                |

## 15. Future Considerations

- Machine learning-based adaptive filtering
- Enhanced schema inference
- More granular metrics
- WebAssembly compilation
- Advanced caching strategies

## See Also

- [Architecture Documentation](./architecture.md)
- [API Specification](./api.md)
- [Deployment Guide](./deployment.md)

---

**Version**: 1.0.0
**Last Updated**: 2025-10-13
