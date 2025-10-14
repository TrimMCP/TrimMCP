<!--
====================================================================================
SYNC IMPACT REPORT - Constitution Update
====================================================================================

Version Change: 1.1.0 → 1.2.0 (Monorepo structure clarification)
Date: 2025-10-14

Summary:
Updated Repository Architecture section to clarify that the monorepo is specifically
for SlimMCP development (not a multi-project OSS MCP umbrella). Aligned with the
updated monorepo README.md that positions this as SlimMCP-specific development hub.

Changes Made:
- Updated monorepo structure diagram to show all SlimMCP assets (mcp/, docs/, specs/, drafts/)
- Clarified that docs/ deploys to slimmcp/docs (not mcp/docs/)
- Enhanced deployment strategy section with specific triggers and purposes
- Updated repository governance to emphasize SlimMCP-specific focus
- Added contribution flow for external contributors

Previous Version Summary (1.0.0 → 1.1.0):
This is the initial ratification of the SlimMCP Constitution, replacing the generic
template with project-specific principles derived from spec.md and plan.md.

Principles Established:
- I. Simplicity First (derived from plan constitution check)
- II. Test-First Development (NON-NEGOTIABLE) (from spec user stories structure)
- III. Production-Ready Observability (from spec SC-003, SC-006, SC-007, SC-013)
- IV. Stateless Architecture (from spec SC-009, plan architecture)
- V. Fail-Safe Defaults (from spec FR-018, edge cases)
- VI. Interface-Driven Design (from plan "Library-First Architecture")
- VII. Zero Breaking Changes (from spec assumptions, MCP compatibility)

Sections Added:
- Performance Standards (from spec success criteria SC-001 to SC-013)
  * Latency Targets (SC-003)
  * Throughput Requirements (SC-004)
  * Resource Constraints (plan technical context)
  * Reliability Standards (SC-013, SC-011, SC-010)
- Development Workflow (from plan design principles)
  * Code Quality Gates (plan test-first principle)
  * Commit Standards (conventional commits)
  * Review Requirements (constitution compliance)
  * Testing Requirements (from quickstart.md test structure)
- Governance (enforcement and amendment process)
  * Amendment Process
  * Compliance Enforcement
  * Complexity Justification (links to plan.md Complexity Tracking)

Templates Requiring Updates:
✅ plan-template.md - Constitution Check section references this file (no changes needed)
✅ spec-template.md - No constitution references (no changes needed)
✅ tasks-template.md - No constitution references (no changes needed)
✅ speckit.plan.md - References constitution.md for gates (no changes needed)
✅ speckit.analyze.md - References constitution.md for validation (no changes needed)

Follow-up TODOs:
None - all placeholders filled with concrete values.

Rationale for Version 1.0.0:
This is the initial ratification of project-specific constitution principles.
Starting at 1.0.0 to indicate it is production-ready and complete.

====================================================================================
-->

# SlimMCP Constitution

## Core Principles

### I. Simplicity First
Start with minimal viable implementations before adding complexity. Favor standard library solutions over third-party dependencies. Each component should have a single, clear responsibility. Complex features must justify their necessity against simpler alternatives.

### II. Test-First Development (NON-NEGOTIABLE)
All features begin with acceptance tests written from user stories. Test structure follows the Red-Green-Refactor cycle strictly: Write test → Verify it fails → Implement → Verify it passes → Refactor. Unit tests must exist before implementation code. Integration tests verify component interactions. Performance benchmarks validate latency and throughput targets.

### III. Production-Ready Observability
Monitoring, metrics, and logging are first-class concerns, not afterthoughts. Every component exposes Prometheus metrics. Structured logging with appropriate levels (debug, info, warn, error). Health checks accurately reflect system state. Request tracing enables debugging production issues. Performance measurement is built-in, not bolted-on.

### IV. Stateless Architecture
Horizontal scalability through zero shared state. All configuration comes from files or environment. No persistent storage dependencies. Each request is independent. Cache invalidation strategies are explicit and safe. This enables load balancing across multiple instances without session affinity.

### V. Fail-Safe Defaults
When in doubt, preserve safety over optimization. If filtering fails, return full unfiltered response (configurable). Authentication passes through unchanged. Schema validation failures fall back to permissive behavior. Circuit breakers prevent cascade failures. Timeout defaults prevent hung requests.

### VI. Interface-Driven Design
Define clear contracts between components. Core logic exists as reusable packages independent of HTTP layer. All public APIs have corresponding tests. Interfaces enable testing with mocks. Components depend on abstractions, not concrete implementations.

### VII. Zero Breaking Changes
Maintain backward compatibility with existing MCP clients and servers. Protocol extensions use optional `_meta` fields. Configuration changes support deprecated options during transition periods. Version new features additively. Document migration paths explicitly.

## Performance Standards

### Latency Targets
- **Minimal strategy**: <10ms p95 filtering latency
- **Smart strategy**: <15ms p95 filtering latency
- **Total request**: <200ms p95 end-to-end (including downstream)
- **Health check**: <100ms response time
- **Metrics collection**: <1ms overhead per request

### Throughput Requirements
- **Single instance**: 500+ requests/second (2 vCPU, 4GB RAM)
- **Horizontal scale**: Linear scaling to 1,500+ req/sec across 3+ instances
- **Concurrent requests**: Handle 100+ simultaneous connections

### Resource Constraints
- **Memory**: Must handle 10MB response objects without excessive allocation
- **CPU**: Filtering should not exceed 10% CPU per request
- **Network**: Support gzip compression for bandwidth efficiency

### Reliability Standards
- **Availability**: 99.9% uptime (three nines) - maximum 43 minutes downtime per month
- **Success rate**: 95% of requests succeed on first attempt (excluding downstream errors)
- **Graceful degradation**: Fallback to full responses on filtering errors
- **Configuration reload**: Complete within 5 seconds without dropping active requests

## Development Workflow

### Code Quality Gates
1. **All code must have tests**: No untested code merges to main branch
2. **Tests must pass**: CI pipeline blocks failing tests
3. **Linting must pass**: `golangci-lint` with project configuration
4. **Benchmarks must not regress**: Performance within 10% of baseline
5. **Documentation required**: All public APIs documented with examples

### Commit Standards
- Commit messages follow conventional commits format (feat:, fix:, docs:, etc.)
- Each commit represents atomic, working change
- No commits breaking CI pipeline
- Merge commits via squash for feature branches

### Review Requirements
- All changes require peer review before merge
- Reviewers verify constitution compliance
- Performance-critical code requires benchmark review
- Breaking changes require architecture review and migration plan

### Testing Requirements
1. **Unit tests**: All packages under `mcp/pkg/` have >80% coverage
2. **Integration tests**: End-to-end scenarios for each user story
3. **Contract tests**: MCP protocol compliance validation
4. **Acceptance tests**: One test per acceptance scenario in spec
5. **Performance tests**: Benchmarks for critical path operations

## Governance

This constitution supersedes all other development practices and standards. When conflicts arise between this document and other guidelines, the constitution takes precedence.

### Amendment Process
1. Propose change with detailed rationale and impact analysis
2. Document affected code, tests, and documentation
3. Gain approval from maintainers
4. Update constitution version number
5. Create migration plan for non-compliant code
6. Execute migration before amendment takes effect

### Compliance Enforcement
- All pull requests must verify constitution compliance
- CI pipeline includes constitution checks (test coverage, linting, benchmarks)
- Reviewers explicitly confirm compliance in approval comments
- Violations block merge until resolved

### Complexity Justification
Any design decision that violates simplicity principles (Principle I) must be documented in `plan.md` under "Complexity Tracking" with:
- Description of the complexity being introduced
- Why simpler alternatives are insufficient
- Long-term maintenance implications
- Plan to reduce complexity in future

## Repository Architecture

### Monorepo Structure

SlimMCP uses a monorepo approach for unified development with automated deployment to separate production repositories:

**Development Monorepo**: `IdeaCraftersHQ/oss-mcp-slim`
```
oss-mcp-slim/
├── mcp/                    # SlimMCP proxy implementation (deployed to slimmcp/slimmcp)
│   ├── cmd/
│   ├── pkg/
│   ├── internal/
│   ├── tests/
│   ├── examples/
│   ├── docs/
│   ├── go.mod
│   ├── README.md
│   └── LICENSE (AGPL-3.0)
├── docs/                   # Documentation site content (deployed to slimmcp/docs)
│   ├── getting-started.md
│   ├── configuration.md
│   ├── architecture.md
│   ├── api.md
│   └── ...
├── specs/                  # Feature specifications (SpecKit, stays in monorepo)
├── drafts/                 # Work-in-progress ideas (stays in monorepo)
├── .specify/               # SpecKit workflow automation (stays in monorepo)
├── .claude/                # Claude Code configuration (stays in monorepo)
├── .codex/                 # Codex configuration (stays in monorepo)
├── .github/workflows/      # CI/CD automation
├── CLAUDE.md               # Development guidelines
└── README.md               # Monorepo overview
```

### Automated Deployment Strategy

GitHub Actions automatically deploy SlimMCP components to separate production repositories:

1. **SlimMCP Proxy**: `slimmcp/slimmcp`
   - Source: `mcp/` directory from monorepo
   - Trigger: Push to main, version tags
   - Contains: Go source code, tests, examples, README, LICENSE
   - Release artifacts: Binaries (Linux, macOS, Windows), Docker images
   - Purpose: Public distribution and consumption

2. **Documentation Site**: `slimmcp/docs`
   - Source: `docs/` directory from monorepo
   - Trigger: Push to main, documentation changes
   - Built with: Static site generator (deployed to slimmcp.com/docs)
   - Purpose: User-facing documentation

3. **Website**: Cloudflare Pages (`slimmcp.com`)
   - Source: Built from `slimmcp/docs` repository
   - Trigger: Automatic on push to `slimmcp/docs`
   - Features: Marketing site, documentation, blog
   - CDN: Cloudflare edge network with SSL

### Build Tools

Consider **Turborepo** for monorepo build orchestration:
- Incremental builds (only rebuild changed packages)
- Caching for faster CI/CD
- Task dependencies and parallelization
- Works well with GitHub Actions

Alternative: **Nx**, **Bazel** (more complex), or custom scripts

### Path Conventions

- **Development**: All work happens in `mcp/` within monorepo
- **Documentation references**: Use paths relative to `mcp/` directory
- **CI/CD**: Scripts handle subdirectory extraction and deployment
- **Testing**: Run tests from `mcp/` context

### Deployment Workflow

```yaml
# Example .github/workflows/deploy-mcp.yml
name: Deploy SlimMCP
on:
  push:
    branches: [main]
    paths: ['mcp/**']

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Deploy to slimmcp/slimmcp
        # Extract mcp/ and push to separate repo
```

### Repository Governance

- **Development Monorepo** (`IdeaCraftersHQ/oss-mcp-slim`):
  - Single source of truth for all SlimMCP development
  - Feature planning and specifications in `specs/`
  - Development tooling (SpecKit, Claude Code, Codex)
  - All development work happens here
  - Private repository (or public based on preference)

- **Production Repositories** (`slimmcp/*`):
  - Auto-generated from monorepo via GitHub Actions
  - No direct commits (read-only from user perspective)
  - Tagged releases for versioning
  - Public repositories for distribution
  - User contributions via PRs (synced back to monorepo)

- **Contribution Flow**:
  - Core team: Work directly in monorepo
  - External contributors: Submit PRs to `slimmcp/slimmcp` or `slimmcp/docs`
  - PR acceptance: Changes synced back to monorepo manually by maintainers

**Version**: 1.2.0 | **Ratified**: 2025-10-13 | **Last Amended**: 2025-10-14
