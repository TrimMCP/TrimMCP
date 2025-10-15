# Data Model: MCP Response Optimizer Proxy

**Feature**: MCP Response Optimizer Proxy
**Branch**: `001-trim`
**Date**: 2025-10-13

## Overview

This document defines all data structures used in SlimMCP. Structures are organized by domain: Request/Response (protocol), Domain Logic (filtering), Configuration, and Metrics.

---

## Request Entities

### MCPRequest

Represents an incoming JSON-RPC 2.0 request from an LLM client.

**Structure**:
```go
type MCPRequest struct {
    JSONRPC string          `json:"jsonrpc"` // Always "2.0"
    Method  string          `json:"method"`  // MCP method (e.g., "tools/call")
    Params  *RequestParams  `json:"params,omitempty"`
    ID      interface{}     `json:"id,omitempty"` // string, number, or null
    Meta    *RequestMeta    `json:"_meta,omitempty"` // SlimMCP extension
}
```

**Fields**:
- `JSONRPC`: Protocol version (must be "2.0")
- `Method`: MCP method name
  - `tools/list`: List available tools
  - `tools/call`: Invoke a tool
  - `resources/list`: List resources
  - `resources/read`: Read a resource
- `Params`: Method-specific parameters
- `ID`: Request identifier for correlation (required for requests expecting response)
- `Meta`: Optional SlimMCP-specific metadata

**Validation Rules**:
- `JSONRPC` must equal "2.0"
- `Method` must be non-empty string
- `ID` required for non-notification requests
- `Params` structure depends on method

---

### RequestParams

Method-specific parameters for tool invocation.

**Structure**:
```go
type RequestParams struct {
    Name      string                 `json:"name"`      // Tool name
    Arguments map[string]interface{} `json:"arguments"` // Tool arguments
}
```

**Fields**:
- `Name`: Name of the tool to invoke (e.g., "notion-fetch")
- `Arguments`: Tool-specific arguments as key-value pairs

**Examples**:
```json
{
  "name": "notion-fetch",
  "arguments": {
    "id": "page-123"
  }
}
```

---

### RequestMeta

SlimMCP-specific metadata for request processing.

**Structure**:
```go
type RequestMeta struct {
    FieldSelection *FieldSelectionSpec `json:"fieldSelection,omitempty"`
}
```

**Fields**:
- `FieldSelection`: Optional field selection configuration

---

### FieldSelectionSpec

Specifies which fields to include/exclude in responses.

**Structure**:
```go
type FieldSelectionSpec struct {
    Include      []string `json:"include,omitempty"`      // Fields to explicitly include
    Exclude      []string `json:"exclude,omitempty"`      // Fields to explicitly exclude
    Strategy     string   `json:"strategy,omitempty"`     // Filtering strategy
    KeepRelated  bool     `json:"keepRelated,omitempty"`  // Include related fields
    MaxDepth     int      `json:"maxDepth,omitempty"`     // Maximum nesting depth
}
```

**Fields**:
- `Include`: Whitelist of field names (supports dot notation: "properties.status")
- `Exclude`: Blacklist of field names
- `Strategy`: One of "smart", "minimal", "explicit", "passthrough" (defaults to proxy config)
- `KeepRelated`: If true, include fields related to requested fields (default: true)
- `MaxDepth`: Maximum nesting level to filter (default: 10, max: 20)

**Validation Rules**:
- Cannot specify both `Include` and `Exclude`
- Strategy must be valid enum value
- Field names in Include/Exclude must be valid identifiers or dot-notation paths
- MaxDepth between 1 and 20

**Examples**:
```json
{
  "include": ["title", "status", "assignee.name"],
  "strategy": "explicit",
  "keepRelated": true,
  "maxDepth": 10
}
```

---

## Response Entities

### MCPResponse

Represents a JSON-RPC 2.0 response to the client.

**Structure**:
```go
type MCPResponse struct {
    JSONRPC string          `json:"jsonrpc"` // Always "2.0"
    Result  json.RawMessage `json:"result,omitempty"`
    Error   *JSONRPCError   `json:"error,omitempty"`
    ID      interface{}     `json:"id"`
    Meta    *ResponseMeta   `json:"_meta,omitempty"` // SlimMCP extension
}
```

**Fields**:
- `JSONRPC`: Protocol version (always "2.0")
- `Result`: Successful response data (filtered if applicable)
- `Error`: Error information if request failed
- `ID`: Request ID for correlation
- `Meta`: SlimMCP-specific response metadata

**Validation Rules**:
- Exactly one of `Result` or `Error` must be present
- `ID` must match request ID
- `JSONRPC` must be "2.0"

---

### JSONRPCError

Standard JSON-RPC 2.0 error structure.

**Structure**:
```go
type JSONRPCError struct {
    Code    int             `json:"code"`
    Message string          `json:"message"`
    Data    json.RawMessage `json:"data,omitempty"`
}
```

**Standard Error Codes**:
- `-32700`: Parse error (Invalid JSON)
- `-32600`: Invalid Request
- `-32601`: Method not found
- `-32602`: Invalid params
- `-32603`: Internal error

**Custom Error Codes** (SlimMCP-specific):
- `-32000`: Downstream server error
- `-32001`: Filtering error
- `-32002`: Schema validation error
- `-32003`: Timeout error

---

### ResponseMeta

SlimMCP-specific response metadata.

**Structure**:
```go
type ResponseMeta struct {
    FieldSelection *FilteringMetadata `json:"fieldSelection,omitempty"`
    Version        string             `json:"version,omitempty"`        // SlimMCP version
    ProtocolVersion string            `json:"protocolVersion,omitempty"` // MCP protocol version
}
```

**Fields**:
- `FieldSelection`: Filtering statistics and metadata
- `Version`: SlimMCP proxy version (e.g., "0.1.0")
- `ProtocolVersion`: MCP protocol version (e.g., "1.0")

---

### FilteringMetadata

Statistics and information about response filtering.

**Structure**:
```go
type FilteringMetadata struct {
    Applied       bool     `json:"applied"`              // Was filtering applied?
    Strategy      string   `json:"strategy"`             // Strategy used
    FilteredBy    string   `json:"filteredBy"`           // "proxy" or "server"
    OriginalSize  int      `json:"originalSize"`         // Original response size (tokens)
    FilteredSize  int      `json:"filteredSize"`         // Filtered response size (tokens)
    Savings       float64  `json:"savings"`              // Savings percentage
    FieldsKept    []string `json:"fields.kept"`          // Fields that were kept
    FieldsRemoved int      `json:"fields.removed"`       // Number of fields removed
    FilteringTime float64  `json:"filteringTime"`        // Time spent filtering (ms)
}
```

**Fields**:
- `Applied`: Whether filtering was actually applied
- `Strategy`: Which strategy was used
- `FilteredBy`: Whether proxy or downstream server did filtering
- `OriginalSize`: Token count before filtering
- `FilteredSize`: Token count after filtering
- `Savings`: Percentage reduction ((original - filtered) / original * 100)
- `FieldsKept`: List of field names included in response
- `FieldsRemoved`: Count of fields removed
- `FilteringTime`: Milliseconds spent on filtering operation

**Example**:
```json
{
  "applied": true,
  "strategy": "smart",
  "filteredBy": "proxy",
  "originalSize": 2847,
  "filteredSize": 421,
  "savings": 85.2,
  "fields": {
    "kept": ["id", "title", "status", "url", "type"],
    "removed": 40
  },
  "filteringTime": 7.3
}
```

---

## Domain Entities

### ToolSchema

Describes a tool's structure, fields, and validation rules.

**Structure**:
```go
type ToolSchema struct {
    Name        string               `json:"name"`
    Description string               `json:"description"`
    Fields      map[string]FieldInfo `json:"fields"`
    Required    []string             `json:"required"`
    Defaults    []string             `json:"defaults"` // Default fields for this tool
}
```

**Fields**:
- `Name`: Tool identifier (e.g., "notion-fetch")
- `Description`: Human-readable tool description
- `Fields`: Map of field name to field metadata
- `Required`: List of required field names
- `Defaults`: Default fields to include if no explicit selection

**Source**:
- Fetched from MCP server via `tools/list` method
- Can be overridden in configuration
- Cached with 1-hour TTL

---

### FieldInfo

Metadata about a specific field in a tool response.

**Structure**:
```go
type FieldInfo struct {
    Type       string   `json:"type"`       // Field type (string, number, object, array)
    Required   bool     `json:"required"`   // Is this field required?
    RelatedTo  []string `json:"relatedTo"`  // Related field names
    TokenCost  int      `json:"tokenCost"`  // Estimated token cost of this field
    Critical   bool     `json:"critical"`   // Is this a critical field (id, url, etc)?
}
```

**Fields**:
- `Type`: JSON Schema type
- `Required`: Whether field must always be present
- `RelatedTo`: Other fields that depend on this field
- `TokenCost`: Approximate token cost (for savings calculation)
- `Critical`: Whether this is a critical field (always kept)

**Critical Fields** (always included):
- `id`: Unique identifier
- `type`: Resource type
- `url`: Resource URL
- `title`: Display title (if present)

---

### ServerCapabilities

Information about what a downstream MCP server supports.

**Structure**:
```go
type ServerCapabilities struct {
    URL                      string    `json:"url"`
    SupportsFieldSelection   bool      `json:"supportsFieldSelection"`
    Version                  string    `json:"version"`
    ProtocolVersion          string    `json:"protocolVersion"`
    SupportedStrategies      []string  `json:"supportedStrategies"`
    LastChecked              time.Time `json:"lastChecked"`
    CacheTTL                 time.Duration `json:"-"`
}
```

**Fields**:
- `URL`: Server base URL
- `SupportsFieldSelection`: Does server natively support field selection?
- `Version`: Server software version
- `ProtocolVersion`: MCP protocol version
- `SupportedStrategies`: List of strategies server supports natively
- `LastChecked`: When capabilities were last queried
- `CacheTTL`: How long to cache (not serialized)

**Detection**:
- Query server's capabilities endpoint on first request
- Cache for 1 hour
- Re-check if server returns error indicating unsupported feature

---

### FilterContext

Combined context for making filtering decisions.

**Structure**:
```go
type FilterContext struct {
    Request       *MCPRequest
    Response      json.RawMessage
    Schema        *ToolSchema
    Strategy      string
    KeepFields    map[string]bool // Set of fields to keep
    ExcludeFields map[string]bool // Set of fields to exclude
    MaxDepth      int
    KeepRelated   bool
}
```

**Fields**:
- `Request`: Original client request
- `Response`: Raw response from downstream server
- `Schema`: Tool schema (if available)
- `Strategy`: Active filtering strategy
- `KeepFields`: Computed set of fields to include
- `ExcludeFields`: Computed set of fields to exclude
- `MaxDepth`: Maximum nesting depth to process
- `KeepRelated`: Whether to include related fields

**Usage**:
- Created by analyzer after request analysis
- Passed to filter engine
- Contains all information needed for filtering decision

---

### FieldCategory

Enum for categorizing fields during filtering.

**Structure**:
```go
type FieldCategory int

const (
    FieldCategoryCritical  FieldCategory = iota // Always keep (id, type, url)
    FieldCategoryRequired                       // Required by schema
    FieldCategoryRequested                      // Explicitly requested
    FieldCategoryRelated                        // Related to requested field
    FieldCategoryOptional                       // Can be removed
)
```

**Values**:
- `Critical`: Must always be present (id, type, url, title)
- `Required`: Schema marks as required
- `Requested`: Client explicitly asked for it
- `Related`: Inferred as related to requested field
- `Optional`: Can be safely removed

**Decision Priority**:
1. Critical → always keep
2. Required → always keep
3. Requested → always keep
4. Related → keep if `KeepRelated` true
5. Optional → remove

---

## Configuration Entities

### Config

Top-level configuration structure.

**Structure**:
```go
type Config struct {
    Proxy        ProxyConfig        `yaml:"proxy"`
    Optimization OptimizationConfig `yaml:"optimization"`
    Monitoring   MonitoringConfig   `yaml:"monitoring"`
    Learning     *LearningConfig    `yaml:"learning,omitempty"`
}
```

---

### ProxyConfig

HTTP proxy server configuration.

**Structure**:
```go
type ProxyConfig struct {
    ListenAddress string                 `yaml:"listenAddress"` // e.g., "0.0.0.0:8080"
    Strategy      string                 `yaml:"strategy"`      // Default strategy
    Servers       map[string]ServerConfig `yaml:"servers"`       // Downstream servers
    CORS          *CORSConfig            `yaml:"cors,omitempty"`
    TLS           *TLSConfig             `yaml:"tls,omitempty"`
}
```

**Fields**:
- `ListenAddress`: Host:port to listen on (default: "0.0.0.0:8080")
- `Strategy`: Default filtering strategy (default: "smart")
- `Servers`: Map of server name to configuration
- `CORS`: CORS policy configuration (optional)
- `TLS`: TLS certificate configuration (optional)

---

### ServerConfig

Configuration for a specific downstream MCP server.

**Structure**:
```go
type ServerConfig struct {
    URL             string            `yaml:"url"`                       // Server URL
    Timeout         time.Duration     `yaml:"timeout"`                   // Request timeout
    Strategy        string            `yaml:"strategy,omitempty"`        // Override global strategy
    Aggressiveness  string            `yaml:"aggressiveness,omitempty"`  // "low", "medium", "high"
    Defaults        map[string][]string `yaml:"defaults,omitempty"`      // Per-tool defaults
}
```

**Fields**:
- `URL`: Base URL of MCP server
- `Timeout`: Request timeout (default: 30s)
- `Strategy`: Override global strategy for this server
- `Aggressiveness`: How aggressive to filter ("low", "medium", "high")
- `Defaults`: Default fields per tool (e.g., `notion-fetch: [id, title, url]`)

---

### OptimizationConfig

Filtering behavior configuration.

**Structure**:
```go
type OptimizationConfig struct {
    AlwaysKeep   []string      `yaml:"alwaysKeep"`   // Always include these fields
    AlwaysRemove []string      `yaml:"alwaysRemove"` // Always exclude these fields
    Safety       SafetyConfig  `yaml:"safety"`       // Safety settings
}
```

**Fields**:
- `AlwaysKeep`: Global whitelist (e.g., ["id", "type", "url", "title"])
- `AlwaysRemove`: Global blacklist (e.g., ["icon", "cover", "created_by"])
- `Safety`: Error handling and fallback configuration

---

### SafetyConfig

Safety and fallback settings.

**Structure**:
```go
type SafetyConfig struct {
    FallbackOnError bool                 `yaml:"fallbackOnError"` // Return full response on error
    CircuitBreaker  CircuitBreakerConfig `yaml:"circuitBreaker"`  // Circuit breaker settings
}
```

**Fields**:
- `FallbackOnError`: If filtering fails, return full unfiltered response (default: true)
- `CircuitBreaker`: Circuit breaker configuration for downstream servers

---

### CircuitBreakerConfig

Circuit breaker settings for fault tolerance.

**Structure**:
```go
type CircuitBreakerConfig struct {
    Enabled   bool          `yaml:"enabled"`   // Enable circuit breaker
    Threshold int           `yaml:"threshold"` // Failures before opening
    Timeout   time.Duration `yaml:"timeout"`   // Time to wait before retry
}
```

**Fields**:
- `Enabled`: Enable circuit breaker (default: true)
- `Threshold`: Number of consecutive failures before opening circuit (default: 10)
- `Timeout`: How long to wait before attempting reset (default: 60s)

---

### MonitoringConfig

Observability configuration.

**Structure**:
```go
type MonitoringConfig struct {
    Enabled        bool                  `yaml:"enabled"`        // Enable monitoring
    MetricsPort    int                   `yaml:"metricsPort"`    // Prometheus metrics port
    LogLevel       string                `yaml:"logLevel"`       // Log level
    Prometheus     PrometheusConfig      `yaml:"prometheus"`     // Prometheus settings
    RequestLogging RequestLoggingConfig  `yaml:"requestLogging"` // Request logging settings
}
```

**Fields**:
- `Enabled`: Enable monitoring (default: true)
- `MetricsPort`: Port for /metrics endpoint (default: 9090)
- `LogLevel`: Logging level ("debug", "info", "warn", "error")
- `Prometheus`: Prometheus-specific settings
- `RequestLogging`: Request/response logging configuration

---

### CORSConfig

CORS policy configuration.

**Structure**:
```go
type CORSConfig struct {
    Enabled      bool     `yaml:"enabled"`
    AllowOrigins []string `yaml:"allowOrigins"`
    AllowMethods []string `yaml:"allowMethods"`
    AllowHeaders []string `yaml:"allowHeaders"`
}
```

---

### TLSConfig

TLS certificate configuration.

**Structure**:
```go
type TLSConfig struct {
    Enabled  bool   `yaml:"enabled"`
    CertFile string `yaml:"certFile"`
    KeyFile  string `yaml:"keyFile"`
}
```

---

## Metrics Entities

### FilteringStatistics

Real-time filtering performance statistics.

**Structure**:
```go
type FilteringStatistics struct {
    TokensOriginal  int           // Original response tokens
    TokensFiltered  int           // Filtered response tokens
    TokensSaved     int           // Tokens saved
    SavingsPercent  float64       // Savings percentage
    FieldsKept      int           // Number of fields kept
    FieldsRemoved   int           // Number of fields removed
    Duration        time.Duration // Filtering duration
    Strategy        string        // Strategy used
    Server          string        // Server name
    Tool            string        // Tool name
}
```

**Usage**:
- Collected during each filtering operation
- Sent to metrics collector
- Used for Prometheus metrics and logging

---

## Data Flow

### Request Processing Flow

```
1. Client → MCPRequest
2. Proxy receives request
3. Extract _meta.fieldSelection (if present)
4. Create FilterContext
   - Analyze request parameters
   - Load tool schema
   - Infer needed fields
   - Determine strategy
5. Forward request to downstream server
6. Receive MCPResponse from server
7. Apply filtering with FilterContext
8. Calculate FilteringStatistics
9. Add ResponseMeta to response
10. Return filtered MCPResponse to client
11. Record metrics
```

### Configuration Flow

```
1. Load Config from YAML file
2. Validate configuration
3. Initialize components:
   - Create HTTP server with ProxyConfig
   - Initialize analyzers with strategies
   - Set up metrics collectors
   - Configure downstream clients
4. Start server
5. On SIGHUP or /admin/reload:
   - Reload Config
   - Revalidate
   - Hot-swap components
```

---

## Validation Rules Summary

| Entity | Required Fields | Constraints |
|--------|----------------|-------------|
| MCPRequest | jsonrpc, method, id | jsonrpc="2.0", method non-empty |
| FieldSelectionSpec | (none, all optional) | strategy ∈ {smart, minimal, explicit, passthrough} |
| MCPResponse | jsonrpc, id, (result XOR error) | Exactly one of result or error |
| ToolSchema | name, fields | name non-empty |
| ServerConfig | url | url must be valid HTTP(S) URL |
| Config | proxy, optimization, monitoring | Complete validation on load |

---

## Next Steps

With data model defined:
1. Generate JSON schemas in `contracts/` directory
2. Create example configuration files
3. Begin implementation of core types in `pkg/types/`
4. Write validation tests for all entities
