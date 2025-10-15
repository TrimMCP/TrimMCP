# Implementation Plan: MCP Response Optimizer Proxy

**Branch**: `001-trim` | **Date**: 2025-10-13 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-trim/spec.md`

## Summary

Build SlimMCP, a transparent HTTP proxy for Model Context Protocol (MCP) servers that intelligently filters response fields to reduce token usage by 70-90%. The proxy sits between LLM clients and MCP servers, analyzing requests to infer needed fields, filtering responses while preserving structure, and exposing comprehensive metrics. Key capabilities include multiple filtering strategies (smart, minimal, explicit, passthrough), support for nested field selection, server capability detection, and production-ready observability with Prometheus metrics and structured logging. Target performance: <15ms p95 latency, 500+ req/sec per instance, horizontal scaling to 1,500+ req/sec (3 instances), 99.9% availability.

## Technical Context

**Language/Version**: Go 1.25+ (minimum 1.24+)
**Primary Dependencies**:
- Standard library (net/http, encoding/json)
- Prometheus client for Go (github.com/prometheus/client_golang v1.23.2)
- MCP Go SDK (github.com/mark3labs/mcp-go v0.42.0-beta.1) for protocol implementation
- YAML parser (github.com/goccy/go-yaml v1.18.0) for configuration
**Storage**: N/A (stateless proxy, optional file-based configuration)
**Testing**: Go testing framework (`go test`), table-driven tests, integration tests with mock MCP servers
**Target Platform**: Linux server (primary), Docker containers, Kubernetes
**Project Type**: Single project (command-line proxy service)
**Performance Goals**:
- <10ms p95 latency for minimal strategy
- <15ms p95 latency for smart strategy
- 500+ requests/second per instance (single instance)
- Horizontal scalability to 1,500+ req/sec (3+ instances with load balancing)
**Constraints**:
- <200ms total p95 latency (including downstream)
- Memory usage scales with response size (must handle 10MB responses)
- Stateless operation (no persistent storage required)
- 99.9% availability (three nines uptime)
**Scale/Scope**:
- 10-100K requests/day per instance
- Support 10+ downstream MCP servers simultaneously
- Handle response objects with 100+ fields
- Filter nested structures up to 10 levels deep

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Note**: No constitution file found with project-specific principles. Using general software engineering best practices:

### Pre-Design Gates ✅

- **Simplicity**: Start with minimal viable filtering (minimal strategy) before adding intelligence (smart strategy)
- **Testability**: All components must be independently testable with clear interfaces
- **Observability**: Metrics and logging built-in from the start (not added later)
- **Production Readiness**: Health checks, graceful shutdown, error handling as first-class concerns

### Design Principles to Follow

1. **Library-First Architecture**: Core filtering logic as reusable packages independent of HTTP layer
2. **Interface-Driven**: Define clear contracts between proxy, analyzer, filter, and client components
3. **Test-First Development**: Write acceptance tests before implementation (per spec scenarios)
4. **Zero Breaking Changes**: Backward compatibility with existing MCP clients and servers
5. **Fail-Safe Defaults**: When in doubt, return full response (configurable fallback behavior)

### Post-Design Re-Check (After Phase 1)

Will validate:
- [ ] Core filtering library can be tested without HTTP server
- [ ] All public APIs have accompanying tests
- [ ] Metrics collection doesn't impact performance (<1ms overhead)
- [ ] Configuration can be reloaded without restart
- [ ] No global state that prevents horizontal scaling

## Project Structure

### Monorepo Overview

**Repository**: `IdeaCraftersHQ/oss-mcp-slim` (SlimMCP development monorepo)

This is the development monorepo specifically for SlimMCP (not a multi-project OSS MCP umbrella). All SlimMCP development happens here with automated deployment to production repositories in the `slimmcp/*` GitHub organization. See constitution.md Repository Architecture section for complete governance model.

### Documentation (this feature)

```
specs/001-trim/
├── spec.md              # Feature specification
├── plan.md              # This file (/speckit.plan output)
├── research.md          # Phase 0 - Technology decisions and patterns
├── data-model.md        # Phase 1 - Data structures and entities
├── quickstart.md        # Phase 1 - Developer onboarding guide
├── contracts/           # Phase 1 - API contracts and schemas
│   ├── mcp-request.json    # JSON-RPC request schema
│   ├── mcp-response.json   # JSON-RPC response schema
│   ├── field-selection.json # _meta.fieldSelection extension schema
│   └── config-schema.yaml  # Configuration file schema
└── tasks.md             # Phase 2 (/speckit.tasks output - NOT YET CREATED)
```

### Source Code (mcp/ subdirectory)

**Note**: All paths below are relative to `mcp/` directory in the monorepo.

```
mcp/
├── cmd/
│   └── slimmcp/
│       └── main.go              # CLI entry point
│
├── pkg/
│   ├── proxy/
│   │   ├── server.go            # HTTP server implementation
│   │   ├── handler.go           # Request/response handling
│   │   ├── router.go            # Server routing logic
│   │   └── middleware.go        # Logging, metrics middleware
│   │
│   ├── analyzer/
│   │   ├── analyzer.go          # Request analysis interface
│   │   ├── smart.go             # Smart field inference
│   │   ├── minimal.go           # Minimal strategy
│   │   ├── explicit.go          # Explicit field selection
│   │   └── passthrough.go       # No filtering strategy
│   │
│   ├── filter/
│   │   ├── filter.go            # Response filtering engine
│   │   ├── recursive.go         # Recursive object filtering
│   │   ├── validator.go         # Field validation logic
│   │   └── metadata.go          # Filtering statistics
│   │
│   ├── schema/
│   │   ├── manager.go           # Tool schema management
│   │   ├── loader.go            # Schema loading (server/config)
│   │   ├── validator.go         # Schema validation
│   │   └── cache.go             # Schema caching
│   │
│   ├── capability/
│   │   ├── detector.go          # Server capability detection
│   │   ├── cache.go             # Capability caching
│   │   └── types.go             # Capability data structures
│   │
│   ├── client/
│   │   ├── mcp.go               # MCP client for downstream servers
│   │   ├── pool.go              # Connection pooling
│   │   └── retry.go             # Retry logic
│   │
│   ├── metrics/
│   │   ├── collector.go         # Metrics collection
│   │   ├── prometheus.go        # Prometheus integration
│   │   └── types.go             # Metric definitions
│   │
│   ├── config/
│   │   ├── config.go            # Configuration structures
│   │   ├── loader.go            # YAML configuration loading
│   │   ├── validator.go         # Configuration validation
│   │   └── reload.go            # Hot reload support
│   │
│   └── types/
│       ├── request.go           # MCP request types
│       ├── response.go          # MCP response types
│       ├── meta.go              # _meta extension types
│       └── errors.go            # Error types
│
├── internal/
│   └── testutil/
│       ├── mock_server.go       # Mock MCP server for tests
│       ├── fixtures.go          # Test fixtures
│       └── assertions.go        # Custom test assertions
│
├── tests/
│   ├── integration/
│   │   ├── proxy_test.go        # End-to-end proxy tests
│   │   ├── filtering_test.go    # Filtering accuracy tests
│   │   ├── strategies_test.go   # Strategy comparison tests
│   │   └── performance_test.go  # Performance benchmarks
│   │
│   ├── contract/
│   │   ├── mcp_compliance_test.go   # MCP protocol compliance
│   │   ├── jsonrpc_test.go          # JSON-RPC 2.0 compliance
│   │   └── extension_test.go        # Field selection extension tests
│   │
│   └── acceptance/
│       ├── user_story_1_test.go # Automatic token reduction tests
│       ├── user_story_2_test.go # Explicit field selection tests
│       ├── user_story_3_test.go # Monitoring tests
│       ├── user_story_4_test.go # Multi-server proxy tests
│       └── user_story_5_test.go # Native capability detection tests
│
├── examples/
│   ├── dev.yaml                 # Development configuration
│   ├── prod.yaml                # Production configuration
│   ├── minimal-example.go       # Minimal usage example
│   └── grafana-dashboard.json   # Grafana dashboard
│
├── docs/                        # Additional MCP-specific documentation
│   └── images/                  # Images for README
│       └── README.md
│
├── go.mod
├── go.sum
├── Makefile                     # Build and test automation
├── Dockerfile
├── .dockerignore
├── README.md                    # (already exists)
├── LICENSE
└── CONTRIBUTING.md
```

**Structure Decision**: Selected **Option 1: Single project** structure because SlimMCP is a focused proxy service without separate frontend/backend concerns. The `pkg/` directory contains reusable library packages, `cmd/` has the CLI entry point, `internal/` has testing utilities, and `tests/` is organized by test type (integration, contract, acceptance). This follows standard Go project layout conventions and keeps the codebase simple while maintaining clear separation of concerns.

**Monorepo Context**:
- **MCP Proxy Code**: All SlimMCP Go code resides in `mcp/` subdirectory, deployed to `slimmcp/slimmcp` repository
- **Documentation Site**: Public docs live in `docs/` at monorepo root (not under `mcp/docs/`), deployed to `slimmcp/docs` repository and hosted at slimmcp.com/docs
- **Feature Specs**: Planning artifacts in `specs/` and `drafts/` stay in monorepo only
- **Development Tools**: `.specify/`, `.claude/`, `.codex/` configuration stays in monorepo
- **Deployment**: GitHub Actions automatically sync `mcp/` and `docs/` to their respective production repos

## Complexity Tracking

*No constitution violations detected - no entries needed.*

All design choices align with simplicity principles:
- Single binary deployment (no microservices needed)
- Stateless architecture (no database or persistent storage)
- Standard HTTP proxy pattern (well-understood)
- Clear package separation (single responsibility per package)

## Phase 0: Outline & Research

### Research Questions to Resolve

The following areas require research and decision documentation:

1. **Go HTTP Server Patterns**
   - Best practices for production HTTP servers in Go
   - Graceful shutdown patterns
   - Connection pooling for downstream requests
   - Request context propagation

2. **JSON-RPC 2.0 Implementation**
   - Existing Go libraries vs. custom implementation
   - Batch request handling patterns
   - Error code standards and handling

3. **Field Inference Algorithms**
   - Parameter parsing strategies (regex, AST, heuristics)
   - Common field naming patterns in MCP servers
   - How to detect field relationships (e.g., assignee.name requires assignee)

4. **Schema Management**
   - MCP schema discovery mechanisms
   - Schema versioning and cache invalidation
   - Fallback strategies when schemas unavailable

5. **Performance Optimization**
   - JSON parsing/filtering performance in Go
   - Memory-efficient response filtering (streaming vs. buffered)
   - Caching strategies for capabilities and schemas

6. **Testing Strategies**
   - Mock MCP server implementation patterns
   - Table-driven testing for multiple strategies
   - Performance benchmarking methodology

7. **Configuration Management**
   - YAML parsing with goccy/go-yaml (selected for better error messages)
   - Hot reload implementation without downtime
   - Configuration validation approaches

8. **Prometheus Integration**
   - Metric types for proxy monitoring (counters, histograms, gauges)
   - Label strategies (cardinality concerns)
   - Performance impact of metrics collection

### Research Outputs

Will produce `research.md` with:
- **Technology Choices**: Libraries and frameworks selected with rationale
- **Design Patterns**: Architectural patterns for each component
- **Implementation Approach**: High-level strategy for core algorithms
- **Tradeoffs**: Documented decisions with alternatives considered

## Phase 1: Design & Contracts

### Prerequisites
- Phase 0 `research.md` complete
- All technology choices made
- Design patterns identified

### Artifacts to Generate

#### 1. Data Model (`data-model.md`)

Document core data structures:

**Request Entities**:
- `MCPRequest`: JSON-RPC 2.0 request structure
- `FieldSelectionSpec`: _meta.fieldSelection extension
- `ToolInvocation`: Tool name and arguments

**Response Entities**:
- `MCPResponse`: JSON-RPC 2.0 response structure
- `FilteringStatistics`: Original size, filtered size, savings
- `ResponseMetadata`: Filtering metadata for client

**Domain Entities**:
- `ToolSchema`: Tool definition with field metadata
- `ServerCapabilities`: Capability flags and protocol version
- `FilterContext`: Combined request + schema + strategy context
- `FieldCategory`: Critical/Required/Requested/Related/Optional

**Configuration Entities**:
- `ProxyConfig`: Top-level configuration
- `ServerConfig`: Per-server settings
- `OptimizationConfig`: Filtering behavior settings
- `MonitoringConfig`: Metrics and logging settings

#### 2. API Contracts (`contracts/`)

Generate JSON schemas and OpenAPI specs:

**JSON-RPC Schemas**:
- `mcp-request.json`: Request structure with _meta extension
- `mcp-response.json`: Response structure with _meta extension
- `field-selection.json`: Field selection extension schema

**Configuration Schema**:
- `config-schema.yaml`: Complete YAML configuration schema with validation rules

**HTTP API Spec**:
- `openapi.yaml`: OpenAPI 3.0 spec for HTTP endpoints (/, /health, /metrics, /admin/*)

#### 3. Developer Quickstart (`quickstart.md`)

Create onboarding guide covering:
- Development environment setup
- Running the proxy locally
- Running tests
- Adding a new filtering strategy
- Debugging request/response flow
- Common troubleshooting

#### 4. Agent Context Update

Run `.specify/scripts/bash/update-agent-context.sh claude` to update `.claude/` with:
- Go 1.25+ as primary language (minimum 1.24+)
- Key dependencies:
  - github.com/mark3labs/mcp-go v0.42.0-beta.1 (MCP protocol implementation)
  - github.com/prometheus/client_golang v1.23.2 (metrics)
  - github.com/goccy/go-yaml v1.18.0 (configuration)
- Project structure and conventions
- Testing approach

## Phase 2: Task Generation

**NOT EXECUTED BY THIS COMMAND**

Phase 2 tasks will be generated by `/speckit.tasks` command, which will:
- Break down implementation into dependency-ordered tasks
- Map user stories to implementation tasks
- Define acceptance criteria per task
- Sequence tasks for iterative delivery

The generated `tasks.md` will provide the implementation roadmap.

## Next Steps After Planning

1. **Review this plan** - Ensure all technical context is accurate
2. **Execute Phase 0** - Research and document technology decisions (this plan command does this)
3. **Execute Phase 1** - Generate data models and contracts (this plan command does this)
4. **Run `/speckit.tasks`** - Generate actionable implementation tasks
5. **Run `/speckit.implement`** - Execute the task-by-task implementation

## Success Criteria Alignment

This plan enables achieving all 13 success criteria from the spec:

- **SC-001** (70-90% token reduction): Smart filtering algorithms + schema awareness
- **SC-002** (85%+ inference accuracy): Field inference research + validation metrics
- **SC-003** (<15ms latency): Performance-focused design + benchmarking
- **SC-004** (500+ req/sec): Concurrent request handling + efficient filtering
- **SC-005** (100% correctness): Schema validation + comprehensive testing
- **SC-006** (Health check <100ms): Simple status endpoint
- **SC-007** (Metrics real-time): Prometheus push metrics
- **SC-008** (Strategy switching): Configuration-driven + reload support
- **SC-009** (Horizontal scaling): Stateless design + no shared state
- **SC-010** (Config reload <5s): Hot reload mechanism
- **SC-011** (95% success rate): Robust error handling + fallbacks
- **SC-012** (Measurable savings): Comprehensive metrics + case studies
- **SC-013** (99.9% availability): Production-ready patterns + monitoring

---

**Status**: Ready for Phase 0 research execution
**Next Command**: Continuing with Phase 0 research generation...
