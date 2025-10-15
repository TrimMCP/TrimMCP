# Token Savings Benchmark PoC

**Spec**: 000-poc | **Status**: ✅ Completed | **Code**: [../poc/](../poc/)

Proof-of-concept benchmark demonstrating **62.6% token reduction** across GitHub and Notion MCP operations.

## Quick Start

```bash
# Navigate to PoC directory
cd docs/poc

# Build
go build -o bin/benchmark ./benchmark/

# Run (quiet mode)
./bin/benchmark

# Run with verbose output
./bin/benchmark --verbose
```

> **See Also**:
> - [PoC Code & Architecture](../poc/) - Complete implementation details
> - [Production Benchmarks](../benchmarks.md) - Expected production performance
> - [Specification](/specs/000-poc/spec.md) - Original feature spec
> - [Implementation Plan](/specs/000-poc/plan.md) - Design decisions

## Results Summary

```
✓ 62.6% average token reduction (tiktoken-cl100k)
✓ 58.2% average size reduction
✓ 12 operations tested (6 GitHub + 6 Notion)
✓ Target exceeded: 62.6% > 50% required
```

### Per-Operation Breakdown

| Server | Operation          | Token Reduction |
|--------|--------------------|--------------------|
| GitHub | get_repository     | **83.0%**          |
| GitHub | create_issue       | 75.0%              |
| GitHub | search_issues      | 74.9%              |
| GitHub | list_pull_requests | 72.4%              |
| GitHub | get_pull_request   | 67.3%              |
| GitHub | list_commits       | 66.4%              |
| Notion | create_page        | 60.3%              |
| Notion | update_page        | 58.5%              |
| Notion | list_pages         | 55.3%              |
| Notion | search_pages       | 48.6%              |
| Notion | query_database     | 47.5%              |
| Notion | create_database    | 42.1%              |

## How It Works

1. **Loads cached MCP responses** from `testdata/` (no API calls required)
2. **Applies minimal filtering** to remove URLs, IDs, timestamps, metadata
3. **Measures tokens** using three methods:
   - `tiktoken-cl100k` - GPT-4 accurate (primary metric)
   - `char-div-4` - Character count / 4 estimation
   - `word-count` - Whitespace-based word count
4. **Calculates reduction** percentages and displays results

### MCP Protocol Handling

The benchmark correctly handles MCP's protocol structure where actual data is nested as a JSON string inside `content[0].text`:

```json
{
  "content": [
    {
      "type": "text",
      "text": "{\"url\":\"...\",\"title\":\"...\"}"  // ← Embedded JSON
    }
  ]
}
```

The filter parses, filters, and re-serializes this embedded JSON transparently.

## Project Structure

```
docs/poc/                   # PoC implementation (moved from mcp/)
├── benchmark/             # CLI entry point
│   └── main.go
├── pkg/
│   ├── token/            # Tokenizer implementations (tiktoken, word-count, char-div-4)
│   ├── filter/           # Minimal filtering strategy
│   └── cache/            # JSON file loader
├── internal/benchmark/   # Runner, reporter, types, validator
├── testdata/             # Cached MCP responses
│   ├── github/           # 6 GitHub operations
│   └── notion/           # 6 Notion operations
├── go.mod                # Go module definition
├── go.sum                # Dependencies
├── .gitignore           # Git ignore
└── README.md            # Architecture details
```

See [../poc/README.md](../poc/README.md) for detailed architecture and component descriptions.

## Dependencies

- `github.com/mark3labs/mcp-go` v0.42.0-beta.1
- `github.com/pkoukk/tiktoken-go`
- `github.com/olekukonko/tablewriter`

## What This Validates

✅ **Core concept**: Filtering MCP responses significantly reduces tokens
✅ **Target feasibility**: 60-85% reduction achievable with minimal strategy
✅ **Implementation approach**: Go + MCP SDK + tiktoken works well
✅ **Performance**: Sub-second execution for 12 operations

## Next Steps

This PoC proves the concept. The next phase (`001-trim`) will implement the production proxy with:

- Real-time MCP proxying (not cached responses)
- Streaming support
- Smart filtering strategies (not just minimal)
- Prometheus metrics
- Production-grade observability

## Test Data

The `testdata/` directory contains realistic mock responses that simulate GitHub and Notion MCP tool output. These are NOT actual API responses but representative samples for benchmark purposes.
