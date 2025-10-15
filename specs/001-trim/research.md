# Phase 0: Research & Technology Decisions

**Feature**: MCP Response Optimizer Proxy
**Branch**: `001-trim`
**Date**: 2025-10-13

## Overview

This document captures research findings and technology decisions for implementing SlimMCP, a transparent MCP proxy with intelligent response filtering. All decisions prioritize performance, simplicity, and production-readiness.

---

## 1. Go HTTP Server Patterns

### Decision: Standard library `net/http` with custom server configuration

**Rationale**:
- Go's standard library HTTP server is battle-tested and production-ready
- Provides all needed features: timeouts, graceful shutdown, TLS support
- No framework overhead for our simple proxy use case
- Better performance than frameworks like Gin or Echo for our workload

**Implementation Approach**:

```go
type Server struct {
    httpServer *http.Server
    config     *config.ProxyConfig
    analyzer   analyzer.Analyzer
    client     client.MCPClient
    metrics    metrics.Collector
}

func (s *Server) Start() error {
    s.httpServer = &http.Server{
        Addr:              s.config.ListenAddress,
        Handler:           s.router(),
        ReadTimeout:       30 * time.Second,
        WriteTimeout:      30 * time.Second,
        IdleTimeout:       120 * time.Second,
        ReadHeaderTimeout: 10 * time.Second,
    }
    return s.httpServer.ListenAndServe()
}
```

**Graceful Shutdown Pattern**:
- Use `context.WithTimeout` for graceful shutdown window (30s)
- Call `httpServer.Shutdown(ctx)` to stop accepting new requests
- Wait for in-flight requests to complete
- Close downstream connections cleanly

**Connection Pooling**:
- Use `http.Client` with custom `Transport` for downstream requests
- Configure connection pool: `MaxIdleConns: 100`, `MaxIdleConnsPerHost: 10`
- Set reasonable timeouts: `DialTimeout: 10s`, `ResponseHeaderTimeout: 30s`

**Request Context Propagation**:
- Extract request context from `http.Request`
- Propagate context to all downstream operations
- Use context for cancellation and timeouts
- Add request ID to context for tracing

**Alternatives Considered**:
- **FastHTTP**: Higher performance but incompatible with standard middleware
- **Gin/Echo**: Unnecessary framework overhead for simple proxy
- **gRPC**: Wrong protocol (MCP uses JSON-RPC over HTTP)

---

## 2. JSON-RPC 2.0 Implementation

### Decision: Custom JSON-RPC implementation using standard `encoding/json`

**Rationale**:
- JSON-RPC 2.0 is simple enough to implement directly
- Existing libraries (gorilla/rpc, ybbus/jsonrpc) are either too heavy or unmaintained
- Custom implementation gives full control over _meta extension handling
- Zero external dependencies for core protocol logic

**Implementation Approach**:

```go
type JSONRPCRequest struct {
    JSONRPC string          `json:"jsonrpc"`
    Method  string          `json:"method"`
    Params  json.RawMessage `json:"params,omitempty"`
    ID      interface{}     `json:"id,omitempty"`
    Meta    *RequestMeta    `json:"_meta,omitempty"`
}

type JSONRPCResponse struct {
    JSONRPC string          `json:"jsonrpc"`
    Result  json.RawMessage `json:"result,omitempty"`
    Error   *JSONRPCError   `json:"error,omitempty"`
    ID      interface{}     `json:"id"`
    Meta    *ResponseMeta   `json:"_meta,omitempty"`
}
```

**Batch Request Handling**:
- Detect batch by checking if request body starts with `[`
- Unmarshal to `[]JSONRPCRequest`
- Process each request independently (can parallelize with goroutines)
- Return `[]JSONRPCResponse` maintaining order by ID

**Error Codes** (JSON-RPC 2.0 standard):
- `-32700`: Parse error (Invalid JSON)
- `-32600`: Invalid Request
- `-32601`: Method not found
- `-32602`: Invalid params
- `-32603`: Internal error
- `-32000 to -32099`: Server error range (custom errors)

**Custom Error Codes**:
- `-32000`: Downstream server error
- `-32001`: Filtering error
- `-32002`: Schema validation error
- `-32003`: Timeout error

**Alternatives Considered**:
- **gorilla/rpc**: Too opinionated about method routing, doesn't support _meta extension easily
- **ybbus/jsonrpc**: Client-focused, not maintained, limited server functionality
- **powerman/rpc-codec**: Good but adds dependency for simple use case

---

## 3. Field Inference Algorithms

### Decision: Multi-strategy approach with heuristics-based smart inference

**Rationale**:
- Different use cases need different inference strategies
- Heuristic approach is fast and doesn't require AST parsing
- Can progressively improve inference accuracy with learning mode later

**Smart Strategy Implementation**:

1. **Parameter Field Extraction**:
   - Parse `params.arguments` object for field references
   - Look for patterns like `"field": "value"` or `["field1", "field2"]`
   - Use regex to find dot-notation references: `properties.status`

2. **Inference Heuristics**:
   - Always include critical fields: `id`, `type`, `url`, `title`
   - If parameter mentions `status`, include status field
   - If parameter has user/assignee reference, include user object
   - If filtering by field, likely need that field in response

3. **Field Relationship Detection**:
   - Use schema to detect parent-child relationships
   - If requesting `assignee.name`, automatically include `assignee`
   - Nested field notation (`properties.status`) preserved in structure

4. **Tool-Specific Defaults**:
   - Configuration can specify default fields per tool
   - Example: `notion-fetch` always includes `[id, title, url, properties.status]`
   - Overridable by explicit field selection

**Minimal Strategy**:
Simple fixed list: `id`, `type`, `url`, `title` only.

**Explicit Strategy**:
Only use fields from `_meta.fieldSelection.include` plus critical fields.

**Passthrough Strategy**:
No filtering applied, return full response.

**Performance Considerations**:
- Regex compilation cached globally
- Parameter parsing: single pass through JSON
- Field set operations using Go maps (O(1) lookups)
- Target: <2ms for inference on typical request

**Alternatives Considered**:
- **AST parsing**: Too slow (10-50ms), overkill for field extraction
- **Machine learning**: Requires training data, adds complexity, slower inference
- **Static analysis**: Can't handle dynamic field references in parameters

---

## 4. Schema Management

### Decision: Dual-source schema loading with server capabilities preferred

**Rationale**:
- Servers should be source of truth for their own schemas
- Configuration overrides allow fixing incorrect server schemas
- Caching prevents repeated schema fetches
- Graceful degradation when schemas unavailable

**Implementation Approach**:

```go
type SchemaManager struct {
    serverCache   map[string]*ToolSchema
    configSchemas map[string]*ToolSchema
    cacheTTL      time.Duration
    mutex         sync.RWMutex
}

func (m *SchemaManager) GetSchema(tool string, serverURL string) (*ToolSchema, error) {
    // 1. Check config override first
    if schema := m.configSchemas[tool]; schema != nil {
        return schema, nil
    }

    // 2. Check cache
    if cached := m.getCached(serverURL, tool); cached != nil {
        return cached, nil
    }

    // 3. Fetch from server
    schema, err := m.fetchFromServer(serverURL, tool)
    if err != nil {
        // 4. Return minimal schema as fallback
        return m.getMinimalSchema(), nil
    }

    // 5. Cache and return
    m.cache(serverURL, tool, schema)
    return schema, nil
}
```

**MCP Schema Discovery**:
- Use `tools/list` method to get available tools
- Tool definitions include input schema (JSON Schema)
- Output schema not standardized - infer from responses
- Cache schemas with 1-hour TTL

**Schema Versioning**:
- Store schema version with cached schema
- Invalidate cache on version mismatch
- Support schema evolution with optional fields
- Required fields always validated

**Fallback Strategy**:
- If no schema available, use minimal critical fields only
- Log warning about missing schema
- Don't fail requests due to schema absence
- Collect usage data to build implicit schemas

**Cache Invalidation**:
- TTL-based expiration (1 hour default)
- Manual invalidation via admin endpoint
- Server disconnect triggers cache clear for that server
- Configuration reload refreshes config-based schemas

**Alternatives Considered**:
- **Database storage**: Adds complexity, not needed for stateless design
- **File-based cache**: Slower than in-memory, complicates deployment
- **No caching**: Too slow (schema fetch every request)

---

## 5. Performance Optimization

### Decision: Buffered JSON filtering with lazy parsing

**Rationale**:
- Full response must be received before filtering can begin
- Lazy parsing (`json.RawMessage`) delays unnecessary JSON operations
- Strategic use of streaming for large responses (>1MB)
- Memory pooling for frequently allocated objects

**JSON Parsing Strategy**:

```go
// Lazy parsing approach
type Response struct {
    ID     string          `json:"id"`
    Result json.RawMessage `json:"result"` // Don't parse yet
}

// Filter after selective parsing
func filterResponse(raw json.RawMessage, fields []string) (map[string]interface{}, error) {
    var data map[string]interface{}
    if err := json.Unmarshal(raw, &data); err != nil {
        return nil, err
    }
    return filterMap(data, fields), nil
}
```

**Filtering Algorithm Performance**:
- Single-pass recursive filtering: O(n) where n = number of fields
- Field lookups using maps: O(1)
- No copying unless field is kept
- Early termination at max depth (default 10)

**Memory Efficiency**:
- Use `sync.Pool` for frequently allocated objects (filter contexts, field sets)
- Reuse byte buffers for JSON marshaling
- Clear maps after use instead of allocating new ones
- Limit response size to prevent OOM (default 10MB max)

**Streaming Consideration**:
For future optimization if needed:
- Responses >1MB could use streaming JSON parser (e.g., `jsoniter`)
- Trade-off: More complex code vs. marginal gain (most responses <100KB)
- Decision: Start buffered, add streaming only if profiling shows need

**Caching Strategies**:
- **Capability cache**: In-memory with 1-hour TTL
- **Schema cache**: In-memory with 1-hour TTL
- **No response caching**: Proxy should be transparent, caching is separate concern

**Benchmarking Targets**:
- Filtering 1KB response: <1ms
- Filtering 100KB response: <5ms
- Filtering 1MB response: <50ms
- Memory allocation: <2x response size

**Alternatives Considered**:
- **Streaming JSON parser** (jsoniter): Premature optimization, adds complexity
- **Protobuf**: Wrong protocol (MCP uses JSON)
- **MessagePack**: Not compatible with MCP specification

---

## 6. Testing Strategies

### Decision: Three-tier testing approach (unit, integration, acceptance)

**Rationale**:
- Unit tests for individual package logic (fast, isolated)
- Integration tests with mock MCP servers (realistic scenarios)
- Acceptance tests verify user stories (end-to-end validation)

**Mock MCP Server Implementation**:

```go
type MockMCPServer struct {
    responses map[string]json.RawMessage
    calls     []CallRecord
    mutex     sync.Mutex
}

func (m *MockMCPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    var req JSONRPCRequest
    json.NewDecoder(r.Body).Decode(&req)

    // Record call for assertions
    m.recordCall(req)

    // Return pre-configured response
    response := m.responses[req.Params.Name]
    json.NewEncoder(w).Encode(JSONRPCResponse{
        JSONRPC: "2.0",
        Result:  response,
        ID:      req.ID,
    })
}
```

**Table-Driven Tests**:
```go
func TestFilterStrategies(t *testing.T) {
    tests := []struct {
        name           string
        strategy       string
        request        MCPRequest
        response       map[string]interface{}
        expectedFields []string
        expectedSavings float64
    }{
        {"minimal_notion", "minimal", notionRequest, notionResponse, []string{"id", "title", "url"}, 85.0},
        {"smart_slack", "smart", slackRequest, slackResponse, []string{"id", "text", "user"}, 68.0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

**Performance Benchmarking**:
```go
func BenchmarkFiltering(b *testing.B) {
    response := generateLargeResponse(100) // 100 fields
    fields := []string{"id", "title", "status"}

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        filterResponse(response, fields)
    }
}
```

**Test Organization**:
- **Unit**: `pkg/*/`*`_test.go` - Test individual functions
- **Integration**: `tests/integration/` - Test component interactions
- **Contract**: `tests/contract/` - Verify protocol compliance
- **Acceptance**: `tests/acceptance/` - Validate user stories

**Coverage Targets**:
- Unit tests: 80%+ coverage
- Critical paths (filtering, routing): 100% coverage
- Integration tests: All user stories covered
- Performance tests: All strategies benchmarked

**Alternatives Considered**:
- **Real MCP servers in tests**: Too slow, requires external dependencies
- **Generated mocks** (gomock): Overkill for simple interfaces
- **BDD frameworks** (Ginkgo): Unnecessary complexity

---

## 7. Configuration Management

### Decision: gopkg.in/yaml.v3 for YAML parsing with manual validation

**Rationale**:
- yaml.v3 is standard, well-maintained, zero dependencies beyond itself
- Viper is overkill (we don't need env vars, remote config, etc.)
- Manual validation gives precise control over error messages
- Simple configuration structure doesn't need complex unmarshaling

**Configuration Structure**:

```go
type Config struct {
    Proxy        ProxyConfig        `yaml:"proxy"`
    Optimization OptimizationConfig `yaml:"optimization"`
    Monitoring   MonitoringConfig   `yaml:"monitoring"`
    Learning     *LearningConfig    `yaml:"learning,omitempty"`
}

type ProxyConfig struct {
    ListenAddress string                 `yaml:"listenAddress"`
    Strategy      string                 `yaml:"strategy"`
    Servers       map[string]ServerConfig `yaml:"servers"`
    CORS          *CORSConfig            `yaml:"cors,omitempty"`
}
```

**Hot Reload Implementation**:

```go
func (s *Server) ReloadConfig() error {
    newConfig, err := config.Load(s.configPath)
    if err != nil {
        return fmt.Errorf("load config: %w", err)
    }

    if err := newConfig.Validate(); err != nil {
        return fmt.Errorf("validate config: %w", err)
    }

    s.mu.Lock()
    defer s.mu.Unlock()

    // Apply new configuration
    s.config = newConfig
    s.analyzer.UpdateStrategy(newConfig.Proxy.Strategy)
    // Update other components...

    return nil
}
```

**Reload Trigger**:
- SIGHUP signal triggers reload
- Admin endpoint: `POST /admin/reload`
- File watcher (optional, using fsnotify)
- Reload completes in <5s (per success criteria)

**Validation Approach**:
- Struct tags for required fields
- Custom validation method checks business rules
- Validate server URLs are reachable (optional check)
- Fail fast with descriptive error messages

**Configuration Schema**:
Will generate `contracts/config-schema.yaml` with full JSON Schema for validation and IDE autocomplete.

**Alternatives Considered**:
- **Viper**: Too heavy (we don't need 12-factor app features)
- **envconfig**: Configuration should be explicit YAML file, not scattered env vars
- **JSON config**: YAML is more human-friendly for configuration

---

## 8. Prometheus Integration

### Decision: prometheus/client_golang with custom collectors

**Rationale**:
- Official Prometheus client, well-maintained
- Rich metric types (Counter, Histogram, Gauge, Summary)
- Efficient collection with minimal overhead (<1ms)
- Standard `/metrics` endpoint format

**Metric Definitions**:

```go
var (
    // Counters
    tokensOriginalTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "mcp_optimizer_tokens_original_total",
            Help: "Total tokens in original (unfiltered) responses",
        },
        []string{"server", "tool"},
    )

    tokensSavedTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "mcp_optimizer_tokens_saved_total",
            Help: "Total tokens saved by filtering",
        },
        []string{"server", "tool"},
    )

    requestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "mcp_optimizer_requests_total",
            Help: "Total requests processed",
        },
        []string{"server", "tool", "status", "strategy"},
    )

    // Histograms
    filteringDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "mcp_optimizer_filtering_duration_seconds",
            Help:    "Time spent filtering responses",
            Buckets: prometheus.DefBuckets, // 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10
        },
        []string{"strategy"},
    )

    // Gauges
    savingsPercentage = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "mcp_optimizer_savings_percentage",
            Help: "Current token savings percentage",
        },
        []string{"server"},
    )
)
```

**Label Strategy**:
- Keep cardinality low (avoid user IDs, request IDs as labels)
- Use dimensions that matter for analysis: server, tool, strategy, status
- Server and tool typically <50 unique values each
- Total unique label combinations: <10K (well within Prometheus limits)

**Performance Impact**:
- Metric collection: <0.5ms per request
- Use fast paths (no string formatting in hot path)
- Pre-allocate metric label sets
- Async metric updates where possible

**Custom Collectors**:
```go
type FilteringCollector struct {
    desc *prometheus.Desc
}

func (c *FilteringCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- c.desc
}

func (c *FilteringCollector) Collect(ch chan<- prometheus.Metric) {
    // Compute current metrics
    ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, value, labels...)
}
```

**Alternatives Considered**:
- **OpenTelemetry**: More complex, we only need metrics (not traces/logs)
- **StatsD**: Requires separate aggregator, Prometheus is pull-based (simpler)
- **Custom metrics format**: Reinventing wheel, Prometheus is standard

---

## Technology Stack Summary

| Component | Technology | Version | Rationale |
|-----------|-----------|---------|-----------|
| Language | Go | 1.21+ | Performance, concurrency, single binary |
| HTTP Server | net/http | stdlib | Production-ready, no framework overhead |
| JSON Parsing | encoding/json | stdlib | Standard, sufficient performance |
| JSON-RPC | Custom | - | Simple protocol, custom implementation best |
| Configuration | yaml.v3 | v3 | Standard YAML parser, lightweight |
| Metrics | prometheus/client_golang | latest | Official client, standard format |
| Testing | testing | stdlib | Built-in, table-driven friendly |
| MCP Protocol | mark3labs/mcp-go | latest | Go SDK for MCP protocol |

**External Dependencies**: Minimal
- `gopkg.in/yaml.v3` - YAML configuration parsing
- `github.com/prometheus/client_golang` - Metrics collection
- `github.com/mark3labs/mcp-go` - MCP protocol support

**Development Dependencies**:
- `golang.org/x/tools/cmd/goimports` - Code formatting
- `github.com/golangci/golangci-lint` - Linting
- Standard Go toolchain for build/test

---

## Design Patterns

### Proxy Pattern
- Server acts as transparent proxy between client and MCP servers
- Intercepts requests/responses for filtering
- Maintains protocol compatibility

### Strategy Pattern
- Multiple filtering strategies (minimal, smart, explicit, passthrough)
- Runtime strategy selection via configuration
- Easy to add new strategies

### Decorator Pattern
- Middleware wraps HTTP handlers for cross-cutting concerns
- Logging, metrics, auth pass-through as decorators

### Observer Pattern
- Metrics collector observes filtering operations
- Logs observer for audit trail
- Decoupled monitoring from core logic

### Factory Pattern
- Analyzer factory creates appropriate strategy analyzer
- Client factory manages downstream connections
- Schema loader factory (server vs. config source)

---

## Tradeoffs & Decisions

### Buffered vs. Streaming JSON
- **Decision**: Buffered (parse full response in memory)
- **Tradeoff**: Simpler code, slightly higher memory vs. streaming complexity
- **Rationale**: 99% of responses <100KB, can optimize later if needed

### In-Memory vs. Persistent Cache
- **Decision**: In-memory with TTL
- **Tradeoff**: Loses cache on restart vs. operational complexity
- **Rationale**: Stateless design, schemas don't change frequently, startup performance acceptable

### Custom vs. Library JSON-RPC
- **Decision**: Custom implementation
- **Tradeoff**: Maintenance burden vs. control and simplicity
- **Rationale**: Protocol is simple, _meta extension needs custom handling anyway

### Sync vs. Async Metrics
- **Decision**: Synchronous metrics collection
- **Tradeoff**: Tiny latency impact vs. complexity of async
- **Rationale**: <0.5ms overhead acceptable, simpler code, no lost metrics

### Configuration Reload: SIGHUP vs. File Watch
- **Decision**: SIGHUP + admin endpoint (file watch optional)
- **Tradeoff**: Manual trigger vs. automatic but complex
- **Rationale**: Explicit operator control preferred, file watch can be added later

---

## Next Steps

With all research complete, proceed to:
1. **Phase 1 Design**: Generate data models and contracts
2. **Validate Approach**: Review with stakeholders if needed
3. **Begin Implementation**: Start with Phase 2 task generation

All major technology decisions are documented and justified. Implementation can proceed with confidence.
