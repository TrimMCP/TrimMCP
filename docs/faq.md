# Frequently Asked Questions (FAQ)

Answers to common questions about SlimMCP.

## General Questions

### What is SlimMCP?

SlimMCP is a transparent proxy that sits between your LLM application and MCP servers. It intelligently filters response fields, reducing token consumption by 60-85% without breaking existing applications.

### How does SlimMCP work?

SlimMCP acts as both an MCP server (to your application) and an MCP client (to actual MCP servers). It:
1. Receives requests from your application
2. Forwards them to the actual MCP server
3. Receives the full response
4. Filters unnecessary fields
5. Returns the optimized response to your application

### Do I need to modify my application?

No! Just change the MCP server URL to point to SlimMCP:
```javascript
// Before
const notion = new MCPClient('http://localhost:3000');

// After
const notion = new MCPClient('http://localhost:8080/notion');
```

### Do I need to modify MCP servers?

No! SlimMCP works with any standard MCP server without modifications.

### Is SlimMCP compatible with the MCP specification?

Yes, SlimMCP is fully compliant with the MCP protocol. It implements an optional extension (`_meta.fieldSelection`) that servers can safely ignore if they don't understand it.

## Functionality Questions

### What filtering strategies are available?

- **Smart** (default) - Analyzes requests to infer needed fields (70-85% savings)
- **Minimal** - Returns only critical fields (85-95% savings)
- **Explicit** - Uses only fields you specify (varies)
- **Learning** - Learns from usage patterns (75-90% savings)
- **Passthrough** - No filtering, for debugging

### Can I control which fields are returned?

Yes! Use the `_meta.fieldSelection` extension:

```json
{
  "_meta": {
    "fieldSelection": {
      "include": ["title", "status", "assignee"],
      "strategy": "explicit"
    }
  }
}
```

### What are "critical fields"?

Critical fields are always included regardless of filtering strategy:
- `id` - Object identifiers
- `type` - Type discriminators  
- `url` - Resource URLs

These ensure responses remain valid and usable.

### Will filtering break my application?

No, SlimMCP is designed to be safe:
- Critical fields are always included
- Schema-required fields are always included
- If filtering fails, the full response is returned
- Circuit breakers prevent cascade failures

### Can I use SlimMCP with multiple MCP servers?

Yes! SlimMCP can proxy multiple servers simultaneously:

```yaml
servers:
  notion:
    url: "http://localhost:3000"
  slack:
    url: "http://localhost:3001"
  drive:
    url: "http://localhost:3002"
```

Access them at:
- `http://localhost:8080/notion`
- `http://localhost:8080/slack`
- `http://localhost:8080/drive`

## Performance Questions

### How much latency does SlimMCP add?

Typical added latency:
- **P50:** 5-6ms
- **P95:** 10-15ms  
- **P99:** 14-20ms

This is 2-3% overhead compared to direct connection.

### What's the throughput?

A single SlimMCP instance can handle:
- **1,000+ req/s** (minimal strategy)
- **500-800 req/s** (smart strategy)

SlimMCP scales horizontally for higher throughput.

### Does SlimMCP use a lot of memory?

No, SlimMCP is lightweight:
- **Idle:** ~50MB
- **100 req/s:** ~150MB
- **1000 req/s:** ~300MB

No memory leaks in 72+ hour tests.

### Can SlimMCP handle high load?

Yes! SlimMCP is production-ready:
- Tested at 10,000 concurrent requests
- 99.9%+ success rate
- Horizontal scaling support
- Circuit breakers for resilience

## Cost Questions

### How much money can I save?

Typical savings:
- **Daily requests:** 100 → **$105/month saved**
- **Daily requests:** 1,000 → **$1,055/month saved**
- **Daily requests:** 10,000 → **$10,548/month saved**

Based on Claude Sonnet pricing and 70% token reduction.

### When does SlimMCP pay for itself?

**Break-even:** ~67 requests/day (2,000/month)

Infrastructure cost: ~$70/month
Savings per 1K requests: ~$35

For any deployment with >100 requests/day, SlimMCP provides net savings.

### Does SlimMCP have licensing costs?

No! SlimMCP is open source (MIT License). Free to use in production.

## Deployment Questions

### How do I install SlimMCP?

Multiple options:

**Docker (recommended):**
```bash
docker pull slimmcp/proxy:latest
```

**Homebrew:**
```bash
brew install slimmcp/tap/slim
```

**Go:**
```bash
go install github.com/slimmcp/slimmcp@latest
```

**Binary:**
Download from [GitHub Releases](https://github.com/slimmcp/slimmcp/releases)

### Can I run SlimMCP in production?

Yes! SlimMCP is production-ready with:
- Circuit breakers
- Health checks
- Prometheus metrics
- Structured logging
- Zero-downtime deployments
- 99.9%+ uptime in tests

### How do I deploy SlimMCP on Kubernetes?

See our [Kubernetes guide](deployment.md#kubernetes). We provide:
- Deployment manifests
- Service definitions
- ConfigMap examples
- Ingress configuration
- Horizontal pod autoscaling

### Can I run multiple SlimMCP instances?

Yes! SlimMCP is stateless and scales horizontally:

```yaml
# docker-compose.yml
services:
  slimmcp:
    image: slimmcp/proxy:latest
    deploy:
      replicas: 3
```

Use a load balancer in front of multiple instances.

### What about high availability?

SlimMCP supports HA deployments:
- Run multiple instances
- Use load balancer (HAProxy, Nginx, cloud LB)
- Health checks for automatic failover
- No single point of failure (stateless)

## Configuration Questions

### Where does SlimMCP look for configuration?

Default locations (in order):
1. `--config` flag
2. `SLIM_CONFIG` environment variable
3. `./slim.yaml`
4. `/etc/slimmcp/slim.yaml`
5. `~/.config/slimmcp/slim.yaml`

### Can I use environment variables for configuration?

Yes! Example:
```bash
SLIM_LISTEN_ADDRESS=0.0.0.0:8080
SLIM_STRATEGY=smart
SLIM_SERVER_NOTION_URL=http://localhost:3000
```

### Can I reload configuration without restart?

Yes, two methods:

1. **Signal:**
```bash
kill -HUP $(pidof slim)
```

2. **API:**
```bash
curl -X POST http://localhost:8080/admin/reload
```

### How do I configure for specific tools?

Use tool-specific defaults:

```yaml
servers:
  notion:
    defaults:
      notion-fetch:
        fields: [id, title, url, status]
      notion-search:
        fields: [id, title, url]
```

## Security Questions

### Does SlimMCP store my data?

No! SlimMCP is stateless and doesn't persist:
- Request data
- Response data
- Credentials
- User information

Optional request logging can be disabled.

### How does authentication work?

SlimMCP passes authentication through unchanged:

```
Client → [Auth: Bearer TOKEN] → SlimMCP → [Auth: Bearer TOKEN] → MCP Server
```

SlimMCP never stores or modifies credentials.

### Is SlimMCP secure?

Yes! Security features:
- No credential storage
- No data persistence
- Optional TLS termination
- Optional client certificates
- Audit logging
- Minimal attack surface

### Can I use SlimMCP with private MCP servers?

Yes! SlimMCP works with:
- Localhost servers
- Private network servers
- VPC/internal servers
- Authenticated servers

### Does SlimMCP collect telemetry?

No. SlimMCP doesn't send any data externally. All monitoring is local (Prometheus metrics, logs).

## Compatibility Questions

### Which MCP servers does SlimMCP support?

All MCP-compliant servers! Tested with:
- ✅ Notion MCP
- ✅ Slack MCP
- ✅ Google Drive MCP
- ✅ Linear MCP
- ✅ GitHub MCP
- ✅ Jira MCP
- ✅ Any standard MCP server

### Does SlimMCP work with Claude Desktop?

Yes! Just update your MCP server URLs in Claude Desktop config:

```json
{
  "mcpServers": {
    "notion": {
      "url": "http://localhost:8080/notion"
    }
  }
}
```

### Does SlimMCP work with Cursor/Windsurf?

Yes! SlimMCP works with any application that uses MCP.

### What about future MCP protocol changes?

SlimMCP is designed to be forward-compatible:
- Optional extension that servers can ignore
- Graceful degradation
- Protocol version detection
- Regular updates to match MCP spec

## Monitoring Questions

### How do I monitor SlimMCP?

SlimMCP provides:
- **Prometheus metrics** at `:9090/metrics`
- **Health endpoint** at `/health`
- **Structured logs** (JSON or text)
- **Admin API** at `/admin/*` (optional)

### What metrics are available?

Key metrics:
- `mcp_optimizer_tokens_saved_total` - Total tokens saved
- `mcp_optimizer_savings_percentage` - Current savings %
- `mcp_optimizer_requests_total` - Request count
- `mcp_optimizer_errors_total` - Error count
- `mcp_optimizer_filtering_duration_seconds` - Filtering time

See [full list](api.md#metrics-endpoint).

### Can I use Grafana?

Yes! We provide a [pre-built dashboard](../examples/grafana-dashboard.json):

```bash
# Import to Grafana
curl -X POST http://grafana:3000/api/dashboards/import \
  -H "Content-Type: application/json" \
  -d @grafana-dashboard.json
```

### How do I track savings?

Check metrics:
```bash
curl http://localhost:9090/metrics | grep savings
```

Or check response metadata:
```json
{
  "_meta": {
    "fieldSelection": {
      "savings": 85.2
    }
  }
}
```

## Troubleshooting Questions

### Why am I not seeing any savings?

Common causes:

1. **Passthrough strategy:**
```yaml
proxy:
  strategy: smart  # Not passthrough!
```

2. **Filtering disabled:**
```yaml
optimization:
  enabled: true
```

3. **Native server support:**
Server might already do field selection. Check:
```bash
curl http://localhost:8080/admin/config | jq '.servers.notion.detectNativeSupport'
```

See [troubleshooting guide](troubleshooting.md#no-savings-observed).

### Why are my responses broken?

This shouldn't happen! SlimMCP includes safety features:
- Critical fields always included
- Required fields always included
- Fallback to full response on error

If it does happen:
1. Enable fallback:
```yaml
optimization:
  safety:
    fallbackOnError: true
```

2. Check logs:
```bash
slim logs | grep error
```

3. [Report the issue](https://github.com/slimmcp/slimmcp/issues)

### Why is SlimMCP slow?

Check:
1. **Strategy:** Try `minimal` for better performance
2. **Token estimation:** Disable if not needed
3. **Connection pools:** Increase for high load
4. **Resources:** Check CPU/memory

See [troubleshooting guide](troubleshooting.md#high-latency).

### Where can I get help?

- 💬 [Discord Community](https://discord.gg/slimmcp)
- 🐛 [GitHub Issues](https://github.com/slimmcp/slimmcp/issues)
- 💡 [GitHub Discussions](https://github.com/slimmcp/slimmcp/discussions)
- 📖 [Documentation](README.md)
- 📧 [Email](mailto:support@slimmcp.io)

## Advanced Questions

### Can I write custom filtering strategies?

Not yet, but it's on the roadmap! Currently, you can:
- Configure tool-specific defaults
- Use explicit field selection
- Adjust aggressiveness levels

Vote for custom strategies: [Feature Request #42](https://github.com/slimmcp/slimmcp/issues/42)

### How does learning mode work?

Learning mode tracks which fields are actually used and optimizes filtering:

1. **Phase 1:** Collect data (100+ samples)
2. **Phase 2:** Analyze patterns
3. **Phase 3:** Apply learned optimizations
4. **Phase 4:** Continuously improve

Accuracy improves from 89% (day 1) to 97% (day 30).

### Can I use SlimMCP as a library?

Yes! SlimMCP core is available as a Go library:

```go
import "github.com/slimmcp/slimmcp/pkg/filter"

f := filter.New(filter.Config{
    Strategy: "smart",
})

filtered := f.Filter(response, fields)
```

See [Go package docs](https://pkg.go.dev/github.com/slimmcp/slimmcp).

### How do I contribute?

We love contributions! See [CONTRIBUTING.md](../CONTRIBUTING.md) for:
- Code contributions
- Documentation improvements
- Bug reports
- Feature requests
- Testing and feedback

### Will this be added to the MCP protocol?

That's the goal! SlimMCP is:
1. Reference implementation
2. Gathering real-world data
3. Proving the concept
4. Building community support

We plan to submit an RFC to the MCP specification once we have sufficient adoption and feedback.

### What's the project roadmap?

See [ROADMAP.md](../ROADMAP.md) for details. Key milestones:
- ✅ Phase 1: Core functionality (Complete)
- 🚧 Phase 2: Smart inference (In Progress)
- 📋 Phase 3: Protocol extension (Next)
- 🎯 Phase 4: Production ready
- 🌍 Phase 5: Community adoption
- 🔮 Phase 6: MCP protocol RFC

### Is there a commercial version?

No. SlimMCP is fully open source (MIT License) and free to use in production.

For enterprise support, consulting, or custom development:
- 📧 [enterprise@slimmcp.io](mailto:enterprise@slimmcp.io)

### How can I support the project?

Many ways to help:
- ⭐ Star on GitHub
- 🐛 Report bugs
- 📝 Improve documentation
- 💬 Help others on Discord
- 📢 Share on social media
- 💰 [Sponsor development](https://github.com/sponsors/slimmcp)

## Comparison Questions

### How is SlimMCP different from caching?

**Caching** reduces redundant requests by storing responses.
**SlimMCP** reduces token usage by filtering responses.

They're complementary! Use both:
- Caching: Avoid duplicate requests
- SlimMCP: Optimize each response

### How is SlimMCP different from compression?

**Compression** (gzip) reduces network transfer size.
**SlimMCP** reduces semantic content (fewer fields).

Benefits:
- Compression: 60-70% smaller over network
- SlimMCP: 60-85% fewer tokens in LLM context

SlimMCP savings apply to LLM processing, not just network transfer.

### Why not just use GraphQL?

**GraphQL requires:**
- Server modifications
- Schema changes
- Client updates
- New protocol

**SlimMCP:**
- No server changes
- No client changes
- No protocol changes
- Drop-in proxy

SlimMCP is much faster to adopt!

### Can I use SlimMCP with non-MCP APIs?

No, SlimMCP is specifically for MCP protocol. For other APIs, consider:
- GraphQL (if API supports it)
- Custom middleware
- API-specific optimization tools

## Still Have Questions?

Can't find your answer? Ask us:

- 💬 [Discord Community](https://discord.gg/slimmcp) - Fast response, community help
- 💡 [GitHub Discussions](https://github.com/slimmcp/slimmcp/discussions) - Threaded discussions
- 📧 [Email](mailto:hello@slimmcp.io) - Direct contact

We typically respond within 24 hours!

---

**Found a mistake in this FAQ?** [Edit on GitHub](https://github.com/slimmcp/slimmcp/edit/main/docs/faq.md)
