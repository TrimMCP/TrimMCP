# Contributing to SlimMCP

Thank you for your interest in contributing to SlimMCP! We welcome contributions from the community.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Commit Message Guidelines](#commit-message-guidelines)

## Code of Conduct

This project adheres to a [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When creating a bug report, include:

- **Clear title and description**
- **Steps to reproduce** the issue
- **Expected vs actual behavior**
- **Environment details** (OS, Go version, SlimMCP version)
- **Relevant logs** or error messages
- **Minimal reproduction** if possible

Use the [bug report template](.github/ISSUE_TEMPLATE/bug_report.yml) when creating an issue.

### Suggesting Features

Feature suggestions are welcome! Please:

- **Check existing feature requests** to avoid duplicates
- **Describe the problem** you're trying to solve
- **Propose a solution** with examples
- **Consider alternatives** you've thought about
- **Explain the benefits** to SlimMCP users

Use the [feature request template](.github/ISSUE_TEMPLATE/feature_request.yml).

### Contributing Code

We love code contributions! Here's how:

1. **Fork the repository**
2. **Create a feature branch** (`git checkout -b feature/amazing-feature`)
3. **Make your changes** following our coding standards
4. **Add tests** for new functionality
5. **Run the test suite** (`make test`)
6. **Commit your changes** following commit guidelines
7. **Push to your fork** (`git push origin feature/amazing-feature`)
8. **Open a Pull Request**

## Development Setup

### Prerequisites

- **Go 1.21+** (Go 1.25+ recommended)
- **Make** (optional, for build automation)
- **Docker** (optional, for containerized testing)
- **Git**

### Clone and Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/slimmcp.git
cd slimmcp

# Install dependencies
go mod download

# Run tests to verify setup
go test ./...

# Build the binary
go build -o slim cmd/slimmcp/main.go

# Run locally
./slim start --config examples/dev.yaml
```

### Project Structure

```
slimmcp/
├── cmd/slimmcp/        # CLI entry point
├── pkg/                # Public packages
│   ├── proxy/         # HTTP proxy server
│   ├── filter/        # Response filtering
│   ├── analyzer/      # Field inference
│   ├── config/        # Configuration management
│   ├── metrics/       # Prometheus metrics
│   └── ...
├── tests/             # Integration and acceptance tests
├── examples/          # Configuration examples
└── docs/              # Additional documentation
```

## Pull Request Process

### Before Submitting

1. ✅ **Tests pass** - Run `go test ./...`
2. ✅ **Linter passes** - Run `golangci-lint run`
3. ✅ **Code formatted** - Run `gofmt -s -w .`
4. ✅ **Documentation updated** - Update relevant docs
5. ✅ **CHANGELOG updated** - Add entry to [CHANGELOG.md](CHANGELOG.md)
6. ✅ **Commits follow guidelines** - See below

### PR Requirements

- **Descriptive title** - Clearly describe what the PR does
- **Linked issue** - Reference related issues (Fixes #123)
- **Clear description** - Explain what changed and why
- **Tests included** - Add tests for new functionality
- **Documentation** - Update docs if behavior changes
- **No breaking changes** - Or clearly marked if unavoidable

### Review Process

1. **Automated checks** - CI/CD must pass
2. **Code review** - Maintainer review required
3. **Approval** - At least one maintainer approval
4. **Merge** - Squash and merge to main

## Coding Standards

### Go Style

Follow standard Go conventions:

- **gofmt** - All code must be formatted
- **golangci-lint** - Linter must pass with zero warnings
- **go vet** - No vet warnings allowed
- **Comments** - Exported functions/types must have doc comments
- **Error handling** - Always handle errors, never ignore

### Package Guidelines

```go
// Package proxy implements an HTTP proxy server for MCP protocol.
//
// The proxy intercepts MCP requests, applies filtering strategies,
// and returns optimized responses while maintaining protocol compatibility.
package proxy

// Handler processes incoming MCP requests and applies filtering.
//
// It implements the http.Handler interface and can be used as a
// standard HTTP handler in any Go HTTP server.
type Handler struct {
    // ...
}

// NewHandler creates a new proxy handler with the given configuration.
//
// Example:
//
//	cfg := config.Load("config.yaml")
//	handler := proxy.NewHandler(cfg)
//	http.ListenAndServe(":8080", handler)
func NewHandler(cfg *config.Config) *Handler {
    // ...
}
```

### Naming Conventions

- **Packages** - Short, lowercase, single word (e.g., `proxy`, `filter`)
- **Interfaces** - Verb or noun ending in -er (e.g., `Analyzer`, `Filter`)
- **Structs** - Descriptive noun (e.g., `ProxyHandler`, `FilterEngine`)
- **Functions** - Descriptive verb (e.g., `ProcessRequest`, `ApplyFilter`)

## Testing Guidelines

### Test Coverage

- **Unit tests** - All public functions in `pkg/`
- **Integration tests** - End-to-end workflows in `tests/integration/`
- **Acceptance tests** - User story validation in `tests/acceptance/`
- **Target coverage** - Aim for 80%+ on new code

### Writing Tests

```go
func TestProxyHandler_BasicFiltering(t *testing.T) {
    // Arrange
    cfg := &config.Config{
        Strategy: config.StrategyMinimal,
    }
    handler := NewHandler(cfg)

    // Act
    resp := handler.ProcessRequest(testRequest)

    // Assert
    assert.NotNil(t, resp)
    assert.Equal(t, 200, resp.StatusCode)
    assert.Less(t, resp.TokenCount, testRequest.TokenCount)
}
```

### Table-Driven Tests

Prefer table-driven tests for multiple scenarios:

```go
func TestAnalyzer_InferFields(t *testing.T) {
    tests := []struct {
        name     string
        request  Request
        expected []string
    }{
        {
            name:     "notion page fetch",
            request:  notionFetchRequest,
            expected: []string{"id", "title", "url"},
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := analyzer.InferFields(tt.request)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

## Commit Message Guidelines

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- **feat** - New feature
- **fix** - Bug fix
- **docs** - Documentation changes
- **style** - Code style changes (formatting, etc.)
- **refactor** - Code refactoring
- **test** - Adding or updating tests
- **chore** - Maintenance tasks
- **perf** - Performance improvements

### Examples

```
feat(filter): add nested field selection support

Implement dot notation for selecting nested fields in responses.
Enables users to specify "properties.Status" to filter nested objects.

Closes #123
```

```
fix(proxy): handle timeout errors gracefully

Previously, timeout errors would crash the proxy. Now they return
a proper error response to the client.

Fixes #456
```

### Guidelines

- **Use imperative mood** - "Add feature" not "Added feature"
- **Limit first line** - 50 characters max
- **Body is optional** - But encouraged for complex changes
- **Reference issues** - Use "Fixes #123" or "Closes #456"

## Community

- 💬 [Discord](https://discord.gg/slimmcp) - Chat with the community
- 🐛 [Issues](https://github.com/slimmcp/slimmcp/issues) - Report bugs
- 💡 [Discussions](https://github.com/slimmcp/slimmcp/discussions) - Ask questions

## License

By contributing to SlimMCP, you agree that your contributions will be licensed under the [GNU AGPL-3.0 License](LICENSE).

---

**Thank you for contributing to SlimMCP!** 🎉
