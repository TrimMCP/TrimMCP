# Architecture Overview

Deep dive into how SlimMCP works internally.

## Table of Contents

- [High-Level Architecture](#high-level-architecture)
- [Component Design](#component-design)
- [Request Flow](#request-flow)
- [Filtering Algorithm](#filtering-algorithm)
- [Protocol Extension](#protocol-extension)
- [Performance Characteristics](#performance-characteristics)

## High-Level Architecture

SlimMCP is a transparent proxy that acts as both an MCP server and MCP client simultaneously.

```
┌─────────────────────────────────────────────────────────┐
│                    Your Application                     │
│                    (LLM / Agent)                        │
└──────────────────────────┬──────────────────────────────┘
                           │
                           │ MCP Protocol
                           │ (JSON-RPC 2.0)
                           │
┌──────────────────────────▼──────────────────────────────┐
│                     SlimMCP Proxy                       │
│  ┌───────────────────────────────────────────────────┐  │
│  │           MCP Server Interface                    │  │
│  │         (Northbound - Port 8080)                  │  │
│  │  • Accepts connections from LLM clients           │  │
│  │  • Implements full MCP protocol                   │  │
│  │  • Returns filtered responses                     │  │
│  └─────────────────────┬─────────────────────────────┘  │
│                        │                                 │
│  ┌─────────────────────▼─────────────────────────────┐  │
│  │          Request Analysis Engine                  │  │
│  │  • Parse incoming requests                        │  │
│  │  • Extract field hints from parameters           │  │
│  │  • Check for _meta.fieldSelection                │  │
│  │  • Select filtering strategy                     │  │
│  └─────────────────────┬─────────────────────────────┘  │
│                        │                                 │
│  ┌─────────────────────▼─────────────────────────────┐  │
│  │      Server Capability Manager                    │  │
│  │  • Detect native field selection support         │  │
│  │  • Cache server capabilities                     │  │
│  │  • Route: native filtering vs proxy filtering   │  │
│  └─────────────────────┬─────────────────────────────┘  │
│                        │                                 │
│  ┌─────────────────────▼─────────────────────────────┐  │
│  │           Schema Manager                          │  │
│  │  • Load MCP server tool schemas                  │  │
│  │  • Understand field types & relationships        │  │
│  │  • Validate filtering operations                 │  │
│  │  • Maintain field dependency graph               │  │
│  └─────────────────────┬─────────────────────────────┘  │
│                        │                                 │
│  ┌─────────────────────▼─────────────────────────────┐  │
│  │           MCP Client Interface                    │  │
│  │         (Southbound - Multiple Servers)           │  │
│  │  • Forwards requests to downstream               │  │
│  │  • Manages connection pools                      │  │
│  │  • Handles timeouts and retries                  │  │
│  └─────────────────────┬─────────────────────────────┘  │
│                        │                                 │
│                        │ Full MCP Request                │
│                        │                                 │
└────────────────────────┼─────────────────────────────────┘
                         │
                         ↓
        ┌────────────────────────────────────┐
        │    Downstream MCP Servers          │
        │  • Notion MCP Server               │
        │  • Slack MCP Server                │
        │  • Google Drive MCP Server         │
        │  • etc.                            │
        └────────────────┬───────────────────┘
                         │
                         │ Full MCP Response
                         │
┌────────────────────────▼─────────────────────────────────┐
│                     SlimMCP Proxy                        │
│  ┌───────────────────────────────────────────────────┐   │
│  │          Response Filter Engine                   │   │
│  │  • Receive full response                          │   │
│  │  • Apply field selection rules                    │   │
│  │  • Preserve structural integrity                  │   │
│  │  • Calculate token savings                        │   │
│  │  • Add _meta with statistics                      │   │
│  └─────────────────────┬─────────────────────────────┘   │
│                        │                                  │
│  ┌─────────────────────▼─────────────────────────────┐   │
│  │       Monitoring & Metrics                        │   │
│  │  • Track token savings                            │   │
│  │  • Record performance metrics                     │   │
│  │  • Update Prometheus counters                     │   │
│  │  • Log filtering decisions                        │   │
│  └─────────────────────┬─────────────────────────────┘   │
└────────────────────────┼──────────────────────────────────┘
                         │
                         │ Filtered MCP Response
                         │
┌────────────────────────▼──────────────────────────────────┐
│                    Your Application                       │
│                    (LLM / Agent)                          │
│  • Receives filtered response                             │
│  • Uses less context window                               │
│  • Saves tokens and money                                 │
└───────────────────────────────────────────────────────────┘
```

## Component Design

### 1. MCP Server Interface

**Responsibility:** Present as an MCP server to clients

```go
type MCPServer struct {
    listenAddr string
    router     *mux.Router
    analyzer   *RequestAnalyzer
}

func (s *MCPServer) HandleToolCall(w http.ResponseWriter, r *http.Request) {
    // 1. Parse MCP request
    var req MCPRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // 2. Route to analyzer
    response := s.analyzer.Process(req)
    
    // 3. Return filtered response
    json.NewEncoder(w).Encode(response)
}
```

**Endpoints:**
- `POST /` - Handle all MCP requests (standard)
- `POST /{server}` - Route to specific downstream server
- `GET /health` - Health check endpoint
- `GET /admin/config` - Configuration endpoint (optional)

### 2. Request Analysis Engine

**Responsibility:** Determine which fields are needed

```go
type RequestAnalyzer struct {
    strategies map[string]FilteringStrategy
    schemas    *SchemaManager
    history    *UsageHistory
}

func (a *RequestAnalyzer) InferFields(req MCPRequest) []string {
    fields := make(map[string]bool)
    
    // 1. Check for explicit field selection
    if meta := req.Meta; meta != nil && meta.FieldSelection != nil {
        for _, field := range meta.FieldSelection.Include {
            fields[field] = true
        }
    }
    
    // 2. Parse request parameters for field names
    for _, param := range req.Params.Arguments {
        if fieldName := extractFieldName(param); fieldName != "" {
            fields[fieldName] = true
        }
    }
    
    // 3. Apply tool-specific defaults
    toolDefaults := a.schemas.GetToolDefaults(req.Params.Name)
    for _, field := range toolDefaults {
        fields[field] = true
    }
    
    // 4. Always include critical fields
    for _, field := range []string{"id", "type", "url"} {
        fields[field] = true
    }
    
    // 5. Smart inference (optional)
    if strategy == "smart" {
        inferred := a.smartInference(req)
        for _, field := range inferred {
            fields[field] = true
        }
    }
    
    // 6. Learning mode (optional)
    if a.history != nil {
        learned := a.history.GetFrequentlyUsed(req.Params.Name)
        for _, field := range learned {
            fields[field] = true
        }
    }
    
    return keys(fields)
}
```

### 3. Server Capability Manager

**Responsibility:** Detect and cache server capabilities

```go
type CapabilityManager struct {
    cache map[string]ServerCapabilities
    mutex sync.RWMutex
}

type ServerCapabilities struct {
    SupportsFieldSelection bool
    Version               string
    Strategies            []string
    LastChecked           time.Time
}

func (m *CapabilityManager) CheckCapabilities(serverURL string) ServerCapabilities {
    // Check cache first
    if caps := m.getFromCache(serverURL); caps != nil {
        return *caps
    }
    
    // Query server for capabilities
    caps := m.queryServer(serverURL)
    
    // Cache result
    m.cache[serverURL] = caps
    
    return caps
}
```

### 4. Schema Manager

**Responsibility:** Understand server schemas and field relationships

```go
type SchemaManager struct {
    schemas map[string]ToolSchema
}

type ToolSchema struct {
    Name       string
    Fields     map[string]FieldInfo
    Required   []string
    Defaults   []string
}

type FieldInfo struct {
    Type         string
    Required     bool
    RelatedTo    []string
    TokenCost    int
}

func (m *SchemaManager) ValidateFieldSelection(
    toolName string,
    fields []string,
) error {
    schema := m.schemas[toolName]
    
    // Check all required fields are included
    for _, required := range schema.Required {
        if !contains(fields, required) {
            return fmt.Errorf("missing required field: %s", required)
        }
    }
    
    // Add related fields
    expanded := m.expandRelatedFields(fields, schema)
    
    return nil
}
```

### 5. Response Filter Engine

**Responsibility:** Apply field selection to responses

```go
type ResponseFilter struct {
    schemas *SchemaManager
}

func (f *ResponseFilter) Filter(
    response interface{},
    keepFields []string,
    schema ToolSchema,
) interface{} {
    switch v := response.(type) {
    case map[string]interface{}:
        return f.filterObject(v, keepFields, schema)
    case []interface{}:
        return f.filterArray(v, keepFields, schema)
    default:
        return response
    }
}

func (f *ResponseFilter) filterObject(
    obj map[string]interface{},
    keepFields []string,
    schema ToolSchema,
) map[string]interface{} {
    filtered := make(map[string]interface{})
    
    for key, value := range obj {
        // Keep if in explicit list
        if contains(keepFields, key) {
            filtered[key] = f.Filter(value, keepFields, schema)
            continue
        }
        
        // Keep if critical field
        if isCriticalField(key) {
            filtered[key] = value
            continue
        }
        
        // Keep if required by schema
        if schema.IsRequired(key) {
            filtered[key] = value
            continue
        }
        
        // Otherwise, drop the field
    }
    
    return filtered
}
```

### 6. Monitoring & Metrics

**Responsibility:** Track performance and savings

```go
type MetricsCollector struct {
    prometheus *prometheus.Registry
    counters   map[string]prometheus.Counter
    histograms map[string]prometheus.Histogram
    gauges     map[string]prometheus.Gauge
}

func (m *MetricsCollector) RecordFiltering(stats FilteringStats) {
    // Update counters
    m.counters["tokens_saved"].Add(float64(stats.TokensSaved))
    m.counters["tokens_original"].Add(float64(stats.TokensOriginal))
    m.counters["requests_total"].Inc()
    
    // Update gauges
    savingsPercent := float64(stats.TokensSaved) / float64(stats.TokensOriginal) * 100
    m.gauges["savings_percentage"].Set(savingsPercent)
    
    // Record timing
    m.histograms["filtering_duration"].Observe(stats.Duration.Seconds())
}
```

## Request Flow

### Complete Request Flow Diagram

```
1. Client Request
   │
   ├─→ Parse JSON-RPC request
   │
   ├─→ Extract tool name and parameters
   │
   ├─→ Check for _meta.fieldSelection
   │
   ├─→ Determine filtering strategy
   │
2. Capability Check
   │
   ├─→ Query capability manager
   │
   ├─→ Does server support native filtering?
   │   │
   │   ├─ YES → Add _meta to request, forward as-is
   │   │
   │   └─ NO → Proceed to field inference
   │
3. Field Inference
   │
   ├─→ Parse request parameters
   │
   ├─→ Extract mentioned field names
   │
   ├─→ Load tool defaults
   │
   ├─→ Apply strategy (smart/minimal/explicit)
   │
   ├─→ Add critical fields (id, type, url)
   │
   ├─→ Expand related fields (if configured)
   │
4. Schema Validation
   │
   ├─→ Load tool schema
   │
   ├─→ Validate field selection
   │
   ├─→ Ensure required fields included
   │
5. Downstream Request
   │
   ├─→ Forward to MCP server (full request)
   │
   ├─→ Wait for response
   │
   ├─→ Handle timeouts/errors
   │
6. Response Processing
   │
   ├─→ Receive full response
   │
   ├─→ If native filtering was used → passthrough
   │
   ├─→ Otherwise → apply filtering
   │
7. Filtering
   │
   ├─→ Recursively filter response
   │
   ├─→ Keep selected fields
   │
   ├─→ Preserve structure
   │
   ├─→ Validate integrity
   │
8. Metrics & Metadata
   │
   ├─→ Calculate token savings
   │
   ├─→ Record metrics
   │
   ├─→ Add _meta to response
   │
   ├─→ Log filtering decision
   │
9. Client Response
   │
   └─→ Return filtered response to client
```

## Filtering Algorithm

### Core Algorithm

```go
func filterResponse(
    data interface{},
    keepFields map[string]bool,
    schema ToolSchema,
) interface{} {
    switch v := data.(type) {
    case map[string]interface{}:
        result := make(map[string]interface{})
        
        for key, value := range v {
            shouldKeep := shouldKeepField(key, keepFields, schema)
            
            if shouldKeep {
                // Recursively filter nested values
                result[key] = filterResponse(value, keepFields, schema)
            }
        }
        
        return result
        
    case []interface{}:
        result := make([]interface{}, len(v))
        
        for i, item := range v {
            result[i] = filterResponse(item, keepFields, schema)
        }
        
        return result
        
    default:
        // Primitive values pass through
        return data
    }
}

func shouldKeepField(
    fieldName string,
    keepFields map[string]bool,
    schema ToolSchema,
) bool {
    // 1. Explicit inclusion
    if keepFields[fieldName] {
        return true
    }
    
    // 2. Critical field
    if isCriticalField(fieldName) {
        return true
    }
    
    // 3. Required by schema
    if schema.IsRequired(fieldName) {
        return true
    }
    
    // 4. Related to kept field
    for keptField := range keepFields {
        if schema.IsRelated(fieldName, keptField) {
            return true
        }
    }
    
    return false
}
```

### Field Categorization

Fields are categorized for filtering decisions:

```go
type FieldCategory int

const (
    Critical FieldCategory = iota  // Always keep (id, type, url)
    Required                        // Required by schema
    Requested                       // Explicitly requested
    Related                         // Related to requested fields
    Optional                        // Can be removed
)

func categorizeField(field string, context FilterContext) FieldCategory {
    // Check in priority order
    if isCriticalField(field) {
        return Critical
    }
    
    if context.Schema.IsRequired(field) {
        return Required
    }
    
    if context.KeepFields[field] {
        return Requested
    }
    
    if isRelatedToRequested(field, context) {
        return Related
    }
    
    return Optional
}
```

## Protocol Extension

### Field Selection Extension

SlimMCP implements the `_meta.fieldSelection` extension:

```typescript
interface MCPRequest {
    jsonrpc: "2.0";
    method: string;
    params: {
        name: string;
        arguments: object;
    };
    id: number;
    
    // Extension
    _meta?: {
        fieldSelection?: {
            include?: string[];
            exclude?: string[];
            strategy?: "smart" | "minimal" | "explicit";
            keepRelated?: boolean;
        }
    };
}
```

### Capability Negotiation

Servers announce field selection support:

```typescript
interface ServerCapabilities {
    fieldSelection?: {
        supported: boolean;
        version: string;
        strategies: string[];
        native: boolean;
    };
}
```

### Response Metadata

Responses include filtering statistics:

```typescript
interface MCPResponse {
    jsonrpc: "2.0";
    result: any;
    id: number;
    
    // Extension
    _meta?: {
        fieldSelection?: {
            applied: boolean;
            strategy: string;
            filteredBy: "proxy" | "server";
            originalSize: number;
            filteredSize: number;
            savings: number;
            fields: {
                kept: string[];
                removed: number;
            };
        }
    };
}
```

## Performance Characteristics

### Time Complexity

| Operation | Complexity | Notes |
|-----------|------------|-------|
| Field inference | O(n) | n = number of request parameters |
| Schema validation | O(m) | m = number of fields |
| Response filtering | O(k * d) | k = response size, d = depth |
| Capability check | O(1) | Cached after first check |

### Space Complexity

| Component | Memory Usage |
|-----------|--------------|
| Schema cache | ~1KB per tool |
| Capability cache | ~100 bytes per server |
| Learning history | ~10MB (configurable) |
| Connection pools | ~50KB per server |
| **Total baseline** | **<10MB** |

### Latency Impact

| Strategy | Added Latency (p50) | Added Latency (p95) |
|----------|---------------------|---------------------|
| Passthrough | <1ms | <2ms |
| Minimal | 2-4ms | 6-8ms |
| Smart | 4-8ms | 10-15ms |
| Learning | 3-6ms | 8-12ms |

### Throughput

Single instance can handle:
- **1,000+ requests/sec** (minimal strategy)
- **500-800 requests/sec** (smart strategy)
- **300-500 requests/sec** (learning strategy)

Horizontally scalable for higher throughput.

## Scalability

### Horizontal Scaling

SlimMCP is stateless and can be easily scaled:

```yaml
# Load balancer configuration
upstream slimmcp {
    server slimmcp-1:8080;
    server slimmcp-2:8080;
    server slimmcp-3:8080;
    
    least_conn;
}
```

### Vertical Scaling

Resource usage per instance:

| Load | CPU | Memory | Network |
|------|-----|--------|---------|
| Idle | <1% | ~50MB | ~1Mbps |
| 100 req/s | 5-10% | ~150MB | ~10Mbps |
| 1000 req/s | 30-50% | ~300MB | ~100Mbps |

## Security Considerations

### Authentication Pass-through

SlimMCP does not authenticate clients:

```
Client → [Auth Header] → SlimMCP → [Auth Header] → MCP Server
                        (passthrough)
```

### Data Privacy

- No credential storage
- No persistent data
- Optional request/response logging
- PII can be excluded from logs

### Network Security

- Optional TLS termination
- Client certificate validation
- IP allowlist/denylist (optional)

## See Also

- [Configuration Reference](configuration.md)
- [API Documentation](api.md)
- [Performance Benchmarks](benchmarks.md)
- [Protocol Extension Spec](protocol-extension.md)
