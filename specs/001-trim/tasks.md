# Tasks: MCP Response Optimizer Proxy (SlimMCP)

**Input**: Design documents from `/specs/001-trim/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Tasks include acceptance tests per user story as specified in the feature spec.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4, US5)
- Include exact file paths in descriptions

## Path Conventions
- **Monorepo structure**: All SlimMCP code in `mcp/` subdirectory of `IdeaCraftersHQ/oss-mcp-slim`
- **Development paths**: `mcp/cmd/`, `mcp/pkg/`, `mcp/internal/`, `mcp/tests/`
- **Task paths**: All file paths below are relative to `mcp/` directory
- **Deployment**: GitHub Actions auto-sync `mcp/` to `slimmcp/slimmcp` repository
- All paths use Go standard project layout within `mcp/` directory

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [X] T001 Create Go project structure per plan.md in mcp/ directory: mcp/cmd/slimmcp/, mcp/pkg/{proxy,analyzer,filter,schema,capability,client,metrics,config,types}/,  mcp/internal/testutil/, mcp/tests/{integration,contract,acceptance}/, mcp/examples/, mcp/docs/
- [X] T002 Initialize Go module in mcp/go.mod: module github.com/slimmcp/slimmcp, Go 1.25+
- [X] T003 [P] Add dependencies: github.com/goccy/go-yaml v1.18.0, github.com/prometheus/client_golang v1.23.2, github.com/mark3labs/mcp-go v0.42.0-beta.1
- [X] T004 [P] Create Makefile in mcp/ with build, test, run, lint targets
- [X] T005 [P] Configure golangci-lint with mcp/.golangci.yml
- [X] T006 [P] Create example configuration files: mcp/examples/dev.yaml, mcp/examples/prod.yaml

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core types, interfaces, and infrastructure that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T007 Define core types in pkg/types/request.go: MCPRequest, RequestParams, RequestMeta, FieldSelectionSpec structs with JSON tags
- [X] T008 [P] Define response types in pkg/types/response.go: MCPResponse, JSONRPCError, ResponseMeta, FilteringMetadata structs
- [X] T009 [P] Define error types in pkg/types/errors.go: Custom error types for filtering, validation, timeout errors with JSON-RPC codes
- [X] T010 [P] Define domain types in pkg/types/meta.go: ToolSchema, FieldInfo, ServerCapabilities, FilterContext, FieldCategory enum
- [X] T011 Create Analyzer interface in pkg/analyzer/analyzer.go: InferFields(req *types.MCPRequest) ([]string, error)
- [X] T012 [P] Create Filter interface in pkg/filter/filter.go: Filter(ctx context.Context, data interface{}, fields []string) (interface{}, error)
- [X] T013 [P] Create SchemaManager interface in pkg/schema/manager.go: GetSchema(tool, serverURL string) (*types.ToolSchema, error), LoadSchema methods
- [X] T014 [P] Create MetricsCollector interface in pkg/metrics/collector.go: RecordFiltering(stats types.FilteringStatistics), RecordRequest methods
- [X] T015 Create configuration structures in pkg/config/config.go: Config, ProxyConfig, ServerConfig, OptimizationConfig, MonitoringConfig structs from data-model.md
- [X] T016 Implement YAML configuration loader in pkg/config/loader.go: LoadConfig(path string) (*Config, error) using github.com/goccy/go-yaml
- [X] T017 [P] Implement configuration validator in pkg/config/validator.go: Validate() error method, check required fields, validate URLs
- [X] T018 [P] Create mock MCP server in internal/testutil/mock_server.go: HTTP server returning pre-configured JSON responses for testing

**Checkpoint**: ✅ Foundation ready - all core types, interfaces, and test infrastructure exist. User story implementation can now begin in parallel.

---

## Phase 3: User Story 1 - Automatic Token Reduction (Priority: P1) 🎯 MVP

**Goal**: Implement minimal filtering strategy that automatically reduces token usage without client configuration

**Independent Test**: Send a single MCP tool/call request through proxy with no _meta.fieldSelection, verify response contains only critical fields (id, type, url, title) and _meta shows 70%+ savings

### Acceptance Tests for User Story 1

**NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T019 [P] [US1] Acceptance test in tests/acceptance/user_story_1_test.go: Test "Given LLM client connected to proxy, When requests Notion page with 45 fields, Then response has only 5 essential fields with 85%+ savings"
- [X] T020 [P] [US1] Acceptance test in tests/acceptance/user_story_1_test.go: Test "Given tool call mentioning specific fields in parameters, When proxy analyzes request, Then infers needed fields and filters accordingly"
- [X] T021 [P] [US1] Acceptance test in tests/acceptance/user_story_1_test.go: Test "Given multiple consecutive requests for similar data, When proxy processes each, Then responses filtered consistently"

### Implementation for User Story 1

- [X] T022 [US1] Implement MinimalAnalyzer in pkg/analyzer/minimal.go: Returns fixed list [id, type, url, title], implements Analyzer interface
- [X] T023 [US1] Implement basic filtering engine in pkg/filter/engine.go: filterObject recursively keeps only specified fields, handles nested objects
- [X] T024 [US1] Implement token counting in pkg/filter/metadata.go: EstimateTokens(data interface{}) int using character count / 4 approximation
- [X] T025 [US1] Implement filtering statistics in pkg/filter/metadata.go: CalculateFilteringStats(original, filtered interface{}) types.FilteringMetadata
- [X] T026 [US1] Create HTTP server in pkg/proxy/server.go: http.Server with custom config (timeouts: Read 30s, Write 30s, Idle 120s), graceful shutdown
- [X] T027 [US1] Implement request handler in pkg/proxy/handler.go: Parse JSON-RPC request, validate structure, handle errors with proper JSON-RPC error codes
- [X] T028 [US1] Implement MCP client in pkg/client/mcp.go: Forward requests to downstream server, handle timeouts (30s default), return raw response
- [X] T029 [US1] Implement request routing in pkg/proxy/server.go: Route POST / and POST /{server} to appropriate handler, add health check GET /health
- [X] T030 [US1] Wire components in cmd/slimmcp/main.go: Load config, create server with minimal analyzer, start HTTP server
- [X] T031 [US1] Add logging in pkg/proxy/handler.go: Request/response logging with log package, log filtering statistics
- [X] T032 [US1] Implement health check in pkg/proxy/handler.go: GET /health returns 200 OK with status JSON including proxy version

**Checkpoint**: ✅ User Story 1 (automatic token reduction with minimal strategy) is fully functional. Binary builds and runs successfully!

---

## Phase 4: User Story 2 - Explicit Field Selection Control (Priority: P2)

**Goal**: Allow clients to specify exactly which fields they want via _meta.fieldSelection

**Independent Test**: Send request with _meta.fieldSelection.include: ["title", "status", "assignee"], verify only those fields plus critical fields are returned

### Acceptance Tests for User Story 2

- [X] T033 [P] [US2] Acceptance test in tests/acceptance/user_story_2_test.go: Test "Given request with _meta.fieldSelection.include, When tool returns data, Then response contains only specified fields plus critical fields"
- [X] T034 [P] [US2] Acceptance test in tests/acceptance/user_story_2_test.go: Test "Given request with _meta.fieldSelection.exclude, When tool returns data, Then response has all fields except excluded ones"
- [X] T035 [P] [US2] Acceptance test in tests/acceptance/user_story_2_test.go: Test "Given nested field selection properties.assignee.name, When tool returns data, Then response preserves nested structure with only specified fields"

### Implementation for User Story 2

- [X] T036 [US2] Implement ExplicitAnalyzer in pkg/analyzer/explicit.go: Extract fields from _meta.fieldSelection.include/exclude, implements Analyzer interface
- [X] T037 [US2] Field selection parsing already implemented in pkg/types/request.go: ParseFieldSelection() method on FieldSelectionSpec
- [X] T038 [US2] Implement nested field filtering in pkg/filter/recursive.go: FilterNested handles dot notation (e.g., "properties.status"), preserves parent structure
- [X] T039 [US2] Add field selection validation in pkg/filter/validator.go: ValidateFieldSelection checks include/exclude not both present, validates field names
- [X] T040 [US2] Update request handler in pkg/proxy/handler.go: Check for _meta.fieldSelection, pass to analyzer, apply explicit strategy if present
- [X] T041 [US2] Add strategy selection logic in pkg/analyzer/factory.go: NewAnalyzer(strategy, config) and NewAnalyzerFromRequest() return appropriate analyzer instance
- [X] T042 [US2] Response metadata already includes fields.kept array and fields.removed count in FilteringMetadata (from Phase 3)

**Checkpoint**: ✅ User Stories 1 AND 2 now work. Can use proxy with automatic filtering (US1) or explicit field selection (US2). Binary builds successfully!

---

## Phase 5: User Story 3 - Performance and Savings Monitoring (Priority: P2)

**Goal**: Provide visibility into token savings and filtering performance through Prometheus metrics

**Independent Test**: Process 100 requests, access /metrics endpoint, verify Prometheus-formatted metrics show total tokens saved, savings percentage, request counts

### Acceptance Tests for User Story 3

- [X] T043 [P] [US3] Acceptance test in tests/acceptance/user_story_3_test.go: Test "Given proxy processed 100 requests, When access /metrics, Then see Prometheus metrics with tokens saved, savings %, request counts by server and tool"
- [X] T044 [P] [US3] Acceptance test in tests/acceptance/user_story_3_test.go: Test "Given single filtered response, When inspect _meta.fieldSelection, Then contains originalSize, filteredSize, savings %, fields kept/removed, filtering time"
- [X] T045 [P] [US3] Acceptance test in tests/acceptance/user_story_3_test.go: Test "Given continuous operation, When query metrics over time, Then can track savings trends and detect performance issues"

### Implementation for User Story 3

- [X] T046 [P] [US3] Define Prometheus metrics in pkg/metrics/prometheus.go: tokensOriginalTotal, tokensSavedTotal, requestsTotal counters; filteringDuration histogram; savingsPercentage gauge
- [X] T047 [P] [US3] Implement metrics collector in pkg/metrics/collector.go: RecordFiltering updates counters/histograms, RecordRequest tracks totals
- [X] T048 [P] [US3] Create metrics server in pkg/metrics/prometheus.go: Separate HTTP server on configurable port (default 9090) with /metrics endpoint
- [X] T049 [US3] Integrate metrics collection in pkg/proxy/handler.go: Record stats after each filtering operation, timing with time.Since
- [X] T050 [US3] Add metrics middleware in pkg/proxy/middleware.go: Wrap handlers to automatically record request metrics (count, duration, status)
- [X] T051 [US3] Update filtering metadata in pkg/filter/metadata.go: Add filteringTime field, calculate in milliseconds
- [X] T052 [US3] Start metrics server in cmd/slimmcp/main.go: Initialize Prometheus registry, start metrics server on configured port
- [X] T053 [US3] Add metrics configuration validation in pkg/config/validator.go: Check metricsPort is valid (1024-65535), enabled flag

**Checkpoint**: All three user stories (US1, US2, US3) now work. Proxy does automatic filtering, explicit field selection, and exposes comprehensive metrics. Test each independently.

---

## Phase 6: User Story 4 - Transparent Multi-Server Proxy (Priority: P1)

**Goal**: Support routing requests to multiple downstream MCP servers through single proxy endpoint

**Independent Test**: Configure proxy with Notion, Slack, Google Drive servers, send requests to /notion, /slack, /gdrive, verify each routes correctly and returns filtered responses

### Acceptance Tests for User Story 4

- [X] T054 [P] [US4] Acceptance test in tests/acceptance/user_story_4_test.go: Test "Given proxy configured with Notion, Slack, Google Drive, When send requests to /notion, /slack, /gdrive, Then each routes to correct server and returns filtered responses"
- [X] T055 [P] [US4] Acceptance test in tests/acceptance/user_story_4_test.go: Test "Given different tools have different schemas, When proxy processes requests for each, Then applies tool-specific filtering rules correctly"
- [X] T056 [P] [US4] Acceptance test in tests/acceptance/user_story_4_test.go: Test "Given downstream server auth requirements, When client includes auth headers, Then proxy passes through unchanged and maintains session"

### Implementation for User Story 4

- [X] T057 [US4] Implement connection pool in pkg/client/pool.go: Manage http.Client instances per server, MaxIdleConns: 100, MaxIdleConnsPerHost: 10
- [X] T058 [US4] Implement server routing in pkg/proxy/router.go: Parse /{server} from URL path, lookup ServerConfig, route to correct downstream URL
- [X] T059 [US4] Implement auth header pass-through in pkg/client/mcp.go: Copy all headers from client request to downstream request (Authorization, etc.)
- [X] T060 [US4] Add per-server configuration in pkg/config/config.go: ServerConfig with URL, timeout, strategy override, aggressiveness, per-tool defaults
- [X] T061 [US4] Implement server-specific strategy override in pkg/analyzer/factory.go: Check ServerConfig for strategy override before using global strategy
- [X] T062 [US4] Add connection health check in pkg/client/mcp.go: Ping downstream servers periodically, track availability in health endpoint
- [X] T063 [US4] Update health endpoint in pkg/proxy/handler.go: Include status of all configured downstream servers with lastCheck timestamps

**Checkpoint**: Proxy now handles multiple servers. Combined with US1-US3, have fully functional multi-server proxy with filtering and metrics.

---

## Phase 7: User Story 5 - Native Field Selection Support Detection (Priority: P3)

**Goal**: Detect when downstream servers natively support field selection and delegate filtering to them

**Independent Test**: Connect to server advertising fieldSelection.supported: true, verify proxy adds _meta.fieldSelection to request and passes response through unfiltered

### Acceptance Tests for User Story 5

- [X] T064 [P] [US5] Acceptance test in tests/acceptance/user_story_5_test.go: Test "Given downstream server advertises fieldSelection support, When proxy receives request, Then adds _meta to downstream request and passes response through"
- [X] T065 [P] [US5] Acceptance test in tests/acceptance/user_story_5_test.go: Test "Given downstream server without field selection, When proxy receives request, Then fetches full response and applies filtering"
- [X] T066 [P] [US5] Acceptance test in tests/acceptance/user_story_5_test.go: Test "Given capability detection ran once, When subsequent requests arrive, Then uses cached capability info without re-querying"

### Implementation for User Story 5

- [X] T067 [P] [US5] Implement capability detector in pkg/capability/detector.go: Query server for capabilities via tools/list or dedicated endpoint, parse ServerCapabilities
- [X] T068 [P] [US5] Implement capability cache in pkg/capability/cache.go: sync.Map with TTL (1 hour), GetCapabilities/SetCapabilities methods
- [ ] T069 [US5] Integrate capability check in pkg/proxy/handler.go: Before filtering, check if server supports native field selection
- [ ] T070 [US5] Implement native field selection request in pkg/client/mcp.go: Add _meta.fieldSelection to downstream request when server supports it
- [X] T071 [US5] Implement passthrough analyzer in pkg/analyzer/passthrough.go: Returns empty field list (no filtering), used when server handles native filtering
- [X] T072 [US5] Update response metadata in pkg/filter/metadata.go: Set filteredBy: "server" when native filtering used, filteredBy: "proxy" when proxy filtered
- [X] T073 [US5] Add capability cache invalidation in pkg/capability/cache.go: TTL-based expiration, manual clear via admin endpoint

**Checkpoint**: All 5 user stories complete. Proxy has automatic filtering, explicit selection, metrics, multi-server support, and native capability detection.

---

## Phase 8: Sm art Filtering Strategy (Enhancement)

**Goal**: Implement intelligent field inference that analyzes request parameters to determine needed fields

**Independent Test**: Send request mentioning "status" in parameters, verify smart analyzer includes status field in response without explicit selection

- [ ] T074 [P] Implement SmartAnalyzer in pkg/analyzer/smart.go: Parse request parameters, extract field references using regex patterns
- [ ] T075 [P] Add heuristic rules in pkg/analyzer/smart.go: If parameter mentions field name, include that field; handle user/assignee relationships
- [ ] T076 [P] Implement parameter parsing in pkg/analyzer/smart.go: Extract JSON paths from arguments, detect dot-notation references
- [ ] T077 Add tool-specific defaults in pkg/analyzer/smart.go: Load defaults from config per tool (e.g., notion-fetch: [id, title, url, status])
- [ ] T078 Implement field relationship detection in pkg/schema/manager.go: Use schema to find parent-child relationships, auto-include parents
- [ ] T079 [P] Add regex pattern cache in pkg/analyzer/smart.go: Pre-compile common field matching patterns globally, reuse across requests
- [ ] T080 Update analyzer factory in pkg/analyzer/factory.go: Add "smart" case, create SmartAnalyzer with config and schema manager
- [ ] T081 [P] Add smart strategy tests in pkg/analyzer/smart_test.go: Table-driven tests for various parameter patterns and field inference scenarios

---

## Phase 9: Schema Management (Enhancement)

**Goal**: Load, cache, and validate tool schemas from servers or configuration

**Independent Test**: Fetch schema from MCP server via tools/list, verify it's cached for 1 hour, validate field selection against schema

- [ ] T082 [P] Implement schema loader from server in pkg/schema/loader.go: Call tools/list method, parse tool definitions, extract field info
- [ ] T083 [P] Implement schema loader from config in pkg/schema/loader.go: Load schemas from YAML configuration files
- [ ] T084 [P] Implement schema cache in pkg/schema/cache.go: sync.Map with TTL (1 hour), GetSchema/SetSchema/Invalidate methods
- [ ] T085 Implement schema precedence in pkg/schema/manager.go: Check config override first, then cache, then fetch from server, fallback to minimal schema
- [ ] T086 [P] Implement schema validator in pkg/schema/validator.go: ValidateFieldSelection checks required fields included, field names valid
- [ ] T087 [P] Add field relationship expansion in pkg/schema/manager.go: ExpandRelatedFields method adds parent fields for nested selections
- [ ] T088 Integrate schema validation in pkg/proxy/handler.go: Validate field selection against schema before filtering, return error if invalid
- [ ] T089 [P] Add schema cache invalidation in pkg/schema/cache.go: TTL expiration, version change detection, manual clear via admin endpoint

---

## Phase 10: Configuration Management (Enhancement)

**Goal**: Support hot configuration reload without restart, proper validation, and error handling

**Independent Test**: Start proxy, send SIGHUP, verify configuration reloads within 5 seconds without dropping requests

- [ ] T090 Implement configuration reload in pkg/config/reload.go: ReloadConfig method loads new config, validates, atomically swaps with sync.RWMutex
- [ ] T091 Add SIGHUP handler in cmd/slimmcp/main.go: Trap SIGHUP signal, trigger config reload, log result
- [ ] T092 [P] Create admin reload endpoint in pkg/proxy/handler.go: POST /admin/reload triggers reload, returns success/error with change summary
- [ ] T093 [P] Add configuration change detection in pkg/config/reload.go: Compare old and new config, generate list of changes for logging
- [ ] T094 Update component reconfiguration in pkg/proxy/server.go: ApplyNewConfig method updates strategy, server settings, doesn't drop connections
- [ ] T095 [P] Add reload timing measurement in pkg/config/reload.go: Track reload duration, fail if exceeds 5 seconds, log performance metrics
- [ ] T096 [P] Add configuration validation tests in pkg/config/validator_test.go: Test all validation rules, invalid configs return descriptive errors

---

## Phase 11: Error Handling & Safety (Enhancement)

**Goal**: Implement robust error handling, circuit breakers, and fallback behavior for production reliability

**Independent Test**: Trigger filtering error, verify fallback returns full response when fallbackOnError: true

- [ ] T097 Implement fallback behavior in pkg/filter/filter.go: On error, return full response if config.fallbackOnError true, log error
- [ ] T098 [P] Implement circuit breaker in pkg/client/retry.go: Track consecutive failures per server, open circuit after threshold (10), timeout (60s)
- [ ] T099 [P] Add timeout handling in pkg/client/mcp.go: Respect configured timeout (30s default), return -32003 error code on timeout
- [ ] T100 [P] Implement retry logic in pkg/client/retry.go: Exponential backoff for transient errors, max 3 retries
- [ ] T101 Add error response generation in pkg/proxy/handler.go: Convert Go errors to JSON-RPC errors with appropriate codes and messages
- [ ] T102 [P] Add max response size limit in pkg/client/mcp.go: Reject responses >10MB, prevent OOM, return error to client
- [ ] T103 Add filtering error recovery in pkg/filter/filter.go: Try-catch (defer-recover) around filtering, log panic, trigger fallback
- [ ] T104 [P] Add circuit breaker metrics in pkg/metrics/prometheus.go: Track circuit state changes, failure counts per server

---

## Phase 12: Contract & Protocol Tests (Quality Assurance)

**Goal**: Verify MCP protocol compliance and contract adherence

- [ ] T105 [P] MCP protocol compliance test in tests/contract/mcp_compliance_test.go: Verify JSON-RPC 2.0 format, method support (tools/list, tools/call, etc.)
- [ ] T106 [P] JSON-RPC error code test in tests/contract/jsonrpc_test.go: Test all standard error codes (-32700 to -32603) and custom codes (-32000 to -32003)
- [ ] T107 [P] Field selection extension test in tests/contract/extension_test.go: Test _meta.fieldSelection request/response format matches spec
- [ ] T108 [P] Batch request test in tests/contract/jsonrpc_test.go: Send batch JSON-RPC requests, verify batch response with correct ID correlation
- [ ] T109 [P] Request validation test in tests/contract/mcp_compliance_test.go: Test invalid requests return proper error responses
- [ ] T110 [P] Response structure test in tests/contract/mcp_compliance_test.go: Verify responses match MCP response schema

---

## Phase 13: Integration & Performance Tests (Quality Assurance)

**Goal**: End-to-end testing and performance validation

- [ ] T111 [P] End-to-end proxy test in tests/integration/proxy_test.go: Start proxy with mock server, send requests, verify filtering works end-to-end
- [ ] T112 [P] Multi-strategy test in tests/integration/strategies_test.go: Test all strategies (minimal, explicit, smart, passthrough) with same data
- [ ] T113 [P] Filtering accuracy test in tests/integration/filtering_test.go: Verify 85%+ field inference accuracy for smart strategy
- [ ] T114 [P] Performance benchmark in tests/integration/performance_test.go: Measure p95 latency <15ms for smart strategy, <10ms for minimal
- [ ] T115 [P] Throughput test in tests/integration/performance_test.go: Verify 500+ requests/sec per instance with concurrent clients, test horizontal scaling to 1,500+ req/sec with 3 instances
- [ ] T116 [P] Token savings test in tests/integration/filtering_test.go: Verify 70-90% token reduction on realistic Notion/Slack/Drive responses
- [ ] T117 [P] Concurrent request test in tests/integration/proxy_test.go: Send 100 concurrent requests, verify all succeed, no race conditions

---

## Phase 14: Polish & Cross-Cutting Concerns

**Purpose**: Final improvements, documentation, and production readiness

- [ ] T118 [P] Add structured logging in pkg/proxy/middleware.go: Use log/slog for JSON structured logs, include request ID, server, tool
- [ ] T119 [P] Implement graceful shutdown in cmd/slimmcp/main.go: Trap SIGTERM/SIGINT, shutdown with 30s timeout, wait for requests to complete
- [ ] T120 [P] Add request ID generation in pkg/proxy/middleware.go: Generate UUID for each request, add to context, include in all logs
- [ ] T121 [P] Create mcp/Dockerfile: Multi-stage build, Go 1.25+ base, copy binary, expose 8080 and 9090, CMD ["slim", "start"]
- [ ] T122 [P] Create docker-compose.yml in mcp/examples/: Proxy + mock servers for local development
- [ ] T123 [P] Add CORS support in pkg/proxy/middleware.go: Configurable CORS headers based on config.proxy.cors settings
- [ ] T124 [P] Add TLS support in pkg/proxy/server.go: Optional TLS termination with cert/key files from config.proxy.tls
- [ ] T125 [P] Create README usage examples in mcp/examples/: Minimal example, explicit field selection example, multi-server example
- [ ] T126 [P] Add CLI flags in mcp/cmd/slimmcp/main.go: --config, --version, --help, --debug flags with cobra or flag package
- [ ] T127 [P] Create getting-started guide in mcp/docs/getting-started.md: Installation, configuration, first request, understanding metrics
- [ ] T128 [P] Update architecture.md in mcp/docs/: Add implementation notes, actual performance numbers, deployment patterns
- [ ] T129 Run quickstart.md validation in mcp/: Verify all instructions work, update as needed based on actual implementation

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phases 3-7)**: All depend on Foundational phase completion
  - User Story 1 (Phase 3): Can start after Foundational - No dependencies on other stories
  - User Story 2 (Phase 4): Can start after Foundational - Independent of US1 but builds on filtering engine
  - User Story 3 (Phase 5): Can start after Foundational - Independent monitoring
  - User Story 4 (Phase 6): Can start after Foundational - Independent routing
  - User Story 5 (Phase 7): Can start after Foundational - Independent capability detection
- **Enhancements (Phases 8-11)**: Can start after User Story 1 complete (core filtering exists)
- **Tests (Phases 12-13)**: Can start after core user stories (3-6) complete
- **Polish (Phase 14)**: Depends on all desired features being complete

### User Story Dependencies

- **User Story 1 (P1)**: Foundation only - No other story dependencies
- **User Story 2 (P2)**: Foundation only - Extends US1 filtering but independently testable
- **User Story 3 (P2)**: Foundation only - Independent metrics collection
- **User Story 4 (P1)**: Foundation only - Independent routing logic
- **User Story 5 (P3)**: Foundation + US4 (multi-server) - Needs routing to detect capabilities per server

### Within Each User Story

- Acceptance tests FIRST (write tests, verify they fail)
- Models/types before services
- Services before handlers
- Core implementation before integration
- Story complete and tested before moving to next priority

### Parallel Opportunities

**Phase 1 (Setup)**: Tasks T003, T004, T005, T006 can run in parallel

**Phase 2 (Foundational)**: Tasks T008, T009, T010, T012, T013, T014, T017, T018 can run in parallel (different files)

**User Story Acceptance Tests**: All tests within a story marked [P] can run in parallel (T019-T021, T033-T035, T043-T045, T054-T056, T064-T066)

**Once Foundational Complete**: All user stories (Phases 3-7) can start in parallel by different developers

**Enhancement Phases**: Tasks within Phases 8-11 marked [P] can run in parallel

**Test Phases**: All contract tests (T105-T110) and integration tests (T111-T117) can run in parallel

**Polish Phase**: Most tasks in Phase 14 marked [P] can run in parallel

---

## Parallel Example: User Story 1

```bash
# Launch all acceptance tests for User Story 1 together:
Task T019: "Acceptance test: 45 fields → 5 fields with 85%+ savings"
Task T020: "Acceptance test: Field inference from parameters"
Task T021: "Acceptance test: Consistent filtering across requests"

# After tests written and failing, launch parallel implementation:
Task T022: "MinimalAnalyzer in pkg/analyzer/minimal.go"
Task T023: "Filtering engine in pkg/filter/filter.go"
Task T024: "Token counting in pkg/filter/metadata.go"

# Then sequential (same file dependencies):
Task T026: "HTTP server in pkg/proxy/server.go"
Task T027: "Request handler in pkg/proxy/handler.go"
# ... etc
```

---

## Parallel Example: Multiple User Stories

```bash
# After Foundational phase complete, launch in parallel:

Developer A (User Story 1):
  - T019-T032: Automatic token reduction

Developer B (User Story 2):
  - T033-T042: Explicit field selection

Developer C (User Story 3):
  - T043-T053: Metrics and monitoring

Developer D (User Story 4):
  - T054-T063: Multi-server routing
```

---

## Implementation Strategy

### MVP First (User Stories 1 & 4 Only)

For minimum viable product focusing on core proxy with automatic filtering:

1. Complete Phase 1: Setup (T001-T006)
2. Complete Phase 2: Foundational (T007-T018) - CRITICAL
3. Complete Phase 3: User Story 1 (T019-T032) - Automatic filtering
4. Complete Phase 6: User Story 4 (T054-T063) - Multi-server support
5. **STOP and VALIDATE**: Test US1+US4 together (multi-server with automatic filtering)
6. Deploy/demo if ready

This gives you a working proxy that handles multiple servers with automatic token reduction.

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US1 (Automatic filtering) → Test independently → MVP v0.1
3. Add US4 (Multi-server) → Test with US1 → MVP v0.2
4. Add US2 (Explicit selection) → Test independently → v0.3
5. Add US3 (Metrics) → Test independently → v0.4
6. Add US5 (Native capabilities) → Test independently → v0.5
7. Add Enhancements (Smart filtering, schemas) → v0.6+
8. Polish and production hardening → v1.0

### Parallel Team Strategy

With 3-4 developers after Foundational complete:

1. Dev 1: User Story 1 (automatic filtering)
2. Dev 2: User Story 2 (explicit selection)
3. Dev 3: User Story 3 (metrics)
4. Dev 4: User Story 4 (multi-server)

All can work in parallel, integrate independently, then combine for full feature set.

---

## Task Summary

**Total Tasks**: 129

**By Phase**:
- Phase 1 (Setup): 6 tasks
- Phase 2 (Foundational): 12 tasks ⚠️ BLOCKS all stories
- Phase 3 (US1 - Automatic Filtering): 14 tasks
- Phase 4 (US2 - Explicit Selection): 10 tasks
- Phase 5 (US3 - Metrics): 11 tasks
- Phase 6 (US4 - Multi-Server): 10 tasks
- Phase 7 (US5 - Capability Detection): 10 tasks
- Phase 8 (Smart Filtering): 8 tasks
- Phase 9 (Schema Management): 8 tasks
- Phase 10 (Config Management): 7 tasks
- Phase 11 (Error Handling): 8 tasks
- Phase 12 (Contract Tests): 6 tasks
- Phase 13 (Integration Tests): 7 tasks
- Phase 14 (Polish): 12 tasks

**Parallel Opportunities**: 68 tasks marked [P] can run in parallel

**MVP Scope**: Phases 1, 2, 3, 6 = 42 tasks (minimum viable proxy with automatic multi-server filtering)

**Independent Test Criteria**:
- US1: Single request → filtered to critical fields → 70%+ savings
- US2: Explicit field request → only requested fields returned
- US3: 100 requests → metrics endpoint shows statistics
- US4: Multi-server requests → all route correctly
- US5: Capability detection → delegates to native filtering when supported

---

## Notes

- [P] tasks = different files or independent functionality, no blocking dependencies
- [Story] label (US1-US5) maps task to specific user story for traceability
- Each user story is independently completable and testable
- Acceptance tests written FIRST, must fail before implementation begins
- Commit after each task or logical group of related tasks
- Stop at checkpoints to validate stories independently before proceeding
- Focus on US1 and US4 for fastest path to working MVP
- All Go code should follow standard project layout and Go best practices
- Use table-driven tests for all unit/integration tests
- Keep test fixtures in internal/testutil/ for reuse across test suites
