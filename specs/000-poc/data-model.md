# Data Model: Token Savings Benchmark PoC

**Feature**: 000-poc | **Date**: 2025-10-14 | **Spec**: [spec.md](./spec.md)

## Overview

This document defines the data structures and entities used in the Token Savings Benchmark PoC. All entities are derived from the Key Entities section in spec.md.

---

## Core Entities

### Benchmark Operation

Represents a single MCP tool call being tested.

**Fields**:
- `OperationName` (string): Human-readable name (e.g., "GitHub: list_pull_requests")
- `Server` (string): MCP server name ("github" or "notion")
- `ToolName` (string): MCP tool name (e.g., "list_pull_requests")
- `UnfilteredResponse` (map[string]interface{}): Original MCP response from cache
- `FilteredResponse` (map[string]interface{}): Response after filtering applied
- `TokenMeasurements` ([]TokenMeasurement): Token counts for this operation

**Relationships**:
- Contains multiple TokenMeasurement entries (one per tokenization method)
- References a CachedResponse as source data

**Validation Rules**:
- `OperationName` must be non-empty
- `Server` must be "github" or "notion"
- `UnfilteredResponse` must be valid JSON
- `FilteredResponse` must pass data integrity validation (required fields present, correct types)

**State Transitions**:
1. **Created**: Operation initialized with name and server
2. **Loaded**: UnfilteredResponse loaded from cache file
3. **Filtered**: FilteredResponse generated from UnfilteredResponse
4. **Measured**: TokenMeasurements calculated for both responses
5. **Validated**: FilteredResponse passes integrity checks
6. **Complete**: Ready for reporting

---

### Token Measurement

Contains token counts for a given text payload using a specific tokenization method.

**Fields**:
- `Method` (string): Tokenization method name ("tiktoken-cl100k", "word-count", "char-div-4")
- `UnfilteredTokens` (int): Token count for original response
- `FilteredTokens` (int): Token count for filtered response
- `ReductionPercent` (float64): Percentage reduction ((Unfiltered - Filtered) / Unfiltered * 100)
- `Timestamp` (time.Time): When measurement was taken

**Relationships**:
- Belongs to a BenchmarkOperation
- Associated with a specific Tokenizer implementation

**Validation Rules**:
- `UnfilteredTokens` must be >= 0
- `FilteredTokens` must be >= 0
- `FilteredTokens` must be <= `UnfilteredTokens` (filtering never adds tokens)
- `ReductionPercent` must be between 0 and 100

**Calculation Logic**:
```go
ReductionPercent = ((UnfilteredTokens - FilteredTokens) / UnfilteredTokens) * 100
```

---

### Cached Response

A recorded MCP server response stored locally as a JSON file.

**Fields**:
- `FilePath` (string): Path to JSON file (e.g., "testdata/github/list_pull_requests.json")
- `Operation` (string): MCP operation name (matches filename without extension)
- `Server` (string): MCP server name ("github" or "notion")
- `Request` (map[string]interface{}): Original MCP request parameters
- `Response` (map[string]interface{}): Full MCP response body
- `CapturedAt` (time.Time): When the response was recorded

**Relationships**:
- Source for BenchmarkOperation.UnfilteredResponse
- One CachedResponse per MCP tool operation

**Validation Rules**:
- `FilePath` must exist on filesystem
- `Operation` must match one of the 12 required operations
- `Server` must be "github" or "notion"
- `Response` must be valid JSON
- File must be readable and parseable

**File Format** (JSON):
```json
{
  "operation": "list_pull_requests",
  "server": "github",
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
        "text": "{...actual GitHub API response...}"
      }
    ]
  },
  "captured_at": "2025-10-14T10:00:00Z"
}
```

**State Transitions**:
1. **Stored**: JSON file exists on disk
2. **Loaded**: Parsed into Go data structure
3. **Validated**: Passes structural checks
4. **Consumed**: Used to create BenchmarkOperation

---

### Filter Strategy

Defines the rules for removing data from MCP responses.

**Fields**:
- `Name` (string): Strategy name (e.g., "minimal")
- `RemoveFields` ([]string): List of field names to remove (e.g., ["_meta", "url", "id"])
- `CompactWhitespace` (bool): Whether to remove JSON formatting whitespace
- `PreserveEssentials` (bool): Whether to validate essential fields remain after filtering

**Relationships**:
- Applied to BenchmarkOperation to generate FilteredResponse
- One FilterStrategy used across all operations (for consistency)

**Validation Rules**:
- `Name` must be non-empty
- `RemoveFields` list must not include essential fields (validated against schema)
- If `PreserveEssentials` is true, must have corresponding validation schema

**Configuration Example**:
```go
strategy := FilterStrategy{
    Name: "minimal",
    RemoveFields: []string{
        "_meta", "jsonrpc", "id",        // MCP protocol overhead
        "url", "html_url", "api_url",    // GitHub URLs
        "node_id",                        // GitHub internal IDs
        "created_at", "updated_at",      // Timestamps
        "object", "parent",               // Notion metadata
    },
    CompactWhitespace: true,
    PreserveEssentials: true,
}
```

---

## Supporting Data Structures

### Benchmark Config

Configuration for the entire benchmark run.

**Fields**:
- `TestDataDir` (string): Path to testdata/ directory (default: "testdata")
- `VerboseMode` (bool): Enable verbose logging (default: false)
- `Operations` ([]OperationConfig): List of operations to benchmark

**Example**:
```go
config := BenchmarkConfig{
    TestDataDir: "testdata",
    VerboseMode: false,
    Operations: []OperationConfig{
        {Server: "github", Tool: "list_pull_requests"},
        {Server: "github", Tool: "search_issues"},
        // ... 10 more operations
    },
}
```

---

### Operation Config

Configuration for a single benchmark operation.

**Fields**:
- `Server` (string): MCP server name ("github" or "notion")
- `Tool` (string): MCP tool name (e.g., "list_pull_requests")
- `CacheFile` (string): Override cache file path (optional, defaults to testdata/{server}/{tool}.json)

---

### Benchmark Report

Results of a complete benchmark run.

**Fields**:
- `TotalOperations` (int): Number of operations tested (should be 12)
- `SuccessfulOperations` (int): Number that passed validation
- `FailedOperations` (int): Number that failed (missing cache, validation error)
- `Operations` ([]BenchmarkOperation): Detailed results per operation
- `AverageReduction` (map[string]float64): Average token reduction per tokenization method
- `Duration` (time.Duration): Total benchmark execution time

**Validation Rules**:
- `TotalOperations` must equal 12 (per SC-006)
- `SuccessfulOperations` + `FailedOperations` must equal `TotalOperations`
- Benchmark must complete in < 10 seconds (per SC-001)

**Output Format** (Terminal Table):
```
┌────────────────────────────────┬────────────┬──────────────┬──────────────┬────────────┐
│ Operation                      │ Method     │ Unfiltered   │ Filtered     │ Reduction  │
├────────────────────────────────┼────────────┼──────────────┼──────────────┼────────────┤
│ GitHub: list_pull_requests     │ tiktoken   │ 4523         │ 1842         │ 59.3%      │
│                                │ word-count │ 3891         │ 1654         │ 57.5%      │
│                                │ char-div-4 │ 4102         │ 1723         │ 58.0%      │
├────────────────────────────────┼────────────┼──────────────┼──────────────┼────────────┤
│ GitHub: search_issues          │ tiktoken   │ 3214         │ 1456         │ 54.7%      │
│ ...                            │ ...        │ ...          │ ...          │ ...        │
└────────────────────────────────┴────────────┴──────────────┴──────────────┴────────────┘

Average Reduction (tiktoken): 56.8%
Average Reduction (word-count): 54.2%
Average Reduction (char-div-4): 55.1%

Duration: 3.42s
Status: ✓ All 12 operations passed validation
```

---

## Entity Relationship Diagram

```
BenchmarkConfig
    │
    ├─> OperationConfig[] (12 operations)
    │
    └─> BenchmarkReport
            │
            └─> BenchmarkOperation[]
                    │
                    ├─> CachedResponse (loads from JSON file)
                    ├─> FilterStrategy (applies filtering rules)
                    └─> TokenMeasurement[] (3 methods per operation)
```

---

## Data Flow

1. **Initialize**: Load BenchmarkConfig with 12 OperationConfigs
2. **Load**: For each operation, load CachedResponse from testdata/{server}/{tool}.json
3. **Filter**: Apply FilterStrategy to generate FilteredResponse
4. **Measure**: Calculate TokenMeasurements using 3 tokenization methods
5. **Validate**: Check FilteredResponse passes data integrity validation
6. **Report**: Aggregate results into BenchmarkReport and display table

---

## File Storage

### Directory Structure
```
testdata/
├── github/
│   ├── list_pull_requests.json
│   ├── search_issues.json
│   ├── get_repository.json
│   ├── create_issue.json
│   ├── list_commits.json
│   └── get_pull_request.json
└── notion/
    ├── create_page.json
    ├── query_database.json
    ├── list_pages.json
    ├── update_page.json
    ├── create_database.json
    └── search_pages.json
```

### File Naming Convention
- Lowercase operation names
- Underscores for word separation
- `.json` extension
- Matches MCP tool names exactly

---

## Summary

All entities from spec.md Key Entities section are fully defined with:
- Field specifications
- Validation rules
- Relationships to other entities
- State transitions where applicable
- Example data formats
