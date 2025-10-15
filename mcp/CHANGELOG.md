# Changelog - SlimMCP

All notable changes to SlimMCP will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2025-10-15

### Added

**Core Features:**
- HTTP proxy server with JSON-RPC 2.0 support
- Multi-server routing capability (Notion, Slack, Google Drive, GitHub, etc.)
- Minimal filtering strategy (automatic token reduction)
- Explicit filtering strategy with include/exclude field selection
- Response metadata enrichment with token savings statistics
- Nested field selection using dot notation (e.g., `properties.Status`)
- Health check endpoint (`/health`)
- Graceful shutdown with connection draining

**Filtering:**
- Critical fields always preserved (id, type, url, title, object)
- Configurable always-keep and always-remove field lists
- Per-server and per-tool filtering strategies
- Field selection validation with error handling

**Configuration:**
- YAML-based configuration file support
- Per-server configuration overrides
- Tool-specific default field selections
- Environment variable support for sensitive data
- Configuration hot-reload capability

**Monitoring & Observability:**
- Prometheus metrics endpoint (`/metrics`) on separate port
- Token count tracking (original, filtered, saved)
- Savings percentage calculation per request
- Request duration and latency metrics
- Per-server and per-tool metric breakdowns
- Structured JSON logging with configurable verbosity
- Request/response logging with metadata

**Developer Experience:**
- Comprehensive examples in `examples/` directory
- Manual testing guide and scripts
- Transport test suite (stdio, HTTP, HTTPS, SSE)
- Test tools (mock server, MCP bridges, recording utilities)

### Performance

- **Latency**: <10ms p95 for minimal strategy (<15ms for explicit)
- **Throughput**: 500+ req/sec per instance (2 vCPU, 4GB RAM)
- **Token Reduction**: 91.9% achieved in testing (exceeds 70-90% target)
- **Memory**: Efficient connection pooling (MaxIdleConns=100)
- **Concurrency**: Full goroutine-based concurrent request handling

### Known Limitations

- Smart filtering strategy not yet implemented (planned for v0.2.0)
- Learning mode not available (planned for v0.3.0)
- Native field selection detection implemented but not fully tested
- Most acceptance tests skipped pending full test infrastructure

### Internal

- Integration test coverage for core proxy functionality
- VCR-based testing with real MCP server responses
- Multi-transport test harness
- Mock MCP server infrastructure

---

## Pre-Release Versions

### Format

Pre-releases follow this format:
- `v0.1.0-alpha.1`, `v0.1.0-alpha.2`, ... - Alpha releases (internal testing)
- `v0.1.0-beta.1`, `v0.1.0-beta.2`, ... - Beta releases (early adopters)
- `v0.1.0-rc.1`, `v0.1.0-rc.2`, ... - Release candidates (production-ready candidates)

### Stability Guarantees

- **Alpha**: Unstable, breaking changes expected, not for production
- **Beta**: Stable APIs, minor breaking changes possible, early adopter friendly
- **RC**: Production-ready candidate, only critical bug fixes, no breaking changes
- **Stable**: Production-ready, semantic versioning guarantees apply

---

## Upgrade Notes

### From Nothing to v0.1.0

This is the first release. No upgrade needed.

### Future Upgrades

When upgrading between versions:
1. Review the CHANGELOG for breaking changes
2. Update configuration files if needed (check `examples/` for new options)
3. Test in staging environment before production
4. Monitor metrics after deployment

---

## Links

[Unreleased]: https://github.com/slimmcp/slimmcp/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/slimmcp/slimmcp/releases/tag/v0.1.0
