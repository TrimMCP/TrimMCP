# Feature Specification: Token Savings Benchmark PoC

**Feature Branch**: `000-poc`
**Created**: 2025-10-14
**Status**: Draft
**Input**: User description: "000-poc a minimal PoC proxying GitHub and Notion MCPs, objective being to get accurate token savings benchmarks on the most common calls made to each by running a single command that would build/start/run (or whatever it takes and verbose)"

## Clarifications

### Session 2025-10-14

- Q: When the benchmark encounters missing or corrupted cached responses, what should happen? → A: Fail immediately with error message listing which cached files are missing
- Q: Which specific tool operations should the benchmark test? → A: GitHub: list_pull_requests, search_issues, get_repository, create_issue, list_commits, get_pull_request; Notion: create_page, query_database, list_pages, update_page, create_database, search_pages
- Q: Which tokenization method should be used for counting tokens? → A: Multiple methods reported side-by-side for comparison
- Q: How should cached responses be stored and organized? → A: Separate JSON files per operation in a testdata directory (e.g., testdata/github/list_pull_requests.json)
- Q: How should data integrity validation work to verify filtered responses preserve essential information? → A: Validate filtered response has required structural elements (valid JSON, required fields present, correct types)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Quick Benchmark Execution (Priority: P1)

A developer wants to validate the token reduction claims of SlimMCP by running a single command that produces concrete before/after metrics for real-world MCP operations.

**Why this priority**: This is the core value proposition - proving the 60-85% token reduction target with real data. Without this, the entire project lacks empirical validation.

**Independent Test**: Can be fully tested by running one command (e.g., `go run cmd/benchmark/main.go`) and receiving a formatted report showing token counts for unfiltered vs. filtered responses.

**Acceptance Scenarios**:

1. **Given** the developer has installed Go and has GitHub/Notion MCP servers configured, **When** they run the benchmark command, **Then** they see a progress log showing each operation being tested
2. **Given** the benchmark completes successfully, **When** the developer reviews the output, **Then** they see a summary table showing token counts (before/after) and percentage reduction for each operation
3. **Given** the benchmark uses cached responses, **When** the developer runs it multiple times, **Then** they get consistent results without hitting external APIs

---

### User Story 2 - GitHub Tool Call Benchmarks (Priority: P2)

A developer wants to see token savings specifically for GitHub MCP tool calls (e.g., listing open PRs, creating issues) to understand optimization impact on write operations.

**Why this priority**: Tool calls represent interactive operations where token efficiency directly impacts response time and cost. This validates the proxy works for both read and write MCP operations.

**Independent Test**: Can be tested by filtering benchmark results to only GitHub tool operations and verifying the response includes tool call metadata and results.

**Acceptance Scenarios**:

1. **Given** the benchmark runs GitHub tool calls using cached responses, **When** it processes "list open pull requests", **Then** it reports token counts for both the tool call request and the response
2. **Given** the filtering strategy is applied, **When** comparing unfiltered vs. filtered tool responses, **Then** metadata fields like API URLs, timestamps, and redundant nested objects are removed
3. **Given** tool call results are filtered, **When** the developer reviews the output, **Then** essential data (PR titles, numbers, states) remains intact while verbose fields are removed

---

### User Story 3 - Notion Tool Call Benchmarks (Priority: P2)

A developer wants to see token savings for Notion MCP tool calls (e.g., creating tasks, querying databases) to validate optimization for structured data operations.

**Why this priority**: Notion responses often contain deeply nested JSON with extensive metadata, making it an ideal test case for aggressive filtering. This validates the proxy handles complex structured data.

**Independent Test**: Can be tested by filtering benchmark results to only Notion tool operations and verifying database/page structures are preserved while metadata is reduced.

**Acceptance Scenarios**:

1. **Given** the benchmark runs Notion tool calls using cached responses, **When** it processes "create task" or "query database", **Then** it reports token counts for the operation
2. **Given** Notion responses contain nested properties, **When** filtering is applied, **Then** Notion-specific metadata (object types, parent references, creation timestamps) are selectively removed
3. **Given** the filtered response, **When** validating data integrity, **Then** task titles, content, and essential properties remain accessible

---

### User Story 4 - Verbose Execution Mode (Priority: P3)

A developer debugging the benchmark wants detailed logs showing exactly what's being proxied, cached, and filtered at each step.

**Why this priority**: Essential for troubleshooting and understanding the optimization process, but not required for basic validation.

**Independent Test**: Can be tested by adding a `--verbose` flag and verifying detailed logs appear showing request/response bodies, filtering decisions, and token calculations.

**Acceptance Scenarios**:

1. **Given** the developer runs the benchmark with verbose mode enabled, **When** each operation executes, **Then** they see logs showing the original MCP request, the cached response, and the filtered result
2. **Given** verbose output is enabled, **When** token counting occurs, **Then** logs show the exact strings being counted and the calculation method
3. **Given** an operation fails, **When** verbose mode is active, **Then** full error details including stack traces are displayed

---

### Edge Cases

- System MUST fail immediately with clear error message listing missing cached response files when required cache data is unavailable or corrupted
- How does the system handle MCP servers that return empty responses?
- Filtered responses MUST pass structural validation (valid JSON, required fields present, correct types) to ensure filtering doesn't break response structure
- How are token counts calculated for non-text content (images, files)?
- What happens if GitHub/Notion API schemas change and cached responses become outdated?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST proxy requests to GitHub MCP server and Notion MCP server using cached/recorded responses
- **FR-002**: System MUST execute these GitHub tool calls: list_pull_requests, search_issues, get_repository, create_issue, list_commits, get_pull_request
- **FR-003**: System MUST execute these Notion tool calls: create_page, query_database, list_pages, update_page, create_database, search_pages
- **FR-004**: System MUST apply a filtering strategy that removes metadata, redundant fields, and verbose formatting from MCP responses
- **FR-005**: System MUST calculate token counts for both unfiltered (original) and filtered responses using multiple tokenization methods reported side-by-side
- **FR-006**: System MUST display a summary report showing operation name, unfiltered tokens, filtered tokens, and percentage reduction
- **FR-007**: System MUST be executable via a single command (e.g., `make benchmark` or `go run cmd/benchmark/main.go`)
- **FR-008**: System MUST provide verbose logging mode that shows request/response bodies and filtering decisions
- **FR-009**: System MUST use cached/recorded responses stored as separate JSON files per operation in a testdata directory to avoid requiring live API credentials during benchmark execution
- **FR-010**: System MUST validate that filtered responses maintain data integrity by checking required structural elements (valid JSON, required fields present, correct types)
- **FR-011**: System MUST fail with a clear error message listing missing cached files when required cached responses are unavailable or corrupted

### Key Entities

- **Benchmark Operation**: Represents a single MCP tool call being tested, including the operation name (e.g., "GitHub: list open PRs"), the unfiltered response, the filtered response, and token counts for each
- **Token Measurement**: Contains token counts for a given text payload using multiple tokenization methods (for side-by-side comparison), and timestamp of measurement
- **Cached Response**: A recorded MCP server response stored as a separate JSON file per operation in a testdata directory (organized by MCP server, e.g., testdata/github/list_pull_requests.json), including the request parameters, the full response body, and metadata about when it was captured
- **Filter Strategy**: Defines the rules for removing data from MCP responses, including which fields to strip, which to preserve, and transformation rules

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Developer can run a single command and receive benchmark results within 10 seconds
- **SC-002**: Benchmark report shows at least 50% average token reduction across all tested operations
- **SC-003**: 100% of filtered responses pass data integrity validation (essential fields preserved)
- **SC-004**: Benchmark can be re-run multiple times with identical results (deterministic using cached responses)
- **SC-005**: Verbose mode provides enough detail for a developer to understand filtering decisions without reading source code
- **SC-006**: Benchmark covers all 12 specified tool call operations (6 GitHub + 6 Notion)
