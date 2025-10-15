# Manual Testing Guide

This directory contains scripts and instructions for manually testing SlimMCP proxy functionality.

## Quick Start

### 1. Build the binaries

```bash
# From mcp/ directory
make build

# Or manually
go build -o bin/slim cmd/slimmcp/main.go
go build -o bin/mockserver tests/tools/mockserver/main.go
```

### 2. Start the mock MCP server

In terminal 1:
```bash
./bin/mockserver --port 3000
```

You should see:
```
Loading Notion fixtures...
Loading GitHub fixtures...
Starting mock MCP server on port 3000...
```

### 3. Start the SlimMCP proxy

In terminal 2:
```bash
./bin/slim -config examples/dev.yaml
```

You should see:
```
SlimMCP v0.1.0 starting...
Configuration loaded from: examples/dev.yaml
Listen address: localhost:8080
Default strategy: smart
Configured servers: 1
SlimMCP proxy server started successfully
```

### 4. Run the test suite

In terminal 3:
```bash
./tests/manual/test_proxy.sh
```

Or run individual tests with curl (see examples below).

## Manual Test Cases

### Test 1: Health Check

```bash
curl http://localhost:8080/health | jq '.'
```

**Expected:**
```json
{
  "status": "ok",
  "version": "0.1.0",
  "proxy": "SlimMCP"
}
```

### Test 2: User Story 1 - Automatic Filtering (Minimal Strategy)

**Scenario:** Send request without `_meta.fieldSelection`, verify proxy applies minimal filtering automatically.

```bash
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
  }' | jq '.'
```

**Expected:**
- Response contains only critical fields: `id`, `type`, `url`, `title`, `object`
- `_meta.fieldSelection.strategy` is `"smart"` (default)
- `_meta.fieldSelection.savings` shows token reduction percentage
- `_meta.fieldSelection.applied` is `true`

### Test 3: User Story 2 - Explicit Include Mode

**Scenario:** Client explicitly requests specific fields via `_meta.fieldSelection.include`.

```bash
curl -X POST http://localhost:8080/mock \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "test-tool",
      "arguments": {"id": "123"}
    },
    "_meta": {
      "fieldSelection": {
        "include": ["id", "title", "url"],
        "strategy": "explicit"
      }
    },
    "id": 2
  }' | jq '.'
```

**Expected:**
- Response contains only: `id`, `title`, `url` (plus any critical fields)
- `_meta.fieldSelection.strategy` is `"explicit"`
- `_meta.fieldSelection.fields.kept` lists the included fields

### Test 4: User Story 2 - Explicit Exclude Mode

**Scenario:** Client excludes specific fields via `_meta.fieldSelection.exclude`.

```bash
curl -X POST http://localhost:8080/mock \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "test-tool",
      "arguments": {"id": "123"}
    },
    "_meta": {
      "fieldSelection": {
        "exclude": ["icon", "cover"],
        "strategy": "explicit"
      }
    },
    "id": 3
  }' | jq '.'
```

**Expected:**
- Response contains all fields EXCEPT `icon` and `cover`
- Excluded fields should not appear in result

### Test 5: User Story 2 - Nested Field Selection

**Scenario:** Client requests nested fields using dot notation.

```bash
curl -X POST http://localhost:8080/mock \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "query_database",
      "arguments": {"database_id": "db123"}
    },
    "_meta": {
      "fieldSelection": {
        "include": ["id", "properties.Status", "properties.Priority"],
        "strategy": "explicit"
      }
    },
    "id": 4
  }' | jq '.'
```

**Expected:**
- Response has `properties` object
- Within `properties`, only `Status` and `Priority` fields are present
- Parent structure is preserved

### Test 6: Error Handling - Invalid Field Selection

**Scenario:** Client sends both `include` and `exclude` (invalid).

```bash
curl -X POST http://localhost:8080/mock \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "test-tool",
      "arguments": {"id": "123"}
    },
    "_meta": {
      "fieldSelection": {
        "include": ["title"],
        "exclude": ["icon"],
        "strategy": "explicit"
      }
    },
    "id": 5
  }' | jq '.'
```

**Expected:**
- Response contains JSON-RPC error
- Error code: `-32600` (Invalid Request)
- Error message mentions field selection validation failure

### Test 7: Token Savings Verification

**Scenario:** Verify that token savings are accurately calculated.

```bash
# Make a request and extract metadata
curl -s -X POST http://localhost:8080/mock \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {"name": "test-tool", "arguments": {}},
    "id": 1
  }' | jq '._meta.fieldSelection'
```

**Expected:**
```json
{
  "applied": true,
  "strategy": "smart",
  "filteredBy": "proxy",
  "originalSize": 100,
  "filteredSize": 25,
  "savings": 75.0,
  "fields.kept": ["id", "type", "url", "title", "object"],
  "fields.removed": 5,
  "filteringTime": 2.5
}
```

## Using Notion Fixtures

The mock server loads real Notion API responses from `tests/fixtures/notion/`:

```bash
# Query database (real Notion structure)
curl -X POST http://localhost:8080/mock \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "query_database",
      "arguments": {"database_id": "db123"}
    },
    "id": 1
  }' | jq '.'
```

Available Notion operations:
- `query_database` - Returns pages from a database
- `create_page` - Page creation response
- `update_page` - Page update response
- `search_pages` - Search results
- `list_pages` - List of pages

## Troubleshooting

### Proxy won't start

**Error:** `Failed to load configuration`
- Check that `examples/dev.yaml` exists
- Verify YAML syntax is correct

**Error:** `listen tcp :8080: bind: address already in use`
- Another process is using port 8080
- Change port in `examples/dev.yaml` or stop other process

### Mock server connection failed

**Error:** `Downstream server error`
- Ensure mock server is running on port 3000
- Check that mock server URL in config is correct: `http://localhost:3000`

### No filtering applied

**Issue:** Responses are not filtered
- Check that request is going through proxy (`:8080`) not directly to mock (`:3000`)
- Verify `_meta` structure is correct in request
- Check proxy logs for error messages

### Tests fail

**Issue:** `test_proxy.sh` reports failures
- Ensure both proxy and mock server are running
- Check that `jq` is installed (for JSON parsing)
- Run individual curl commands to debug

## Performance Testing

To test performance under load:

```bash
# Install Apache Bench (if not already installed)
# brew install httpd (macOS)
# apt-get install apache2-utils (Linux)

# Run 1000 requests with 10 concurrent connections
ab -n 1000 -c 10 -T 'application/json' \
  -p tests/manual/requests/us1_minimal.json \
  http://localhost:8080/mock
```

**Expected:** >500 requests/second

## Next Steps

After manual testing:
1. Verify all tests pass
2. Check logs for any errors or warnings
3. Inspect filtering metadata for accuracy
4. Test with different strategies (minimal, explicit)
5. Measure performance and token savings
