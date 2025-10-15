# Research & Technology Decisions: Token Savings Benchmark PoC

**Feature**: 000-poc | **Date**: 2025-10-14 | **Plan**: [plan.md](./plan.md)

## Overview

This document consolidates research findings and technology decisions for the Token Savings Benchmark PoC. All NEEDS CLARIFICATION items from Technical Context are resolved here.

---

## Decision: Tokenization Libraries for Side-by-Side Comparison

**Context**: FR-005 requires "multiple tokenization methods reported side-by-side" to validate token reduction across different counting approaches.

### Decision

Implement three tokenization methods:

1. **tiktoken-go** (GPT-4 cl100k_base encoding) - Primary/most accurate
2. **Simple word-based** (whitespace splitting) - Language-agnostic approximation
3. **Character-based** (length/4) - Quick estimation method

**Rationale**:
- **tiktoken-go** (`github.com/pkoukk/tiktoken-go`): Industry standard for GPT tokenization, provides accurate counts matching OpenAI's billing. Essential for validating real-world cost savings.
- **Simple word-based**: No external dependency, language-agnostic, provides rough approximation useful for non-English content or quick estimates.
- **Character-based**: Fastest method, useful baseline. Common rule-of-thumb in NLP (1 token ≈ 4 characters for English).

**Alternatives Considered**:
- **sentencepiece** (Google): Overkill for PoC, requires model files, harder to integrate with Go
- **anthropic/tokenizers** (Claude): No official Go library, would require Python bridge
- **gpt-3-encoder** (JS port): Would require Node.js bridge, unnecessary complexity

**Implementation**:
```go
// pkg/token/tokenizer.go
type Tokenizer interface {
    Count(text string) int
    Name() string // "tiktoken-cl100k", "word-count", "char-div-4"
}

// Multiple tokenizers used in benchmark
tokenizers := []Tokenizer{
    NewTiktoken(),     // Primary - GPT-4 accurate
    NewWordCount(),    // Secondary - simple approximation
    NewCharDivFour(),  // Tertiary - quick estimate
}
```

---

## Decision: MCP Response Caching Strategy

**Context**: FR-009 requires cached responses stored as separate JSON files to avoid live API calls.

### Decision

Use **manual JSON fixtures** (not govcr or go-vcr).

**Rationale**:
- **Explicit control**: JSON files are easy to inspect, modify, and version control
- **Simplicity**: No recording infrastructure needed, just save MCP responses as JSON
- **Reproducibility**: Fixtures ensure deterministic benchmarks across environments
- **Documentation**: JSON files serve as examples of real MCP responses for future development

**Alternatives Considered**:
- **govcr**: HTTP recording library, but MCP responses don't come over HTTP in benchmark (we're testing filtering logic, not HTTP)
- **go-vcr**: Same limitation, designed for HTTP testing not JSON fixture management
- **Live recording mode**: Rejected because it would require GitHub/Notion API credentials, defeating FR-009's "no credentials required" goal

**Implementation Approach**:
1. Manually capture MCP responses from real GitHub/Notion MCP servers (one-time setup by maintainer)
2. Save each response as `testdata/{server}/{operation}.json`
3. Benchmark loads JSON files directly using `encoding/json`
4. No HTTP recording library needed

**File Format**:
```json
{
  "operation": "list_pull_requests",
  "request": {
    "method": "tools/call",
    "params": {
      "name": "list_pull_requests",
      "arguments": {"state": "open"}
    }
  },
  "response": {
    "content": [
      {
        "type": "text",
        "text": "[actual GitHub API response JSON]"
      }
    ]
  },
  "captured_at": "2025-10-14T10:00:00Z"
}
```

---

## Decision: Filtering Strategy for Token Reduction

**Context**: FR-004 requires filtering that "removes metadata, redundant fields, and verbose formatting."

### Decision

Implement **minimal filtering strategy** with configurable field removal rules.

**Filtering Rules**:
1. **Remove MCP metadata fields**: `_meta`, `jsonrpc`, `id` (protocol overhead)
2. **Remove GitHub-specific metadata**: `url`, `node_id`, `html_url`, `api_url`, timestamps (non-essential for content)
3. **Remove Notion-specific metadata**: `object`, `id`, `created_time`, `last_edited_time`, `parent` (preserving only content)
4. **Compact whitespace**: Remove indentation/pretty-printing from JSON

**Rationale**:
- Targets highest token-count offenders (URLs, IDs, timestamps)
- Preserves essential data (PR titles, issue numbers, task content)
- Simple rule-based approach (no ML/NLP needed for PoC)
- Measurable impact (can track tokens removed per field type)

**Alternatives Considered**:
- **Smart filtering with ML**: Overkill for PoC, deferred to future iteration
- **Zero filtering (pass-through only)**: Wouldn't validate token reduction claims
- **Aggressive filtering (content-only)**: Risk of removing too much, failing data integrity validation

**Implementation**:
```go
// pkg/filter/minimal.go
type MinimalFilter struct {
    RemoveFields []string // configurable field blacklist
}

func (f *MinimalFilter) Apply(response map[string]interface{}) (map[string]interface{}, error) {
    filtered := make(map[string]interface{})
    for k, v := range response {
        if !contains(f.RemoveFields, k) {
            filtered[k] = v
        }
    }
    return filtered, nil
}
```

---

## Decision: Data Integrity Validation Approach

**Context**: FR-010 requires structural validation (valid JSON, required fields present, correct types).

### Decision

Implement **JSON schema validation** using manual schema definitions.

**Validation Rules**:
1. **Valid JSON**: Filtered response must parse successfully
2. **Required fields**: Check presence of essential fields (e.g., `title`, `number` for PRs)
3. **Type correctness**: Verify field types match expected schema (string, int, array, etc.)

**Rationale**:
- JSON schema is standard for API validation
- Can define schemas per MCP operation
- Catches both missing data and type errors
- Testable with known-good and known-bad inputs

**Alternatives Considered**:
- **Semantic equivalence testing**: Compare parsed data structures, too complex for PoC
- **Manual field checks**: Not scalable, error-prone
- **Go struct validation**: Would require defining structs for every MCP response type

**Implementation**:
```go
// internal/benchmark/validator.go
type Validator struct {
    schemas map[string]Schema // operation -> required schema
}

func (v *Validator) Validate(operation string, filtered map[string]interface{}) error {
    schema := v.schemas[operation]
    if err := schema.CheckRequiredFields(filtered); err != nil {
        return fmt.Errorf("missing required fields: %w", err)
    }
    if err := schema.CheckTypes(filtered); err != nil {
        return fmt.Errorf("type mismatch: %w", err)
    }
    return nil
}
```

---

## Best Practices Research

### Go CLI Development
- Use `flag` package for argument parsing (stdlib, simple)
- Use `log` package with configurable output for verbose mode
- Exit with proper status codes (0 = success, 1 = error)
- Provide help text via `--help` flag

### Go Testing Patterns
- Table-driven tests for tokenizer comparison
- Test fixtures in `testdata/` directory (Go convention)
- Subtests for different scenarios (`t.Run()`)
- Golden files for expected outputs

### Token Counting Best Practices
- Always use tiktoken for GPT model accuracy
- Cache tokenizer initialization (expensive)
- Handle Unicode correctly (count by tokens, not bytes/characters)
- Report multiple methods to show variance

### MCP Protocol Handling
- Parse MCP responses as generic `map[string]interface{}` for flexibility
- Preserve original structure in testdata for future reference
- Document MCP protocol version in cache files

---

## Summary of Resolved Clarifications

| Original Question | Decision | Source |
|-------------------|----------|--------|
| Additional tokenization libraries for side-by-side comparison | tiktoken-go, word-count, char/4 | This research |
| How to cache MCP responses | Manual JSON fixtures in testdata/ | This research |
| Filtering strategy details | Minimal filtering (remove metadata, URLs, timestamps) | This research |
| Data integrity validation method | JSON schema validation with required fields + type checks | This research |

All NEEDS CLARIFICATION items from plan.md Technical Context are now resolved.
