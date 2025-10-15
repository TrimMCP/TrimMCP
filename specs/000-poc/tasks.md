# Tasks: Token Savings Benchmark PoC

**Input**: Design documents from `/specs/000-poc/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Included (Test-First Development per Constitution II)

**Organization**: Tasks grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4)
- Exact file paths included in descriptions

## Path Conventions
- Base: `mcp/` (within monorepo root)
- Packages: `mcp/pkg/{token,filter,cache}/`
- Internal: `mcp/internal/benchmark/`
- CLI: `mcp/cmd/benchmark/`
- Tests: `mcp/tests/{acceptance,integration}/`
- Fixtures: `mcp/testdata/{github,notion}/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] **T001** [P] [Setup] Create directory structure: `mcp/cmd/benchmark/`, `mcp/pkg/{token,filter,cache}/`, `mcp/internal/benchmark/`, `mcp/tests/{acceptance,integration}/`, `mcp/testdata/{github,notion}/`
- [ ] **T002** [P] [Setup] Initialize Go module in `mcp/go.mod` if not exists with `go mod init github.com/slimmcp/slimmcp`
- [ ] **T003** [Setup] Install dependencies: `github.com/mark3labs/mcp-go@v0.42.0-beta.1`, `github.com/pkoukk/tiktoken-go`, `github.com/olekukonko/tablewriter` using `go get`
- [ ] **T004** [P] [Setup] Create `.gitignore` in `mcp/` to exclude build artifacts and IDE files

**Checkpoint**: Basic project structure and dependencies ready

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core interfaces and test data that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Test Data Foundation

- [ ] **T005** [P] [Foundation] Capture real GitHub MCP response for `list_pull_requests` and save to `mcp/testdata/github/list_pull_requests.json` (follows data-model.md CachedResponse format)
- [ ] **T006** [P] [Foundation] Capture real GitHub MCP response for `search_issues` and save to `mcp/testdata/github/search_issues.json`
- [ ] **T007** [P] [Foundation] Capture real GitHub MCP response for `get_repository` and save to `mcp/testdata/github/get_repository.json`
- [ ] **T008** [P] [Foundation] Capture real GitHub MCP response for `create_issue` and save to `mcp/testdata/github/create_issue.json`
- [ ] **T009** [P] [Foundation] Capture real GitHub MCP response for `list_commits` and save to `mcp/testdata/github/list_commits.json`
- [ ] **T010** [P] [Foundation] Capture real GitHub MCP response for `get_pull_request` and save to `mcp/testdata/github/get_pull_request.json`
- [ ] **T011** [P] [Foundation] Capture real Notion MCP response for `create_page` and save to `mcp/testdata/notion/create_page.json`
- [ ] **T012** [P] [Foundation] Capture real Notion MCP response for `query_database` and save to `mcp/testdata/notion/query_database.json`
- [ ] **T013** [P] [Foundation] Capture real Notion MCP response for `list_pages` and save to `mcp/testdata/notion/list_pages.json`
- [ ] **T014** [P] [Foundation] Capture real Notion MCP response for `update_page` and save to `mcp/testdata/notion/update_page.json`
- [ ] **T015** [P] [Foundation] Capture real Notion MCP response for `create_database` and save to `mcp/testdata/notion/create_database.json`
- [ ] **T016** [P] [Foundation] Capture real Notion MCP response for `search_pages` and save to `mcp/testdata/notion/search_pages.json`

### Core Interfaces

- [ ] **T017** [P] [Foundation] Define `Tokenizer` interface in `mcp/pkg/token/tokenizer.go` (methods: Count, Name)
- [ ] **T018** [P] [Foundation] Define `Filter` interface in `mcp/pkg/filter/filter.go` (methods: Apply, Name)
- [ ] **T019** [P] [Foundation] Define `CacheLoader` interface in `mcp/pkg/cache/loader.go` (methods: Load, Exists) and `CachedResponse` struct
- [ ] **T020** [P] [Foundation] Define `Validator` interface and `Schema` interface in `mcp/internal/benchmark/validator.go` (methods: Validate, RegisterSchema, CheckRequiredFields, CheckTypes)

### Core Data Structures

- [ ] **T021** [P] [Foundation] Define `BenchmarkOperation` struct in `mcp/internal/benchmark/types.go` (fields: OperationName, Server, ToolName, UnfilteredResponse, FilteredResponse, TokenMeasurements)
- [ ] **T022** [P] [Foundation] Define `TokenMeasurement` struct in `mcp/internal/benchmark/types.go` (fields: Method, UnfilteredTokens, FilteredTokens, ReductionPercent, Timestamp)
- [ ] **T023** [P] [Foundation] Define `BenchmarkConfig` and `OperationConfig` structs in `mcp/internal/benchmark/types.go`
- [ ] **T024** [P] [Foundation] Define `BenchmarkReport` struct in `mcp/internal/benchmark/types.go`

**Checkpoint**: Foundation ready - all interfaces, structs, and test data files created. User story implementation can now begin in parallel.

---

## Phase 3: User Story 1 - Quick Benchmark Execution (Priority: P1) 🎯 MVP

**Goal**: Developer runs single command and sees summary table showing token reduction across 12 operations

**Independent Test**: Run `go run mcp/cmd/benchmark/main.go` and verify formatted output with token counts and reductions appears in <10 seconds

### Tests for User Story 1 (Write FIRST, ensure they FAIL before implementation)

- [ ] **T025** [US1] Write acceptance test `TestQuickBenchmarkExecution` in `mcp/tests/acceptance/quick_benchmark_test.go`:
  - Subtest: `displays_progress_log` - verify progress messages appear for each operation
  - Subtest: `shows_summary_table_with_token_counts` - verify table headers (Operation, Unfiltered, Filtered, Reduction)
  - Subtest: `produces_consistent_results_on_multiple_runs` - verify deterministic output
  - Subtest: `completes_within_10_seconds` - verify SC-001 performance target
- [ ] **T026** [US1] Run `go test mcp/tests/acceptance/quick_benchmark_test.go` and VERIFY ALL TESTS FAIL (no implementation yet - Red phase)

### Tokenizer Implementations for User Story 1

- [ ] **T027** [P] [US1] Implement `TiktokenTokenizer` in `mcp/pkg/token/tiktoken.go`:
  - Method `Count(text string) (int, error)` using tiktoken-go with cl100k_base encoding
  - Method `Name() string` returns "tiktoken-cl100k"
  - Handle encoder initialization in constructor `NewTiktoken()`
- [ ] **T028** [P] [US1] Implement `WordCountTokenizer` in `mcp/pkg/token/simple.go`:
  - Method `Count(text string) (int, error)` using `strings.Fields()`
  - Method `Name() string` returns "word-count"
  - Constructor `NewWordCount()`
- [ ] **T029** [P] [US1] Implement `CharDivFourTokenizer` in `mcp/pkg/token/charcount.go`:
  - Method `Count(text string) (int, error)` returns `len(text) / 4`
  - Method `Name() string` returns "char-div-4"
  - Constructor `NewCharDivFour()`
- [ ] **T030** [P] [US1] Write unit tests for all tokenizers in `mcp/pkg/token/token_test.go`:
  - Table-driven tests with known inputs
  - Contract tests verifying non-negative counts, determinism, unique names

### Filter Implementation for User Story 1

- [ ] **T031** [US1] Implement `MinimalFilter` in `mcp/pkg/filter/minimal.go`:
  - Constructor `NewMinimal()` with predefined RemoveFields list (\_meta, jsonrpc, id, url, html_url, api_url, node_id, created_at, updated_at, object, parent)
  - Method `Apply(response map[string]interface{}) (map[string]interface{}, error)` - recursively remove fields
  - Method `removeFields()` helper for nested map filtering
  - Method `filterArray()` helper for array processing
  - Method `Name() string` returns "minimal"
- [ ] **T032** [P] [US1] Write unit tests for `MinimalFilter` in `mcp/pkg/filter/filter_test.go`:
  - Test field removal (verify blacklisted fields gone)
  - Test nested object handling
  - Test array handling
  - Test immutability (original not modified)

### Cache Loader Implementation for User Story 1

- [ ] **T033** [US1] Implement `JSONFileLoader` in `mcp/pkg/cache/json.go`:
  - Constructor `NewJSONFileLoader(testDataDir string)`
  - Method `Load(server, operation string) (*CachedResponse, error)` - read and parse JSON file
  - Method `Exists(server, operation string) bool` - check file existence
  - Helper `filePath(server, operation string) string` - construct path `{testDataDir}/{server}/{operation}.json`
  - Return clear error if file missing (per FR-011)
- [ ] **T034** [P] [US1] Write unit tests for `JSONFileLoader` in `mcp/pkg/cache/loader_test.go`:
  - Test successful load with real fixture
  - Test error on missing file
  - Test error on corrupted JSON
  - Test Exists() behavior

### Benchmark Runner Implementation for User Story 1

- [ ] **T035** [US1] Implement `Runner` struct in `mcp/internal/benchmark/runner.go`:
  - Fields: CacheLoader, Filter, Tokenizers ([]Tokenizer), Verbose (bool)
  - Method `Run(operations []OperationConfig) ([]Result, error)` - main execution loop
  - Method `runOperation(op OperationConfig) (Result, error)` - process single operation
  - Helper `toString(data interface{}) string` - convert to JSON string for token counting
  - Handle errors gracefully, collect results
- [ ] **T036** [US1] Implement `Report` struct and formatter in `mcp/internal/benchmark/report.go`:
  - Method `displayResults(results []Result)` using tablewriter
  - Format table with columns: Operation, Method, Unfiltered, Filtered, Reduction
  - Calculate and display average reduction per tokenization method
  - Display total duration and success/failure counts

### Data Integrity Validator for User Story 1

- [ ] **T037** [US1] Implement basic `Validator` in `mcp/internal/benchmark/validator.go`:
  - Simple structural validation: check valid JSON, non-empty response
  - Method `Validate(operation string, filtered map[string]interface{}) error`
  - For MVP: verify response is not nil, has expected top-level structure
  - (Full schema validation will be added in later story if needed)

### CLI Entry Point for User Story 1

- [ ] **T038** [US1] Implement `main.go` in `mcp/cmd/benchmark/main.go`:
  - Parse `--verbose` flag using stdlib `flag` package
  - Initialize CacheLoader with `testdata` path
  - Initialize MinimalFilter
  - Initialize all 3 Tokenizers (tiktoken, word-count, char-div-4)
  - Create Runner with initialized components
  - Define 12 OperationConfigs (6 GitHub + 6 Notion from FR-002, FR-003)
  - Call `Runner.Run()` and handle errors
  - Display results using `displayResults()`
  - Exit with code 0 on success, 1 on error
- [ ] **T039** [US1] Run `go test mcp/tests/acceptance/quick_benchmark_test.go` and VERIFY ALL TESTS PASS (Green phase)
- [ ] **T040** [US1] Run `go run mcp/cmd/benchmark/main.go` manually and verify output shows 12 operations with token reductions

**Checkpoint**: User Story 1 complete and independently testable. MVP is functional!

---

## Phase 4: User Story 2 - GitHub Tool Call Benchmarks (Priority: P2)

**Goal**: Validate filtering strategy specifically for GitHub responses - verify metadata removed while essential data preserved

**Independent Test**: Run `go run mcp/cmd/benchmark/main.go` and filter output to GitHub operations, verify PR titles/numbers/states present but URLs/IDs/timestamps absent

### Tests for User Story 2 (Write FIRST)

- [ ] **T041** [P] [US2] Write acceptance test `TestGitHubToolCallBenchmarks` in `mcp/tests/acceptance/github_tools_test.go`:
  - Subtest: `reports_token_counts_for_list_pull_requests` - verify GitHub: list_pull_requests appears in output
  - Subtest: `filtering_removes_metadata_fields` - verify reduction % > 0 for GitHub operations
  - Subtest: `essential_data_preserved` - parse filtered response, verify PR title/number/state fields exist
  - Subtest: `metadata_fields_removed` - parse filtered response, verify url/node_id/created_at fields absent
- [ ] **T042** [US2] Run `go test mcp/tests/acceptance/github_tools_test.go` and VERIFY TESTS FAIL (need enhanced validation)

### Enhanced Validation for User Story 2

- [ ] **T043** [US2] Create GitHub response schemas in `mcp/internal/benchmark/schemas.go`:
  - Define required fields for each GitHub operation (list_pull_requests: title, number, state, user.login)
  - Define forbidden fields after filtering (url, html_url, api_url, node_id, created_at, updated_at)
  - Implement schema checking logic
- [ ] **T044** [US2] Enhance `Validator` in `mcp/internal/benchmark/validator.go`:
  - Add `RegisterSchema(operation string, schema Schema)` method
  - Implement `CheckRequiredFields(data map[string]interface{}) error` for GitHub schemas
  - Implement `CheckTypes(data map[string]interface{}) error` for type validation
  - Register all 6 GitHub operation schemas in validator initialization
- [ ] **T045** [US2] Update `Runner.runOperation()` in `mcp/internal/benchmark/runner.go`:
  - Call validator after filtering
  - Return validation errors clearly
  - Update main.go to initialize validator with GitHub schemas
- [ ] **T046** [US2] Run `go test mcp/tests/acceptance/github_tools_test.go` and VERIFY ALL TESTS PASS

**Checkpoint**: User Story 2 complete. GitHub filtering validated with schema checks.

---

## Phase 5: User Story 3 - Notion Tool Call Benchmarks (Priority: P2)

**Goal**: Validate filtering strategy for Notion responses - handle deeply nested JSON, preserve page/database content

**Independent Test**: Run benchmark, filter to Notion operations, verify task titles/content present but Notion metadata (object, id, timestamps) absent

### Tests for User Story 3 (Write FIRST)

- [ ] **T047** [P] [US3] Write acceptance test `TestNotionToolCallBenchmarks` in `mcp/tests/acceptance/notion_tools_test.go`:
  - Subtest: `reports_token_counts_for_create_page` - verify Notion: create_page appears in output
  - Subtest: `filtering_removes_notion_metadata` - verify object/id/created_time/last_edited_time/parent fields removed
  - Subtest: `page_content_preserved` - verify properties and title fields remain
  - Subtest: `nested_properties_handled` - verify deeply nested Notion properties filtered correctly
- [ ] **T048** [US3] Run `go test mcp/tests/acceptance/notion_tools_test.go` and VERIFY TESTS FAIL

### Enhanced Filtering and Validation for User Story 3

- [ ] **T049** [US3] Create Notion response schemas in `mcp/internal/benchmark/schemas.go`:
  - Define required fields for each Notion operation (create_page: properties; query_database: results array)
  - Define forbidden fields (object, id, created_time, last_edited_time, parent)
  - Handle nested property structures specific to Notion
- [ ] **T050** [US3] Enhance `MinimalFilter.removeFields()` in `mcp/pkg/filter/minimal.go`:
  - Add special handling for Notion nested properties
  - Ensure recursive filtering depth handles Notion's complex nesting
  - Test with real Notion fixtures from testdata/
- [ ] **T051** [US3] Register Notion schemas in `Validator`:
  - Add all 6 Notion operation schemas
  - Implement Notion-specific field checking (handle property objects)
  - Update main.go validator initialization
- [ ] **T052** [US3] Run `go test mcp/tests/acceptance/notion_tools_test.go` and VERIFY ALL TESTS PASS

**Checkpoint**: User Story 3 complete. Notion filtering validated with deep nesting support.

---

## Phase 6: User Story 4 - Verbose Execution Mode (Priority: P3)

**Goal**: Developer can debug benchmark with `--verbose` flag showing full request/response bodies and token calculations

**Independent Test**: Run `go run mcp/cmd/benchmark/main.go --verbose` and verify detailed logs show unfiltered response, filtered response, and tokenization method details

### Tests for User Story 4 (Write FIRST)

- [ ] **T053** [P] [US4] Write acceptance test `TestVerboseExecutionMode` in `mcp/tests/acceptance/verbose_mode_test.go`:
  - Subtest: `shows_request_response_bodies_in_verbose_mode` - verify "Unfiltered response:" and "Filtered response:" appear in output
  - Subtest: `shows_token_calculation_method_in_verbose_mode` - verify "tiktoken", "word-count", "char-div-4" method names appear
  - Subtest: `shows_full_error_details_on_failure` - simulate missing cache file, verify stack trace appears in verbose mode
- [ ] **T054** [US4] Run `go test mcp/tests/acceptance/verbose_mode_test.go` and VERIFY TESTS FAIL

### Verbose Logging Implementation for User Story 4

- [ ] **T055** [US4] Add verbose logging to `Runner.runOperation()` in `mcp/internal/benchmark/runner.go`:
  - If `r.Verbose == true`, log full unfiltered response JSON
  - Log full filtered response JSON after filtering
  - Log token counts for each method with method name
  - Use `log.Printf()` with clear prefixes ("Unfiltered response:", "Filtered response:", "Token count [tiktoken-cl100k]:")
- [ ] **T056** [US4] Add verbose logging to tokenizers in `mcp/pkg/token/*.go`:
  - In `Count()` methods, log input text length if verbose mode enabled
  - Log calculated token count with method name
  - (Note: verbose flag needs to be passed or use global logger)
- [ ] **T057** [US4] Add verbose error logging to `JSONFileLoader` in `mcp/pkg/cache/json.go`:
  - On error, log full file path attempted
  - Log JSON parse errors with snippet of corrupt data
- [ ] **T058** [US4] Update `main.go` in `mcp/cmd/benchmark/main.go`:
  - Configure `log` package based on `--verbose` flag
  - Pass verbose flag to Runner
  - Ensure errors display stack traces if verbose enabled
- [ ] **T059** [US4] Run `go test mcp/tests/acceptance/verbose_mode_test.go` and VERIFY ALL TESTS PASS

**Checkpoint**: User Story 4 complete. Verbose mode enables full debugging visibility.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final integration, documentation, and validation

- [ ] **T060** [P] [Polish] Run full test suite `go test ./...` from `mcp/` and verify all tests pass
- [ ] **T061** [P] [Polish] Run benchmark and verify SC-001 (completes in <10s): `time go run mcp/cmd/benchmark/main.go`
- [ ] **T062** [P] [Polish] Verify SC-002 (≥50% average token reduction): Check output average reduction percentages
- [ ] **T063** [P] [Polish] Verify SC-003 (100% data integrity): Confirm no validation errors in benchmark output
- [ ] **T064** [P] [Polish] Verify SC-004 (deterministic results): Run twice and diff outputs: `diff <(go run mcp/cmd/benchmark/main.go) <(go run mcp/cmd/benchmark/main.go)`
- [ ] **T065** [P] [Polish] Verify SC-005 (verbose mode useful): Review `--verbose` output for debugging detail
- [ ] **T066** [P] [Polish] Verify SC-006 (12 operations covered): Count operations in output: `go run mcp/cmd/benchmark/main.go | grep -c "GitHub:\|Notion:"`
- [ ] **T067** [P] [Polish] Add README.md to `mcp/cmd/benchmark/` with usage instructions
- [ ] **T068** [P] [Polish] Add package documentation (doc.go) to `mcp/pkg/token/`, `mcp/pkg/filter/`, `mcp/pkg/cache/`
- [ ] **T069** [Polish] Code cleanup: run `gofmt -w mcp/` and `golangci-lint run mcp/...`
- [ ] **T070** [Polish] Create Makefile in `mcp/` with targets: `make benchmark`, `make test`, `make lint`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational completion
  - Can proceed in parallel if staffed
  - Or sequentially: US1 (MVP) → US2 → US3 → US4
- **Polish (Phase 7)**: Depends on all user stories complete

### User Story Dependencies

- **US1 (P1 - MVP)**: Foundational only - No other story dependencies
- **US2 (P2 - GitHub)**: Foundational + uses US1 components (Runner, Filter) - Extends validation
- **US3 (P2 - Notion)**: Foundational + uses US1 components (Runner, Filter) - Extends validation and filtering
- **US4 (P3 - Verbose)**: Foundational + extends US1 Runner with logging - No new components

### Within Each User Story (TDD Order)

1. **Tests first** (write, verify FAIL)
2. **Interfaces/Models** (if needed for story)
3. **Implementation** (make tests PASS)
4. **Integration** (wire into main.go)
5. **Validation** (run tests, verify GREEN)

### Parallel Opportunities

**Setup Phase**:
- T001, T002, T003, T004 can all run in parallel

**Foundational Phase**:
- T005-T016 (all cache file captures) can run in parallel
- T017-T020 (all interface definitions) can run in parallel
- T021-T024 (all struct definitions) can run in parallel

**User Story 1**:
- T027, T028, T029 (tokenizer implementations) can run in parallel
- T030 and T032 (unit tests) can run after their implementations in parallel

**User Story 2**:
- T041 and T043 can run in parallel (test + schema definition)

**User Story 3**:
- T047 and T049 can run in parallel (test + schema definition)

**User Story 4**:
- All verbose logging additions (T055-T058) can be done in parallel by different developers

**Polish Phase**:
- T060-T069 (all verification and documentation tasks) can run in parallel

---

## Parallel Example: User Story 1

```bash
# After Foundational phase complete, launch US1 in parallel:

# Developer 1: Tokenizers
Task T027: Implement TiktokenTokenizer in mcp/pkg/token/tiktoken.go
Task T028: Implement WordCountTokenizer in mcp/pkg/token/simple.go
Task T029: Implement CharDivFourTokenizer in mcp/pkg/token/charcount.go

# Developer 2: Filter
Task T031: Implement MinimalFilter in mcp/pkg/filter/minimal.go

# Developer 3: Cache Loader
Task T033: Implement JSONFileLoader in mcp/pkg/cache/json.go

# Then all developers write unit tests in parallel:
Task T030: Token unit tests
Task T032: Filter unit tests
Task T034: Cache loader unit tests

# Finally, one developer integrates:
Task T035: Implement Runner
Task T036: Implement Report formatter
Task T038: Wire up main.go
```

---

## Implementation Strategy

### MVP First (User Story 1 Only - Fastest Path to Value)

1. ✅ Complete **Phase 1**: Setup (T001-T004) - ~30 minutes
2. ✅ Complete **Phase 2**: Foundational (T005-T024) - ~2-4 hours (cache capture is longest)
3. ✅ Complete **Phase 3**: User Story 1 (T025-T040) - ~4-6 hours
4. **STOP and VALIDATE**: Run acceptance tests, verify SC-001 to SC-006
5. **Demo MVP**: Show working benchmark with 12 operations, token reductions

**Total MVP Time Estimate**: 1-2 days for single developer

### Incremental Delivery (Add Stories Progressively)

1. MVP (US1) → Validate → Demo
2. Add US2 (GitHub validation) → Validate → Demo enhanced GitHub reporting
3. Add US3 (Notion validation) → Validate → Demo full coverage
4. Add US4 (Verbose mode) → Validate → Demo debugging capability
5. Polish → Final validation → Production-ready

### Parallel Team Strategy (3 Developers)

**Day 1**:
- All: Phase 1 + Phase 2 together (setup and foundation)

**Day 2** (after Foundational complete):
- Dev A: US1 (MVP) - T025-T040
- Dev B: US2 (GitHub validation) - T041-T046
- Dev C: US3 (Notion validation) - T047-T052

**Day 3**:
- Dev A: US4 (Verbose mode) - T053-T059
- Dev B + Dev C: Polish (T060-T070) in parallel

**Total Time with 3 Devs**: 2-3 days

---

## Notes

- **[P]** = Parallel-safe (different files, no dependencies)
- **[Story]** = US1/US2/US3/US4 for traceability
- **TDD**: Tests written FIRST, verified to FAIL, then implemented
- **Checkpoints**: Stop after each user story phase to validate independently
- **Constitution Compliance**: Follows Principle II (Test-First Development)
- **MVP Strategy**: US1 alone provides complete, demonstrable value
- **Incremental**: Each story adds capability without breaking previous stories
- **Error Handling**: FR-011 enforced - fail fast with clear messages on missing cache files

## Success Metrics

**Total Tasks**: 70
- Setup: 4 tasks
- Foundational: 20 tasks
- US1 (MVP): 16 tasks
- US2 (GitHub): 6 tasks
- US3 (Notion): 6 tasks
- US4 (Verbose): 7 tasks
- Polish: 11 tasks

**Parallel Opportunities**: 45+ tasks can run in parallel with proper staffing

**MVP Scope**: 40 tasks (Setup + Foundational + US1) = Minimum viable benchmark

**Test Coverage**: 4 acceptance test files (one per user story) with 15+ subtests total
