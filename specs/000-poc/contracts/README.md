# Contracts: Token Savings Benchmark PoC

**Feature**: 000-poc | **Date**: 2025-10-14

## Overview

This directory contains interface contracts and validation schemas for the benchmark PoC. Since this is a CLI tool (not an API), contracts define Go interfaces and JSON schemas for data validation.

---

## Go Interfaces

### Tokenizer Interface

**Purpose**: Abstract token counting implementations to support multiple tokenization methods.

**File**: `mcp/pkg/token/tokenizer.go`

```go
package token

// Tokenizer counts tokens in text using a specific tokenization method.
type Tokenizer interface {
    // Count returns the number of tokens in the given text.
    Count(text string) (int, error)

    // Name returns the human-readable name of this tokenization method.
    // Examples: "tiktoken-cl100k", "word-count", "char-div-4"
    Name() string
}
```

**Implementations**:
- `TiktokenTokenizer` - GPT-4 cl100k_base encoding
- `WordCountTokenizer` - Simple whitespace-based word counting
- `CharDivFourTokenizer` - Character count divided by 4

**Contract Requirements**:
- `Count()` must return non-negative values
- `Count()` must be deterministic (same input = same output)
- `Name()` must return unique identifier for the method
- Implementations must be safe for concurrent use

---

### Filter Interface

**Purpose**: Abstract response filtering logic to support different strategies.

**File**: `mcp/pkg/filter/filter.go`

```go
package filter

// Filter transforms MCP responses by removing unnecessary data.
type Filter interface {
    // Apply filters the given response, returning a new filtered response.
    // Original response is not modified (immutable).
    Apply(response map[string]interface{}) (map[string]interface{}, error)

    // Name returns the human-readable name of this filtering strategy.
    // Example: "minimal"
    Name() string
}
```

**Implementations**:
- `MinimalFilter` - Removes metadata, URLs, timestamps

**Contract Requirements**:
- `Apply()` must not modify the input (immutable)
- `Apply()` must return valid JSON-serializable output
- `Apply()` must preserve essential data (validated by schema)
- Implementations must be safe for concurrent use

---

### CacheLoader Interface

**Purpose**: Abstract cache file loading to support different storage formats.

**File**: `mcp/pkg/cache/loader.go`

```go
package cache

import "time"

// CachedResponse represents a stored MCP response.
type CachedResponse struct {
    Operation   string
    Server      string
    Request     map[string]interface{}
    Response    map[string]interface{}
    CapturedAt  time.Time
}

// CacheLoader loads cached MCP responses from storage.
type CacheLoader interface {
    // Load retrieves a cached response for the given server and operation.
    // Returns error if cache file missing or corrupted.
    Load(server, operation string) (*CachedResponse, error)

    // Exists checks if a cached response exists without loading it.
    Exists(server, operation string) bool
}
```

**Implementations**:
- `JSONFileLoader` - Loads from testdata/{server}/{operation}.json

**Contract Requirements**:
- `Load()` must return error if file missing (not nil response)
- `Load()` must validate JSON structure before returning
- `Exists()` must be fast (no full file parsing)
- Implementations must handle file system errors gracefully

---

### Validator Interface

**Purpose**: Validate filtered responses maintain data integrity.

**File**: `internal/benchmark/validator.go`

```go
package benchmark

// Validator checks filtered responses for data integrity.
type Validator interface {
    // Validate checks if the filtered response meets integrity requirements.
    // Returns error describing any validation failures.
    Validate(operation string, filtered map[string]interface{}) error

    // RegisterSchema adds a validation schema for a specific operation.
    RegisterSchema(operation string, schema Schema)
}

// Schema defines required fields and types for a response.
type Schema interface {
    // CheckRequiredFields verifies all required fields are present.
    CheckRequiredFields(data map[string]interface{}) error

    // CheckTypes verifies field values have correct types.
    CheckTypes(data map[string]interface{}) error
}
```

**Contract Requirements**:
- `Validate()` must return nil if validation passes
- `Validate()` must return descriptive error on failure
- Schemas must be registered before validation
- Missing schema for operation is an error

---

## JSON Schemas

### Cached Response Schema

**Purpose**: Define the structure of cache files in testdata/

**File**: `contracts/cached-response-schema.json`

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Cached MCP Response",
  "description": "Schema for cached MCP responses stored in testdata/",
  "type": "object",
  "required": ["operation", "server", "request", "response", "captured_at"],
  "properties": {
    "operation": {
      "type": "string",
      "description": "MCP tool operation name",
      "enum": [
        "list_pull_requests", "search_issues", "get_repository",
        "create_issue", "list_commits", "get_pull_request",
        "create_page", "query_database", "list_pages",
        "update_page", "create_database", "search_pages"
      ]
    },
    "server": {
      "type": "string",
      "description": "MCP server name",
      "enum": ["github", "notion"]
    },
    "request": {
      "type": "object",
      "description": "Original MCP request that produced this response",
      "required": ["method", "params"],
      "properties": {
        "method": {"type": "string"},
        "params": {"type": "object"}
      }
    },
    "response": {
      "type": "object",
      "description": "Full MCP response body"
    },
    "captured_at": {
      "type": "string",
      "format": "date-time",
      "description": "ISO 8601 timestamp when response was recorded"
    }
  }
}
```

---

### GitHub Response Schemas

**Purpose**: Define required fields for GitHub MCP tool responses after filtering.

**Example**: `contracts/github-list-prs-schema.json`

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "GitHub list_pull_requests Response",
  "description": "Required fields after filtering",
  "type": "object",
  "required": ["content"],
  "properties": {
    "content": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["type", "text"],
        "properties": {
          "type": {"const": "text"},
          "text": {
            "type": "string",
            "description": "JSON string containing PR array"
          }
        }
      }
    }
  }
}
```

**Essential PR Fields** (parsed from `content[0].text`):
- `number` (int) - PR number
- `title` (string) - PR title
- `state` (string) - open/closed
- `user.login` (string) - Author username

**Removed Fields** (should NOT be present after filtering):
- `url`, `html_url`, `api_url` - URLs
- `node_id` - Internal ID
- `created_at`, `updated_at` - Timestamps

---

### Notion Response Schemas

**Purpose**: Define required fields for Notion MCP tool responses after filtering.

**Example**: `contracts/notion-create-page-schema.json`

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Notion create_page Response",
  "description": "Required fields after filtering",
  "type": "object",
  "required": ["content"],
  "properties": {
    "content": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["type", "text"],
        "properties": {
          "type": {"const": "text"},
          "text": {
            "type": "string",
            "description": "JSON string containing page object"
          }
        }
      }
    }
  }
}
```

**Essential Page Fields** (parsed from `content[0].text`):
- `properties` (object) - Page properties
- `title` or page title property - Page title

**Removed Fields** (should NOT be present after filtering):
- `object` - Type identifier ("page")
- `id` - Page ID
- `created_time`, `last_edited_time` - Timestamps
- `parent` - Parent reference

---

## Contract Testing

### Test Requirements

Each interface implementation must have:
1. **Unit tests**: Test each method independently
2. **Contract tests**: Verify implementation satisfies interface contract
3. **Integration tests**: Test with real cache files

### Example Contract Test

```go
// pkg/token/contract_test.go
func TestTokenizerContract(t *testing.T) {
    tokenizers := []Tokenizer{
        NewTiktoken(),
        NewWordCount(),
        NewCharDivFour(),
    }

    for _, tok := range tokenizers {
        t.Run(tok.Name(), func(t *testing.T) {
            // Contract: Count must return non-negative
            count, err := tok.Count("hello world")
            if err != nil {
                t.Fatalf("Count failed: %v", err)
            }
            if count < 0 {
                t.Errorf("Count returned negative value: %d", count)
            }

            // Contract: Count must be deterministic
            count2, _ := tok.Count("hello world")
            if count != count2 {
                t.Errorf("Count not deterministic: %d != %d", count, count2)
            }

            // Contract: Name must be unique
            if tok.Name() == "" {
                t.Error("Name returned empty string")
            }
        })
    }
}
```

---

## Summary

- **4 Go interfaces**: Tokenizer, Filter, CacheLoader, Validator
- **3+ JSON schemas**: Cached response, GitHub responses, Notion responses
- **Contract tests**: Required for all implementations
- **Validation**: Ensures filtered responses maintain data integrity

All contracts support the test-first development principle (Constitution II).
