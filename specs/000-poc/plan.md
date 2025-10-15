# Implementation Plan: Token Savings Benchmark PoC

**Branch**: `000-poc` | **Date**: 2025-10-14 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/000-poc/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Build a minimal benchmark tool that validates SlimMCP's token reduction claims by proxying GitHub and Notion MCP tool calls using cached responses. The benchmark will measure token counts using multiple tokenization methods, apply filtering strategies, and report side-by-side comparisons showing percentage reduction. Executable via single command with verbose logging support for debugging.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**:
- `github.com/mark3labs/mcp-go` (v0.42.0-beta.1) - MCP Go SDK for protocol handling
- `github.com/pkoukk/tiktoken-go` - GPT tokenization (cl100k_base encoding)
- NEEDS CLARIFICATION: Additional tokenization libraries for side-by-side comparison
- `github.com/olekukonko/tablewriter` - Terminal table formatting for reports

**Storage**: JSON files in testdata/ directory (no database)
**Testing**: Standard Go testing (`go test`), table-driven tests
**Target Platform**: macOS/Linux command-line tool
**Project Type**: Single CLI project
**Performance Goals**: Complete 12 operations in <10 seconds
**Constraints**:
- Deterministic results (no live API calls)
- No external credentials required
- Single-command execution

**Scale/Scope**:
- 12 MCP tool operations (6 GitHub + 6 Notion)
- Multiple tokenization methods per operation
- Cached responses stored as separate JSON files per operation

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Simplicity First ✅
- **Compliant**: Using standard library JSON parsing, simple file I/O
- **Compliant**: CLI tool with single entry point (no HTTP server needed for PoC)
- **Compliant**: Minimal dependencies (MCP SDK, tiktoken, table formatter only)

### Principle II: Test-First Development ✅
- **Compliant**: Acceptance tests will be written for each user story before implementation
- **Compliant**: Table-driven tests for token counting with known inputs/outputs
- **Compliant**: Integration tests for file loading and filtering pipeline

### Principle III: Production-Ready Observability ⚠️
- **Deviation justified**: This is a PoC/benchmark tool, not production service
- **Mitigation**: Verbose logging mode provides debugging capability (FR-008)
- **Future**: Production proxy will have full Prometheus metrics (per constitution)

### Principle IV: Stateless Architecture ✅
- **Compliant**: No persistent state, all data from testdata/ files
- **Compliant**: Each benchmark run is independent and deterministic

### Principle V: Fail-Safe Defaults ✅
- **Compliant**: Fail immediately with clear error on missing cache files (FR-011)
- **Compliant**: Data integrity validation ensures filtering doesn't break structure (FR-010)

### Principle VI: Interface-Driven Design ✅
- **Compliant**: Core logic in packages: `filter`, `token`, `cache`, `benchmark`
- **Compliant**: Interfaces for Tokenizer, Filter, CacheLoader enable testing

### Principle VII: Zero Breaking Changes ✅
- **Compliant**: PoC doesn't modify MCP protocol, only analyzes responses
- **Future**: Production proxy will maintain protocol compatibility

**Gate Status**: PASS (observability deviation justified for PoC scope)

## Project Structure

### Documentation (this feature)

```
specs/000-poc/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
mcp/                              # Main SlimMCP project (deployed to slimmcp/slimmcp)
├── cmd/
│   └── benchmark/                # PoC benchmark CLI (temporary, for validation)
│       └── main.go
├── pkg/                          # Reusable packages (will be used by main proxy)
│   ├── filter/                   # Response filtering logic
│   │   ├── filter.go             # Filter interface
│   │   ├── minimal.go            # Minimal filtering strategy
│   │   └── filter_test.go
│   ├── token/                    # Token counting
│   │   ├── tokenizer.go          # Tokenizer interface
│   │   ├── tiktoken.go           # GPT tokenization
│   │   ├── simple.go             # Simple word-count tokenization
│   │   ├── charcount.go          # Character/4 estimation
│   │   └── token_test.go
│   └── cache/                    # Cache loading
│       ├── loader.go             # CacheLoader interface
│       ├── json.go               # JSON file loader
│       └── loader_test.go
├── internal/                     # Private packages
│   └── benchmark/                # Benchmark orchestration
│       ├── runner.go             # Benchmark execution
│       ├── report.go             # Results formatting
│       └── validator.go          # Data integrity validation
├── testdata/                     # Cached MCP responses
│   ├── github/
│   │   ├── list_pull_requests.json
│   │   ├── search_issues.json
│   │   ├── get_repository.json
│   │   ├── create_issue.json
│   │   ├── list_commits.json
│   │   └── get_pull_request.json
│   └── notion/
│       ├── create_page.json
│       ├── query_database.json
│       ├── list_pages.json
│       ├── update_page.json
│       ├── create_database.json
│       └── search_pages.json
└── tests/
    ├── acceptance/               # User story acceptance tests
    │   ├── quick_benchmark_test.go
    │   ├── github_tools_test.go
    │   ├── notion_tools_test.go
    │   └── verbose_mode_test.go
    └── integration/              # End-to-end integration tests
        └── benchmark_integration_test.go
```

**Structure Decision**: Single CLI project under `mcp/cmd/benchmark/` using reusable packages from `mcp/pkg/`. This structure allows the PoC to validate filtering and token counting logic that will later be integrated into the main SlimMCP proxy. The benchmark CLI is temporary (for validation purposes) while the packages are permanent and will be reused.

## Complexity Tracking

No violations requiring justification. PoC maintains simplicity principles.

---

## Post-Design Constitution Re-Check

*Re-evaluated after Phase 1 design completion*

### Principle I: Simplicity First ✅
- **Verified**: Design uses 4 interfaces (Tokenizer, Filter, CacheLoader, Validator) - each with single responsibility
- **Verified**: No additional dependencies beyond those identified in research.md
- **Verified**: Manual JSON fixtures chosen over complex recording libraries (govcr rejected)

### Principle II: Test-First Development ✅
- **Verified**: quickstart.md specifies acceptance tests BEFORE implementation
- **Verified**: Contract tests defined for all interfaces in contracts/README.md
- **Verified**: Test structure follows Red-Green-Refactor cycle

### Principle III: Production-Ready Observability ⚠️
- **Deviation remains justified**: PoC scope, verbose mode provides debugging
- **Note**: Design does NOT add Prometheus metrics (appropriate for PoC)

### Principle IV: Stateless Architecture ✅
- **Verified**: Runner has no mutable state, each benchmark run independent
- **Verified**: All configuration from files (testdata/) or CLI flags

### Principle V: Fail-Safe Defaults ✅
- **Verified**: Error handling design - fail-fast on missing cache files
- **Verified**: Validation layer prevents silent data loss

### Principle VI: Interface-Driven Design ✅
- **Verified**: 4 Go interfaces defined in contracts/
- **Verified**: Implementations can be swapped (e.g., different tokenizers)
- **Verified**: Testable with mocks

### Principle VII: Zero Breaking Changes ✅
- **Verified**: Design only reads MCP responses, doesn't modify protocol
- **Verified**: Cache format extensible (can add fields without breaking)

**Final Gate Status**: PASS

All principles compliant or justified deviations documented. Design ready for `/speckit.tasks`.
