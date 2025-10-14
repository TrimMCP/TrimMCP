# Configuration Reference

Complete guide to configuring SlimMCP for your needs.

## Configuration File

SlimMCP uses YAML for configuration. By default, it looks for `slim.yaml` in the current directory, but you can specify a different file:

```bash
slim start --config /path/to/config.yaml
```

## Complete Configuration Example

```yaml
# Server configuration
proxy:
  # Address and port to listen on
  listenAddress: "0.0.0.0:8080"
  
  # Read timeout for requests
  readTimeout: 30s
  
  # Write timeout for responses
  writeTimeout: 30s
  
  # Idle timeout for connections
  idleTimeout: 120s
  
  # Default filtering strategy
  strategy: smart  # smart | minimal | explicit | learning | passthrough
  
  # Downstream MCP servers
  servers:
    notion:
      url: "http://localhost:3000"
      timeout: 30s
      strategy: smart  # Override global strategy
      aggressiveness: medium  # low | medium | high
      
      # Per-tool field defaults
      defaults:
        notion-fetch:
          fields: [id, title, url, status, properties]
        notion-search:
          fields: [id, title, url]
        notion-create-pages:
          fields: [id, url]
      
      # Enable native support detection
      detectNativeSupport: true
      
      # Connection pool settings
      maxConnections: 100
      maxIdleConnections: 10
      
    slack:
      url: "http://localhost:3001"
      timeout: 30s
      strategy: minimal
      aggressiveness: low

# Optimization settings
optimization:
  # Global field rules
  alwaysKeep:
    - id
    - type
    - url
    - title
  
  alwaysRemove:
    - icon
    - cover
    - created_by
    - last_edited_by
    - created_time
    - last_edited_time
  
  # Token estimation
  tokenEstimation:
    enabled: true
    charactersPerToken: 4
  
  # Learning mode
  learning:
    enabled: false
    historySize: 10000
    updateInterval: 1h
    minSamplesForLearning: 100
  
  # Safety settings
  safety:
    # Always return full response on error
    fallbackOnError: true
    
    # Circuit breaker configuration
    circuitBreaker:
      enabled: true
      threshold: 10        # errors before opening
      timeout: 60s         # recovery time
      halfOpenRequests: 3  # test requests in half-open
    
    # Max response size (bytes)
    maxResponseSize: 10485760  # 10MB
    
    # Enable request validation
    validateRequests: true

# Monitoring and observability
monitoring:
  enabled: true
  
  # Prometheus metrics
  metricsPort: 9090
  prometheus:
    enabled: true
    path: /metrics
    
  # Logging
  logLevel: info  # debug | info | warn | error
  logFormat: json  # json | text
  logOutput: stdout  # stdout | file
  logFile: /var/log/slimmcp/slimmcp.log
  
  # Request logging
  requestLogging:
    enabled: true
    logHeaders: false
    logBodies: false  # Be careful with PII
    logMetadata: true
  
  # Health check endpoint
  healthCheck:
    enabled: true
    path: /health

# TLS configuration (optional)
tls:
  enabled: false
  certFile: /path/to/cert.pem
  keyFile: /path/to/key.pem
  clientAuth: false  # Require client certificates

# Development mode
dev:
  enabled: false
  prettyPrint: true
  verboseLogging: true
  disableCircuitBreaker: true
```

## Configuration Sections

### Proxy Configuration

#### `proxy.listenAddress`

The address and port SlimMCP listens on.

```yaml
proxy:
  listenAddress: "0.0.0.0:8080"
```

- Default: `"0.0.0.0:8080"`
- Format: `"host:port"`
- Examples:
  - `"0.0.0.0:8080"` - Listen on all interfaces
  - `"127.0.0.1:8080"` - Localhost only
  - `":8080"` - All interfaces (short form)

#### `proxy.strategy`

Global filtering strategy applied to all servers unless overridden.

```yaml
proxy:
  strategy: smart
```

**Options:**

- `smart` (default) - Analyzes requests to infer needed fields
- `minimal` - Returns only critical fields (id, title, url, type)
- `explicit` - Only uses fields specified in `_meta.fieldSelection`
- `learning` - Learns from usage patterns over time
- `passthrough` - No filtering (bypass mode)

**When to use each:**

| Strategy | Use Case | Savings | Accuracy |
|----------|----------|---------|----------|
| `smart` | General purpose | 70-85% | High |
| `minimal` | Maximum savings | 85-95% | Medium |
| `explicit` | Full control | Varies | Perfect |
| `learning` | Long-running production | 75-90% | Very High |
| `passthrough` | Debugging | 0% | N/A |

#### `proxy.servers`

Configure downstream MCP servers.

```yaml
proxy:
  servers:
    notion:
      url: "http://localhost:3000"
      timeout: 30s
      strategy: smart
      aggressiveness: medium
```

**Server Options:**

- `url` (required) - Downstream server URL
- `timeout` - Request timeout (default: 30s)
- `strategy` - Override global strategy
- `aggressiveness` - How aggressive to filter: `low`, `medium`, `high`
- `defaults` - Per-tool field defaults
- `detectNativeSupport` - Auto-detect native filtering (default: true)
- `maxConnections` - Connection pool size (default: 100)
- `maxIdleConnections` - Idle connection limit (default: 10)

### Optimization Configuration

#### `optimization.alwaysKeep`

Fields that are always included in responses.

```yaml
optimization:
  alwaysKeep:
    - id
    - type
    - url
    - title
```

**Common fields to always keep:**
- `id` - Object identifiers
- `type` - Type discriminators
- `url` - Resource URLs
- `title` - Human-readable names

#### `optimization.alwaysRemove`

Fields that are removed unless explicitly requested.

```yaml
optimization:
  alwaysRemove:
    - icon
    - cover
    - created_by
    - last_edited_by
    - created_time
    - last_edited_time
```

**Common fields to remove:**
- Timestamps (unless needed)
- User metadata
- Icons and images
- Large text fields (descriptions)

#### `optimization.learning`

Enable learning mode to adapt filtering based on actual usage.

```yaml
optimization:
  learning:
    enabled: true
    historySize: 10000
    updateInterval: 1h
    minSamplesForLearning: 100
```

- `enabled` - Enable learning mode
- `historySize` - Number of requests to track
- `updateInterval` - How often to update learned patterns
- `minSamplesForLearning` - Minimum samples before applying learned patterns

#### `optimization.safety`

Safety and fallback configuration.

```yaml
optimization:
  safety:
    fallbackOnError: true
    circuitBreaker:
      enabled: true
      threshold: 10
      timeout: 60s
    maxResponseSize: 10485760
```

**Circuit Breaker:**

The circuit breaker prevents cascading failures:

1. **Closed** (normal) - Requests pass through
2. **Open** (tripped) - Fast-fail after threshold errors
3. **Half-Open** (testing) - Allow test requests to check recovery

### Monitoring Configuration

#### `monitoring.logLevel`

Set logging verbosity.

```yaml
monitoring:
  logLevel: info
```

**Levels:**
- `debug` - Verbose, all details
- `info` - General operational messages
- `warn` - Warnings and potential issues
- `error` - Errors only

**Recommendation:** Use `info` for production, `debug` for troubleshooting.

#### `monitoring.prometheus`

Enable Prometheus metrics.

```yaml
monitoring:
  prometheus:
    enabled: true
    path: /metrics
  metricsPort: 9090
```

Access metrics at: `http://localhost:9090/metrics`

**Available Metrics:**

| Metric | Type | Description |
|--------|------|-------------|
| `mcp_optimizer_tokens_saved_total` | Counter | Total tokens saved |
| `mcp_optimizer_tokens_original_total` | Counter | Total original tokens |
| `mcp_optimizer_savings_percentage` | Gauge | Current savings % |
| `mcp_optimizer_requests_total` | Counter | Total requests |
| `mcp_optimizer_errors_total` | Counter | Total errors |
| `mcp_optimizer_filtering_duration_seconds` | Histogram | Filtering time |
| `mcp_optimizer_request_duration_seconds` | Histogram | Total request time |

## Environment Variables

Override configuration with environment variables:

```bash
# Server settings
SLIM_LISTEN_ADDRESS=0.0.0.0:8080
SLIM_STRATEGY=smart

# Downstream server
SLIM_SERVER_NOTION_URL=http://localhost:3000
SLIM_SERVER_SLACK_URL=http://localhost:3001

# Monitoring
SLIM_METRICS_PORT=9090
SLIM_LOG_LEVEL=info

# Safety
SLIM_FALLBACK_ON_ERROR=true
SLIM_CIRCUIT_BREAKER_ENABLED=true
```

Format: `SLIM_<SECTION>_<KEY>=<VALUE>`

Nested keys use underscores: `SLIM_OPTIMIZATION_SAFETY_FALLBACK_ON_ERROR`

## Configuration Profiles

### Minimal Configuration

For quick testing:

```yaml
proxy:
  listenAddress: ":8080"
  servers:
    notion:
      url: "http://localhost:3000"
```

### Development Configuration

For local development:

```yaml
proxy:
  listenAddress: "127.0.0.1:8080"
  strategy: smart
  servers:
    notion:
      url: "http://localhost:3000"

monitoring:
  enabled: true
  logLevel: debug
  
dev:
  enabled: true
  prettyPrint: true
  verboseLogging: true
```

### Production Configuration

For production deployment:

```yaml
proxy:
  listenAddress: "0.0.0.0:8080"
  readTimeout: 30s
  writeTimeout: 30s
  idleTimeout: 120s
  strategy: smart
  
  servers:
    notion:
      url: "http://notion-mcp:3000"
      timeout: 30s
      maxConnections: 100
    slack:
      url: "http://slack-mcp:3001"
      timeout: 30s
      maxConnections: 100

optimization:
  alwaysKeep: [id, type, url, title]
  safety:
    fallbackOnError: true
    circuitBreaker:
      enabled: true
      threshold: 10
      timeout: 60s
    maxResponseSize: 10485760

monitoring:
  enabled: true
  metricsPort: 9090
  logLevel: info
  logFormat: json
  prometheus:
    enabled: true
  requestLogging:
    enabled: true
    logMetadata: true
  healthCheck:
    enabled: true

tls:
  enabled: true
  certFile: /certs/server.crt
  keyFile: /certs/server.key
```

### High-Performance Configuration

For maximum throughput:

```yaml
proxy:
  listenAddress: "0.0.0.0:8080"
  strategy: minimal  # Fastest filtering
  
  servers:
    notion:
      url: "http://localhost:3000"
      timeout: 10s
      maxConnections: 200
      maxIdleConnections: 50

optimization:
  tokenEstimation:
    enabled: false  # Disable for speed
  learning:
    enabled: false
  safety:
    fallbackOnError: false  # Fail fast
    maxResponseSize: 5242880  # 5MB limit

monitoring:
  logLevel: warn  # Minimal logging
  requestLogging:
    enabled: false
```

## Docker Configuration

### Docker Compose

```yaml
version: '3.8'

services:
  slimmcp:
    image: slimmcp/proxy:latest
    ports:
      - "8080:8080"
      - "9090:9090"
    volumes:
      - ./slim.yaml:/slim.yaml:ro
      - ./logs:/var/log/slimmcp
    environment:
      - SLIM_LOG_LEVEL=info
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

### Kubernetes ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: slimmcp-config
data:
  slim.yaml: |
    proxy:
      listenAddress: "0.0.0.0:8080"
      strategy: smart
      servers:
        notion:
          url: "http://notion-mcp-service:3000"
    
    monitoring:
      enabled: true
      metricsPort: 9090
```

## Advanced Topics

### Custom Field Rules per Tool

```yaml
proxy:
  servers:
    notion:
      url: "http://localhost:3000"
      defaults:
        notion-fetch:
          fields: [id, title, url, status, properties]
          keepRelated: true
        notion-search:
          fields: [id, title, url]
          keepRelated: false
        notion-create-pages:
          fields: [id, url]
```

### Dynamic Configuration Reload

SlimMCP supports hot-reloading configuration without restart:

```bash
# Send HUP signal to reload
kill -HUP $(pidof slim)

# Or via API (if enabled)
curl -X POST http://localhost:8080/admin/reload
```

### Multi-Tenant Configuration

Run multiple SlimMCP instances with different configs:

```bash
# Instance 1 - Production
slim start --config prod.yaml --port 8080

# Instance 2 - Staging
slim start --config staging.yaml --port 8081

# Instance 3 - Development
slim start --config dev.yaml --port 8082
```

## Configuration Validation

Validate your configuration before starting:

```bash
slim validate --config slim.yaml
```

Output:
```
✓ Configuration is valid
✓ All server URLs are reachable
✓ Ports are available
✓ File permissions are correct
```

## Troubleshooting Configuration

### Check Active Configuration

```bash
# View current configuration
curl http://localhost:8080/admin/config

# View parsed configuration (with defaults)
slim config show --file slim.yaml
```

### Common Configuration Errors

**Error:** `address already in use`

```yaml
# Solution: Change port
proxy:
  listenAddress: "0.0.0.0:8081"  # Use different port
```

**Error:** `connection refused to downstream server`

```yaml
# Solution: Verify downstream URL
proxy:
  servers:
    notion:
      url: "http://host.docker.internal:3000"  # For Docker
      # or
      url: "http://localhost:3000"  # For local
```

**Error:** `circuit breaker open`

```yaml
# Solution: Adjust circuit breaker settings
optimization:
  safety:
    circuitBreaker:
      threshold: 20  # Allow more errors
      timeout: 30s   # Recover faster
```

## Best Practices

1. **Start with defaults** - The default configuration works well for most cases
2. **Enable monitoring** - Always enable Prometheus metrics in production
3. **Use smart strategy** - Best balance of savings and accuracy
4. **Configure safety** - Enable circuit breakers and fallback
5. **Test before production** - Validate configuration changes in staging
6. **Monitor savings** - Track metrics to optimize configuration
7. **Document changes** - Keep notes on why settings were changed

## See Also

- [Getting Started Guide](getting-started.md)
- [Architecture Overview](architecture.md)
- [API Documentation](api.md)
- [Troubleshooting Guide](troubleshooting.md)
