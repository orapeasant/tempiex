# Tempiex

Tempiex is a durable workflow orchestration engine — a clean-room implementation inspired by the design of Temporal. It provides reliable, fault-tolerant execution of long-running business processes using an event-sourced history model.

## Contents

- [Architecture](#architecture)
- [Quick Start (Docker)](#quick-start-docker)
- [Quick Start (Local)](#quick-start-local)
- [Python SDK](#python-sdk)
- [CLI](#cli)
- [Web UI](#web-ui)
- [Configuration](#configuration)
- [Project Layout](#project-layout)
- [Development](#development)

---

## Architecture

```
┌──────────────────────────────────────────────────────┐
│                    Clients / Workers                 │
│          Python SDK · CLI · Web UI · pi agent        │
└───────────┬──────────────────────┬───────────────────┘
            │ gRPC :8133           │ HTTP :8080
            ▼                      ▼
   ┌─────────────────┐   ┌───────────────────────┐
   │  Tempiex Core   │   │  UI Server (server/)  │
   │  (tempiex/)     │   │  REST API + Web UI    │
   │                 │   └───────────────────────┘
   │  Frontend svc   │
   │  History svc    │
   │  Matching svc   │
   └────────┬────────┘
            │
            ▼
   ┌─────────────────┐
   │   PostgreSQL    │
   │   (persists     │
   │   all state)    │
   └─────────────────┘
```

| Component | Folder | Port | Language |
|-----------|--------|------|----------|
| Core engine | `tempiex/` | `8133` (gRPC) | Go |
| UI server | `server/` | `8080` (HTTP) | Go |
| Web UI | `web/` | — (embedded in server) | React + TypeScript |
| CLI | `cli/` | — | TypeScript + Ink |
| Python SDK | `sdk-python/` | — | Python 3.10+ |
| Generic worker | `worker/` | — | Go |
| Pi agent harness | `pi/` | `8090` | Go |

---

## Quick Start (Docker)

The fastest way to run the full stack. Requires [Docker Desktop](https://www.docker.com/products/docker-desktop/) or Docker Engine with Compose v2.

```bash
git clone https://github.com/orapeasant/tempiex.git
cd tempiex

docker compose up -d
```

This starts:

| Service | URL |
|---------|-----|
| **Web UI** | http://localhost:8080 |
| **Tempiex gRPC API** | `localhost:8133` |
| **PostgreSQL** | `localhost:5432` |

Check everything is healthy:

```bash
docker compose ps
curl http://localhost:8080/health
```

Stop (data persists):

```bash
docker compose down
```

Stop and wipe all data:

```bash
docker compose down -v
```

### Data persistence

PostgreSQL data is stored in the named Docker volume `tempiex_postgres_data`. This volume survives `docker compose down` and container removal. It is only deleted when you explicitly run `docker compose down -v`.

---

## Quick Start (Local)

### Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.24+ | https://go.dev/dl |
| Node.js | 22+ | https://nodejs.org |
| pnpm | 9+ | `npm install -g pnpm` |
| Python | 3.10+ | https://python.org |
| PostgreSQL | 14+ | https://postgresql.org |

### 1. PostgreSQL

```bash
# Create the database and user
psql -U postgres <<SQL
CREATE USER tempiex WITH PASSWORD 'tempiex';
CREATE DATABASE tempiex OWNER tempiex;
SQL
```

### 2. Tempiex core engine

```bash
cd tempiex
go build -o bin/tempiex-server ./cmd/server/
./bin/tempiex-server          # listens on :8133
```

The server automatically applies the schema on first start.

### 3. UI server + web UI

Build the web UI first so the server can embed it:

```bash
cd web
pnpm install
pnpm build
cp -r dist/ ../server/internal/server/ui/
```

Then start the UI server:

```bash
cd server
go build -o bin/tempiex-ui-server ./cmd/server/
./bin/tempiex-ui-server       # listens on :8080
```

Open http://localhost:8080 in your browser.

### 4. CLI

```bash
cd cli
pnpm install
pnpm build
node dist/tempiex.js --help

# Optional: install globally
npm link
tempiex --help
```

---

## Python SDK

The Tempiex Python SDK lets you write and run workflow and activity workers in pure Python.

### Install

```bash
cd sdk-python
pip install grpcio protobuf          # runtime deps only
# or with uv:
uv pip install grpcio protobuf
```

### Write a workflow

```python
import asyncio
from datetime import timedelta
from tempiex.client import Client
from tempiex.worker import Worker
from tempiex import workflow, activity


@activity.defn
async def greet(name: str) -> str:
    return f"Hello, {name}!"


@workflow.defn
class GreetingWorkflow:

    @workflow.run
    async def run(self, name: str) -> str:
        return await workflow.execute_activity(
            greet, name,
            start_to_close_timeout=timedelta(seconds=10),
        )


async def main():
    client = await Client.connect("localhost:8133", namespace="default")

    worker = Worker(client, "greetings",
                    workflows=[GreetingWorkflow],
                    activities=[greet])

    async with asyncio.TaskGroup() as tg:
        tg.create_task(worker.run())
        result = await client.execute_workflow(
            GreetingWorkflow, "World",
            id="hello-world", task_queue="greetings",
        )
        print(result)   # Hello, World!
        worker.shutdown()


asyncio.run(main())
```

### Annotations reference

| Decorator | Target | Description |
|-----------|--------|-------------|
| `@workflow.defn` | class | Marks a class as a workflow |
| `@workflow.defn(name="x")` | class | Workflow with a custom registered name |
| `@workflow.run` | async method | Main entry-point of a workflow (required, exactly one) |
| `@workflow.signal` | async method | Receives inbound signals while the workflow runs |
| `@workflow.signal(name="x")` | async method | Signal with a custom name |
| `@workflow.query` | sync method | Read-only state query answered synchronously |
| `@workflow.query(name="x")` | sync method | Query with a custom name |
| `@workflow.update` | async method | Synchronous RPC that mutates state and returns a value |
| `@activity.defn` | async function | Marks a function as an activity |
| `@activity.defn(name="x")` | async function | Activity with a custom registered name |

### Workflow helpers

```python
# Execute an activity from inside a workflow
result = await workflow.execute_activity(
    my_activity, arg1, arg2,
    start_to_close_timeout=timedelta(minutes=5),
)

# Block until a condition is true (e.g. waiting for a signal)
await workflow.wait_condition(lambda: self._approved)

# With timeout — returns False if timed out
completed = await workflow.wait_condition(
    lambda: self._done, timeout=timedelta(hours=1)
)
```

### Signals and queries

```python
@workflow.defn
class ApprovalWorkflow:

    def __init__(self):
        self._approved = False

    @workflow.signal
    async def approve(self, approver: str = ""):
        self._approved = True

    @workflow.query
    def get_status(self) -> str:
        return "approved" if self._approved else "pending"

    @workflow.run
    async def run(self, request_id: str) -> str:
        await workflow.wait_condition(lambda: self._approved)
        return "approved"
```

Sending a signal from the client:

```python
handle = await client.start_workflow(
    ApprovalWorkflow, "req-123",
    id="approval-req-123", task_queue="approvals",
)

# From another process / later:
await handle.signal("approve", "alice")
result = await handle.result()
```

### WorkflowHandle methods

```python
handle = await client.start_workflow(...)

result  = await handle.result()             # wait for completion
status  = await handle.describe()           # DescribeWorkflowExecution
await handle.signal("my-signal", payload)   # send a signal
value   = await handle.query("get-status")  # synchronous query
await handle.cancel()                       # request graceful cancel
await handle.terminate(reason="cleanup")    # force terminate
```

### Sample app

A complete order-processing workflow demonstrating all annotations is in:

```
sdk-python/samples/order_workflow.py
```

Run it (server must be running):

```bash
cd sdk-python
python samples/order_workflow.py
```

---

## CLI

The `tempiex` CLI provides command-mode and interactive TUI access to a running Tempiex server.

```bash
# Point at a server
export TEMPIEX_ADDRESS=http://localhost:8080

tempiex workflow list -n default
tempiex workflow describe -n default -w my-workflow-id
tempiex workflow show    -n default -w my-workflow-id   # event history
tempiex namespace list
tempiex cluster health
```

### Global flags

| Flag | Env var | Default | Description |
|------|---------|---------|-------------|
| `--address` | `TEMPIEX_ADDRESS` | `http://localhost:8080` | UI server HTTP address |
| `-n, --namespace` | `TEMPIEX_NAMESPACE` | `default` | Namespace |
| `-o, --output` | — | `text` | `text` \| `json` \| `yaml` |
| `--profile` | `TEMPIEX_PROFILE` | `default` | Config profile |

### Command reference

```
tempiex workflow  list | describe | show | start | count | signal | terminate | cancel
tempiex namespace list | describe | create | delete
tempiex cluster   describe | health
tempiex taskqueue describe
tempiex config    list | get | set | delete
tempiex server    start-dev
tempiex worker    start-dev
tempiex pi        start-dev
```

### Config profiles

```bash
# Store connection settings
tempiex config set --prop address   --value http://prod.example.com:8080
tempiex config set --prop namespace --value prod-ns

# Use a named profile
tempiex config set --profile staging --prop address --value http://staging:8080
tempiex workflow list --profile staging
```

Config file: `~/.config/tempiex/config.toml`

---

## Web UI

The web UI is served at `http://localhost:8080` by the UI server.

| Page | Path | Description |
|------|------|-------------|
| Namespace list | `/` | All namespaces |
| Workflow list | `/namespaces/:ns/workflows` | Workflows in a namespace |
| Workflow detail | `/namespaces/:ns/workflows/:id/:runId` | Status, history events, attributes |

---

## Configuration

### Core engine (`tempiex/`)

Default path: `config/development.yaml`. Override with `TEMPIEX_CONFIG` env var.

```yaml
server:
  grpcPort: 8133

database:
  host: localhost
  port: 5432
  user: tempiex
  password: tempiex
  dbName: tempiex
  sslMode: disable
```

### UI server (`server/`)

Default path: `config/development.yaml`. Override with `TEMPIEX_CONFIG` env var.

```yaml
tempiexGrpcAddress: "localhost:8133"
host: "0.0.0.0"
port: 8080
dbDsn: "postgres://tempiex:tempiex@localhost:5432/tempiex?sslmode=disable"
auth:
  enabled: false
cors:
  cookieInsecure: true
```

---

## Project Layout

```
tempiex/                  # Go module — core engine (gRPC :8133)
  api/proto/              # Protobuf definitions
  api/gen/                # Generated Go stubs
  internal/
    config/               # Config loader
    persistence/          # PostgreSQL store
    history/              # Workflow state machine + event sourcing
    matching/             # Task queue with long-poll
    frontend/             # gRPC service handler
  schema/postgresql/      # DDL (auto-applied on startup)
  cmd/server/             # Binary entry point

server/                   # Go module — HTTP/gRPC gateway + web UI (:8080)
  internal/api/           # REST handlers (namespaces, workflows, cluster)
  internal/rpc/           # gRPC client to core engine
  internal/server/        # Echo setup, routes, embedded SPA

web/                      # React 18 + TypeScript + Tailwind web UI
  src/pages/              # NamespaceList, WorkflowList, WorkflowDetail
  src/components/         # Layout, Sidebar, StatusBadge, EventPayload
  src/hooks/              # React Query data-fetching hooks
  src/api/                # API client + types

sdk-python/               # Python SDK — pure Python + grpcio
  tempiex/
    client.py             # Client, WorkflowHandle
    worker.py             # Worker poll loop + history replay
    workflow.py           # @workflow.defn/run/signal/query/update
    activity.py           # @activity.defn
  samples/
    order_workflow.py     # Full example with all annotations

cli/                      # TypeScript + Ink v5 CLI
  src/
    main.tsx              # Commander root
    commands/             # workflow/, namespace/, cluster/, config/, server/, worker/, pi/
    components/           # Table, StatusBadge, Spinner, ErrorPanel
    hooks/                # Data-fetching hooks

worker/                   # Go — generic Tempiex activity worker (task-queue polling/dispatch)

pi/                       # Go — Pi AI agent harness (edge runner), depends on worker/

docs/spec/                # Design specifications
  01-tempiex.md           # Core engine
  02-proxy.md             # gRPC proxy
  03-server.md            # UI server
  04-web.md               # Web UI
  05-cli.md               # CLI
  06-pi.md                # Pi agent harness
  07-sdk-python.md        # Python SDK
  08-worker.md            # Generic worker
```

---

## Development

### Run tests

```bash
# Core engine (Go)
cd tempiex && go test ./...

# UI server (Go)
cd server && go test ./...

# Web UI
cd web && pnpm test

# CLI
cd cli && pnpm test

# Python SDK (requires server running on :8133)
cd sdk-python && python -m pytest tests/ -v
```

### Lint

```bash
cd tempiex  && golangci-lint run ./...
cd server   && golangci-lint run ./...
cd web      && pnpm lint && pnpm typecheck
cd cli      && pnpm lint && pnpm typecheck
cd sdk-python && uv run ruff check . && uv run ruff format --check .
```

### Regenerate protobuf stubs

```bash
# Go stubs
cd tempiex/api/proto && buf generate

# Python stubs
cd sdk-python
python -m grpc_tools.protoc \
  --proto_path=../tempiex/api/proto \
  --python_out=tempiex/api \
  --grpc_python_out=tempiex/api \
  --pyi_out=tempiex/api \
  tempiex/api/{enums,common,failure,history,workflowservice}/v1/*.proto
```

### Docker build (individual images)

All images are built from the repo root:

```bash
docker build -f tempiex/Dockerfile    -t tempiex/core:latest    .
docker build -f server/Dockerfile     -t tempiex/server:latest  .
docker build -f cli/Dockerfile        -t tempiex/cli:latest     .
docker build -f sdk-python/Dockerfile -t tempiex/sdk-python:latest .
```

Or use the helper script:

```bash
./scripts/docker-build.sh
./scripts/docker-build.sh v1.0.0     # with a tag
```

### Port reference

| Port | Service | Protocol |
|------|---------|----------|
| `8133` | Tempiex core engine | gRPC |
| `8080` | UI server + Web UI | HTTP |
| `5432` | PostgreSQL | TCP |
| `8090` | Pi agent HTTP API | HTTP |

---

## License

MIT
