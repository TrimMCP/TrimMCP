# Troubleshooting Guide

Common issues and solutions for SlimMCP.

## Table of Contents

- [Installation Issues](#installation-issues)
- [Connection Issues](#connection-issues)
- [Filtering Issues](#filtering-issues)
- [Performance Issues](#performance-issues)
- [Configuration Issues](#configuration-issues)
- [Monitoring Issues](#monitoring-issues)
- [Debugging Tools](#debugging-tools)

## Installation Issues

### Issue: Command Not Found

**Symptom:**
```bash
$ slim start
bash: slim: command not found
```

**Solutions:**

1. **Verify installation:**
```bash
which slim
```

2. **Check PATH:**
```bash
echo $PATH
export PATH=$PATH:/usr/local/bin
```

3. **Reinstall:**
```bash
# Homebrew
brew reinstall slimmcp

# Go
go install github.com/slimmcp/slimmcp@latest

# Binary
curl -L https://github.com/slimmcp/slimmcp/releases/latest/download/slim-$(uname-s)-$(uname -m) -o /usr/local/bin/slim
chmod +x /usr/local/bin/slim
```

### Issue: Permission Denied

**Symptom:**
```bash
$ slim start
Error: permission denied: /slim.yaml
```

**Solutions:**

1. **Check file permissions:**
```bash
ls -la slim.yaml
chmod 644 slim.yaml
```

2. **Run with sudo (not recommended):**
```bash
sudo slim start
```

3. **Change ownership:**
```bash
chown $USER:$USER slim.yaml
```

### Issue: Docker Image Pull Failed

**Symptom:**
```bash
Error response from daemon: manifest for slimmcp/proxy:latest not found
```

**Solutions:**

1. **Check image name:**
```bash
docker pull slimmcp/proxy:latest
docker pull slimmcp/proxy:v0.1.0  # Specific version
```

2. **Check Docker Hub status:**
```bash
curl https://hub.docker.com/v2/repositories/slimmcp/proxy/tags
```

3. **Build locally:**
```bash
git clone https://github.com/slimmcp/slimmcp
cd slimmcp
docker build -t slimmcp/proxy:latest .
```

## Connection Issues

### Issue: Connection Refused

**Symptom:**
```bash
$ curl http://localhost:8080/health
curl: (7) Failed to connect to localhost port 8080: Connection refused
```

**Diagnosis:**

1. **Check if SlimMCP is running:**
```bash
ps aux | grep slim
docker ps | grep slimmcp
```

2. **Check port binding:**
```bash
netstat -tuln | grep 8080
lsof -i :8080
```

3. **Check logs:**
```bash
slim logs
docker logs slimmcp
```

**Solutions:**

1. **Start SlimMCP:**
```bash
slim start
```

2. **Change port if in use:**
```yaml
# slim.yaml
proxy:
  listenAddress: "0.0.0.0:8081"
```

3. **Check firewall:**
```bash
sudo ufw allow 8080
sudo iptables -A INPUT -p tcp --dport 8080 -j ACCEPT
```

### Issue: Can't Connect to Downstream Server

**Symptom:**
```
Error: Failed to connect to downstream server: notion
```

**Diagnosis:**

1. **Verify downstream server is running:**
```bash
curl http://localhost:3000/health
```

2. **Check network connectivity:**
```bash
telnet localhost 3000
nc -zv localhost 3000
```

3. **Check SlimMCP logs:**
```bash
slim logs | grep "downstream"
```

**Solutions:**

1. **Fix URL in configuration:**
```yaml
servers:
  notion:
    url: "http://localhost:3000"  # Correct URL
```

2. **For Docker, use host.docker.internal:**
```yaml
servers:
  notion:
    url: "http://host.docker.internal:3000"
```

3. **Increase timeout:**
```yaml
servers:
  notion:
    url: "http://localhost:3000"
    timeout: 60s  # Increase from 30s
```

### Issue: Intermittent Connection Failures

**Symptom:**
```
Error: i/o timeout
Error: connection reset by peer
```

**Diagnosis:**

1. **Check circuit breaker status:**
```bash
curl http://localhost:9090/metrics | grep circuit_breaker
```

2. **Monitor error rate:**
```bash
curl http://localhost:9090/metrics | grep errors_total
```

**Solutions:**

1. **Adjust circuit breaker:**
```yaml
optimization:
  safety:
    circuitBreaker:
      threshold: 20  # More tolerant
      timeout: 30s   # Faster recovery
```

2. **Increase connection pool:**
```yaml
servers:
  notion:
    maxConnections: 200  # From 100
    maxIdleConnections: 50  # From 10
```

3. **Enable retries (if not already):**
```yaml
servers:
  notion:
    maxRetries: 3
    retryBackoff: 1s
```

## Filtering Issues

### Issue: No Savings Observed

**Symptom:**
```bash
$ curl http://localhost:9090/metrics | grep savings
mcp_optimizer_savings_percentage{server="notion"} 0
```

**Diagnosis:**

1. **Check strategy:**
```bash
$ curl http://localhost:8080/admin/config | jq '.proxy.strategy'
"passthrough"  # Problem: no filtering
```

2. **Check if filtering is being applied:**
```bash
slim logs | grep "Request filtered"
```

3. **Verify downstream responses:**
```bash
# Make a request and check _meta
curl -X POST http://localhost:8080/notion \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"notion-fetch","arguments":{"id":"page-123"}},"id":1}' \
  | jq '._meta.fieldSelection'
```

**Solutions:**

1. **Change strategy from passthrough:**
```yaml
proxy:
  strategy: smart  # Not passthrough
```

2. **Ensure filtering is enabled:**
```yaml
optimization:
  enabled: true
```

3. **Check for native support interference:**
```yaml
servers:
  notion:
    detectNativeSupport: false  # Force proxy filtering
```

### Issue: Broken Responses

**Symptom:**
```json
{
  "error": "Invalid response structure",
  "details": "Missing required field: id"
}
```

**Diagnosis:**

1. **Check which fields are being kept:**
```bash
slim logs | grep "fields_kept"
```

2. **Verify schema validation:**
```bash
slim validate-schema --server notion --tool notion-fetch
```

**Solutions:**

1. **Add to alwaysKeep:**
```yaml
optimization:
  alwaysKeep:
    - id  # Critical!
    - type
    - url
    - title
```

2. **Use less aggressive filtering:**
```yaml
servers:
  notion:
    aggressiveness: low  # From high
```

3. **Enable fallback:**
```yaml
optimization:
  safety:
    fallbackOnError: true  # Return full response on error
```

### Issue: False Negatives (Missing Needed Fields)

**Symptom:**
Client receives response but required field is missing.

**Diagnosis:**

1. **Check filtering accuracy:**
```bash
curl http://localhost:9090/metrics | grep filtering_accuracy
```

2. **Review logs for specific request:**
```bash
slim logs --request-id req_abc123
```

**Solutions:**

1. **Use explicit strategy:**
```json
{
  "_meta": {
    "fieldSelection": {
      "include": ["id", "title", "status", "assignee"],
      "strategy": "explicit"
    }
  }
}
```

2. **Add to tool defaults:**
```yaml
servers:
  notion:
    defaults:
      notion-fetch:
        fields: [id, title, url, status, assignee, properties]
```

3. **Enable learning mode:**
```yaml
optimization:
  learning:
    enabled: true
```

## Performance Issues

### Issue: High Latency

**Symptom:**
```
Requests taking >100ms longer than direct connection
```

**Diagnosis:**

1. **Measure filtering time:**
```bash
curl http://localhost:9090/metrics | grep filtering_duration
```

2. **Check strategy:**
```bash
$ curl http://localhost:8080/admin/config | jq '.proxy.strategy'
"learning"  # May be slower
```

3. **Check CPU usage:**
```bash
top -p $(pidof slim)
```

**Solutions:**

1. **Use faster strategy:**
```yaml
proxy:
  strategy: minimal  # Fastest
```

2. **Disable token estimation:**
```yaml
optimization:
  tokenEstimation:
    enabled: false
```

3. **Disable learning:**
```yaml
optimization:
  learning:
    enabled: false
```

4. **Scale horizontally:**
```bash
# Run multiple instances behind load balancer
```

### Issue: High Memory Usage

**Symptom:**
```bash
$ docker stats slimmcp
CONTAINER    MEM USAGE
slimmcp      2.5GB / 4GB
```

**Diagnosis:**

1. **Check for memory leak:**
```bash
# Monitor over time
watch -n 5 'docker stats slimmcp --no-stream'
```

2. **Check learning history size:**
```bash
$ curl http://localhost:8080/admin/config | jq '.optimization.learning.historySize'
100000  # Very large
```

**Solutions:**

1. **Reduce learning history:**
```yaml
optimization:
  learning:
    historySize: 10000  # From 100000
```

2. **Restart periodically:**
```bash
# Add to cron
0 2 * * * docker restart slimmcp
```

3. **Increase container memory:**
```yaml
# docker-compose.yml
services:
  slimmcp:
    mem_limit: 4g
```

### Issue: Low Throughput

**Symptom:**
Can only handle 100 req/s instead of expected 500+

**Diagnosis:**

1. **Check connection pool:**
```bash
curl http://localhost:9090/metrics | grep connection_pool
```

2. **Check CPU usage:**
```bash
top -p $(pidof slim)
```

3. **Check for bottlenecks:**
```bash
go tool pprof http://localhost:9090/debug/pprof/profile
```

**Solutions:**

1. **Increase connection pools:**
```yaml
servers:
  notion:
    maxConnections: 200
    maxIdleConnections: 50
```

2. **Scale horizontally:**
```bash
docker-compose up --scale slimmcp=4
```

3. **Optimize configuration:**
```yaml
proxy:
  strategy: minimal  # Fastest filtering
  
optimization:
  tokenEstimation:
    enabled: false
  learning:
    enabled: false
```

## Configuration Issues

### Issue: Configuration Not Loading

**Symptom:**
```
Error: Failed to load configuration from slim.yaml
```

**Diagnosis:**

1. **Validate YAML syntax:**
```bash
slim validate --config slim.yaml
```

2. **Check file path:**
```bash
ls -la slim.yaml
```

**Solutions:**

1. **Fix YAML syntax:**
```bash
# Use YAML validator
yamllint slim.yaml
```

2. **Specify config path:**
```bash
slim start --config /path/to/slim.yaml
```

3. **Use environment variables:**
```bash
export SLIM_CONFIG=/path/to/slim.yaml
slim start
```

### Issue: Configuration Changes Not Applied

**Symptom:**
Changed configuration but SlimMCP still uses old settings

**Solutions:**

1. **Restart SlimMCP:**
```bash
docker restart slimmcp
# or
systemctl restart slimmcp
```

2. **Trigger hot reload:**
```bash
kill -HUP $(pidof slim)
# or
curl -X POST http://localhost:8080/admin/reload
```

3. **Verify configuration:**
```bash
curl http://localhost:8080/admin/config | jq
```

## Monitoring Issues

### Issue: Metrics Not Available

**Symptom:**
```bash
$ curl http://localhost:9090/metrics
curl: (7) Failed to connect
```

**Diagnosis:**

1. **Check if metrics are enabled:**
```bash
$ curl http://localhost:8080/admin/config | jq '.monitoring.enabled'
false  # Problem!
```

2. **Check metrics port:**
```bash
netstat -tuln | grep 9090
```

**Solutions:**

1. **Enable monitoring:**
```yaml
monitoring:
  enabled: true
  metricsPort: 9090
```

2. **Check firewall:**
```bash
sudo ufw allow 9090
```

3. **Bind to correct interface:**
```yaml
monitoring:
  metricsAddress: "0.0.0.0:9090"  # Not 127.0.0.1
```

### Issue: Logs Not Appearing

**Symptom:**
No logs visible in stdout or log file

**Solutions:**

1. **Check log level:**
```yaml
monitoring:
  logLevel: debug  # More verbose
```

2. **Check log output:**
```yaml
monitoring:
  logOutput: stdout  # Or file
  logFile: /var/log/slimmcp/slimmcp.log
```

3. **Check file permissions:**
```bash
ls -la /var/log/slimmcp/
sudo chmod 755 /var/log/slimmcp
```

## Debugging Tools

### Enable Debug Mode

```yaml
dev:
  enabled: true
  prettyPrint: true
  verboseLogging: true
```

### View Active Configuration

```bash
curl http://localhost:8080/admin/config | jq
```

### Trace Specific Request

```bash
# Add request ID to trace
curl -X POST http://localhost:8080/notion \
  -H "X-Request-ID: trace-123" \
  -H "Content-Type: application/json" \
  -d '{...}'

# Check logs for trace-123
slim logs | grep trace-123
```

### Check Health of All Components

```bash
curl http://localhost:8080/health | jq
```

Output:
```json
{
  "status": "healthy",
  "servers": {
    "notion": {
      "status": "connected",
      "latency": "23ms"
    },
    "slack": {
      "status": "error",
      "error": "connection timeout"
    }
  }
}
```

### Analyze Metrics

```bash
# Token savings
curl -s http://localhost:9090/metrics | grep savings_percentage

# Error rate
curl -s http://localhost:9090/metrics | grep errors_total

# Latency percentiles
curl -s http://localhost:9090/metrics | grep filtering_duration
```

### Use Verbose Mode

```bash
slim start --config slim.yaml --verbose
```

### Export Diagnostic Bundle

```bash
slim diagnostics --output diag.tar.gz
```

Includes:
- Current configuration
- Recent logs
- Metrics snapshot
- Health check results
- System info

### Test Filtering Locally

```bash
# Test what fields would be kept
slim test-filter --tool notion-fetch --strategy smart --input test.json
```

## Getting Help

### Before Asking for Help

Gather this information:

1. **Version:**
```bash
slim version
```

2. **Configuration:**
```bash
cat slim.yaml
```

3. **Logs:**
```bash
slim logs --tail 100
```

4. **Metrics:**
```bash
curl http://localhost:9090/metrics
```

5. **Health:**
```bash
curl http://localhost:8080/health
```

### Where to Get Help

- 💬 [Discord Community](https://discord.gg/slimmcp)
- 🐛 [GitHub Issues](https://github.com/slimmcp/slimmcp/issues)
- 💡 [GitHub Discussions](https://github.com/slimmcp/slimmcp/discussions)
- 📧 [Email Support](mailto:support@slimmcp.io)

### Creating a Bug Report

Include:

1. **Environment:**
   - OS and version
   - SlimMCP version
   - Deployment method (Docker, binary, etc.)

2. **Configuration:**
   - Sanitized `slim.yaml`
   - Relevant environment variables

3. **Reproduction:**
   - Steps to reproduce
   - Expected behavior
   - Actual behavior

4. **Logs:**
   - Relevant log entries
   - Error messages

5. **Metrics:**
   - Relevant metrics
   - Screenshots if applicable

## See Also

- [Getting Started Guide](getting-started.md)
- [Configuration Reference](configuration.md)
- [FAQ](faq.md)
- [Performance Benchmarks](benchmarks.md)
