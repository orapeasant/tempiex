# Tempiex — Build Plan

## Overview

Seven components built in strict dependency order. Each phase produces a working, testable artifact before the next begins. Phases are not parallelized — each gate must pass before proceeding.

**Tempiex server is a clean-room implementation** of a durable workflow orchestration engine. It defines its own gRPC API (`tempiex.api.*`) — there is no compatibility requirement with the Temporal wire protocol. Tempiex provides its own SDKs; `sdk-python/` is the first. The Temporal reference source at `~/app/temporal/temporal` is used as design material only — state machines, schema structure, and protocol semantics. PostgreSQL is the only supported database.

---

## Phase 1 — tempiex (Workflow Server Core)

**Folder**: `tempiex/`  
**Depends on**: nothing (PostgreSQL must be running)  
**Gate**: `go test ./...` passes; server starts; gRPC health check on :8133 passes; `sdk-python` client can start and complete a simple workflow end-to-end

### Steps

1. **Repo scaffold**
   - Initialize git repo at monorepo root
   - Create `tempiex/go.mod` (`module github.com/tempiex/tempiex`)
   - Add `golangci-lint` config (`.golangci.yaml`) at monorepo root
   - Add `scripts/build-all.sh`, `scripts/test-all.sh`, `scripts/lint-all.sh`
   - Add `.github/workflows/ci.yml` (lint + test on push/PR)

2. **Proto definitions** (`api/proto/`, `api/gen/`)
   - Write Tempiex proto definitions from scratch in `api/proto/tempiex/api/*/v1/`
   - Use the Temporal proto files at `~/app/temporal/temporal/proto/` as a structural reference for what messages and fields the workflow protocol needs — do not copy verbatim
   - All package names are `tempiex.api.*`; no compatibility with the Temporal wire protocol
   - Set up `buf.yaml` and `buf.gen.yaml`; run `buf generate` to produce Go stubs into `api/gen/`
   - Commit generated stubs; regenerate only on proto changes

3. **Configuration layer** (`internal/config/`)
   - YAML config struct covering all services, persistence, membership, OTel
   - `config/development.yaml` — PostgreSQL local, single-node, OTel → localhost:4317
   - `config/docker.yaml` — PostgreSQL via service name, OTel → otel-collector:4317
   - Env-var override with `TEMPIEX_` prefix

4. **OTel bootstrap** (`internal/metrics/`)
   - Initialize OTel SDK at startup: OTLP gRPC exporter to collector (`:4317`)
   - Shared `TracerProvider` and `MeterProvider` via fx
   - No Prometheus client — metrics go exclusively to OTel collector
   - gRPC interceptor: per-RPC spans on Frontend; OTel context propagation via gRPC metadata

5. **Persistence layer** (`internal/persistence/`)
   - Interface definitions: `ExecutionStore`, `VisibilityStore`, `NamespaceStore` (see spec for full interface)
   - PostgreSQL implementation via `pgx/v5` with connection pool
   - History event storage: serialized proto blobs per branch node
   - Transfer task table: workflow/activity task routing from History → Matching
   - Timer task table: fire-time-ordered; polled by History timer processor
   - Schema files: `schema/postgresql/temporal/versioned/` (execution) and `schema/postgresql/visibility/versioned/` (visibility)
   - Serialization layer: `ProtoSerializer` marshals/unmarshals protobuf to `[]byte`

6. **Schema migration tool** (`cmd/tools/`)
   - `tempiex-tool` binary; applies versioned SQL migrations
   - `create-default-db` command: creates exec store schema
   - `create-visibility-db` command: creates visibility store schema
   - Dev mode: server auto-runs migrations on startup

7. **Membership manager** (`internal/membership/`)
   - Single-node mode: localhost, all shards owned by this process (no ring needed)
   - Ring interface: `GetHostsByService(service)` → returns hosts for shard routing
   - Future: consistent hash ring for multi-node

8. **Namespace registry** (`internal/namespace/`)
   - In-memory cache of namespace config loaded from `NamespaceStore`
   - Refresh every 10s
   - `GetNamespace(id)` used by all services

9. **History service** (`internal/history/`) — core of the system
   - Per-shard execution engine: owns a set of shards; shard = hash(workflowID) % numHistoryShards
   - Mutable state per workflow: event history (append-only log), activity infos, timer infos, child workflow infos
   - All state mutations: append events + update mutable state in a single DB transaction; per-shard mutex
   - Workflow task processing: apply worker Commands → update state → enqueue transfer tasks
   - Command handlers: ScheduleActivityTask, StartTimer, CompleteWorkflowExecution, FailWorkflowExecution, ContinueAsNewWorkflowExecution, SignalExternalWorkflow, StartChildWorkflow (see spec for full list)
   - Transfer task processor (background): reads DB transfer tasks, calls Matching.AddWorkflowTask / AddActivityTask
   - Timer task processor (background): reads timer tasks by fire time; creates timeout events; schedules workflow tasks
   - Signal, query, update delivery to running executions
   - gRPC handler: `HistoryServiceServer` on `:8134`

10. **Matching service** (`internal/matching/`)
    - Per-task-queue manager: workflow tasks + activity tasks
    - Long-poll dispatch: if poller waiting, dispatch immediately; else buffer task in DB
    - Sticky task queues: workflow tasks returned to the same worker
    - Task queue partitioning: N partitions; forwarder from leaf partitions to root
    - gRPC handler: `MatchingServiceServer` on `:8135`

11. **Frontend service** (`internal/frontend/`)
    - `WorkflowService`, `OperatorService`, `AdminService` on `:8133`
    - HTTP gateway on `:8243` via grpc-gateway
    - gRPC interceptor chain: OTel → Auth → RateLimit → Logging → handler
    - Request validation: field checks, namespace existence, size limits
    - Shard routing: workflow RPCs → History instance owning the shard
    - Poll routing: `PollWorkflowTaskQueue`, `PollActivityTaskQueue` → Matching
    - Visibility routing: `ListWorkflowExecutions`, `CountWorkflowExecutions` → VisibilityStore

12. **Auth** (`internal/auth/`)
    - gRPC interceptor: JWT bearer (JWKS), API key header, mTLS client cert, no-auth dev mode
    - RBAC: namespace-level roles (Reader, Writer, Admin)

13. **Rate limiter** (`internal/quotas/`)
    - Token bucket per namespace + per API
    - Returns `codes.ResourceExhausted` on limit exceeded

14. **Worker service** (`internal/worker/`)
    - Internal Tempiex worker using the same poll loop pattern as `sdk-python/` (polls `tempiex-sys-*` task queues via the Tempiex gRPC API)
    - Schedule workflow implementation
    - Archival workflow stub (no-op initially)
    - gRPC handler: `WorkerServiceServer` on `:8139`

15. **Binary entry point** (`cmd/server/main.go`)
    - `fx.New(...)` wiring all modules
    - Graceful shutdown on SIGINT/SIGTERM: drain in-flight tasks, flush OTel spans

16. **Tests**
    - Unit tests alongside every package (use `testify/require`, `testify/assert`)
    - Integration tests under `tests/integration/` (`//go:build integration`); require PostgreSQL
    - Functional smoke test: start server → create namespace → start workflow → poll + respond → workflow completes
    - Replay correctness test: simulate a sequence of WFT history events, verify History applies correct state transitions

### Deliverable

```
tempiex/
├── cmd/
│   ├── server/             # tempiex-server binary
│   └── tools/              # tempiex-tool binary (schema migrations)
├── api/
│   ├── proto/              # proto definitions (Temporal-compatible API contract)
│   └── gen/                # generated Go stubs
├── internal/
│   ├── config/
│   ├── metrics/            # OTel bootstrap
│   ├── persistence/        # PostgreSQL driver + interfaces + serialization
│   ├── membership/
│   ├── namespace/          # namespace registry cache
│   ├── frontend/           # gRPC handler, auth interceptor, rate limiter
│   ├── history/            # engine, events, state machines, transfer/timer processors
│   ├── matching/           # task queue manager, poll manager, forwarder
│   ├── worker/             # internal system worker (schedules, archival)
│   ├── auth/
│   ├── quotas/
│   ├── archiver/           # stub initially
│   └── dynamicconfig/      # file-based runtime flags
├── schema/
│   └── postgresql/
│       ├── temporal/versioned/
│       └── visibility/versioned/
├── config/
│   ├── development.yaml
│   └── docker.yaml
└── tests/
    ├── integration/
    └── functional/
```

---

## Phase 2 — proxy (gRPC Proxy)

**Folder**: `proxy/`  
**Depends on**: tempiex (gRPC API types via `go.tempiex.com/api`)  
**Gate**: `go test ./...` passes; proxy starts, routes requests to tempiex on :8133

### Steps

1. **Module scaffold** (`proxy/go.mod`: `module github.com/tempiex/proxy`)

2. **Config layer** (`internal/config/`)
   - YAML structs: gateway, upstreams[], routing rules, auth, TLS, OTel
   - Validation with descriptive errors

3. **Router** (`internal/router/`)
   - Namespace match: exact, glob prefix, regex
   - gRPC metadata match
   - System upstream for namespace-less RPCs
   - Default fallback

4. **Upstream proxy instances** (`internal/proxy/`)
   - Per-upstream connection pool to Tempiex service
   - Namespace translation (prefix / suffix / exact map) — symmetric on request + response
   - Credential injection: API key header, mTLS client cert
   - Outbound TLS config

5. **Crypto module** (`internal/crypto/`, `internal/kms/`)
   - AES-256-GCM envelope encryption
   - DEK generation per operation; KEK wrapping via KMS
   - KMS backends: AWS KMS, Azure Key Vault, GCP Cloud KMS
   - Per-namespace key override

6. **Inbound auth** (`internal/auth/`)
   - Static bearer token
   - JWT / JWKS validation (RS256, ES256)
   - Extension server delegation

7. **Extension server framework** (`pkg/ext/`)
   - gRPC server scaffold for custom KMS / auth
   - TLS, health check, graceful shutdown built-in

8. **Gateway** (`internal/server/`)
   - Single inbound gRPC listener (:8133 default, configurable)
   - TLS termination (optional)
   - Codec-transparent: never parses workflow payloads

9. **OTel** — metrics + traces (OTLP gRPC)

10. **Tests**
    - Unit: router matching, namespace translation, crypto round-trip
    - Integration (`e2e/`): full proxy → tempiex round-trip

### Deliverable

```
proxy/
├── cmd/proxy/          # tempiex-proxy binary
├── internal/
│   ├── config/
│   ├── router/
│   ├── proxy/
│   ├── crypto/
│   ├── kms/
│   ├── auth/
│   ├── server/
│   └── metrics/
├── pkg/
│   └── ext/            # extension server framework
└── e2e/
```

---

## Phase 3 — server (UI Server)

**Folder**: `server/`  
**Depends on**: tempiex gRPC API (:8133)  
**Gate**: `go test ./...` passes; `GET /health` returns 200; workflow list endpoint returns valid JSON

### Steps

1. **Module scaffold** (`server/go.mod`: `module github.com/tempiex/server`)

2. **Config layer** (`internal/config/`)
   - `tempiexGrpcAddress`, `host`, `port`, TLS, auth, CORS, CSRF, codec endpoint, OTel

3. **gRPC client** (`internal/rpc/`)
   - WorkflowService + OperatorService clients
   - Connection pool, reconnect, TLS, timeout

4. **Middleware stack** (`internal/cors/`, `internal/csrf/`, `internal/headers/`)
   - Logger → Recover → CORS → CSRF → Auth → SecurityHeaders

5. **Auth module** (`internal/auth/`)
   - OIDC / OAuth 2.0 with `go-oidc/v3` and `golang.org/x/oauth2`
   - Secure session cookie (`gorilla/securecookie`)
   - Login / callback / logout routes
   - No-auth dev mode

6. **HTTP route handlers** (`internal/api/`)
   - Namespace, workflow, schedule, cluster, auth, settings routes
   - HTTP → gRPC translation: JSON ↔ protobuf marshal/unmarshal
   - Pagination token pass-through

7. **Codec server proxy** (`internal/codec/`)
   - Forward payload decode requests to external codec server URL

8. **Static asset serving**
   - `//go:embed ui/*` — serves built `web/dist/` as SPA
   - SPA fallback: unknown routes → `index.html`

9. **OTel** — request metrics + traces (OTLP gRPC)

10. **Binary entry point** (`cmd/server/main.go`) — fx wiring

11. **Tests**
    - Unit: middleware, auth session, JSON↔gRPC transforms
    - Integration: full HTTP request → mocked gRPC → response

### Deliverable

```
server/
├── cmd/server/         # tempiex-ui-server binary
├── internal/
│   ├── config/
│   ├── rpc/
│   ├── api/
│   ├── auth/
│   ├── cors/
│   ├── csrf/
│   ├── headers/
│   ├── codec/
│   └── metrics/
└── ui/                 # placeholder for embedded web/dist
```

---

## Phase 4 — web (React UI)

**Folder**: `web/`  
**Depends on**: server HTTP API (:8080)  
**Gate**: `pnpm lint && pnpm typecheck` pass; `pnpm test` passes; `pnpm build` produces `dist/`

### Steps

1. **pnpm workspace scaffold**
   - `pnpm-workspace.yaml` at monorepo root
   - `web/package.json` with all dependencies pinned to exact versions
   - `web/vite.config.ts` — dev proxy `/api` → `localhost:8080`
   - `web/tsconfig.json` — strict mode
   - `web/tailwind.config.ts`
   - `web/eslint.config.mjs` (flat config, `@typescript-eslint`, `react-hooks`)
   - `web/.prettierrc`

2. **shadcn/ui setup**
   - Initialize shadcn (`pnpm dlx shadcn@latest init`)
   - Add required components: `button`, `table`, `dialog`, `sheet`, `select`, `badge`, `input`, `tabs`, `dropdown-menu`, `toast`

3. **API client layer** (`src/lib/api/`)
   - `client.ts` — fetch wrapper: base URL, CSRF cookie header, error normalization
   - `workflows.ts`, `namespaces.ts`, `schedules.ts`, `cluster.ts`

4. **Type definitions** (`src/lib/types/`)
   - `workflow.ts`, `namespace.ts`, `schedule.ts`, `cluster.ts`
   - Derived from server API response shapes

5. **State layer**
   - React Query setup (`QueryClientProvider`) in `App.tsx`
   - Hooks: `useWorkflows`, `useWorkflowDetail`, `useEventHistory`, `useNamespaces`, `useCluster`, `useSchedules`
   - Zustand stores: `namespaceStore` (selected ns, persisted), `uiStore` (sidebar, theme)

6. **Router** (`App.tsx` + `src/routes/`)
   - React Router v6 with nested routes
   - Route layout: sidebar + namespace selector + main content

7. **Pages**
   - Workflow List — filterable, paginated table; status badges
   - Workflow Detail — tabs: Event History, Pending Activities, Stack Trace, Queries, Input/Result
   - Event History — expandable rows with CodeMirror JSON payload viewer
   - Schedule List + Detail
   - Namespace Settings
   - Cluster Dashboard
   - Login page (when auth enabled)

8. **Shared components** (`src/components/common/`)
   - `DataTable`, `SearchBar`, `Pagination`, `CodeEditor` (CodeMirror 6), `DateTimeDisplay`, `ErrorBoundary`

9. **Tests**
   - Vitest unit tests for hooks and utils
   - Playwright E2E: list workflows, inspect detail, send signal, terminate

10. **Build integration**
    - `pnpm build` → `web/dist/`
    - CI step copies `web/dist/` → `server/ui/` before building server binary

### Deliverable

```
web/
├── src/
│   ├── routes/
│   ├── components/
│   └── lib/
│       ├── api/
│       ├── hooks/
│       ├── store/
│       ├── types/
│       └── utils/
├── test/e2e/
├── dist/               # build output → copied to server/ui/
├── vite.config.ts
├── tailwind.config.ts
└── package.json
```

---

## Phase 5 — cli (Command-Line Tool)

**Folder**: `cli/`  
**Depends on**: tempiex gRPC API, server HTTP API  
**Gate**: `go test ./...` passes; `tempiex workflow list` returns valid output against running tempiex

### Steps

1. **Module scaffold** (`cli/go.mod`: `module github.com/tempiex/cli`)

2. **Config** (`internal/config/`)
   - `~/.tempiex/config.yaml` loader via viper
   - Priority: flag > env (`TEMPIEX_*`) > config file

3. **gRPC client factory** (`internal/client/`)
   - WorkflowService, OperatorService, AdminService
   - TLS, API key, timeout configurable

4. **Output formatters** (`internal/output/`)
   - `table.go` — human-readable tabular output
   - `json.go` — raw JSON (for `jq` piping)
   - `yaml.go` — YAML

5. **Command implementations**
   - `workflow`: list, describe, start, signal, query, update, terminate, cancel, reset, history
   - `namespace`: list, describe, register, update
   - `schedule`: list, describe, create, update, trigger, pause, unpause, delete
   - `taskqueue`: describe, list-members
   - `cluster`: info, health
   - `proxy`: start, validate
   - `server`: start, version
   - `dev`: orchestrate tempiex + proxy + server as child processes

6. **Shell completion** (`cobra` built-in)
   - bash, zsh, fish, powershell

7. **Tests**
   - Unit: output formatters, config loading, flag parsing
   - Integration: commands against running tempiex

### Deliverable

```
cli/
├── cmd/tempiex/        # tempiex binary
└── internal/
    ├── config/
    ├── client/
    ├── output/
    ├── workflow/
    ├── namespace/
    ├── schedule/
    ├── taskqueue/
    ├── cluster/
    ├── proxy/
    ├── server/
    └── dev/
```

---

## Phase 6 — pi (Agent Harness)

**Folder**: `pi/`  
**Depends on**: tempiex SDK (`go.tempiex.com/sdk`), server HTTP API (:8080)  
**Gate**: `go test ./...` passes; pi registers worker, executes `shell` activity, emits SSE events

### Steps

1. **Module scaffold** (`pi/go.mod`: `module github.com/tempiex/pi`)

2. **Config** (`internal/config/`)
   - YAML: temporal address, namespace, workers[], API port, session store path, OTel, log level
   - Env override with `PI_` prefix

3. **Tool registry** (`internal/tool/`)
   - Generic `Tool[I, O any]` struct
   - `ToolRegistry`: Register, Get, List
   - Enforces unique names at registration time

4. **Event bus** (`internal/event/`)
   - `EventBus`: Publish, Subscribe (fan-out)
   - Event types: `execution_start`, `execution_update`, `execution_end`, `execution_error`
   - In-memory with channel-based dispatch

5. **Session manager** (`internal/session/`)
   - SQLite-backed session store (`modernc.org/sqlite`)
   - Open, Close, Get, List, Cancel
   - Persists events per session for SSE replay on reconnect

6. **Tempiex worker wrapper** (`internal/worker/`)
   - `WorkerPool`: one Tempiex worker per configured task queue
   - Activity shim: wraps `Tool.Execute` with BeforeExec / AfterExec hooks and event emission
   - Progress streaming via `context`-attached helper

7. **HTTP API server** (`internal/api/`)
   - `GET /health`
   - `GET /api/sessions`, `GET /api/sessions/:id`, `POST /api/sessions/:id/cancel`
   - `GET /api/sessions/:id/events` — SSE with history replay
   - `GET /api/tools`

8. **Built-in tools** (`tools/`)
   - `shell` — exec shell command; streams stdout/stderr via `execution_update`
   - `http_request` — HTTP client tool
   - `file_read`, `file_write`

9. **OTel** — activity metrics + traces, per-activity span under Tempiex workflow trace context

10. **Binary entry point** (`cmd/pi/main.go`) — fx wiring, graceful drain on SIGINT/SIGTERM

11. **Tests**
    - Unit: tool registry, event bus, session manager
    - Integration: full Tempiex activity dispatch → event → SSE

### Deliverable

```
pi/
├── cmd/pi/             # pi binary
├── internal/
│   ├── config/
│   ├── tool/
│   ├── event/
│   ├── session/
│   ├── worker/
│   ├── api/
│   └── metrics/
└── tools/
    ├── shell/
    ├── http/
    └── file/
```

---

## Phase 7 — sdk-python (Python SDK)

**Folder**: `sdk-python/`  
**Depends on**: tempiex server running on gRPC :8133  
**Gate**: `uv run pytest` passes end-to-end workflow round-trip (start → activity → result)

### Steps

1. **Package scaffold**
   - `sdk-python/pyproject.toml` — package `tempiex`, Python ≥3.10, hatchling build backend
   - `uv sync --all-extras` — install deps and dev group
   - `scripts/gen_protos.py` — generate gRPC stubs from Tempiex proto files into `tempiex/api/`
   - Add `py.typed` PEP 561 marker

2. **Proto stub generation** (`scripts/gen_protos.py`)
   - Locate Tempiex `.proto` files (from tempiex module or separately vendored)
   - Run `grpc_tools.protoc` for: workflowservice/v1, common/v1, history/v1, enums/v1, failure/v1, namespace/v1, schedule/v1
   - Output to `tempiex/api/tempiex/api/`
   - Committed stubs; regenerate only when protos change

3. **Service layer** (`tempiex/service.py`)
   - `ServiceClient` dataclass: host, port, namespace, TLS config, API key, metadata interceptor
   - `async connect(target: str, ...) -> ServiceClient`
   - Wraps generated gRPC stub (`WorkflowServiceStub`) with asyncio channel
   - `TLSConfig`: server_root_ca_cert, client_cert, client_private_key, domain

4. **Common types** (`tempiex/common.py`)
   - `RetryPolicy`, `WorkflowIDReusePolicy` (enum), `WorkflowInfo`, `ActivityInfo`
   - `Schedule`, `ScheduleSpec`, `ScheduleActionStartWorkflow`, `ScheduleHandle` stubs
   - `SearchAttributes`, `Memo` type aliases

5. **Exceptions** (`tempiex/exceptions.py`)
   - `ApplicationError(message, type, non_retryable, details)`
   - `CancelledError`, `ActivityError(cause)`, `ChildWorkflowError(cause)`, `TimeoutError`, `TerminatedError`
   - `RPCError(status_code, message)`, `RPCStatusCode` (mirrors gRPC status codes)

6. **Data converter** (`tempiex/converter.py`)
   - `DataConverter` dataclass with `payload_codec: PayloadCodec | None`
   - `DataConverter.default()` — JSON encoding
   - `JSONPayloadConverter`: encodes Python → `Payload(encoding="json/plain", data=json.dumps(...).encode())`
   - `PayloadCodec` abstract base: `encode(payloads) -> payloads`, `decode(payloads) -> payloads`
   - Pluggable codec chain: encode on send, decode on receive

7. **Activity decorator** (`tempiex/activity.py`)
   - `@activity.defn` / `@activity.defn(name="...", no_thread_cancel_exception=False)`
   - `activity.info() -> ActivityInfo` — reads from `contextvars.ContextVar`
   - `activity.heartbeat(*details)` — sends `RecordActivityTaskHeartbeat` gRPC call
   - `activity.is_cancelled() -> bool`

8. **Workflow decorators** (`tempiex/workflow.py`)
   - `@workflow.defn(name=None, sandboxed=True)`
   - `@workflow.run` — marks the entry-point coroutine method
   - `@workflow.signal`, `@workflow.query`, `@workflow.update`
   - In-workflow context APIs: `workflow.info()`, `workflow.now()`, `workflow.memo()`, `workflow.random()`, `workflow.logger`
   - Workflow-safe async operations: `workflow.execute_activity(fn, *args, ...)`, `workflow.sleep(td)`, `workflow.wait_condition(fn, timeout)`, `workflow.start_child_workflow(wf, ...)`, `workflow.get_external_workflow_handle(id)`
   - Versioning: `workflow.patch(id)`, `workflow.deprecate_patch(id)`
   - `workflow.continue_as_new(fn, *args)`
   - All context APIs raise `RuntimeError` when called outside a workflow task

9. **Workflow execution engine** (`tempiex/_internal/`)
   - `_WorkflowInstance`: holds the workflow coroutine, current command sequence, pending futures keyed by sequence number
   - `_WorkflowTask.execute(history_events, new_events)`:
     1. Replay existing events — resolve futures for completed activities, timers, signals, child workflows
     2. Advance coroutine until it blocks or completes
     3. Collect new commands (ScheduleActivityTask, StartTimer, CompleteWorkflowExecution, etc.)
   - `_command.py`: builder functions that produce gRPC `Command` protos from Python-side operations
   - `_proto_helpers.py`: `payload_to_python(payload, dc)`, `python_to_payload(obj, dc)` using DataConverter
   - Sandbox: override asyncio event-loop `call_soon` / `call_later` for workflow coroutine to block disallowed I/O; raise `SandboxViolationError`

10. **Activity executor** (`tempiex/_internal/_activity_task.py`)
    - `execute_activity(fn, input, info, heartbeat_cb)`:
      - `asyncio.iscoroutinefunction(fn)` → `await fn(*args)` directly in the event loop
      - Sync function → `loop.run_in_executor(thread_pool, fn, *args)`
    - Cancellation: `asyncio.Task.cancel()` for async; threading.Event for sync (polled in heartbeat)
    - Result serialized with DataConverter; failure serialized as `Failure` proto

11. **Poll loop** (`tempiex/_internal/_poll_loop.py`)
    - Two concurrent asyncio tasks: WFT poller, AT poller
    - WFT poller: `PollWorkflowTaskQueue` (long poll, 60s timeout) → dispatch to `_WorkflowTask.execute()` → `RespondWorkflowTaskCompleted` or `RespondWorkflowTaskFailed`
    - AT poller: `PollActivityTaskQueue` (long poll, 60s timeout) → dispatch to `_ActivityTask.execute()` → `RespondActivityTaskCompleted`, `RespondActivityTaskFailed`, or `RespondActivityTaskCanceled`
    - Graceful drain: on shutdown signal, stop polling, wait for in-flight tasks to complete

12. **Worker** (`tempiex/worker.py`)
    - `WorkerConfig` dataclass: task_queue, workflows, activities, max_concurrent_workflow_tasks, max_concurrent_activities, interceptors, data_converter
    - `Worker(client, task_queue, workflows, activities, ...)` — validates registry, starts poll loop
    - `await worker.run()` — blocks until `asyncio.CancelledError`
    - Async context manager support

13. **Client** (`tempiex/client.py`)
    - `await Client.connect(target, namespace, tls, api_key, data_converter) -> Client`
    - `await client.start_workflow(fn, *args, id, task_queue, ...) -> WorkflowHandle`
    - `await client.execute_workflow(fn, *args, id, task_queue, ...) -> T`
    - `client.get_workflow_handle(id, run_id=None) -> WorkflowHandle`
    - `async for wf in client.list_workflows(query)` — uses `ListWorkflowExecutions`
    - `await client.count_workflows(query) -> int`
    - Schedule methods: `create_schedule`, get handle → `trigger`, `pause`, `unpause`, `delete`
    - `WorkflowHandle`: `result()`, `cancel()`, `terminate()`, `signal()`, `query()`, `start_update()`, `describe()`

14. **OTel interceptor** (`tempiex/opentelemetry.py`) — optional extra
    - `OpenTelemetryInterceptor(tracer_provider=None)`
    - Wraps `ExecuteActivityInput` and `ExecuteWorkflowInput` with spans
    - Propagates trace context via workflow memo / activity header

15. **Testing utilities** (`tempiex/testing.py`)
    - `WorkflowEnvironment.start_local(tempiex_path=None, postgres_dsn=None) -> AsyncContextManager`
      - Launches `tempiex-server` subprocess with a dev PostgreSQL config (uses `postgres_dsn` or `TEMPIEX_TEST_DB_DSN` env var)
      - Waits for gRPC health check; yields ready client
      - Tears down subprocess on exit
    - `WorkflowEnvironment.start_time_skipping() -> AsyncContextManager`
      - Same but patches `workflow.sleep` and `workflow.wait_condition` to resolve immediately (skips real time)

16. **Tests** (`tests/`)
    - `test_converter.py` — round-trip encode/decode for int, str, list, dict, dataclass, None
    - `test_replay.py` — replay correctness: simulate WFT history, verify commands match expectations
    - `test_activity.py` — async activity, sync activity, heartbeat, cancellation
    - `test_workflow.py` — full round-trip using `WorkflowEnvironment.start_local()`
    - `test_client.py` — start_workflow, execute_workflow, list_workflows, signal, query
    - `test_worker.py` — graceful shutdown drain, max concurrency enforcement

17. **Examples** (`examples/`)
    - `order_worker.py` — basic workflow + activity example
    - `hello_world.py` — minimal 10-line example

### Deliverable

```
sdk-python/
├── tempiex/
│   ├── __init__.py
│   ├── client.py
│   ├── worker.py
│   ├── workflow.py
│   ├── activity.py
│   ├── converter.py
│   ├── exceptions.py
│   ├── service.py
│   ├── testing.py
│   ├── common.py
│   ├── opentelemetry.py        # optional OTel interceptor
│   ├── py.typed
│   └── _internal/
│       ├── _workflow_instance.py
│       ├── _workflow_task.py
│       ├── _activity_task.py
│       ├── _poll_loop.py
│       ├── _command.py
│       └── _proto_helpers.py
├── tempiex/api/                # generated protobuf + gRPC stubs
├── tests/
├── examples/
├── scripts/
│   └── gen_protos.py
├── pyproject.toml
└── uv.lock
```

---

## CI/CD Pipeline (`.github/workflows/`)

| Workflow | Trigger | Steps |
|----------|---------|-------|
| `ci.yml` | push / PR | lint + unit test all Go modules; pnpm lint + typecheck + test web; ruff + pyright + pytest (unit only) sdk-python |
| `build.yml` | push to main | build all binaries + web; copy `web/dist` → `server/ui`; build server binary with embedded UI; `uv build` sdk-python wheel |
| `integration.yml` | push to main | spin up PostgreSQL (docker service); run tempiex, proxy, server, pi; run integration + E2E tests; run sdk-python `WorkflowEnvironment.start_local()` tests |
| `docker.yml` | tag `v*` | build + push Docker images for tempiex, proxy, server, pi |

---

## Monorepo Root Files (to create)

```
/
├── AGENTS.md                   ✓ done
├── PLAN.md                     ✓ done
├── pnpm-workspace.yaml         (web/ workspace)
├── .golangci.yaml              (shared linter config)
├── .github/
│   └── workflows/
│       ├── ci.yml
│       ├── build.yml
│       ├── integration.yml
│       └── docker.yml
└── scripts/
    ├── build-all.sh
    ├── test-all.sh
    └── lint-all.sh
```

---

## Summary

| Phase | Folder | Language | Gate |
|-------|--------|----------|------|
| 1 | `tempiex/` | Go | server starts, gRPC :8133 up, SDK round-trip passes |
| 2 | `proxy/` | Go | proxy routes to tempiex |
| 3 | `server/` | Go | HTTP API returns workflow list |
| 4 | `web/` | TypeScript/React | build passes, E2E green |
| 5 | `cli/` | TypeScript/Ink | `tempiex workflow list` works |
| 6 | `pi/` | Go | activity executes, SSE streams |
| 7 | `sdk-python/` | Python | workflow round-trip test passes |
