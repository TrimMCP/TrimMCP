# Repository Guidelines

## Project Structure & Module Organization
SlimMCP lives in `mcp/`, a self-contained Go module with `cmd/slimmcp` for the CLI entry point, reusable packages under `pkg/`, private helpers in `internal/`, and system-level suites in `tests/`. Configuration examples sit in `mcp/examples/`. Docs rendered to the public site live in `docs/`, while plan artifacts stay in `specs/` and drafts in `drafts/`. SpecKit automation and AI assistant prompts live in `.specify/`, `.claude/`, and `.codex/`; keep them updated when workflows change.

## Build, Test, and Development Commands
From the repo root run `cd mcp` before executing toolchain commands. Typical workflow:
```bash
go mod tidy
go fmt ./...
go test ./...
go build -o slim cmd/slimmcp/main.go
```
Use `./slim start --config examples/dev.yaml` to verify the proxy locally. For planning, run `.specify/scripts/bash/create-new-feature.sh` to scaffold new specs.

## Coding Style & Naming Conventions
Format Go code with `gofmt` (tabs equivalent to 2 spaces) and enforce linting via `golangci-lint run`. Exported symbols use PascalCase, packages stay lower_snake, and tests end with `_test.go`. Keep public APIs in `pkg/` free of MCP-specific assumptions; stash implementation details in `internal/`. Markdown docs should follow sentence-case headings and one-sentence-per-line edits for clean diffs.

## Testing Guidelines
Unit and integration tests belong alongside source in `pkg/*_test.go` and `tests/` respectively; prefer table-driven cases. Run `go test -cover ./...` before PRs and add benchmarks under `pkg/filter` when optimizing paths that affect latency targets. Document fixtures inside `mcp/tests/data/` and avoid baking credentials into configs.

## Commit & Pull Request Guidelines
Follow Conventional Commits (`feat:`, `fix:`, `docs:`). Reference the relevant spec (`specs/001-trim`) or issue in the body and note any config changes. PRs should include a short changelog, test evidence (`go test` output or screenshots for docs), and updated SpecKit tasks if automation scripts ran. Request review from the domain owner when touching `pkg/` or protocol docs.

## SpecKit & Agent Workflow
Keep SpecKit state in sync by running `.specify/scripts/bash/update-agent-context.sh` after major changes. Agents should refresh CLAUDE and Codex sessions (`.codex/sessions/`) post-merge so the assistants stay aligned with the latest architecture and guidelines.
