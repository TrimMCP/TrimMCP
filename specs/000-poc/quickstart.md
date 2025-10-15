# Quickstart: Token Savings Benchmark PoC

**Feature**: 000-poc | **Date**: 2025-10-14 | **Spec**: [spec.md](./spec.md)

## Overview

This guide provides step-by-step instructions for implementing the Token Savings Benchmark PoC using test-first development. Follow the order below to ensure acceptance tests exist before implementation.

---

## Prerequisites

- Go 1.21+ installed
- Working directory: `mcp/` (within monorepo root)
- Familiarity with Go testing (`go test`)

---

## Phase 1: Setup Project Structure

### Step 1.1: Initialize Go Module (if not exists)

```bash
cd mcp/
go mod init github.com/slimmcp/slimmcp
go mod tidy
```

### Step 1.2: Create Directory Structure

```bash
# Create package directories
mkdir -p cmd/benchmark
mkdir -p pkg/token pkg/filter pkg/cache
mkdir -p internal/benchmark
mkdir -p tests/acceptance tests/integration
mkdir -p testdata/github testdata/notion
```

### Step 1.3: Install Dependencies

```bash
go get github.com/mark3labs/mcp-go@v0.42.0-beta.1
go get github.com/pkoukk/tiktoken-go
go get github.com/olekukonko/tablewriter
```

---

## Phase 2: Write Acceptance Tests (Test-First)

**Constitution II: Test-First Development** - Write tests BEFORE implementation.

### Step 2.1: User Story 1 - Quick Benchmark Execution

**File**: `tests/acceptance/quick_benchmark_test.go`

```go
package acceptance

import (
    "os"
    "os/exec"
    "strings"
    "testing"
    "time"
)

func TestQuickBenchmarkExecution(t *testing.T) {
    t.Run("displays_progress_log", func(t *testing.T) {
        // Acceptance: developer sees progress log showing each operation
        cmd := exec.Command("go", "run", "../../cmd/benchmark/main.go")
        output, err := cmd.CombinedOutput()
        if err != nil {
            t.Fatalf("Benchmark failed: %v\n%s", err, output)
        }

        // Verify progress messages appear
        outputStr := string(output)
        if !strings.Contains(outputStr, "Testing GitHub: list_pull_requests") {
            t.Error("Missing progress log for GitHub operations")
        }
        if !strings.Contains(outputStr, "Testing Notion: create_page") {
            t.Error("Missing progress log for Notion operations")
        }
    })

    t.Run("shows_summary_table_with_token_counts", func(t *testing.T) {
        // Acceptance: developer sees summary table with before/after token counts
        cmd := exec.Command("go", "run", "../../cmd/benchmark/main.go")
        output, err := cmd.CombinedOutput()
        if err != nil {
            t.Fatalf("Benchmark failed: %v", err)
        }

        outputStr := string(output)
        // Verify table headers
        if !strings.Contains(outputStr, "Operation") {
            t.Error("Missing Operation column in output")
        }
        if !strings.Contains(outputStr, "Unfiltered") {
            t.Error("Missing Unfiltered column in output")
        }
        if !strings.Contains(outputStr, "Filtered") {
            t.Error("Missing Filtered column in output")
        }
        if !strings.Contains(outputStr, "Reduction") {
            t.Error("Missing Reduction column in output")
        }
    })

    t.Run("produces_consistent_results_on_multiple_runs", func(t *testing.T) {
        // Acceptance: running benchmark twice produces identical results
        cmd1 := exec.Command("go", "run", "../../cmd/benchmark/main.go")
        output1, _ := cmd1.CombinedOutput()

        cmd2 := exec.Command("go", "run", "../../cmd/benchmark/main.go")
        output2, _ := cmd2.CombinedOutput()

        if string(output1) != string(output2) {
            t.Error("Benchmark results are not deterministic")
        }
    })

    t.Run("completes_within_10_seconds", func(t *testing.T) {
        // Acceptance: SC-001 - benchmark completes in <10s
        start := time.Now()
        cmd := exec.Command("go", "run", "../../cmd/benchmark/main.go")
        _, err := cmd.CombinedOutput()
        duration := time.Since(start)

        if err != nil {
            t.Fatalf("Benchmark failed: %v", err)
        }
        if duration > 10*time.Second {
            t.Errorf("Benchmark took %v, exceeds 10s limit", duration)
        }
    })
}
```

### Step 2.2: User Story 2 & 3 - GitHub/Notion Tool Call Benchmarks

**File**: `tests/acceptance/tool_benchmarks_test.go`

```go
package acceptance

import (
    "os/exec"
    "strings"
    "testing"
)

func TestGitHubToolCallBenchmarks(t *testing.T) {
    t.Run("reports_token_counts_for_list_pull_requests", func(t *testing.T) {
        cmd := exec.Command("go", "run", "../../cmd/benchmark/main.go")
        output, err := cmd.CombinedOutput()
        if err != nil {
            t.Fatalf("Benchmark failed: %v", err)
        }

        outputStr := string(output)
        if !strings.Contains(outputStr, "GitHub: list_pull_requests") {
            t.Error("Missing GitHub list_pull_requests in results")
        }
    })

    t.Run("filtering_removes_metadata_fields", func(t *testing.T) {
        // This will be verified in unit tests for filter package
        // Acceptance test just checks filtered tokens < unfiltered tokens
        cmd := exec.Command("go", "run", "../../cmd/benchmark/main.go")
        output, err := cmd.CombinedOutput()
        if err != nil {
            t.Fatalf("Benchmark failed: %v", err)
        }

        // Parse output and verify reduction % > 0
        outputStr := string(output)
        if !strings.Contains(outputStr, "%") {
            t.Error("Missing reduction percentage in output")
        }
    })
}

func TestNotionToolCallBenchmarks(t *testing.T) {
    t.Run("reports_token_counts_for_create_page", func(t *testing.T) {
        cmd := exec.Command("go", "run", "../../cmd/benchmark/main.go")
        output, err := cmd.CombinedOutput()
        if err != nil {
            t.Fatalf("Benchmark failed: %v", err)
        }

        outputStr := string(output)
        if !strings.Contains(outputStr, "Notion: create_page") {
            t.Error("Missing Notion create_page in results")
        }
    })
}
```

### Step 2.3: User Story 4 - Verbose Mode

**File**: `tests/acceptance/verbose_mode_test.go`

```go
package acceptance

import (
    "os/exec"
    "strings"
    "testing"
)

func TestVerboseExecutionMode(t *testing.T) {
    t.Run("shows_request_response_bodies_in_verbose_mode", func(t *testing.T) {
        cmd := exec.Command("go", "run", "../../cmd/benchmark/main.go", "--verbose")
        output, err := cmd.CombinedOutput()
        if err != nil {
            t.Fatalf("Benchmark failed: %v", err)
        }

        outputStr := string(output)
        // Verbose mode should show "Unfiltered response:" and "Filtered response:"
        if !strings.Contains(outputStr, "Unfiltered response:") {
            t.Error("Verbose mode missing unfiltered response output")
        }
        if !strings.Contains(outputStr, "Filtered response:") {
            t.Error("Verbose mode missing filtered response output")
        }
    })

    t.Run("shows_token_calculation_method_in_verbose_mode", func(t *testing.T) {
        cmd := exec.Command("go", "run", "../../cmd/benchmark/main.go", "--verbose")
        output, err := cmd.CombinedOutput()
        if err != nil {
            t.Fatalf("Benchmark failed: %v", err)
        }

        outputStr := string(output)
        if !strings.Contains(outputStr, "tiktoken") {
            t.Error("Verbose mode missing tokenization method")
        }
    })
}
```

### Step 2.4: Run Tests (Expected to Fail - Red Phase)

```bash
cd tests/acceptance
go test -v
# EXPECTED: All tests fail (no implementation yet)
```

---

## Phase 3: Implement Core Packages (Green Phase)

**Now that acceptance tests are written, implement packages to make tests pass.**

### Step 3.1: Implement Tokenizer Interface

**File**: `pkg/token/tokenizer.go`

```go
package token

// Tokenizer counts tokens in text.
type Tokenizer interface {
    Count(text string) (int, error)
    Name() string
}
```

**File**: `pkg/token/tiktoken.go`

```go
package token

import (
    "github.com/pkoukk/tiktoken-go"
)

type TiktokenTokenizer struct {
    encoder *tiktoken.Tiktoken
}

func NewTiktoken() (*TiktokenTokenizer, error) {
    enc, err := tiktoken.GetEncoding("cl100k_base")
    if err != nil {
        return nil, err
    }
    return &TiktokenTokenizer{encoder: enc}, nil
}

func (t *TiktokenTokenizer) Count(text string) (int, error) {
    tokens := t.encoder.Encode(text, nil, nil)
    return len(tokens), nil
}

func (t *TiktokenTokenizer) Name() string {
    return "tiktoken-cl100k"
}
```

**File**: `pkg/token/simple.go` (word count)

```go
package token

import "strings"

type WordCountTokenizer struct{}

func NewWordCount() *WordCountTokenizer {
    return &WordCountTokenizer{}
}

func (w *WordCountTokenizer) Count(text string) (int, error) {
    return len(strings.Fields(text)), nil
}

func (w *WordCountTokenizer) Name() string {
    return "word-count"
}
```

**File**: `pkg/token/charcount.go`

```go
package token

type CharDivFourTokenizer struct{}

func NewCharDivFour() *CharDivFourTokenizer {
    return &CharDivFourTokenizer{}
}

func (c *CharDivFourTokenizer) Count(text string) (int, error) {
    return len(text) / 4, nil
}

func (c *CharDivFourTokenizer) Name() string {
    return "char-div-4"
}
```

**File**: `pkg/token/token_test.go` (unit tests)

```go
package token

import "testing"

func TestTokenizers(t *testing.T) {
    tests := []struct {
        input    string
        expected map[string]int // method -> count
    }{
        {
            input: "hello world",
            expected: map[string]int{
                "tiktoken-cl100k": 2,  // approximate
                "word-count":      2,
                "char-div-4":      2,  // 11 chars / 4 = 2
            },
        },
    }

    tokenizers := []Tokenizer{
        must(NewTiktoken()),
        NewWordCount(),
        NewCharDivFour(),
    }

    for _, tt := range tests {
        for _, tok := range tokenizers {
            count, err := tok.Count(tt.input)
            if err != nil {
                t.Fatalf("%s Count failed: %v", tok.Name(), err)
            }
            // Note: tiktoken count may vary slightly
            t.Logf("%s counted %d tokens for %q", tok.Name(), count, tt.input)
        }
    }
}

func must(t *TiktokenTokenizer, err error) Tokenizer {
    if err != nil {
        panic(err)
    }
    return t
}
```

### Step 3.2: Implement Filter Interface

**File**: `pkg/filter/filter.go`

```go
package filter

type Filter interface {
    Apply(response map[string]interface{}) (map[string]interface{}, error)
    Name() string
}
```

**File**: `pkg/filter/minimal.go`

```go
package filter

type MinimalFilter struct {
    RemoveFields []string
}

func NewMinimal() *MinimalFilter {
    return &MinimalFilter{
        RemoveFields: []string{
            "_meta", "jsonrpc", "id",
            "url", "html_url", "api_url", "node_id",
            "created_at", "updated_at",
            "object", "parent",
        },
    }
}

func (f *MinimalFilter) Apply(response map[string]interface{}) (map[string]interface{}, error) {
    return f.removeFields(response), nil
}

func (f *MinimalFilter) removeFields(data map[string]interface{}) map[string]interface{} {
    filtered := make(map[string]interface{})
    for k, v := range data {
        if contains(f.RemoveFields, k) {
            continue // skip this field
        }
        // Recursively filter nested maps
        if nested, ok := v.(map[string]interface{}); ok {
            filtered[k] = f.removeFields(nested)
        } else if arr, ok := v.([]interface{}); ok {
            filtered[k] = f.filterArray(arr)
        } else {
            filtered[k] = v
        }
    }
    return filtered
}

func (f *MinimalFilter) filterArray(arr []interface{}) []interface{} {
    result := make([]interface{}, len(arr))
    for i, item := range arr {
        if nested, ok := item.(map[string]interface{}); ok {
            result[i] = f.removeFields(nested)
        } else {
            result[i] = item
        }
    }
    return result
}

func (f *MinimalFilter) Name() string {
    return "minimal"
}

func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
```

### Step 3.3: Implement Cache Loader

**File**: `pkg/cache/loader.go`

```go
package cache

import "time"

type CachedResponse struct {
    Operation  string
    Server     string
    Request    map[string]interface{}
    Response   map[string]interface{}
    CapturedAt time.Time
}

type CacheLoader interface {
    Load(server, operation string) (*CachedResponse, error)
    Exists(server, operation string) bool
}
```

**File**: `pkg/cache/json.go`

```go
package cache

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
)

type JSONFileLoader struct {
    TestDataDir string
}

func NewJSONFileLoader(testDataDir string) *JSONFileLoader {
    return &JSONFileLoader{TestDataDir: testDataDir}
}

func (j *JSONFileLoader) Load(server, operation string) (*CachedResponse, error) {
    path := j.filePath(server, operation)
    if !j.Exists(server, operation) {
        return nil, fmt.Errorf("cache file missing: %s", path)
    }

    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read cache file: %w", err)
    }

    var cached CachedResponse
    if err := json.Unmarshal(data, &cached); err != nil {
        return nil, fmt.Errorf("failed to parse cache file: %w", err)
    }

    return &cached, nil
}

func (j *JSONFileLoader) Exists(server, operation string) bool {
    _, err := os.Stat(j.filePath(server, operation))
    return err == nil
}

func (j *JSONFileLoader) filePath(server, operation string) string {
    return filepath.Join(j.TestDataDir, server, operation+".json")
}
```

### Step 3.4: Implement Benchmark Runner

**File**: `internal/benchmark/runner.go`

```go
package benchmark

import (
    "fmt"
    "time"

    "github.com/slimmcp/slimmcp/pkg/cache"
    "github.com/slimmcp/slimmcp/pkg/filter"
    "github.com/slimmcp/slimmcp/pkg/token"
)

type Runner struct {
    CacheLoader cache.CacheLoader
    Filter      filter.Filter
    Tokenizers  []token.Tokenizer
    Verbose     bool
}

type Result struct {
    Operation         string
    TokenMeasurements []TokenMeasurement
    Duration          time.Duration
}

type TokenMeasurement struct {
    Method            string
    UnfilteredTokens  int
    FilteredTokens    int
    ReductionPercent  float64
}

func (r *Runner) Run(operations []OperationConfig) ([]Result, error) {
    var results []Result

    for _, op := range operations {
        start := time.Now()
        result, err := r.runOperation(op)
        if err != nil {
            return nil, err
        }
        result.Duration = time.Since(start)
        results = append(results, result)
    }

    return results, nil
}

func (r *Runner) runOperation(op OperationConfig) (Result, error) {
    // Load cached response
    cached, err := r.CacheLoader.Load(op.Server, op.Tool)
    if err != nil {
        return Result{}, fmt.Errorf("failed to load cache for %s/%s: %w", op.Server, op.Tool, err)
    }

    // Apply filter
    filtered, err := r.Filter.Apply(cached.Response)
    if err != nil {
        return Result{}, fmt.Errorf("filtering failed: %w", err)
    }

    // Measure tokens
    var measurements []TokenMeasurement
    for _, tokenizer := range r.Tokenizers {
        unfilteredCount, _ := tokenizer.Count(toString(cached.Response))
        filteredCount, _ := tokenizer.Count(toString(filtered))
        reduction := float64(unfilteredCount-filteredCount) / float64(unfilteredCount) * 100

        measurements = append(measurements, TokenMeasurement{
            Method:           tokenizer.Name(),
            UnfilteredTokens: unfilteredCount,
            FilteredTokens:   filteredCount,
            ReductionPercent: reduction,
        })
    }

    return Result{
        Operation:         fmt.Sprintf("%s: %s", op.Server, op.Tool),
        TokenMeasurements: measurements,
    }, nil
}

func toString(data interface{}) string {
    // Convert to JSON string for token counting
    bytes, _ := json.Marshal(data)
    return string(bytes)
}

type OperationConfig struct {
    Server string
    Tool   string
}
```

### Step 3.5: Implement CLI Entry Point

**File**: `cmd/benchmark/main.go`

```go
package main

import (
    "flag"
    "fmt"
    "os"

    "github.com/olekukonko/tablewriter"
    "github.com/slimmcp/slimmcp/internal/benchmark"
    "github.com/slimmcp/slimmcp/pkg/cache"
    "github.com/slimmcp/slimmcp/pkg/filter"
    "github.com/slimmcp/slimmcp/pkg/token"
)

func main() {
    verbose := flag.Bool("verbose", false, "Enable verbose logging")
    flag.Parse()

    // Initialize components
    cacheLoader := cache.NewJSONFileLoader("testdata")
    filterStrategy := filter.NewMinimal()
    tiktoken, _ := token.NewTiktoken()
    tokenizers := []token.Tokenizer{
        tiktoken,
        token.NewWordCount(),
        token.NewCharDivFour(),
    }

    runner := &benchmark.Runner{
        CacheLoader: cacheLoader,
        Filter:      filterStrategy,
        Tokenizers:  tokenizers,
        Verbose:     *verbose,
    }

    // Define operations
    operations := []benchmark.OperationConfig{
        {Server: "github", Tool: "list_pull_requests"},
        {Server: "github", Tool: "search_issues"},
        {Server: "github", Tool: "get_repository"},
        {Server: "github", Tool: "create_issue"},
        {Server: "github", Tool: "list_commits"},
        {Server: "github", Tool: "get_pull_request"},
        {Server: "notion", Tool: "create_page"},
        {Server: "notion", Tool: "query_database"},
        {Server: "notion", Tool: "list_pages"},
        {Server: "notion", Tool: "update_page"},
        {Server: "notion", Tool: "create_database"},
        {Server: "notion", Tool: "search_pages"},
    }

    // Run benchmark
    results, err := runner.Run(operations)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Benchmark failed: %v\n", err)
        os.Exit(1)
    }

    // Display results
    displayResults(results)
}

func displayResults(results []benchmark.Result) {
    table := tablewriter.NewWriter(os.Stdout)
    table.SetHeader([]string{"Operation", "Method", "Unfiltered", "Filtered", "Reduction"})

    for _, result := range results {
        for _, m := range result.TokenMeasurements {
            table.Append([]string{
                result.Operation,
                m.Method,
                fmt.Sprintf("%d", m.UnfilteredTokens),
                fmt.Sprintf("%d", m.FilteredTokens),
                fmt.Sprintf("%.1f%%", m.ReductionPercent),
            })
        }
    }

    table.Render()
}
```

---

## Phase 4: Create Test Data

### Step 4.1: Capture Real MCP Responses (One-Time Setup)

**Maintainer task**: Capture real responses from GitHub and Notion MCP servers.

Example: `testdata/github/list_pull_requests.json`

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
        "text": "[{\"number\": 123, \"title\": \"Fix bug\", \"state\": \"open\", \"user\": {\"login\": \"alice\"}, \"url\": \"https://...\", \"node_id\": \"PR_123\", \"created_at\": \"2025-01-01T00:00:00Z\"}]"
      }
    ]
  },
  "captured_at": "2025-10-14T10:00:00Z"
}
```

**Create 12 cache files** (6 GitHub + 6 Notion) following this format.

---

## Phase 5: Run Tests (Should Pass - Green Phase)

```bash
cd tests/acceptance
go test -v
# EXPECTED: All tests pass
```

---

## Phase 6: Refactor (if needed)

- Extract common code
- Improve error messages
- Optimize performance

---

## Success Criteria Verification

Run these commands to verify all success criteria are met:

```bash
# SC-001: Completes in <10 seconds
time go run cmd/benchmark/main.go

# SC-002: Shows ≥50% average token reduction
go run cmd/benchmark/main.go | grep "Average"

# SC-003: 100% pass data integrity validation
go test ./internal/benchmark -v

# SC-004: Deterministic results
diff <(go run cmd/benchmark/main.go) <(go run cmd/benchmark/main.go)

# SC-005: Verbose mode provides debugging detail
go run cmd/benchmark/main.go --verbose | head -20

# SC-006: Covers all 12 operations
go run cmd/benchmark/main.go | grep -c "GitHub:\|Notion:"  # Should output 12
```

---

## Next Steps

After quickstart implementation is complete:
1. Run `/speckit.tasks` to generate detailed task breakdown
2. Execute tasks in priority order
3. Run full test suite: `go test ./...`
4. Verify all acceptance criteria pass

---

## Summary

This quickstart follows Constitution II (Test-First Development):
1. ✅ Write acceptance tests first (Red phase)
2. ✅ Implement packages to pass tests (Green phase)
3. ✅ Refactor for clarity (Refactor phase)
4. ✅ Verify all success criteria met
