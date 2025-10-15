# Recording Real MCP Server Responses with govcr

This guide explains how to record real HTTP interactions from Notion and GitHub MCP servers using govcr (Go VCR).

## Overview

SlimMCP uses **govcr** (Go VCR) to record real HTTP interactions with MCP servers for testing. This approach:

- ✅ Uses **real API responses** from actual MCP servers
- ✅ Records responses **once**, then replays them for fast tests
- ✅ **Sanitizes sensitive data** (API keys, tokens) before saving
- ✅ Enables **offline testing** without live API access
- ✅ Ensures **deterministic tests** with consistent data

## Prerequisites

### 1. Get API Credentials

#### Notion API Key
1. Go to https://www.notion.so/my-integrations
2. Click **"+ New integration"**
3. Name it "SlimMCP Test Integration"
4. Copy the **Internal Integration Token** (starts with `secret_`)
5. Create a test database in Notion
6. Click **"Share"** on the database and invite your integration
7. Copy the database ID from the URL: `https://notion.so/YOUR_DATABASE_ID?v=...`

#### GitHub Personal Access Token
1. Go to https://github.com/settings/tokens
2. Click **"Generate new token (classic)"**
3. Name it "SlimMCP Testing"
4. Select scopes: `repo`, `read:org`, `read:user`
5. Click **"Generate token"**
6. Copy the token (starts with `ghp_`)

### 2. Set Up MCP Servers

You need to run the actual MCP servers locally:

```bash
# Install Notion MCP server (example - adjust based on actual server)
npm install -g @modelcontextprotocol/server-notion

# Start Notion MCP server on port 3000
mcp-server-notion --port 3000 --api-key YOUR_NOTION_API_KEY &

# Install GitHub MCP server
npm install -g @modelcontextprotocol/server-github

# Start GitHub MCP server on port 3001
mcp-server-github --port 3001 --token YOUR_GITHUB_TOKEN &
```

## Setup Instructions

### Step 1: Configure Environment

Copy the example env file and fill in your credentials:

```bash
cd mcp/
cp .env.test.example .env.test
```

Edit `.env.test`:

```bash
# Notion Configuration
NOTION_API_KEY=secret_YOUR_ACTUAL_TOKEN_HERE
NOTION_DATABASE_ID=YOUR_DATABASE_ID_HERE

# GitHub Configuration
GITHUB_TOKEN=ghp_YOUR_ACTUAL_TOKEN_HERE
GITHUB_REPO_OWNER=your-username
GITHUB_REPO_NAME=your-test-repo

# MCP Server URLs (adjust ports if needed)
NOTION_MCP_URL=http://localhost:3000
GITHUB_MCP_URL=http://localhost:3001

# Recording mode
VCR_MODE=record
```

### Step 2: Configure Recording Scenarios

Edit `tests/recording-scenarios.json` to customize what gets recorded:

```json
[
  {
    "server": "notion",
    "operation": "query_database",
    "url": "http://localhost:3000",
    "method": "tools/call",
    "params": {
      "name": "query_database",
      "arguments": {
        "database_id": "YOUR_DATABASE_ID_HERE"
      }
    }
  }
]
```

**Replace placeholders:**
- `YOUR_DATABASE_ID_HERE` → Your actual Notion database ID
- `YOUR_GITHUB_USERNAME` → Your GitHub username
- `YOUR_REPO_NAME` → Your test repository name

### Step 3: Record Real Responses

Build and run the recording tool:

```bash
# Build the recording tool
go build -o bin/record-mcp tools/record-mcp/main.go

# Record all scenarios (or use: make record-cassettes)
./bin/record-mcp -env .env.test -config tests/recording-scenarios.json -output tests/cassettes

# Output:
# Recording 10 scenarios to tests/cassettes
# [1/10] Recording notion/query_database
#   SUCCESS: Saved to tests/cassettes/notion_query_database.json
# [2/10] Recording notion/search_pages
#   SUCCESS: Saved to tests/cassettes/notion_search_pages.json
# ...
# Recording complete!
```

### Step 4: Verify Cassettes

Check that cassettes were created:

```bash
ls -lh tests/cassettes/
# notion_query_database.json
# notion_search_pages.json
# github_get_repository.json
# ...
```

Inspect a cassette (sensitive data should be removed):

```bash
cat tests/cassettes/notion_query_database.json | jq '.interactions[0].request.headers'
# Should NOT contain Authorization: Bearer secret_...
```

### Step 5: Use in Tests

Use VCR in your integration tests:

```go
package integration

import (
	"testing"
	"github.com/slimmcp/slimmcp/internal/testutil"
)

func TestRealNotionQuery(t *testing.T) {
	// Load environment (only needed for initial recording)
	testutil.LoadEnvFile("../../.env.test")

	// Create VCR client
	client, cleanup, err := testutil.NewVCRHTTPClient(t, testutil.VCRConfig{
		CassetteName:    "notion_query_database",
		Mode:            testutil.VCRModeAuto, // Auto: replay if exists, else record
		SanitizeHeaders: true,
	})
	if err != nil {
		t.Fatalf("Failed to create VCR client: %v", err)
	}
	defer cleanup()

	// Use client for MCP requests
	// ... your test code here using the VCR client
}
```

## VCR Modes

Set `VCR_MODE` in `.env.test` or pass to recorder:

- **`record`**: Always record new interactions (overwrites existing cassettes)
- **`replay`**: Only replay existing cassettes (fails if not found)
- **`auto`**: Smart mode - replay if cassette exists, otherwise record

## Recording Workflow

### First Time Recording

```bash
# 1. Set mode to record
export VCR_MODE=record

# 2. Start MCP servers
mcp-server-notion --port 3000 --api-key $NOTION_API_KEY &
mcp-server-github --port 3001 --token $GITHUB_TOKEN &

# 3. Record responses
./bin/record-mcp -env .env.test

# 4. Stop servers
killall mcp-server-notion mcp-server-github

# 5. Commit cassettes (sensitive data removed)
git add tests/cassettes/
git commit -m "Add VCR cassettes for integration tests"
```

### Re-recording After API Changes

```bash
# Delete old cassettes
rm tests/cassettes/*.json

# Re-record
export VCR_MODE=record
./bin/record-mcp -env .env.test
```

### Running Tests (Replay Mode)

```bash
# Tests will use recorded cassettes (no live API needed)
export VCR_MODE=replay
go test ./tests/integration/...

# Or use auto mode (recommended)
export VCR_MODE=auto
go test ./tests/integration/...
```

## Security Notes

### What Gets Sanitized

The recorder automatically removes sensitive data before saving:

- ✅ `Authorization` headers
- ✅ `Cookie` headers
- ✅ `Set-Cookie` response headers
- ✅ Notion API version headers

### What You Should NOT Commit

**NEVER commit to git:**
- `.env.test` (contains real API keys)
- Any files with `secret_*` or `ghp_*` tokens

**Safe to commit:**
- `.env.test.example` (template with placeholders)
- `tests/cassettes/*.json` (sanitized recordings)

### Verify Before Committing

Always check cassettes don't contain secrets:

```bash
# Search for potential secrets
grep -r "secret_\|ghp_\|Authorization: Bearer" tests/cassettes/

# Should return nothing!
```

## Troubleshooting

### "NOTION_API_KEY not set"

Copy `.env.test.example` to `.env.test` and fill in your credentials.

### "cassette not found"

Set `VCR_MODE=record` or `VCR_MODE=auto` to record new cassettes.

### "connection refused" on localhost:3000

Start the MCP servers before recording:

```bash
mcp-server-notion --port 3000 --api-key $NOTION_API_KEY
```

### Invalid database_id error

Update `tests/recording-scenarios.json` with your actual database ID from Notion.

## CI/CD Integration

In CI pipelines, always use replay mode (cassettes committed to repo):

```yaml
# .github/workflows/test.yml
- name: Run integration tests
  env:
    VCR_MODE: replay  # Use recorded cassettes only
  run: go test ./tests/integration/...
```

## Related Files

- `testutil/vcr.go` - VCR recorder utilities
- `tools/record-mcp/main.go` - Recording CLI tool
- `tests/recording-scenarios.json` - Scenarios to record
- `.env.test.example` - Environment template
- `tests/cassettes/*.json` - Recorded interactions (committed)
- `.env.test` - Your credentials (NEVER commit)

## Resources

- [govcr Documentation](https://github.com/dnaeon/go-vcr)
- [Notion API Docs](https://developers.notion.com/)
- [GitHub API Docs](https://docs.github.com/en/rest)
- [MCP Protocol Spec](https://spec.modelcontextprotocol.io/)
