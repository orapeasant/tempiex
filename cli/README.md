# @tempiex/cli

Tempiex CLI. A terminal interface for managing Tempiex workflows. Renders TUI components using [Ink](https://github.com/vadimdemedes/ink) (React for CLIs).

Two modes:
- **Command mode** — run a command, print output, exit.
- **Interactive TUI** — `tempiex ui` launches a full keyboard-navigable dashboard.

## Quick Start

### Prerequisites

- Node.js 22+
- pnpm
- A running `server` instance on `http://localhost:8080`

### Installation

```bash
# From the monorepo root:
pnpm install

# Or install globally from npm (once published):
npm install -g @tempiex/cli
```

### Run (development)

```bash
cd cli/
pnpm build         # compile TypeScript → dist/
node dist/main.js  # or via the bin alias
```

### Configure

The CLI reads `TEMPIEX_SERVER_URL` (default: `http://localhost:8080`):

```bash
export TEMPIEX_SERVER_URL=http://localhost:8080
```

## Commands

### Workflow

```bash
# List running workflows
tempiex workflow list

# Start a workflow
tempiex workflow start --workflow-type MyWorkflow --task-queue my-queue --namespace default

# Describe a workflow
tempiex workflow describe --workflow-id <id> --run-id <run-id>

# Show workflow history
tempiex workflow show --workflow-id <id>

# Signal a workflow
tempiex workflow signal --workflow-id <id> --signal-name my-signal --input '{"key":"value"}'

# Cancel a workflow
tempiex workflow cancel --workflow-id <id>

# Terminate a workflow
tempiex workflow terminate --workflow-id <id> --reason "manual stop"
```

### Namespace

```bash
tempiex namespace list
tempiex namespace describe --namespace default
```

### Task Queue

```bash
tempiex taskqueue list
```

### Interactive TUI

```bash
tempiex ui
```

## Development

```bash
cd cli/

# Install deps
pnpm install

# Watch mode (rebuilds on change)
pnpm dev

# Type check
pnpm typecheck

# Lint
pnpm lint

# Run tests
pnpm test

# Production build
pnpm build
```

## Project Structure

```
cli/
├── src/
│   ├── app/          # Ink app entry point
│   ├── commands/     # Command handlers (workflow/, namespace/, etc.)
│   ├── components/   # Reusable Ink UI components
│   ├── client/       # HTTP client for server API
│   ├── hooks/        # React hooks
│   └── config/       # Config loading
└── test/             # Unit tests
```
