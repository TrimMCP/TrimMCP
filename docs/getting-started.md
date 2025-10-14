# Getting Started with SlimMCP

Welcome! This guide will help you get SlimMCP up and running in less than 5 minutes.

## What You'll Need

- An MCP server you want to optimize (e.g., Notion, Slack, Google Drive)
- Basic command line knowledge
- 5 minutes of your time

## Installation

Choose your preferred installation method:

### Option 1: Docker (Recommended)

Fastest way to get started:

```bash
# Pull the image
docker pull slimmcp/proxy:latest

# Run it
docker run -d \
  --name slimmcp \
  -p 8080:8080 \
  -p 9090:9090 \
  -v $(pwd)/slim.yaml:/slim.yaml \
  slimmcp/proxy:latest
```

### Option 2: Homebrew (macOS/Linux)

```bash
brew install slimmcp/tap/slim
slim start
```

### Option 3: Go Install

```bash
go install github.com/slimmcp/slimmcp@latest
slimmcp start
```

### Option 4: Binary Download

Download from [GitHub Releases](https://github.com/slimmcp/slimmcp/releases):

```bash
# Linux
curl -L https://github.com/slimmcp/slimmcp/releases/latest/download/slim-linux-amd64 -o slim
chmod +x slim
./slim start

# macOS
curl -L https://github.com/slimmcp/slimmcp/releases/latest/download/slim-darwin-amd64 -o slim
chmod +x slim
./slim start

# Windows
# Download slim-windows-amd64.exe from releases page
```

## Basic Configuration

Create a file called `slim.yaml` in your working directory:

```yaml
proxy:
  # Address to listen on
  listenAddress: "0.0.0.0:8080"
  
  # Point to your existing MCP servers
  servers:
    notion:
      url: "http://localhost:3000"
    slack:
      url: "http://localhost:3001"
  
  # Use smart filtering (default)
  strategy: smart

# Enable monitoring
monitoring:
  enabled: true
  metricsPort: 9090
  logLevel: info
```

## Start SlimMCP

### With Docker

```bash
docker run -d \
  --name slimmcp \
  -p 8080:8080 \
  -p 9090:9090 \
  -v $(pwd)/slim.yaml:/slim.yaml \
  slimmcp/proxy:latest
```

### With Binary

```bash
slim start --config slim.yaml
```

You should see:

```
INFO  SlimMCP starting...
INFO  Loaded configuration from slim.yaml
INFO  Server listening on 0.0.0.0:8080
INFO  Metrics server listening on 0.0.0.0:9090
INFO  Connected to downstream server: notion (http://localhost:3000)
INFO  Connected to downstream server: slack (http://localhost:3001)
INFO  SlimMCP ready! 🪶
```

## Update Your Application

Change your MCP client to point to SlimMCP instead of the direct server:

### JavaScript/TypeScript

**Before:**
```javascript
import { MCPClient } from '@modelcontextprotocol/sdk';

const notion = new MCPClient('http://localhost:3000');
```

**After:**
```javascript
import { MCPClient } from '@modelcontextprotocol/sdk';

// Just change the URL - that's it!
const notion = new MCPClient('http://localhost:8080/notion');
```

### Python

**Before:**
```python
from mcp import Client

notion = Client("http://localhost:3000")
```

**After:**
```python
from mcp import Client

# Just change the URL - that's it!
notion = Client("http://localhost:8080/notion")
```

### Claude Desktop (config)

**Before:**
```json
{
  "mcpServers": {
    "notion": {
      "url": "http://localhost:3000"
    }
  }
}
```

**After:**
```json
{
  "mcpServers": {
    "notion": {
      "url": "http://localhost:8080/notion"
    }
  }
}
```

## Verify It's Working

### 1. Check Health

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "version": "0.1.0",
  "uptime": "2m30s"
}
```

### 2. Make a Request

Use your application normally. SlimMCP will automatically filter responses.

### 3. Check Metrics

```bash
curl http://localhost:9090/metrics | grep savings
```

You should see something like:
```
mcp_optimizer_savings_percentage{server="notion"} 82.3
mcp_optimizer_tokens_saved_total{server="notion"} 12847
```

### 4. View Logs

**Docker:**
```bash
docker logs slimmcp
```

**Binary:**
```bash
# Logs are output to stdout
# Or check the log file if configured
tail -f slimmcp.log
```

You should see entries like:
```json
{
  "level": "info",
  "message": "Request filtered successfully",
  "server": "notion",
  "tool": "notion-fetch",
  "tokens_before": 2847,
  "tokens_after": 421,
  "savings_pct": 85.2
}
```

## Test Your Setup

Let's make a test request to verify everything is working:

```bash
# Using curl to test the proxy
curl -X POST http://localhost:8080/notion \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "notion-fetch",
      "arguments": {
        "id": "your-page-id"
      }
    },
    "id": 1
  }'
```

Check the response for the `_meta.fieldSelection` section to see filtering stats.

## Understanding Your Savings

After running for a few minutes, check your savings dashboard:

```bash
# Total tokens saved
curl -s http://localhost:9090/metrics | grep "mcp_optimizer_tokens_saved_total"

# Savings percentage by server
curl -s http://localhost:9090/metrics | grep "mcp_optimizer_savings_percentage"

# Request count
curl -s http://localhost:9090/metrics | grep "mcp_optimizer_requests_total"
```

## Common First-Time Issues

### Issue: Connection Refused

**Symptom:** `Connection refused` when accessing SlimMCP

**Solution:**
1. Check SlimMCP is running: `docker ps` or `ps aux | grep slim`
2. Verify port 8080 is not in use: `lsof -i :8080`
3. Check firewall rules

### Issue: Can't Connect to Downstream Server

**Symptom:** `Failed to connect to downstream server: notion`

**Solution:**
1. Verify your MCP server is running: `curl http://localhost:3000/health`
2. Check the URL in `slim.yaml` is correct
3. If using Docker, ensure network connectivity: use `host.docker.internal` instead of `localhost`

```yaml
servers:
  notion:
    url: "http://host.docker.internal:3000"  # For Docker
```

### Issue: No Savings Showing

**Symptom:** Metrics show 0% savings

**Possible causes:**
1. Not enough requests yet (wait for a few requests)
2. Strategy too conservative (try `strategy: smart`)
3. Downstream server already returns minimal data

**Debug:**
```bash
# Check if requests are being processed
curl http://localhost:9090/metrics | grep requests_total

# Enable debug logging
# In slim.yaml:
monitoring:
  logLevel: debug
```

## Next Steps

Now that SlimMCP is running:

1. **Monitor Your Savings** - Set up [Grafana dashboards](configuration.md#grafana-setup)
2. **Optimize Settings** - Tune [filtering strategies](configuration.md#filtering-strategies)
3. **Add More Servers** - Configure [multiple MCP servers](configuration.md#multiple-servers)
4. **Production Deployment** - See [deployment guide](deployment.md)

## Quick Reference

### Essential Commands

```bash
# Start SlimMCP
slim start

# Start with config file
slim start --config slim.yaml

# Check version
slim version

# Check health
curl http://localhost:8080/health

# View metrics
curl http://localhost:9090/metrics

# Stop SlimMCP (Docker)
docker stop slimmcp

# View logs (Docker)
docker logs -f slimmcp
```

### Essential Configuration

```yaml
# Minimum viable configuration
proxy:
  listenAddress: "0.0.0.0:8080"
  servers:
    myserver:
      url: "http://localhost:3000"

# Recommended production configuration
proxy:
  listenAddress: "0.0.0.0:8080"
  strategy: smart
  servers:
    notion:
      url: "http://localhost:3000"
      timeout: 30s
  
optimization:
  safety:
    fallbackOnError: true
    circuitBreaker:
      enabled: true

monitoring:
  enabled: true
  metricsPort: 9090
  logLevel: info
```

## Getting Help

- 📖 [Full Documentation](README.md)
- 💬 [Discord Community](https://discord.gg/slimmcp)
- 🐛 [Report Issues](https://github.com/slimmcp/slimmcp/issues)
- 💡 [Discussions](https://github.com/slimmcp/slimmcp/discussions)

## What's Next?

- Learn about [configuration options](configuration.md)
- Understand [how SlimMCP works](architecture.md)
- Review [performance benchmarks](benchmarks.md)
- Read the [troubleshooting guide](troubleshooting.md)
- Check the [FAQ](faq.md)

---

**Congratulations!** You've successfully set up SlimMCP. You're now saving tokens and money with every request. 🎉
