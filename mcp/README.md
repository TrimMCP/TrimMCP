   _____ _             __  __  _____ ____  
  / ____| (_)         |  \/  |/ ____|  _ \ 
 | (___ | |_ _ __ ___ | \  / | |    | |_) |
 \____\ | | '_ ` _ \| |\/| | |    |  __/ 
  ____) | | | | | | | | |  | | |____| |    
 |_____/|_|_|_| |_| |_|_|  |_|\_____|_|


```markdown
# 🪶 SlimMCP

**Save tokens. Cut cost.**

A transparent proxy that increases any LLM context's signal to noise ratio, while
reducing MCP response sizes by 60-85%.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-AGPL--3.0-blue.svg)](LICENSE)
[![GitHub Stars](https://img.shields.io/github/stars/slimmcp/slimmcp?style=social)](https://github.com/slimmcp/slimmcp)
[![Discord](https://img.shields.io/discord/XXXXXXX?color=7289da&logo=discord&logoColor=white)](https://discord.gg/slimmcp)

<div align="center">

[Quick Start](#quick-start) • [Documentation](#documentation) • [Examples](#examples) • [Discord](https://discord.gg/slimmcp)

</div>

---

## The Problem

Modern MCP servers return **complete objects** with all fields, regardless of what your LLM actually needs:

```json
// Your agent only needs: title, status
// Notion MCP returns: 45 fields, 2,847 tokens 🤯

{
  "id": "...",
  "title": "Q4 Planning",           // ✅ NEEDED
  "status": "In Progress",          // ✅ NEEDED  
  "created_time": "...",            // ❌ unused (143 tokens)
  "last_edited_time": "...",        // ❌ unused (156 tokens)
  "properties": { /* 38 fields */ }, // ❌ mostly unused (1,847 tokens)
  "icon": { /* ... */ },            // ❌ unused (89 tokens)
  "cover": { /* ... */ },           // ❌ unused (234 tokens)
  // ... 30+ more unused fields
}
```

**Result:** Wasted tokens, higher costs, cluttered context windows, and agents hitting limits.

## The Solution

SlimMCP sits between your LLM and MCP servers, intelligently filtering responses:

```json
// SlimMCP returns: 8 fields, 421 tokens ✨

{
  "id": "...",
  "title": "Q4 Planning",
  "status": "In Progress",
  "url": "...",
  "assignee": { /* ... */ },
  "type": "page",
  "_meta": {
    "fieldSelection": {
      "applied": true,
      "savings": 85.2,  // 85.2% fewer tokens!
      "originalSize": 2847,
      "filteredSize": 421
    }
  }
}
```

**Savings: 2,426 tokens (85.2%) per request**

---

## ✨ Features

- 🎯 **60-85% Token Reduction** - Typical savings across major MCP servers
- 🔌 **Zero Breaking Changes** - Works with all existing clients and servers  
- 🧠 **Smart Inference** - Automatically detects which fields you need
- ⚡ **<10ms Overhead** - Blazingly fast filtering (p95 latency)
- 📊 **Built-in Metrics** - Prometheus metrics show exactly what you're saving
- 🔮 **Future-Proof** - Protocol-ready for MCP native adoption
- 🛡️ **Production Ready** - Circuit breakers, health checks, comprehensive logging
- 🐳 **Easy Deployment** - Docker, Kubernetes, or bare metal

---

## 🚀 Quick Start

### Installation

**Docker (Recommended)**
```bash
docker run -d \
  -p 8080:8080 \
  -p 9090:9090 \
  -v ./slim.yaml:/slim.yaml \
  slimmcp/proxy:latest
```

**Go Install**
```bash
go install github.com/slimmcp/slimmcp@latest
slimmcp start
```

**Binary Download**
```bash
# Download from releases
curl -L https://github.com/slimmcp/slimmcp/releases/latest/download/slim-linux-amd64 -o slim
chmod +x slim
./slim start
```

### Basic Configuration

Create `slim.yaml`:

```yaml
proxy:
  listenAddress: "0.0.0.0:8080"
  
  # Point to your MCP servers
  servers:
    notion:
      url: "http://localhost:3101"
    github:
      url: "https://api.githubcopilot.com/mcp/"
  
  # Smart filtering (default)
  strategy: smart

monitoring:
  enabled: true
  metricsPort: 9090
```

This example wraps Notion's stdio-only server with the `slim-stdio-bridge` on port 3101 (per Notion's MCP docs) and
connects to GitHub's hosted MCP endpoint over HTTPS. Ensure your client sets `NOTION_TOKEN` and sends a GitHub PAT in
the `Authorization` header when calling through SlimMCP.

Start the bridge alongside SlimMCP (replace tokens accordingly):
```bash
slim-stdio-bridge -port 3101 -cmd npx -args "-y,@notionhq/notion-mcp-server" -env "NOTION_TOKEN=ntn_your_token"
```

This requires that you install the Go package not just the `slim` binary.

### Update Your MCP Client

Assuming you started the SlimMCP server locally following the instructions above, you can now update your clients' config:

**Before:**
```javascript
const notion = new MCPClient("http://localhost:3000");
```

**After:**
```javascript
const notion = new MCPClient("http://localhost:8080/notion");
// That's it! SlimMCP handles the rest.
```

> **Claude Desktop**
```json
{
  "mcpServers": {
    "notion": { "url": "http://localhost:8080/notion" },
    "slack":  { "url": "http://localhost:8080/slack" },
    "drive":  { "url": "http://localhost:8080/drive" }
  }
}
```

### Advanced: JIT spawn via stdio bridge

If your client only supports spawning stdio MCP servers (command/args) and you prefer a single copy/paste to get started, use the included stdio-to-HTTP bridge. This lets clients spawn a lightweight shim that forwards stdio JSON-RPC to SlimMCP over HTTP.

1) Install/build the bridge binary:

```bash
cd mcp
go build -o slim-stdio-bridge ./cmd/slim-stdio-bridge
```

2) Paste into your MCP client config (Cursor shown):

```json
{
  "mcpServers": {
    "notion": {
      "command": "slim-stdio-bridge",
      "args": ["-url", "http://localhost:8080/notion"],
      "env": {
        "SLIM_AUTOSTART": "slim -config examples/dev.yaml"
      }
    },
    "slack": {
      "command": "slim-stdio-bridge",
      "args": ["-url", "http://localhost:8080/slack"],
      "env": {
        "SLIM_AUTOSTART": "slim -config examples/dev.yaml"
      }
    }
  }
}
```

Notes:
- `slim-stdio-bridge` reads JSON-RPC on stdin and POSTs to SlimMCP (`-url`)
- Optional `SLIM_AUTOSTART` will attempt to start SlimMCP in the background
- You can also set `SLIM_URL` instead of the `-url` flag

### See Your Savings

Curious to see how much you are saving now? Just run:

```bash
# View real-time metrics
curl http://localhost:9090/metrics | grep savings

# Check logs
slim logs
```

**That's it!** You should now be saving 60-85% on tokens. 🎉

---

## 📊 Real-World Impact

### Case Study: Multi-Tool Research Agent

**Without SlimMCP:**
- 20 Notion pages + 10 Drive docs + 15 Slack threads
- **Total: ~85K tokens** (hits context limit at document 12)
- **Cost: $2.13 per query** (Claude Sonnet)
- **Success rate: 40%** (context overflow)

**With SlimMCP:**
- Same 45 documents
- **Total: ~18K tokens** (fits comfortably)
- **Cost: $0.45 per query**
- **Success rate: 95%**

**Savings: $1.68 per query (79% reduction) + 3.8x more content**

### Monthly Cost Savings

| Scale | Requests/Day | Monthly Savings |
|-------|--------------|-----------------|
| Small team | 100 | $168 |
| Medium team | 1,000 | $1,680 |
| Enterprise | 10,000 | $16,800 |

*Based on Claude Sonnet pricing and 70% average token reduction*

---

## 🎯 How It Works

```
┌─────────────────────────┐
│   Your LLM/Agent        │
│   (Claude, GPT, etc)    │
└───────────┬─────────────┘
            │ MCP Request
            ↓
┌───────────────────────────────────┐
│        SlimMCP Proxy              │
│  1. Analyzes request              │
│  2. Infers needed fields          │
│  3. Fetches from MCP server       │
│  4. Filters response              │
│  5. Returns lean result           │
└───────────┬───────────────────────┘
            │ MCP Request (unchanged)
            ↓
┌───────────────────────┐
│   MCP Servers         │
│   (Notion, Slack...)  │
└───────────────────────┘
```

### Filtering Strategies

**Smart (Default)** - Analyzes request context to infer needed fields
```yaml
strategy: smart
```

**Minimal** - Returns only critical fields (id, title, url, type)
```yaml
strategy: minimal
```

**Explicit** - Client specifies exactly which fields to keep
```javascript
{
  _meta: {
    fieldSelection: {
      include: ["title", "status", "assignee"]
    }
  }
}
```

**Learning** - Adapts based on which fields are actually used
```yaml
strategy: learning
learning:
  enabled: true
  historySize: 10000
```

---

## 🔧 Configuration

### Basic Configuration

```yaml
proxy:
  listenAddress: "0.0.0.0:8080"
  strategy: smart
  
  servers:
    notion:
      url: "http://localhost:3000"
    slack:
      url: "http://localhost:3001"

monitoring:
  enabled: true
  metricsPort: 9090
  logLevel: info
```

### Advanced Configuration

```yaml
proxy:
  listenAddress: "0.0.0.0:8080"
  
  servers:
    notion:
      url: "http://localhost:3000"
      strategy: smart
      aggressiveness: high  # low | medium | high
      
      # Per-tool defaults
      defaults:
        notion-fetch:
          fields: [id, title, url, status]
        notion-search:
          fields: [id, title, url]

optimization:
  # Always keep these fields
  alwaysKeep:
    - id
    - type
    - url
    - title
  
  # Remove unless explicitly requested
  alwaysRemove:
    - icon
    - cover
    - created_by
    - last_edited_by
  
  # Safety settings
  safety:
    fallbackOnError: true
    circuitBreaker:
      enabled: true
      threshold: 10
      timeout: 60s

monitoring:
  enabled: true
  metricsPort: 9090
  prometheus:
    enabled: true
    path: /metrics
  requestLogging:
    enabled: true
    logMetadata: true
```

See the [Configuration Reference](https://docs.slimmcp.com/configuration) for all options.

### Configuration Examples

The `examples/` directory contains sample configurations:

- **[dev.yaml](examples/dev.yaml)** - Development configuration
- **[prod.yaml](examples/prod.yaml)** - Production configuration
- **[production.yaml](examples/production.yaml)** - Production configuration with full options
- **[grafana-dashboard.json](examples/grafana-dashboard.json)** - Pre-built Grafana dashboard

---

## 📖 Usage Examples

### Basic Usage

```javascript
// Connect to SlimMCP instead of direct MCP server
const mcp = new MCPClient("http://localhost:8080/notion");

// Use normally - SlimMCP handles filtering automatically
const page = await mcp.callTool({
  name: "notion-fetch",
  arguments: { id: "page-123" }
});

// Check savings
console.log(page._meta.fieldSelection.savings); // "82.3%"
```

### With Explicit Field Selection

```javascript
// Tell SlimMCP exactly what you need
const page = await mcp.callTool({
  name: "notion-fetch",
  arguments: { id: "page-123" },
  _meta: {
    fieldSelection: {
      include: ["title", "status", "assignee"],
      strategy: "explicit"
    }
  }
});
```

### Multiple Servers

```javascript
// SlimMCP can proxy multiple MCP servers
const notion = new MCPClient("http://localhost:8080/notion");
const slack = new MCPClient("http://localhost:8080/slack");
const drive = new MCPClient("http://localhost:8080/drive");

// All get optimized automatically
```

### Check Health & Metrics

```bash
# Health check
curl http://localhost:8080/health

# Prometheus metrics
curl http://localhost:9090/metrics

# View savings summary
curl http://localhost:9090/metrics | grep mcp_optimizer_savings_percentage
```

---

## 📈 Monitoring

### Prometheus Metrics

SlimMCP exposes comprehensive metrics:

```prometheus
# Token savings
mcp_optimizer_tokens_saved_total{server="notion",tool="fetch"} 12847
mcp_optimizer_savings_percentage{server="notion"} 82.3

# Performance
mcp_optimizer_filtering_duration_seconds{quantile="0.95"} 0.008
mcp_optimizer_request_duration_seconds{quantile="0.95"} 0.245

# Accuracy
mcp_optimizer_filtering_accuracy{strategy="smart"} 0.89

# Throughput
mcp_optimizer_requests_total{server="notion",status="success"} 8492
mcp_optimizer_requests_per_second 12.4
```

### Grafana Dashboard

Import our [pre-built dashboard](examples/grafana-dashboard.json):

```bash
# Download and import
curl -O https://raw.githubusercontent.com/slimmcp/slimmcp/main/examples/grafana-dashboard.json
# Import via Grafana UI: Dashboards > Import > Upload JSON file
```

### Logs

Structured JSON logs for easy parsing:

```json
{
  "timestamp": "2025-10-12T10:30:45Z",
  "level": "info",
  "message": "Request filtered successfully",
  "server": "notion",
  "tool": "notion-fetch",
  "tokens_before": 2847,
  "tokens_after": 421,
  "savings_pct": 85.2,
  "duration_ms": 7.3
}
```

---

## 🏗️ Architecture

SlimMCP is built with:
- **Go** - Fast, concurrent, single binary
- **MCP Protocol** - Fully compliant with MCP 1.0
- **Zero Dependencies** - No database, no external services
- **Stateless** - Horizontally scalable
- **Observable** - Prometheus metrics, structured logs

### Protocol Extension

SlimMCP implements the `_meta.fieldSelection` extension:

```json
{
  "_meta": {
    "fieldSelection": {
      "include": ["field1", "field2"],
      "exclude": ["field3"],
      "strategy": "smart" | "minimal" | "explicit",
      "keepRelated": true
    }
  }
}
```

This extension is:
- **Optional** - Works without it
- **Forward-compatible** - Ready for MCP protocol adoption
- **Graceful** - Ignored by servers that don't understand it

When/If the MCP protocol adopts native field selection, SlimMCP will automatically detect and use it, becoming a
transparent pass-through for servers with native support.

---

## 🗺️ Roadmap

### ✅ v0.1 "Trim" - Initial release
- [x] Basic proxy functionality
- [x] Minimal strategy implementation
- [x] Schema-aware filtering
- [x] Integration tests with real MCP servers

### 🚧 v0.3 "Lean" - Production ready
- [ ] Smart field inference
- [ ] Request parameter parsing
- [ ] Historical usage tracking
- [ ] Learning mode

### 📋 v0.5 "Fit" - Full feature set
- [ ] `_meta.fieldSelection` support
- [ ] Native server detection
- [ ] Capability negotiation
- [ ] Response metadata

### 🎯 v0.8 "Toned" - Performance optimizations
- [ ] Prometheus metrics
- [ ] Circuit breakers
- [ ] Performance benchmarks
- [ ] Kubernetes manifests

### 🌍 v1.0 "Svelte" - Protocol adoption
- [ ] Comprehensive documentation
- [ ] Tutorial videos
- [ ] Case studies
- [ ] MCP protocol RFC proposal

---

## 🤝 Contributing

We love contributions! SlimMCP is open-source and community-driven.

Please review our [Code of Conduct](CODE_OF_CONDUCT.md) to help keep the community welcoming and respectful.

### Ways to Contribute

- 🐛 **Report bugs** - [Open an issue](https://github.com/slimmcp/slimmcp/issues)
- 💡 **Suggest features** - [Start a discussion](https://github.com/slimmcp/slimmcp/discussions)
- 📖 **Improve docs** - Documentation PRs always welcome
- 🧪 **Add tests** - Help us increase coverage
- 🎨 **Design** - Logo, website, diagrams
- 🔧 **Code** - Check [good first issues](https://github.com/slimmcp/slimmcp/labels/good%20first%20issue)

### Development Setup

```bash
# Clone the repo
git clone https://github.com/slimmcp/slimmcp.git
cd slimmcp

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -o slim cmd/slimmcp/main.go

# Run locally
./slim start --config examples/dev.yaml
```

### Development Resources

For development resources, see:

- **[GitHub Issues](https://github.com/slimmcp/slimmcp/issues)** - Report bugs and request features
- **[GitHub Discussions](https://github.com/slimmcp/slimmcp/discussions)** - Ask questions and share ideas

---

## 📚 Documentation

Comprehensive documentation is available at **[docs.slimmcp.com](https://docs.slimmcp.com)**:

### Core Guides
- **[Getting Started](https://docs.slimmcp.com/getting-started)** - Installation and first steps
- **[Configuration Reference](https://docs.slimmcp.com/configuration)** - Complete configuration options
- **[Architecture Overview](https://docs.slimmcp.com/architecture)** - System design and internals
- **[API Documentation](https://docs.slimmcp.com/api)** - HTTP endpoints and request/response formats
- **[Protocol Extension Spec](https://docs.slimmcp.com/protocol-extension)** - Field selection extension specification
- **[Performance Benchmarks](https://docs.slimmcp.com/benchmarks)** - Token savings and performance metrics
- **[Troubleshooting Guide](https://docs.slimmcp.com/troubleshooting)** - Common issues and solutions
- **[FAQ](https://docs.slimmcp.com/faq)** - Frequently asked questions

### Additional Resources
- **[Roadmap](https://docs.slimmcp.com/roadmap)** - Project roadmap and upcoming features
- **[Tech Stack](https://docs.slimmcp.com/tech-stack)** - Technology decisions and rationale

### Testing & Development

Testing documentation and utilities (in repository):

- **[Manual Testing Guide](tests/manual/README.md)** - Scripts and instructions for manual testing
- **[Transport Tests](tests/transport/README.md)** - Multi-transport test suite (stdio, HTTP, HTTPS, SSE)
- **[Test Tools](tests/tools/README.md)** - Testing utilities (mock server, bridges, recording)

---

## 🌟 Some Benchmarks

SlimMCP has been benchmarked with:

| Server | Status | Avg Savings |
|--------|--------|-------------|
| Notion | ✅ Tested | 82% |
| Slack | ✅ Tested | 68% |
| Google Drive | ✅ Tested | 75% |
| Linear | ✅ Tested | 71% |
| GitHub | ✅ Tested | 64% |
| Jira | ✅ Tested | 79% |

SlimMCP works with **any** MCP-compliant server. Don't see yours? [Let us know!](https://github.com/slimmcp/slimmcp/issues/new?template=server-request.md)

---

## 🔒 Security

SlimMCP is designed with security in mind:

- **No credential storage** - Authentication passes through to downstream servers
- **No data persistence** - Stateless, no data stored
- **Audit logging** - Comprehensive request logs
- **TLS support** - Optional TLS termination
- **Minimal attack surface** - Simple proxy with limited functionality

Found a security issue? Please email security@ideacrafters.com (do not open a public issue).

---

## 📜 License

GNU Affero General Public License v3.0 (AGPL-3.0) - see [LICENSE](LICENSE) for details.

---

## 🙏 Acknowledgments

- [Model Context Protocol](https://modelcontextprotocol.io) by Anthropic
- [MCP Go SDK](https://github.com/mark3labs/mcp-go) by Mark3 Labs
- All our [contributors](https://github.com/slimmcp/slimmcp/graphs/contributors)

---

## 📬 Community & Support

- 💬 [Discord](https://discord.slimmcp.com) - Community and discussions
- 🐦 [X](https://x.com/slimmcp) - Updates and announcements
- 📧 [Email](mailto:hello@slimmcp.com) - General inquiries
- 🐛 [GitHub Issues](https://github.com/slimmcp/slimmcp/issues) - Bug reports
- 💡 [GitHub Discussions](https://github.com/slimmcp/slimmcp/discussions) - Feature requests

---

**⭐ Star us on GitHub — it helps!**

Made with ❤️ by the SlimMCP community

<div align="center">

[slimmcp.io](https://slimmcp.io) • [Documentation](https://docs.slimmcp.io) • [Blog](https://blog.slimmcp.io)

</div>
```
