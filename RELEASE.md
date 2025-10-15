# SlimMCP Release Process

This document describes the complete release process for SlimMCP, including versioning strategy, deployment workflows, and release checklists.

## Version Format

SlimMCP uses semantic versioning with support for pre-release versions:

```
v0.1.0-alpha.1    # Alpha release (internal testing)
v0.1.0-alpha.2    # Second alpha
v0.1.0-beta.1     # Beta release (early adopters)
v0.1.0-rc.1       # Release candidate (production-ready candidate)
v0.1.0            # Stable release
```

## Creating a Release

### 1. Pre-Release (Alpha/Beta)

Alpha and beta releases are used for internal testing and early adopter feedback:

```bash
# Update CHANGELOG.md files
# - Add entry to CHANGELOG.md (monorepo root)
# - Add entry to mcp/CHANGELOG.md (product)

git add CHANGELOG.md mcp/CHANGELOG.md
git commit -m "Update CHANGELOG for v0.1.0-alpha.1"

# Create and push tag
git tag v0.1.0-alpha.1
git push origin v0.1.0-alpha.1
```

**What happens:**
- ✅ Tests run automatically
- ✅ Binaries built for all platforms (Linux, macOS, Windows)
- ✅ Docker images built and tagged: `slimmcp/proxy:0.1.0-alpha.1`, `slimmcp/proxy:alpha`
- ✅ GitHub release created (marked as pre-release)
- ❌ **NOT deployed** to slimmcp/slimmcp (alpha/beta only)

### 2. Release Candidate

Release candidates are production-ready candidates that undergo final validation:

```bash
# Final testing complete, ready for production candidate

git tag v0.1.0-rc.1
git push origin v0.1.0-rc.1
```

**What happens:**
- ✅ All pre-release steps above
- ✅ **Deployed to slimmcp/slimmcp** production repository
- ✅ Docker tagged: `slimmcp/proxy:0.1.0-rc.1`, `slimmcp/proxy:rc`

### 3. Stable Release

Stable releases are fully validated and ready for production use:

```bash
# RC validated, ready for stable release

git tag v0.1.0
git push origin v0.1.0
```

**What happens:**
- ✅ All pre-release steps above
- ✅ **Deployed to slimmcp/slimmcp** production repository
- ✅ Docker tagged: `slimmcp/proxy:0.1.0`, `slimmcp/proxy:0.1`, `slimmcp/proxy:latest`
- ✅ Release notes from `.github/RELEASE_TEMPLATE.md`

## Release Checklist

Before creating a release, verify:

- [ ] All tests passing (`cd mcp && go test ./...`)
- [ ] CHANGELOG.md updated (both monorepo and mcp/)
- [ ] Version references updated if needed
- [ ] Documentation reviewed for accuracy
- [ ] Examples tested and working
- [ ] Performance targets met
- [ ] Breaking changes documented
- [ ] Migration guide provided (if applicable)

## Manual Release (Emergency)

If automated release fails, you can trigger manually:

```bash
# Via GitHub UI
# 1. Go to Actions tab
# 2. Select "Release SlimMCP" workflow
# 3. Click "Run workflow"
# 4. Enter version (e.g., v0.1.0-alpha.1)
```

## Documentation Deployment

Documentation is automatically deployed when `docs/` changes:

```bash
# Make changes to documentation
git add docs/
git commit -m "docs: update configuration guide"
git push origin main

# Automatically triggers docs deployment to slimmcp/docs
```

## Hotfix Releases

For urgent production fixes:

```bash
# Create fix on main or release branch
git commit -m "fix: critical security issue"

# Tag as patch release
git tag v0.1.1
git push origin v0.1.1

# Automatically builds and deploys
```

## Versioning Strategy

- **Monorepo**: May have separate tags for infrastructure updates (`monorepo-v1.0.0`)
- **SlimMCP Product**: Follows semantic versioning (`v0.1.0`, `v0.2.0`, `v1.0.0`)
- **Pre-releases**: Use suffixes (`-alpha.N`, `-beta.N`, `-rc.N`)

## Release Workflow Diagram

```
┌─────────────────────────────────────┐
│  IdeaCraftersHQ/oss-mcp-slim        │
│  (This Monorepo)                    │
└──────────────┬──────────────────────┘
               │
               │ Push tag (v0.1.0)
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

## Version History

See [CHANGELOG.md](CHANGELOG.md) for complete version history.
