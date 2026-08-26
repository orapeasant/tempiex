# Tempiex — Development Rules

## Repository Overview

Tempiex is a monorepo containing a Tempiex workflow platform distribution with enterprise-grade features: a core Tempiex server, a smart gRPC proxy, a UI server, a React/shadcn web UI, a CLI, and a Go-based agent harness (pi).

**Monorepo structure:**

```
tempiex/                  # Go module: Tempiex server core
proxy/                    # Go module: gRPC proxy
server/                   # Go module: UI HTTP/gRPC gateway server
web/                      # pnpm package: React + shadcn Web UI
cli/                      # Go module: CLI tool
pi/                       # Go module: agent harness (Raspberry Pi / edge runner)
sdk-python/               # Python package: Tempiex Python SDK (pure Python + grpcio)
docs/                     # Project documentation and specs
scripts/                  # Monorepo-level scripts
.github/                  # CI/CD workflows
```

## Conversational Style

- Keep answers short and concise.
- No emojis in commits, issues, PR comments, or code.
- No fluff or cheerful filler text.
- Technical prose only; be direct.
- Use concise, clear, simple language. Define unavoidable jargon before using it.
- Explain non-trivial designs as: problem → concrete example/trace → solution → why it is necessary vs. optional.
- Prefer concrete behavior and small illustrations over abstract summaries or unexplained lists.
- When the user asks a question, answer it first before making edits or running implementation commands.
- When responding to user feedback or analysis, explicitly say whether you agree or disagree before describing what you changed.

## Code Quality

- Read files in full before wide-ranging changes or when asked to audit. Do not rely on search snippets for broad changes.
- No `any` (Go: `interface{}` / TS: `any`) unless absolutely necessary; prefer typed alternatives.
- Inline single-line helpers that have only one call site.
- Check `node_modules` or `go.sum` for external API types; don't guess signatures.
- Always ask before removing functionality that appears intentional.
- Do not preserve backward compatibility unless requested.

### Go (tempiex/, proxy/, server/, cli/, pi/)

- Target **Go 1.24+**. Use standard library idioms; prefer `errors.Is`/`errors.As` over string matching.
- Use `go.uber.org/fx` for dependency injection in all server binaries.
- Use `github.com/rs/zerolog` for structured logging.
- Use OpenTelemetry (`go.opentelemetry.io/otel`) for metrics and tracing; **do not use Prometheus directly** — export via the OTel collector.
- gRPC services use `google.golang.org/grpc` with generated protobuf types.
- Return errors from library code; wrap with `fmt.Errorf("context: %w", err)`. Validate inputs early.
- Prefer `sync.Mutex` for synchronization; avoid holding locks across I/O.
- Use `context.Context` as the first parameter of every function that performs I/O.
- Format with `gofmt`; lint with `golangci-lint`.

### TypeScript / React (web/)

- Target **React 18+** with TypeScript strict mode.
- UI components use **shadcn/ui** on top of Radix UI primitives and **Tailwind CSS**.
- State management: React Query (`@tanstack/react-query`) for server state; Zustand for client state if needed.
- No `any`. No inline dynamic imports. Top-level imports only.
- Use only erasable TypeScript syntax (no parameter properties, `enum`, `namespace/module`, `import =`, `export =`).
- Format with **Prettier**; lint with **ESLint** (+ `@typescript-eslint/eslint-plugin`).

### Python (sdk-python/)

- Target **Python 3.10+**. Pure asyncio; no Rust bridge.
- Package manager: **uv**. Do not use `pip`, `poetry`, or `conda` directly.
- No `Any` (use typed alternatives). Use `dataclasses` or `pydantic` models, not plain dicts.
- Format and lint with **ruff** (`ruff check` + `ruff format`). Type-check with **pyright**.
- Use `pytest` + `pytest-asyncio` for tests; no bare `assert` — use pytest assertions.
- Do not use `time.sleep` or blocking I/O inside `@workflow.defn` methods.
- All gRPC calls use `grpcio` with generated stubs from Tempiex proto files.
- OTel tracing via `opentelemetry-exporter-otlp-proto-grpc`; no Prometheus.

## Package Manager

- **pnpm** for all JavaScript/TypeScript packages. Do not use `npm` or `yarn`.
- Install with `pnpm install --frozen-lockfile` in CI. Locally: `pnpm install`.
- Workspace config at `pnpm-workspace.yaml`.

## Commands

### Go modules

```bash
# From any Go module subdirectory:
go build ./...
go test ./...
golangci-lint run ./...
```

### web/

```bash
pnpm install
pnpm dev          # start Vite dev server
pnpm build        # production build
pnpm lint         # ESLint
pnpm test         # Vitest unit tests
pnpm test:e2e     # Playwright E2E
pnpm typecheck    # tsc --noEmit
```

### Root convenience (scripts/):

```bash
./scripts/build-all.sh     # build every Go module + web
./scripts/test-all.sh      # run all Go + web tests
./scripts/lint-all.sh      # lint all modules
```

After Go code changes: `golangci-lint run ./...` in the affected module. Fix all errors before committing.
After web changes: `pnpm lint && pnpm typecheck` from `web/`. Fix all errors before committing.
After Python changes: `uv run ruff check . && uv run ruff format --check . && uv run pyright` from `sdk-python/`. Fix all errors before committing.
Never run builds or tests unnecessarily — only the smallest targeted command that covers the changed code.

## Git

Multiple sessions may run in the same workspace concurrently. Follow these rules:

**Committing:**
- Only commit files changed in this session.
- Stage explicit paths (`git add <path1> <path2>`); never `git add -A` or `git add .`.
- Verify with `git status` before committing.
- Commit message format: `{feat,fix,docs,chore,refactor}[(tempiex,proxy,server,web,cli,pi)]: <concise message>`.

**Never run:**
- `git reset --hard`, `git checkout .`, `git clean -fd`, `git stash`, `git add -A`, `git add .`, `git commit --no-verify`

**Rebase conflicts:**
- Resolve only in files you modified. If a conflict is in an untouched file, abort and ask the user.
- Never force push.

## Dependency & Security

- Pin Go module dependencies to exact versions in `go.sum`.
- Pin npm/pnpm packages to exact versions in `pnpm-lock.yaml`.
- Never commit secrets or credentials.
- Treat lockfile changes as reviewed code.
- Review external dep changelogs before upgrading security-sensitive packages (grpc, oidc, jwt, otel).

## Testing

- Write unit tests alongside every new exported function/method.
- Integration and E2E tests live in `*_test.go` files tagged with `//go:build integration` (Go) or under `test/e2e/` (web).
- For Go: use `testify/require` and `testify/assert`. For web: use Vitest + React Testing Library.
- If you create or modify a test file, run it and iterate until it passes.
- Never commit failing tests.

## OpenTelemetry Observability

All server components (tempiex, proxy, server, pi) export telemetry via the OTel Collector:

- **Traces**: OTLP gRPC to collector (`localhost:4317` by default, configurable).
- **Metrics**: OTLP gRPC to collector (replaces Prometheus scrape endpoints).
- **Logs**: Structured JSON via zerolog; optionally forwarded to collector.

Do not add direct Prometheus client instrumentation. The OTel Collector handles metric translation to any backend (Prometheus, Grafana, Datadog, etc.).

## Implementation Sequence

Build in this order — each depends on the previous:

1. **tempiex** — Tempiex server core (foundation)
2. **proxy** — gRPC proxy (depends on tempiex API types)
3. **server** — UI gateway server (depends on tempiex gRPC API)
4. **web** — React UI (depends on server HTTP API)
5. **cli** — CLI tool (depends on tempiex + proxy APIs)
6. **pi** — Agent harness (depends on tempiex SDK + server API)

Detailed specs for each component are in `docs/spec/`:
- `01-tempiex.md` — Tempiex server core
- `02-proxy.md` — gRPC proxy
- `03-server.md` — UI server
- `04-web.md` — Web UI
- `05-cli.md` — CLI
- `06-pi.md` — Pi agent harness
- `07-sdk-python.md` — Python SDK

## Issues and PRs

When creating issues, add `pkg:*` labels for affected packages (`pkg:tempiex`, `pkg:proxy`, `pkg:server`, `pkg:web`, `pkg:cli`, `pkg:pi`).

When posting AI-generated issue/PR comments, end with: `This comment is AI-generated.`

## User Override

If the user's instructions conflict with any rule in this document, ask for explicit confirmation before overriding.
