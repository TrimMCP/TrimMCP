# SlimMCP Development Monorepo

This is the development monorepo for **SlimMCP** - a transparent MCP proxy that reduces response sizes by 60-85% while maintaining protocol compatibility. Sponsored by [IdeaCraftersHQ](https://github.com/IdeaCraftersHQ).

## Purpose

This monorepo serves as the **single source of truth** for all SlimMCP development, planning, and documentation. The actual production code, documentation site, and web application are automatically built and deployed to dedicated repositories in the [SlimMCP organization](https://github.com/slimmcp) through GitHub Actions CI/CD pipelines.

## Why a Monorepo?

The monorepo approach provides several advantages for SlimMCP development:

- **Unified Development Environment** - All specs, drafts, source code, docs, and web assets in one place
- **Atomic Changes** - Make changes across MCP code, documentation, and website in a single commit
- **Shared Tooling** - Common build scripts, CI/CD workflows, and development tools
- **Version Alignment** - Keep SlimMCP proxy, docs, and website perfectly synchronized
- **Simplified Planning** - Centralized project planning with SpecKit feature specifications

## Repository Structure

```
.
├── mcp/                    # SlimMCP proxy source code
│   ├── README.md          # Project README (deployed to slimmcp/slimmcp)
│   ├── LICENSE            # AGPL-3.0 license
│   ├── examples/          # Configuration examples
│   └── docs/              # Additional documentation
│
├── docs/                  # SlimMCP documentation site
│   ├── README.md          # Documentation index
│   ├── getting-started.md # Quick start guides
│   ├── configuration.md   # Configuration reference
│   ├── architecture.md    # Technical architecture
│   ├── api.md            # API documentation
│   ├── benchmarks.md     # Performance benchmarks
│   ├── protocol-extension.md  # MCP protocol extension spec
│   ├── roadmap.md        # Project roadmap
│   ├── tech-stack.md     # Technology decisions
│   ├── troubleshooting.md # Common issues and solutions
│   ├── faq.md            # Frequently asked questions
│   └── commercial-hosted.md  # Commercial offering info
│
├── specs/                 # Feature specifications
│   └── 001-docs-architecture-md/  # Example feature spec
│       ├── spec.md       # Feature specification
│       ├── plan.md       # Implementation plan
│       ├── tasks.md      # Task breakdown
│       ├── research.md   # Research notes
│       ├── data-model.md # Data models
│       ├── dependencies.md # Dependencies
│       ├── quickstart.md # Quick start guide
│       ├── PERFORMANCE-TARGETS.md # Performance goals
│       ├── checklists/   # Quality checklists
│       └── contracts/    # API contracts and schemas
│
├── drafts/                # Draft documents and ideas
│   └── dynamic-mcp-requests.md
│
├── .specify/              # SpecKit workflow automation
│   └── scripts/          # Automation scripts
│
├── .claude/               # Claude Code configuration
├── .codex/                # Codex configuration
├── CLAUDE.md             # Development guidelines
└── README.md             # This file
```

## SlimMCP Assets

This monorepo contains all SlimMCP-related assets:

- **MCP Proxy (`mcp/`)** - Go source code for the SlimMCP proxy
- **Documentation (`docs/`)** - Technical documentation, guides, and API reference
- **Specifications (`specs/`)** - Feature specs and implementation plans
- **Drafts (`drafts/`)** - Work-in-progress ideas and proposals

### Production Deployments

- **Production Code:** [slimmcp/slimmcp](https://github.com/slimmcp/slimmcp)
- **Documentation Site:** [slimmcp/docs](https://github.com/slimmcp/docs)
- **Website:** [slimmcp.com](https://slimmcp.com) (Cloudflare Pages)
- **License:** GNU AGPL-3.0

## Build & Deployment Workflow

### Automated Deployment

All deployments are handled automatically through GitHub Actions:

```
┌─────────────────────────────────────┐
│  IdeaCraftersHQ/oss-mcp-slim        │
│  (This Monorepo)                    │
└──────────────┬──────────────────────┘
               │
               │ Push to main
               ↓
    ┌──────────────────────────┐
    │   GitHub Actions         │
    │   - Build Go binaries    │
    │   - Run tests            │
    │   - Generate docs        │
    │   - Create releases      │
    └──────────┬───────────────┘
               │
               ├──────────────────────┐
               │                      │
               ↓                      ↓
    ┌──────────────────┐   ┌─────────────────┐
    │  slimmcp/slimmcp │   │  slimmcp/docs   │
    │  (Binary, Code)  │   │  (Documentation)│
    └──────────────────┘   └────────┬────────┘
                                    │
                                    ↓
                           ┌─────────────────┐
                           │ Cloudflare Pages│
                           │   slimmcp.io    │
                           └─────────────────┘
```

### Deployment Targets

1. **Source Code Repository** (`slimmcp/slimmcp`)
   - Go source code from `mcp/`
   - README and LICENSE
   - Release binaries (Linux, macOS, Windows)
   - Docker images
   - Triggered on: Push to `main`, version tags

2. **Documentation Repository** (`slimmcp/docs`)
   - Markdown documentation from `docs/`
   - Built with static site generator
   - Versioned documentation
   - Triggered on: Push to `main`, documentation changes

3. **Website** (Cloudflare Pages)
   - Marketing website
   - Built from `slimmcp/docs` repository
   - Automatic deployments on push
   - SSL, CDN, edge optimization

## Development Workflow

### Setup

```bash
# Clone the monorepo
git clone https://github.com/IdeaCraftersHQ/oss-mcp-slim.git
cd oss-mcp-slim

# Install dependencies
cd mcp
go mod download

# Run tests
go test ./...

# Build locally
go build -o slim cmd/slimmcp/main.go
```

### Making Changes

1. **Create a Feature Spec** (optional for large features)
   ```bash
   mkdir specs/002-new-feature
   # Add spec.md, plan.md, tasks.md
   ```

2. **Develop in the Monorepo**
   ```bash
   # Make changes in mcp/, docs/, etc.
   git add .
   git commit -m "feat: add smart filtering strategy"
   ```

3. **Push to Trigger Deployment**
   ```bash
   git push origin main
   # GitHub Actions will automatically build and deploy
   ```

4. **Create a Release**
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   # Triggers release build with binaries
   ```

## Monorepo Tools

This monorepo uses several tools to manage the workflow:

- **SpecKit** - Feature specification and planning workflow
- **Claude Code** - AI-assisted development with project context
- **GitHub Actions** - CI/CD automation
- **Go Modules** - Dependency management for Go code
- **Make** - Build automation (coming soon)

## Branch Strategy

- `main` - Production-ready code, triggers deployments
- `develop` - Development branch for integration
- `feature/*` - Feature branches
- `fix/*` - Bug fix branches
- `docs/*` - Documentation-only changes

## Contributing

### For Core Team

1. Clone this monorepo
2. Create a feature branch
3. Make changes in appropriate directories
4. Push and create a PR
5. After merge, GitHub Actions deploys automatically

### For External Contributors

For external contributions, please contribute directly to the production repositories:

- Code contributions: [slimmcp/slimmcp](https://github.com/slimmcp/slimmcp)
- Documentation: [slimmcp/docs](https://github.com/slimmcp/docs)

See `CONTRIBUTING.md` in each repository for guidelines.

## Feature Planning

This monorepo uses **SpecKit** for feature planning:

```bash
# Create a new feature spec using Claude or Codex
/speckit.constitution  # Define project principles
/speckit.specify      # Create feature specification
/speckit.plan         # Generate implementation plan
/speckit.tasks        # Break down into tasks
/speckit.implement    # Execute the plan
```

All specs live in `specs/` and follow a structured format.

## Documentation

- **Project Documentation:** `docs/` → deployed to [docs.slimmcp.io](https://docs.slimmcp.io)
- **API Documentation:** Auto-generated from Go code
- **Specs:** `specs/` for feature planning
- **Drafts:** `drafts/` for work-in-progress ideas

## Directory Organization

The monorepo is organized into distinct areas:

- **`mcp/`** - SlimMCP proxy source code (deployed to `slimmcp/slimmcp`)
- **`docs/`** - Documentation content (deployed to `slimmcp/docs` → slimmcp.com/docs)
- **`specs/`** - SpecKit feature specifications and implementation plans
- **`drafts/`** - Work-in-progress ideas and proposals
- **`.specify/`** - SpecKit workflow automation scripts
- **`.claude/` / `.codex/`** - AI assistant configuration
- **`.github/`** - CI/CD workflows and GitHub Actions

## License

- **SlimMCP Source Code:** GNU AGPL-3.0 (see `mcp/LICENSE`)
- **Documentation:** CC BY 4.0

## Links

- **Sponsor Organization:** [github.com/IdeaCraftersHQ](https://github.com/IdeaCraftersHQ)
- **SlimMCP Organization:** [github.com/slimmcp](https://github.com/slimmcp)
- **SlimMCP Repository:** [github.com/slimmcp/slimmcp](https://github/slimmcp/slimmcp)
- **Documentation:** [slimmcp.com/docs](https://slimmcp.com/docs)
- **Website:** [slimmcp.com](https://slimmcp.com)
- **Discord:** [discord.gg/slimmcp](https://discord.gg/slimmcp)

## Contact

- **General:** hello@ideacrafters.hq
- **Security:** security@slimmcp.io
- **Discord:** [Join our community](https://discord.gg/slimmcp)

---

**This is a development monorepo for SlimMCP.** End users should visit [slimmcp.com](https://slimmcp.com) or the production repositories at [github.com/slimmcp](https://github.com/slimmcp).
