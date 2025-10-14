# API Documentation

Complete reference for SlimMCP's HTTP API and MCP protocol extensions.

## Table of Contents

- [HTTP Endpoints](#http-endpoints)
- [MCP Protocol](#mcp-protocol)
- [Field Selection Extension](#field-selection-extension)
- [Response Format](#response-format)
- [Error Handling](#error-handling)
- [Examples](#examples)

## HTTP Endpoints

### MCP Request Endpoint

```
POST /
POST /{server}
```

Main endpoint for MCP requests.

**Request Format:**
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "notion-fetch",
    "arguments": {
      "id": "page-123"
    }
  },
  "id": 1,
  "_meta": {
    "fieldSelection": {
      "include": ["title", "status"],
      "strategy": "explicit"
    }
  }
}
```

**Response Format:**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "id": "page-123",
    "title": "Q4 Planning",
    "status": "In Progress",
    "url": "https://...",
    "type": "page"
  },
  "id": 1,
  "_meta": {
    "fieldSelection": {
      "applied": true,
      "strategy": "explicit",
      "filteredBy": "proxy",
      "originalSize": 2847,
      "filteredSize": 421,
      "savings": 85.2,
      "fields": {
        "kept": ["id", "title", "status", "url", "type"],
        "removed": 40
      }
    }
  }
}
```

### Health Check

```
GET /health
```

Returns the health status of SlimMCP.

**Response:**
```json
{
  "status": "healthy",
  "version": "0.1.0",
  "uptime": "2h15m30s",
  "servers": {
    "notion": {
      "status": "connected",
      "url": "http://localhost:3000",
      "lastCheck": "2025-10-12T10:30:45Z"
    },
    "slack": {
      "status": "connected",
      "url": "http://localhost:3001",
      "lastCheck": "2025-10-12T10:30:45Z"
    }
  }
}
```

**Status Codes:**
- `200 OK` - Service is healthy
- `503 Service Unavailable` - Service is unhealthy

### Metrics Endpoint

```
GET /metrics
```

Prometheus-compatible metrics endpoint.

**Response:**
```prometheus
# HELP mcp_optimizer_tokens_saved_total Total tokens saved by filtering
# TYPE mcp_optimizer_tokens_saved_total counter
mcp_optimizer_tokens_saved_total{server="notion",tool="fetch"} 12847

# HELP mcp_optimizer_savings_percentage Current token savings percentage
# TYPE mcp_optimizer_savings_percentage gauge
mcp_optimizer_savings_percentage{server="notion"} 82.3

# HELP mcp_optimizer_requests_total Total number of requests processed
# TYPE mcp_optimizer_requests_total counter
mcp_optimizer_requests_total{server="notion",status="success"} 8492

# HELP mcp_optimizer_filtering_duration_seconds Time spent filtering responses
# TYPE mcp_optimizer_filtering_duration_seconds histogram
mcp_optimizer_filtering_duration_seconds_bucket{le="0.005"} 7234
mcp_optimizer_filtering_duration_seconds_bucket{le="0.01"} 8103
```

### Configuration Endpoint (Optional)

```
GET /admin/config
```

Returns the active configuration (requires admin access).

**Response:**
```json
{
  "proxy": {
    "listenAddress": "0.0.0.0:8080",
    "strategy": "smart",
    "servers": {
      "notion": {
        "url": "http://localhost:3000",
        "timeout": "30s"
      }
    }
  },
  "monitoring": {
    "enabled": true,
    "metricsPort": 9090
  }
}
```

### Reload Configuration (Optional)

```
POST /admin/reload
```

Triggers a configuration reload without restarting.

**Response:**
```json
{
  "status": "reloaded",
  "timestamp": "2025-10-12T10:30:45Z",
  "changes": [
    "Updated server timeout for notion",
    "Added new server: github"
  ]
}
```

## MCP Protocol

SlimMCP implements the [Model Context Protocol (MCP)](https://modelcontextprotocol.io) specification.

### Supported Methods

| Method | Description | Status |
|--------|-------------|--------|
| `tools/list` | List available tools | ✅ Supported |
| `tools/call` | Call a tool | ✅ Supported |
| `resources/list` | List available resources | ✅ Supported |
| `resources/read` | Read a resource | ✅ Supported |
| `prompts/list` | List available prompts | ✅ Supported |
| `prompts/get` | Get a prompt | ✅ Supported |

### Request Format

All MCP requests follow JSON-RPC 2.0 format:

```json
{
  "jsonrpc": "2.0",
  "method": "METHOD_NAME",
  "params": {
    // Method-specific parameters
  },
  "id": REQUEST_ID
}
```

### Response Format

Standard JSON-RPC 2.0 response:

```json
{
  "jsonrpc": "2.0",
  "result": {
    // Method-specific result
  },
  "id": REQUEST_ID
}
```

### Error Response

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32600,
    "message": "Invalid Request",
    "data": {
      "details": "Additional error information"
    }
  },
  "id": REQUEST_ID
}
```

## Field Selection Extension

SlimMCP implements the `_meta.fieldSelection` extension for fine-grained control over response filtering.

### Request Extension

Add `_meta.fieldSelection` to any MCP request:

```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "notion-fetch",
    "arguments": {
      "id": "page-123"
    }
  },
  "id": 1,
  "_meta": {
    "fieldSelection": {
      "include": ["title", "status", "assignee"],
      "exclude": ["icon", "cover"],
      "strategy": "explicit",
      "keepRelated": true,
      "maxDepth": 5
    }
  }
}
```

### Extension Parameters

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `include` | `string[]` | Fields to explicitly include | `null` |
| `exclude` | `string[]` | Fields to explicitly exclude | `null` |
| `strategy` | `string` | Filtering strategy | `"smart"` |
| `keepRelated` | `boolean` | Include fields related to included fields | `true` |
| `maxDepth` | `number` | Maximum nesting depth to filter | `10` |

### Filtering Strategies

**explicit**
```json
{
  "strategy": "explicit",
  "include": ["id", "title", "status"]
}
```
Only includes explicitly listed fields (plus critical fields).

**smart**
```json
{
  "strategy": "smart"
}
```
Analyzes request to infer needed fields intelligently.

**minimal**
```json
{
  "strategy": "minimal"
}
```
Returns only critical fields: `id`, `type`, `url`, `title`.

**passthrough**
```json
{
  "strategy": "passthrough"
}
```
No filtering applied (full response).

### Response Extension

Filtered responses include `_meta.fieldSelection` with statistics:

```json
{
  "jsonrpc": "2.0",
  "result": {
    // Filtered data
  },
  "id": 1,
  "_meta": {
    "fieldSelection": {
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
  }
}
```

### Metadata Fields

| Field | Type | Description |
|-------|------|-------------|
| `applied` | `boolean` | Was filtering applied? |
| `strategy` | `string` | Strategy used |
| `filteredBy` | `string` | Who filtered: `"proxy"` or `"server"` |
| `originalSize` | `number` | Original response size (tokens) |
| `filteredSize` | `number` | Filtered response size (tokens) |
| `savings` | `number` | Savings percentage |
| `fields.kept` | `string[]` | Fields that were kept |
| `fields.removed` | `number` | Number of fields removed |
| `filteringTime` | `number` | Time spent filtering (ms) |

## Error Handling

### Error Codes

SlimMCP uses standard JSON-RPC error codes plus custom codes:

| Code | Message | Description |
|------|---------|-------------|
| `-32700` | Parse error | Invalid JSON |
| `-32600` | Invalid Request | Invalid request object |
| `-32601` | Method not found | Method doesn't exist |
| `-32602` | Invalid params | Invalid parameters |
| `-32603` | Internal error | Internal server error |
| `-32000` | Downstream error | Error from MCP server |
| `-32001` | Filtering error | Error during filtering |
| `-32002` | Schema error | Schema validation failed |
| `-32003` | Timeout | Request timeout |

### Error Response Format

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32001,
    "message": "Filtering error",
    "data": {
      "details": "Failed to filter field: properties.custom_field",
      "tool": "notion-fetch",
      "server": "notion",
      "fallback": "full_response_returned"
    }
  },
  "id": 1
}
```

### Fallback Behavior

When filtering fails, SlimMCP can:

1. **Return full response** (default, safe)
2. **Return error** (fail fast)
3. **Return partial response** (best effort)

Configured via:
```yaml
optimization:
  safety:
    fallbackOnError: true  # Return full response on error
```

## Examples

### Example 1: Basic Request

```bash
curl -X POST http://localhost:8080/notion \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "notion-fetch",
      "arguments": {
        "id": "page-123"
      }
    },
    "id": 1
  }'
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "id": "page-123",
    "title": "Q4 Planning",
    "url": "https://notion.so/page-123",
    "type": "page"
  },
  "id": 1,
  "_meta": {
    "fieldSelection": {
      "applied": true,
      "strategy": "smart",
      "savings": 87.5
    }
  }
}
```

### Example 2: Explicit Field Selection

```bash
curl -X POST http://localhost:8080/notion \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "notion-fetch",
      "arguments": {
        "id": "page-123"
      }
    },
    "id": 1,
    "_meta": {
      "fieldSelection": {
        "include": ["title", "status", "assignee"],
        "strategy": "explicit"
      }
    }
  }'
```

### Example 3: Multiple Fields with Nesting

```json
{
  "_meta": {
    "fieldSelection": {
      "include": [
        "id",
        "title",
        "properties.status",
        "properties.assignee.name",
        "properties.assignee.email"
      ],
      "strategy": "explicit",
      "keepRelated": false
    }
  }
}
```

### Example 4: Exclusion List

```json
{
  "_meta": {
    "fieldSelection": {
      "exclude": ["icon", "cover", "created_by", "last_edited_by"],
      "strategy": "smart"
    }
  }
}
```

### Example 5: Minimal Strategy

```json
{
  "_meta": {
    "fieldSelection": {
      "strategy": "minimal"
    }
  }
}
```

Returns only: `id`, `type`, `url`, `title`

### Example 6: JavaScript Client

```javascript
import { MCPClient } from '@modelcontextprotocol/sdk';

const client = new MCPClient('http://localhost:8080/notion');

// With explicit field selection
const response = await client.callTool({
  name: 'notion-fetch',
  arguments: {
    id: 'page-123'
  },
  _meta: {
    fieldSelection: {
      include: ['title', 'status'],
      strategy: 'explicit'
    }
  }
});

console.log(`Saved ${response._meta.fieldSelection.savings}% tokens`);
```

### Example 7: Python Client

```python
from mcp import Client

client = Client("http://localhost:8080/notion")

# With smart filtering (automatic)
response = client.call_tool(
    name="notion-fetch",
    arguments={"id": "page-123"}
)

# Check savings
if response.get("_meta", {}).get("fieldSelection", {}).get("applied"):
    savings = response["_meta"]["fieldSelection"]["savings"]
    print(f"Saved {savings}% tokens")
```

### Example 8: Batch Requests

```bash
curl -X POST http://localhost:8080/notion \
  -H "Content-Type: application/json" \
  -d '[
    {
      "jsonrpc": "2.0",
      "method": "tools/call",
      "params": {
        "name": "notion-fetch",
        "arguments": {"id": "page-1"}
      },
      "id": 1
    },
    {
      "jsonrpc": "2.0",
      "method": "tools/call",
      "params": {
        "name": "notion-fetch",
        "arguments": {"id": "page-2"}
      },
      "id": 2
    }
  ]'
```

## Rate Limiting

SlimMCP respects downstream rate limits and adds minimal overhead.

**Headers:**
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 950
X-RateLimit-Reset: 1634567890
```

## Authentication

SlimMCP passes authentication through to downstream servers.

**Example with Bearer Token:**
```bash
curl -X POST http://localhost:8080/notion \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{...}'
```

The `Authorization` header is forwarded to the downstream MCP server unchanged.

## Compression

SlimMCP supports gzip compression for both requests and responses.

**Request:**
```
Content-Encoding: gzip
```

**Response:**
```
Content-Encoding: gzip
```

## CORS

CORS is configurable in `slim.yaml`:

```yaml
proxy:
  cors:
    enabled: true
    allowOrigins: ["*"]
    allowMethods: ["GET", "POST", "OPTIONS"]
    allowHeaders: ["Content-Type", "Authorization"]
```

## Versioning

API version is included in responses:

```json
{
  "_meta": {
    "version": "0.1.0",
    "protocolVersion": "1.0"
  }
}
```

## See Also

- [Configuration Reference](configuration.md)
- [Protocol Extension Spec](protocol-extension.md)
- [Architecture Overview](architecture.md)
- [Examples Repository](https://github.com/slimmcp/examples)
