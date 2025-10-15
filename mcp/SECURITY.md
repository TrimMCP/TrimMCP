# Security Policy

## Supported Versions

Security updates are provided for the following versions of SlimMCP:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |
| < 0.1   | :x:                |

Once SlimMCP reaches 1.0, we will support:
- Latest major version (e.g., 1.x)
- Previous major version for 6 months after new major release

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

### How to Report

**Email:** security@slimmcp.io

Please include:

1. **Description** - Detailed description of the vulnerability
2. **Impact** - Potential impact if exploited
3. **Reproduction** - Steps to reproduce the issue
4. **Proof of Concept** - Code or configuration that demonstrates the issue (if applicable)
5. **Suggested Fix** - If you have ideas on how to fix it (optional)

### What to Expect

- **Acknowledgment** - Within 48 hours of your report
- **Initial Assessment** - Within 5 business days
- **Status Updates** - Every 7 days until resolution
- **Resolution Timeline** - Depends on severity:
  - **Critical** - 7 days
  - **High** - 30 days
  - **Medium** - 90 days
  - **Low** - 120 days

### Disclosure Policy

- We follow **coordinated disclosure**
- Security advisories published after fix is available
- Credit given to reporter (unless they prefer anonymity)
- CVE assigned for significant vulnerabilities

## Security Best Practices

### Deployment

When deploying SlimMCP:

1. **Use HTTPS** - Always terminate TLS at SlimMCP or reverse proxy
2. **Network Isolation** - Deploy in isolated network segment
3. **Least Privilege** - Run with minimal required permissions
4. **Monitoring** - Enable audit logging and metrics
5. **Updates** - Keep SlimMCP and dependencies up to date

### Configuration

Secure configuration recommendations:

```yaml
proxy:
  # Bind to localhost or internal network only
  listenAddress: "127.0.0.1:8080"

  # Enable request logging for audit trail
  requestLogging:
    enabled: true
    logLevel: info

  # Configure timeouts to prevent DoS
  timeouts:
    read: 10s
    write: 10s
    idle: 60s

monitoring:
  # Enable metrics for security monitoring
  enabled: true
  metricsPort: 9090
```

### Secrets Management

**Never commit:**
- API keys or tokens
- Passwords or credentials
- Private keys or certificates

**Use environment variables or secret management systems:**

```yaml
proxy:
  servers:
    notion:
      url: "http://localhost:3000"
      # Good: Use environment variable
      token: ${NOTION_TOKEN}
      # Bad: Never hardcode
      # token: "secret_abc123"
```

## Known Security Considerations

### Architecture

SlimMCP is designed as a **stateless proxy**:

- ✅ **No data persistence** - No stored credentials or data
- ✅ **Pass-through authentication** - Auth tokens passed to upstream servers
- ✅ **No credential storage** - SlimMCP never stores API keys
- ✅ **Audit logging** - All requests logged for forensics

### Threat Model

**In Scope:**
- Request/response tampering
- Injection attacks (JSON-RPC, header injection)
- Denial of service
- Information disclosure
- Authentication bypass

**Out of Scope:**
- Vulnerabilities in upstream MCP servers
- TLS/SSL configuration (handled by reverse proxy)
- Operating system vulnerabilities
- Network-level attacks (DDoS at network layer)

### Dependency Security

We use **Dependabot** to monitor dependencies:

- Automated security updates for Go modules
- Weekly vulnerability scanning
- Security advisories tracked in GitHub

## Security Features

### Built-in Protections

1. **Input Validation**
   - JSON-RPC request validation
   - Field selection syntax validation
   - Request size limits

2. **Error Handling**
   - No stack traces in production responses
   - Sensitive data scrubbed from error messages
   - Safe fallback on errors

3. **Rate Limiting** (planned for v0.2)
   - Per-client request limits
   - Circuit breakers for upstream failures
   - Adaptive rate limiting

4. **Audit Logging**
   - Structured JSON logs
   - Request/response metadata
   - Timestamp and client identification

## Compliance

### GDPR

SlimMCP processes data in a privacy-preserving manner:

- **No persistent storage** - Data not retained after request
- **Minimal logging** - Only metadata logged by default
- **Data minimization** - Filtering reduces data volume
- **Transparency** - Open-source for audit

### SOC 2 / ISO 27001

For organizations requiring compliance:

- Enable comprehensive audit logging
- Implement access controls at network level
- Use secure deployment practices
- Monitor security metrics via Prometheus

## Security Roadmap

Planned security enhancements:

- **v0.2** - Rate limiting and circuit breakers
- **v0.3** - mTLS support for server authentication
- **v0.4** - Request signing and verification
- **v1.0** - Full security audit and penetration testing

## Hall of Fame

We recognize security researchers who responsibly disclose vulnerabilities:

<!-- Security researchers will be listed here -->

*Be the first to contribute!*

## Contact

- **Security Issues:** security@ideacrafters.com
- **General Questions:** hello@slimmcp.com
- **Discord:** [Join our community](https://discord.slimmcp.com)

---

**Thank you for helping keep SlimMCP secure!** 🔒
