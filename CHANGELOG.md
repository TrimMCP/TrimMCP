# Changelog - SlimMCP Monorepo

All notable changes to the SlimMCP development monorepo will be documented in this file.

This changelog tracks changes to the **monorepo infrastructure**, documentation, and specifications. For SlimMCP product releases, see [mcp/CHANGELOG.md](mcp/CHANGELOG.md).

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Monorepo
- Initial monorepo structure
- SpecKit integration for feature specifications
- CI/CD workflows for automated deployment

## [0.1.0] - 2025-10-15

### SlimMCP Product
- See [mcp/CHANGELOG.md](mcp/CHANGELOG.md) for v0.1.0 release notes

### Monorepo
- Initial monorepo setup
- Documentation structure in `docs/`
- Specification workflow in `specs/001-trim/`
- Development guidelines in `CLAUDE.md`, `AGENTS.md`, `.specify/memory/constitution.md`

### Documentation
- Added getting-started.md
- Added configuration.md
- Added architecture.md
- Added api.md
- Added protocol-extension.md
- Added benchmarks.md
- Added troubleshooting.md
- Added faq.md
- Added roadmap.md
- Added tech-stack.md

### Infrastructure
- GitHub Actions workflows for release automation
- Docker build configuration
- Multi-platform binary builds
- Documentation deployment pipeline

---

## Versioning Strategy

This monorepo uses a **hybrid versioning approach**:

### Product Versions (SlimMCP)
Product releases are tagged with semantic versioning:
- `v0.1.0-alpha.1`, `v0.1.0-alpha.2` - Alpha releases
- `v0.1.0-beta.1`, `v0.1.0-beta.2` - Beta releases
- `v0.1.0-rc.1`, `v0.1.0-rc.2` - Release candidates
- `v0.1.0` - Stable release

### Monorepo Versions
The monorepo may have additional tags for infrastructure/docs-only changes:
- `monorepo-v1.0.0` - Major monorepo infrastructure update
- Not typically tagged - changes tracked in this CHANGELOG

[Unreleased]: https://github.com/IdeaCraftersHQ/oss-mcp-slim/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/IdeaCraftersHQ/oss-mcp-slim/releases/tag/v0.1.0
