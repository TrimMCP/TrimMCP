# SlimMCP Development Guidelines

**Monorepo**: SlimMCP development monorepo (IdeaCraftersHQ/oss-mcp-slim)
**Project**: Transparent MCP proxy for 60-85% token reduction
**Last Updated**: 2025-10-14

## Monorepo Structure

This is the **SlimMCP-specific development monorepo**, not a multi-project OSS MCP umbrella.

```
oss-mcp-slim/                    # Monorepo root
├── mcp/                         # SlimMCP proxy (Go) → deploys to slimmcp/slimmcp
│   ├── cmd/slimmcp/            # CLI entry point
│   ├── pkg/                    # Public packages (proxy, analyzer, filter, etc.)
│   ├── internal/               # Private packages (testutil)
│   ├── tests/                  # Integration, contract, acceptance tests
│   ├── examples/               # Configuration examples
│   ├── docs/                   # MCP-specific docs and images
│   ├── go.mod                  # Go module definition
│   ├── README.md               # Main project README (deployed)
│   └── LICENSE                 # AGPL-3.0
│
├── docs/                        # Documentation site → deploys to slimmcp/docs
│   ├── getting-started.md
│   ├── configuration.md
│   ├── architecture.md
│   ├── api.md
│   ├── protocol-extension.md
│   └── ...
│
├── specs/                       # Feature specifications (SpecKit, stays in monorepo)
│   └── 001-docs-architecture-md/
│       ├── spec.md
│       ├── plan.md
│       ├── tasks.md
│       └── ...
│
├── drafts/                      # WIP ideas (stays in monorepo)
├── .specify/                    # SpecKit automation (stays in monorepo)
├── .claude/                     # Claude Code config (stays in monorepo)
├── .codex/                      # Codex config (stays in monorepo)
├── CLAUDE.md                    # This file
└── README.md                    # Monorepo overview
```

## Active Technologies

- **Go 1.21+** (minimum 1.21, target 1.25+)
- **MCP Go SDK**: github.com/mark3labs/mcp-go v0.42.0-beta.1
- **Prometheus**: github.com/prometheus/client_golang v1.23.2
- **YAML Parser**: github.com/goccy/go-yaml v1.18.0

## Development Workflow

### For Core Team (Working in Monorepo)

```bash
# Clone monorepo
git clone https://github.com/IdeaCraftersHQ/oss-mcp-slim.git
cd oss-mcp-slim

# Work in mcp/ directory
cd mcp/
go mod download
go test ./...
go build -o slim cmd/slimmcp/main.go
```

### Deployment

All deployments are automated via GitHub Actions:
- `mcp/` → `slimmcp/slimmcp` (on push to main, version tags)
- `docs/` → `slimmcp/docs` → slimmcp.com/docs (on documentation changes)

## Code Style

### Go Standards
- Follow standard Go conventions (gofmt, golangci-lint)
- Table-driven tests for all packages
- Interfaces in separate files (e.g., `analyzer.go` defines `Analyzer` interface)
- Package-level documentation in `doc.go` files

### Project Conventions
- **Stateless architecture**: No persistent storage, horizontally scalable
- **Library-first**: Core logic in `pkg/` packages independent of HTTP layer
- **Test-first development**: Acceptance tests before implementation
- **Observability built-in**: Prometheus metrics, structured logging from start

### File Organization
- `pkg/`: Public, reusable packages
- `internal/`: Private packages (not importable by external projects)
- `cmd/`: CLI entry points
- `tests/`: Integration, contract, and acceptance tests (separate from unit tests)

## Common Commands

```bash
# Build
cd mcp/
go build -o slim cmd/slimmcp/main.go

# Test (all)
go test ./...

# Test with coverage
go test -cover ./...

# Run linter
golangci-lint run

# Run proxy locally
./slim start --config examples/dev.yaml

# Run benchmarks
go test -bench=. ./pkg/filter/
```

## Constitution Principles

See `.specify/memory/constitution.md` for full details. Key principles:

1. **Simplicity First**: Start minimal before adding complexity
2. **Test-First Development**: Write tests before implementation (NON-NEGOTIABLE)
3. **Production-Ready Observability**: Metrics and logging are first-class concerns
4. **Stateless Architecture**: Zero shared state, horizontally scalable
5. **Fail-Safe Defaults**: When in doubt, preserve safety over optimization
6. **Interface-Driven Design**: Clear contracts between components
7. **Zero Breaking Changes**: Maintain MCP protocol compatibility

### Performance Targets

- <10ms p95 latency (minimal strategy)
- <15ms p95 latency (smart strategy)
- 500+ req/sec per instance (2 vCPU, 4GB RAM)
- 99.9% availability (three nines)

## Recent Changes

- 2025-10-14: Updated CLAUDE.md for SlimMCP-specific monorepo structure
- 2025-10-13: Added Go 1.21+ with MCP SDK dependencies
- 2025-10-13: Initial feature spec (001-docs-architecture-md)

## SpecKit Workflow

Feature planning uses SpecKit commands:
- `/speckit.constitution` - Define project principles
- `/speckit.specify` - Create feature specification
- `/speckit.plan` - Generate implementation plan
- `/speckit.tasks` - Break down into tasks
- `/speckit.implement` - Execute implementation

All specs in `specs/` directory, never committed to production repos.

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->