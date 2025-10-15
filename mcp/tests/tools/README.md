# Test Tools

Testing utilities for SlimMCP development.

## MCP Transport Types

MCP (Model Context Protocol) servers support three transport mechanisms:

| Transport | Description | Use Case | Examples |
|-----------|-------------|----------|----------|
| **stdio** | Standard input/output (stdin/stdout) | CLI tools, local processes | GitHub MCP, Notion MCP |
| **SSE** | Server-Sent Events | Streaming, real-time updates | Live data feeds |
| **HTTP** | Direct HTTP/REST | Web services, APIs | Remote MCP servers |

## Tools

### 1. `mockserver/`

**Purpose**: Standalone mock MCP server with pre-loaded fixtures

**Usage**:
```bash
# Via Makefile
make test-server

# Direct
go run ./tests/tools/mockserver --port 3000
```

**Features**:
- Loads Notion and GitHub fixtures from `tests/fixtures/`
- Returns realistic MCP responses
- Used for integration and acceptance tests

---

### 2. `record-mcp/`

**Purpose**: Records real MCP server responses as VCR cassettes

**Usage**:
```bash
# Via Makefile
make record-cassettes

# Direct
go run ./tests/tools/record-mcp \
  -env .env.test \
  -config tests/recording-scenarios.json \
  -output tests/cassettes
```

**Features**:
- Records HTTP interactions using govcr
- Sanitizes auth tokens and sensitive data
- Creates YAML cassettes for replay in tests

---

### 3. `mcp-stdio-http-bridge/`

**Purpose**: Wraps stdio-based MCP servers with HTTP endpoints

**Why needed**: Many MCP servers (GitHub, Notion) use stdio transport, but testing tools like govcr and curl expect HTTP. This bridge translates between them.

**Usage**:
```bash
# Generic bridge
make mcp-bridge PORT=3100 CMD=npx ARGS="-y,@modelcontextprotocol/server-github" ENV="GITHUB_TOKEN=..."

# GitHub convenience shortcut
make github-bridge

# Notion convenience shortcut
make notion-bridge

# Custom MCP server
go run ./tests/tools/mcp-stdio-http-bridge \
  -port 3200 \
  -cmd /path/to/mcp-server \
  -args "arg1,arg2" \
  -env "KEY=value,KEY2=value2"
```

**How it works**:
1. Accepts HTTP POST requests with JSON-RPC
2. Spawns the MCP server process
3. Writes request to server's stdin
4. Reads response from server's stdout
5. Returns response as HTTP

**Architecture**:
```
Client (curl/test)
       │
       ▼ HTTP POST
  [mcp-stdio-http-bridge]
       │
       ▼ stdin/stdout
  MCP Server (npx @modelcontextprotocol/server-github)
```

## Environment Variables

Required for recording/bridging with real servers:

```bash
# GitHub
GITHUB_PERSONAL_ACCESS_TOKEN=github_pat_...
# or
GITHUB_TOKEN=github_pat_...

# Notion
NOTION_TOKEN=secret_...
# or
NOTION_API_KEY=secret_...
```

## Common Workflows

### Testing with Mock Server
```bash
# Terminal 1: Start mock server
make test-server

# Terminal 2: Send test requests
curl -X POST http://localhost:3000 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"test-tool"},"id":1}'
```

### Testing with Real GitHub MCP
```bash
# Terminal 1: Start bridge
export GITHUB_TOKEN=github_pat_...
make github-bridge

# Terminal 2: Send requests through bridge
curl -X POST http://localhost:3100 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"get_file_contents","arguments":{"owner":"user","repo":"repo","path":"README.md"}},"id":1}'
```

### Recording Cassettes
```bash
# 1. Configure scenarios in tests/recording-scenarios.json
# 2. Set environment variables
export GITHUB_TOKEN=github_pat_...
export NOTION_TOKEN=secret_...

# 3. Record
make record-cassettes

# 4. Cassettes saved to tests/cassettes/
# 5. Use in tests with VCR_MODE=replay
```

## Development

All tools are built with Go and can be run with:
```bash
go run ./tests/tools/<tool-name> [flags]
```

Or built as binaries:
```bash
go build -o bin/mockserver ./tests/tools/mockserver
go build -o bin/record-mcp ./tests/tools/record-mcp
go build -o bin/mcp-bridge ./tests/tools/mcp-stdio-http-bridge
```
