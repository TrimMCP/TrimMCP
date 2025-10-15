# Feature Specification: MCP Response Optimizer Proxy

**Feature Branch**: `001-trim`
**Created**: 2025-10-13
**Status**: Draft
**Input**: User description: "@docs/architecture.md @docs/api.md"

## Clarifications

### Session 2025-10-13

- Q: What availability and reliability target should the proxy meet? → A: Standard availability (99.9% / three nines) - production-grade with reasonable downtime allowance
- Q: What are the TLS/HTTPS requirements for client connections to the proxy? → A: HTTPS optional but recommended - flexible deployment, TLS can be handled by load balancer
- Q: When tool schemas are available from both server capabilities and configuration, which takes precedence? → A: Server capabilities first, configuration override - trust server but allow operator control

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Automatic Token Reduction for MCP Responses (Priority: P1)

As an LLM application developer, I need my MCP tool responses to automatically exclude unnecessary fields so that I can reduce context window usage and API costs without manually filtering each response.

**Why this priority**: This is the core value proposition - automatic optimization without code changes. Every MCP request benefits from this immediately.

**Independent Test**: Can be fully tested by sending a single MCP tool call through the proxy and verifying the response is smaller than the original while maintaining correctness. Delivers immediate cost savings.

**Acceptance Scenarios**:

1. **Given** an LLM client connected to SlimMCP proxy, **When** it requests a Notion page with 45 fields, **Then** the response contains only the 5 essential fields (id, type, url, title, status) and metadata shows 85%+ token savings
2. **Given** a tool call that mentions specific fields in parameters, **When** the proxy analyzes the request, **Then** it infers which fields are needed and filters the response accordingly
3. **Given** multiple consecutive requests for similar data, **When** the proxy processes each request, **Then** each response is filtered consistently without manual configuration

---

### User Story 2 - Explicit Field Selection Control (Priority: P2)

As an LLM application developer with specific data requirements, I need to explicitly specify which fields to include or exclude in responses so that I have precise control over context usage for critical workflows.

**Why this priority**: Provides fine-grained control for advanced users who know exactly what they need. Enables optimization of specialized workflows.

**Independent Test**: Can be tested by sending requests with `_meta.fieldSelection.include` array and verifying only those fields (plus critical fields) are returned.

**Acceptance Scenarios**:

1. **Given** a request with `_meta.fieldSelection.include: ["title", "status", "assignee"]`, **When** the tool returns data, **Then** the response contains only id, type, url, title, status, and assignee fields
2. **Given** a request with `_meta.fieldSelection.exclude: ["icon", "cover", "created_by"]`, **When** the tool returns data, **Then** the response contains all fields except the excluded ones
3. **Given** nested field selection like `properties.assignee.name`, **When** the tool returns data, **Then** the response preserves the nested structure but only includes specified nested fields

---

### User Story 3 - Performance and Savings Monitoring (Priority: P2)

As a platform administrator, I need visibility into token savings and filtering performance so that I can measure ROI and optimize proxy configuration for my workloads.

**Why this priority**: Essential for proving value and identifying optimization opportunities. Required for production deployments.

**Independent Test**: Can be tested by making several requests and verifying metrics endpoint returns accurate token counts, savings percentages, and timing data.

**Acceptance Scenarios**:

1. **Given** the proxy has processed 100 requests, **When** I access the `/metrics` endpoint, **Then** I see Prometheus-formatted metrics including total tokens saved, savings percentage, and request counts by server and tool
2. **Given** a single filtered response, **When** I inspect the `_meta.fieldSelection` object, **Then** it contains originalSize, filteredSize, savings percentage, fields kept/removed count, and filtering time
3. **Given** continuous operation, **When** I query metrics over time, **Then** I can track savings trends, identify highest-value optimizations, and detect performance issues

---

### User Story 4 - Transparent Proxy for Multiple MCP Servers (Priority: P1)

As a developer using multiple MCP servers (Notion, Slack, Google Drive), I need a single proxy endpoint that handles all my servers so that I can apply optimization uniformly without changing my application code for each integration.

**Why this priority**: Enables centralized optimization across all integrations. Critical for multi-server deployments.

**Independent Test**: Can be tested by configuring multiple downstream servers and routing requests to each through the proxy, verifying responses are correctly filtered for each server type.

**Acceptance Scenarios**:

1. **Given** proxy configured with Notion, Slack, and Google Drive servers, **When** client sends requests to `/notion`, `/slack`, `/gdrive`, **Then** each request routes to the correct downstream server and returns filtered responses
2. **Given** different tools have different schemas, **When** the proxy processes requests for each tool, **Then** it applies tool-specific field inference and filtering rules correctly
3. **Given** downstream server authentication requirements, **When** client includes auth headers, **Then** the proxy passes them through unchanged and maintains session state

---

### User Story 5 - Native Field Selection Support Detection (Priority: P3)

As a system operator, I want the proxy to detect when downstream MCP servers natively support field selection so that filtering happens at the source when possible, reducing proxy processing and network transfer.

**Why this priority**: Performance optimization for advanced scenarios. Not required for basic functionality but improves efficiency.

**Independent Test**: Can be tested by connecting to a server that advertises field selection support and verifying the proxy passes `_meta.fieldSelection` to the server instead of doing post-processing.

**Acceptance Scenarios**:

1. **Given** a downstream server that advertises `fieldSelection.supported: true`, **When** the proxy receives a request, **Then** it adds `_meta.fieldSelection` to the downstream request and passes the response through without filtering
2. **Given** a downstream server without field selection support, **When** the proxy receives a request, **Then** it fetches the full response and applies filtering in the proxy layer
3. **Given** capability detection has run once, **When** subsequent requests arrive, **Then** the proxy uses cached capability information without re-querying

---

### Edge Cases

- What happens when a requested field doesn't exist in the response schema?
  - System logs a warning and continues filtering other fields
  - Response metadata indicates which requested fields were not found

- How does the system handle malformed `_meta.fieldSelection` objects?
  - System validates the structure and returns a JSON-RPC error (-32602 Invalid params) with details
  - Falls back to smart strategy if validation fails and `fallbackOnError: true`

- What happens when filtering fails due to an internal error?
  - Default behavior: return full unfiltered response with error logged
  - Configurable: can be set to return error response instead (fail-fast mode)
  - Response metadata indicates fallback was triggered

- How does the system handle deeply nested structures (>10 levels)?
  - Honors `_meta.fieldSelection.maxDepth` setting (default: 10)
  - Stops filtering at max depth and preserves remaining structure
  - Logs warning if depth limit reached

- What happens when downstream MCP server times out?
  - Proxy respects configured timeout (default 30s)
  - Returns JSON-RPC error (-32003 Timeout) to client
  - Metrics track timeout events per server

- How does the system handle batch requests?
  - Accepts JSON-RPC batch request format (array of requests)
  - Processes each request independently with filtering
  - Returns batch response array maintaining request ID correlation

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST accept MCP protocol requests in JSON-RPC 2.0 format via HTTP POST endpoints
- **FR-002**: System MUST support routing requests to multiple configured downstream MCP servers
- **FR-003**: System MUST analyze request parameters to infer which response fields are likely needed (smart strategy)
- **FR-004**: System MUST filter responses to include only selected fields while preserving response structure and critical fields (id, type, url)
- **FR-005**: System MUST support explicit field selection via `_meta.fieldSelection.include` and `_meta.fieldSelection.exclude` arrays
- **FR-006**: System MUST support multiple filtering strategies: explicit, smart, minimal, and passthrough
- **FR-007**: System MUST calculate and report token savings by comparing original vs filtered response sizes
- **FR-008**: System MUST include filtering metadata in responses under `_meta.fieldSelection` object
- **FR-009**: System MUST detect downstream server field selection capabilities and use native filtering when available
- **FR-010**: System MUST cache server capability information to avoid repeated capability checks
- **FR-011**: System MUST validate field selection requests against tool schemas when available, preferring server-advertised schemas with configuration-based overrides
- **FR-012**: System MUST preserve all authentication headers and pass them through to downstream servers unchanged
- **FR-013**: System MUST expose health check endpoint showing proxy status and downstream server connectivity
- **FR-014**: System MUST expose Prometheus-compatible metrics endpoint with token savings, request counts, and performance data
- **FR-015**: System MUST support nested field selection using dot notation (e.g., `properties.assignee.name`)
- **FR-016**: System MUST handle JSON-RPC batch requests (array of requests)
- **FR-017**: System MUST return appropriate JSON-RPC error codes for various failure scenarios
- **FR-018**: System MUST support fallback behavior when filtering fails (configurable: return full response or error)
- **FR-019**: System MUST respect downstream server rate limits and propagate rate limit headers
- **FR-020**: System MUST support gzip compression for requests and responses
- **FR-021**: System MUST support configurable CORS policies
- **FR-026**: System MUST support both HTTP and HTTPS for client connections, with HTTPS recommended for production deployments
- **FR-022**: System MUST maintain request/response correlation using JSON-RPC id field
- **FR-023**: System MUST support optional request/response logging with PII filtering
- **FR-024**: System MUST support configuration reload without restart
- **FR-025**: System MUST handle tool-specific default field selections based on tool schemas

### Key Entities

- **MCP Request**: Represents an incoming JSON-RPC 2.0 request with method, params, id, and optional _meta.fieldSelection extension
- **MCP Response**: Represents the response from downstream server or proxy with result/error and optional _meta.fieldSelection statistics
- **Field Selection Spec**: Defines which fields to include/exclude, filtering strategy, and options like keepRelated and maxDepth
- **Tool Schema**: Describes available tools, their fields, required fields, field types, and relationships between fields
- **Server Capabilities**: Contains information about downstream server's native field selection support and protocol version
- **Filter Context**: Combines request, inferred fields, tool schema, and strategy to guide filtering decisions
- **Filtering Statistics**: Tracks original size, filtered size, savings percentage, fields kept/removed, and filtering time

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users deploying the proxy achieve 70-90% reduction in MCP response token usage for typical workflows (Notion, Slack, Google Drive integrations)
- **SC-002**: Smart strategy correctly infers needed fields with 85%+ accuracy (measured by comparing inferred fields to fields actually used in subsequent operations)
- **SC-003**: Proxy adds less than 10ms latency (p95) for minimal strategy and less than 15ms (p95) for smart strategy
- **SC-004**: Single proxy instance handles 500+ requests per second with smart strategy on standard cloud compute (2 vCPU, 4GB RAM)
- **SC-005**: Filtered responses maintain 100% correctness - all required fields for successful downstream operations are preserved
- **SC-006**: Health check endpoint responds within 100ms and accurately reflects downstream server connectivity status
- **SC-007**: Metrics endpoint updates in real-time (within 1 second) of request completion
- **SC-008**: Users can switch between filtering strategies without application code changes
- **SC-009**: Proxy operates statelessly, enabling horizontal scaling to 3+ instances with load balancing
- **SC-010**: Configuration reloads complete within 5 seconds without dropping active requests
- **SC-011**: 95% of requests complete successfully on first attempt (excluding downstream server errors)
- **SC-012**: Token cost reduction translates to measurable API cost savings for users (documented in case studies)
- **SC-013**: Proxy achieves 99.9% availability (three nines uptime) - allowing approximately 43 minutes of downtime per month

## Assumptions

- MCP protocol follows JSON-RPC 2.0 specification as documented
- Downstream MCP servers return well-formed JSON responses matching their advertised schemas
- Token counting uses approximate character-based estimation (response size in bytes / 4)
- Critical fields (id, type, url, title) are universally important and always preserved
- Most workflows benefit from field reduction without breaking functionality
- Developers can identify needed fields when explicit strategy is required
- Standard web performance expectations apply (sub-second response times)
- Horizontal scaling uses stateless architecture with external load balancer
- Authentication is handled by downstream servers, not the proxy
- Prometheus is the standard metrics collection system for monitoring
- Field relationships in schemas are defined explicitly or inferred from naming patterns
- Server-advertised schemas are authoritative unless explicitly overridden by configuration
- Default timeout of 30 seconds is sufficient for most MCP operations
- Gzip compression is supported by clients and provides meaningful bandwidth savings
- TLS termination can be handled by external load balancer or API gateway for production deployments
- HTTP mode is acceptable for local development and trusted internal networks

## Dependencies

- Downstream MCP servers must be accessible via HTTP/HTTPS
- Tool schemas should be available from server capabilities (preferred) or configuration files for optimal filtering
- Monitoring infrastructure should support Prometheus metrics format
- Load balancer required for horizontal scaling scenarios
- Configuration file or environment variables needed for proxy setup

## Constraints

- Proxy cannot filter fields that are required by downstream consumers (but should detect this from schemas)
- Filtering accuracy limited by quality of field inference and schema information
- Cannot optimize protocols other than MCP over HTTP
- Adds processing latency (though minimal) compared to direct server connection
- Memory usage scales with response size during filtering operations
- Learning mode (if implemented) requires persistent storage for usage history

## Out of Scope

- Authentication and authorization - proxy passes through credentials
- Protocol translation (only MCP-to-MCP supported)
- Response caching (separate concern, can be layered)
- Request transformation beyond field selection
- Complex query optimization or query rewriting
- Custom business logic or data transformation
- Multi-tenant isolation (handled at infrastructure level)
- Backup and disaster recovery (infrastructure concern)
