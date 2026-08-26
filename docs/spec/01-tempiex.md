# 01 — Tempiex: Server Core Specification

## Overview

**Component**: Tempiex Server Core  
**Folder**: `tempiex/`  
**Language**: Go 1.24+  
**Module**: `github.com/tempiex/tempiex`  
**Type**: Distributed Workflow Orchestration Engine — clean-room implementation  
**Implementation Priority**: 1 — Foundation (all other components depend on this)

## Purpose

Tempiex is a clean-room implementation of a durable workflow orchestration engine. It defines its own gRPC API (`tempiex.api.*`) — there is no compatibility requirement with the Temporal wire protocol. All Tempiex SDKs are written by this project; the first is `sdk-python/`. Additional language SDKs will be added based on need.

The design of workflows, activities, task queues, event sourcing, and state machines follows the same conceptual model as Temporal (documented in the reference specs at `~/app/temporal/`), but the wire protocol, proto packages, and SDK interfaces are Tempiex-native.

No source code from `go.temporal.io/server` is used. The Temporal reference source and specs serve as design material only.

**PostgreSQL is the only supported database.** Both the execution store and visibility store use PostgreSQL via `pgx/v5`.

---

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    tempiex server                       │
│                                                         │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐       │
│  │  Frontend  │  │  History   │  │  Matching  │       │
│  │  :8133     │  │  :8134     │  │  :8135     │       │
│  └─────┬──────┘  └─────┬──────┘  └─────┬──────┘       │
│        └───────────────┴───────────────┘                │
│                        │ internal gRPC                   │
│                 ┌──────┴──────┐                         │
│                 │   Worker    │                         │
│                 │   :8139     │                         │
│                 └──────┬──────┘                         │
│        ┌───────────────┴───────────────┐                │
│        │                               │                │
│  ┌─────▼──────┐                 ┌──────▼──────┐        │
│  │Persistence │                 │  Membership │        │
│  │   Layer    │                 │   Manager   │        │
│  └────────────┘                 └─────────────┘        │
└─────────────────────────────────────────────────────────┘
         │
         ▼
┌─────────────────┐
│   PostgreSQL    │
│ - exec store    │
│ - visibility    │
└─────────────────┘
```

---

## Core Services

| Service | Port | Purpose |
|---------|------|---------|
| Frontend | 8133 (gRPC), 8243 (HTTP) | API gateway for all SDK/worker clients |
| History | 8134 (gRPC) | Workflow state machine and event sourcing |
| Matching | 8135 (gRPC) | Task queue management and routing |
| Worker | 8139 (gRPC) | Internal system workflows (schedules, archival) |
| OTel Collector | 4317 | OTLP gRPC export target |

Membership ports (ring-based discovery, internal):

| Service | Membership Port |
|---------|----------------|
| Frontend | 6933 |
| History | 6934 |
| Matching | 6935 |
| Worker | 6939 |

---

## Technical Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| Language | Go | 1.24+ |
| RPC | gRPC | 1.80+ |
| API | Protobuf | 1.36+ |
| DI | uber-go/fx | 1.24+ |
| Logging | zerolog | 1.35+ |
| Observability | OpenTelemetry | 1.44+ |
| Persistence | PostgreSQL | pgx/v5 5.10+ |
| Auth | JWT (golang-jwt/jwt v4) | — |

---

## Database Support

PostgreSQL is the only supported database. Both stores run on PostgreSQL via `pgx/v5`.

| Store | Database | Schema path |
|-------|----------|-------------|
| Execution store | `tempiex` | `schema/postgresql/tempiex/versioned/` |
| Visibility store | `tempiex_visibility` | `schema/postgresql/visibility/versioned/` |

Schema is versioned. `tempiex-tool` applies migrations. Dev mode applies them automatically on startup.

---

## Folder Structure

```
tempiex/
├── cmd/
│   ├── server/              # main binary: starts all services
│   └── tools/               # tempiex-tool: schema migration
├── config/
│   ├── development.yaml     # PostgreSQL local, single-node
│   └── docker.yaml          # PostgreSQL via service name
├── api/
│   ├── proto/               # Tempiex proto definitions (tempiex.api.* namespace)
│   └── gen/                 # generated Go stubs (buf/protoc output)
├── internal/
│   ├── config/              # Config loader (YAML + env override)
│   ├── metrics/             # OTel bootstrap (TracerProvider, MeterProvider)
│   ├── membership/          # Ring-based cluster membership
│   ├── persistence/         # DB interfaces + PostgreSQL driver
│   │   ├── datainterfaces/  # ExecutionStore, VisibilityStore interfaces
│   │   ├── postgresql/      # pgx/v5 implementation
│   │   └── serialization/   # event history serialization (proto ↔ DB)
│   ├── frontend/            # Frontend service (gRPC handler + auth + rate limit)
│   ├── history/             # History service (state machine, event sourcing)
│   │   ├── engine/          # workflow execution engine
│   │   ├── events/          # event history builder
│   │   ├── statemachines/   # workflow/activity/timer state machines
│   │   └── transfer/        # transfer/timer task processors
│   ├── matching/            # Matching service (task queues, poller management)
│   │   ├── taskqueue/       # per-queue state and poller management
│   │   └── forwarder/       # task forwarding across partitions
│   ├── worker/              # Internal Worker service (schedules, archival)
│   ├── namespace/           # Namespace registry and lifecycle
│   ├── quotas/              # Rate limiting (per-namespace, per-API)
│   ├── auth/                # JWT validation, mTLS, API key, RBAC
│   ├── archiver/            # Workflow history archival (filesystem, S3, GCS)
│   └── dynamicconfig/       # Runtime-configurable flags (file-based)
├── schema/
│   └── postgresql/
│       ├── tempiex/versioned/      # execution store migrations
│       └── visibility/versioned/   # visibility store migrations
├── tests/
│   ├── integration/               # //go:build integration
│   └── functional/                # end-to-end workflow tests
├── go.mod
└── go.sum
```

---

## Proto / API Strategy

Tempiex defines its own gRPC API under the `tempiex.api.*` proto package namespace. There is no compatibility requirement with the Temporal wire protocol — gRPC service names, message names, and field names are chosen for Tempiex and may diverge from Temporal at any point.

The Temporal proto files at `~/app/temporal/temporal/proto/` are used as a **structural reference** — to understand what messages and fields are needed for the workflow protocol — but they are not copied verbatim. Proto definitions are written fresh in `api/proto/`.

Steps:
1. Write Tempiex proto definitions in `api/proto/tempiex/api/*/v1/`.
2. Set up `buf.yaml` and `buf.gen.yaml`; run `buf generate` to produce Go stubs into `api/gen/`.
3. Commit generated stubs; regenerate only on proto changes.

Key proto packages:

| Tempiex package | Purpose |
|-----------------|---------|
| `tempiex.api.workflowservice.v1` | Main public API (start WF, poll, signal, query, etc.) |
| `tempiex.api.operatorservice.v1` | Namespace management, search attributes |
| `tempiex.api.adminservice.v1` | Admin operations (internal) |
| `grpc.health.v1` | gRPC health check (standard, unchanged) |
| `tempiex.api.history.v1` | History event types |
| `tempiex.api.common.v1` | Payload, retry policy, headers |
| `tempiex.api.enums.v1` | All enum types |
| `tempiex.api.failure.v1` | Failure type |
| `tempiex.api.schedule.v1` | Schedule messages |
| `tempiex.api.taskqueue.v1` | Task queue messages |

---

## Service Implementations

### 1. Frontend Service (`internal/frontend/`)

The public entry point. All SDK clients and workers connect here.

**Responsibilities:**
- Expose `WorkflowService`, `OperatorService`, `AdminService` gRPC endpoints on `:8133`
- HTTP gateway on `:8243` via grpc-gateway (`GET /health`, `GET /ready`, REST endpoints)
- Validate all inbound requests (field checks, namespace existence, rate limits)
- Authenticate requests: JWT bearer, API key header, mTLS client cert
- Forward most RPCs to the History or Matching service via internal gRPC clients
- Enforce per-namespace and per-API rate limits (token bucket)
- Return structured gRPC errors (`codes.InvalidArgument`, `codes.NotFound`, etc.)

**Key RPCs handled locally (no forwarding):**
- `RegisterNamespace`, `DescribeNamespace`, `ListNamespaces`, `UpdateNamespace`
- `GetClusterInfo`, `GetSystemInfo`
- `ListWorkflowExecutions`, `CountWorkflowExecutions` (query visibility store directly)
- `DescribeWorkflowExecution` (query persistence)

**Key RPCs forwarded to History:**
- `StartWorkflowExecution` → History (shard-routed by workflow ID)
- `SignalWorkflowExecution`, `RequestCancelWorkflowExecution`, `TerminateWorkflowExecution`
- `QueryWorkflow`, `UpdateWorkflowExecution`
- `GetWorkflowExecutionHistory`
- `RespondWorkflowTaskCompleted`, `RespondWorkflowTaskFailed`
- `RespondActivityTaskCompleted`, `RespondActivityTaskFailed`, `RecordActivityTaskHeartbeat`

**Key RPCs forwarded to Matching:**
- `PollWorkflowTaskQueue`, `PollActivityTaskQueue`

**Shard routing:**
Workflow requests are routed to the History service instance owning the shard for the workflow ID. Shard ID = `hash(workflowID) % numHistoryShards`. Frontend maintains a shard-to-host mapping via the membership ring.

```
internal/frontend/
├── fx.go               # fx module: registers gRPC server, handlers, middleware
├── handler.go          # WorkflowServiceServer implementation
├── operator_handler.go # OperatorServiceServer implementation
├── admin_handler.go    # AdminServiceServer implementation
├── auth.go             # JWT / API key / mTLS interceptor
├── rate_limiter.go     # per-namespace token bucket
└── configs/            # per-service rate limit config
```

---

### 2. History Service (`internal/history/`)

The core of the system. Implements the workflow execution state machine via event sourcing.

**Responsibilities:**
- Own a set of shards (configurable per `numHistoryShards`)
- For each workflow execution: maintain an ordered event history in PostgreSQL
- Process Workflow Task completions: apply commands from worker, advance state machine, schedule next tasks
- Process Activity Task completions: apply results, schedule next workflow task
- Drive timers (workflow timers, schedule-to-start, start-to-close timeouts)
- Apply signals, queries, updates to running executions
- Enqueue transfer tasks (to Matching) and timer tasks (internal)

**Event sourcing model:**

```
WorkflowExecution
  └── EventHistory [event1, event2, ..., eventN]    # append-only log in DB
  └── ExecutionState {status, taskQueue, ...}        # mutable, snapshotted
  └── ActivityInfos map[scheduledEventID]ActivityInfo
  └── TimerInfos    map[timerID]TimerInfo
  └── ChildInfos    map[initiatedEventID]ChildInfo
  └── RequestCancelInfos
  └── SignalInfos
```

Every state change appends one or more events to the history and updates mutable state atomically in a single DB transaction.

**Workflow task processing:**

```
1. Worker calls PollWorkflowTaskQueue (via Frontend → Matching)
2. Matching returns a WorkflowTask token
3. Worker replays history, executes workflow code, produces Commands
4. Worker calls RespondWorkflowTaskCompleted with Commands
5. Frontend routes to History (shard owner)
6. History applies each Command:
   - ScheduleActivityTask    → write ActivityInfo, enqueue transfer task
   - RequestCancelActivity   → write cancellation request
   - StartTimer              → write TimerInfo
   - CompleteWorkflowExecution → write final event, set status=Completed
   - FailWorkflowExecution   → write final event, set status=Failed
   - ContinueAsNewWorkflowExecution → close current run, start new run
   - SignalExternalWorkflow  → enqueue signal transfer task
   - StartChildWorkflowExecution → write ChildInfo, enqueue start task
7. History writes all events + state to DB atomically
8. History enqueues transfer tasks for Matching (activity tasks, workflow tasks)
```

**Transfer task processor:**
Background goroutine per shard. Reads transfer tasks from DB, fans out to Matching service to enqueue the actual task queue entries. Retries on failure.

**Timer task processor:**
Background goroutine per shard. Reads timer tasks ordered by fire time. On fire: creates timeout events (schedule-to-start expired, start-to-close expired, workflow execution timeout), schedules a new workflow task.

```
internal/history/
├── fx.go
├── handler.go               # HistoryServiceServer gRPC handler
├── engine/
│   ├── engine.go            # per-shard execution engine
│   ├── workflow_context.go  # mutable workflow execution context
│   └── commands.go          # command application logic
├── events/
│   ├── builder.go           # constructs History events from operations
│   └── mutable_state.go     # in-memory mutable state snapshot
├── statemachines/
│   ├── workflow.go          # workflow lifecycle state machine
│   ├── activity.go          # activity lifecycle state machine
│   ├── timer.go             # timer lifecycle
│   └── child_workflow.go    # child workflow state machine
├── transfer/
│   ├── processor.go         # transfer task processor (fan-out to Matching)
│   └── timer_processor.go   # timer task processor
└── shard/
    ├── controller.go        # shard assignment and ownership
    └── context.go           # per-shard state: DB handle, task queues
```

---

### 3. Matching Service (`internal/matching/`)

Task queue management. Buffers workflow tasks and activity tasks, dispatches them to pollers.

**Responsibilities:**
- Maintain in-memory poller queues per task queue name + task type (workflow / activity)
- Receive AddWorkflowTask and AddActivityTask from History transfer processor
- Match tasks to waiting pollers (long-poll): if a poller is waiting, dispatch immediately; otherwise buffer in DB
- Support sticky task queues (workflow tasks dispatched back to the same worker)
- Implement task queue partitioning for high-throughput queues
- Track task queue backlog metrics per queue

**Poller long-poll:**
```
Worker → PollWorkflowTaskQueue (60s timeout)
    → Matching receives poll
    → If task available: return immediately
    → If no task: park poller goroutine
    → When History enqueues task: wake parked poller, return task
    → If 60s elapsed without task: return empty response
```

**Task queue partitioning:**
Each task queue can be spread across N partitions. The root partition (partition 0) aggregates. The forwarder sub-package forwards tasks from leaf partitions to the root when no local poller is waiting.

```
internal/matching/
├── fx.go
├── handler.go               # MatchingServiceServer gRPC handler
├── taskqueue/
│   ├── manager.go           # per-queue task manager
│   ├── poll_manager.go      # poller registration and wakeup
│   └── db_task_manager.go   # DB-backed task buffering
└── forwarder/
    └── forwarder.go         # inter-partition forwarding
```

---

### 4. Worker Service (`internal/worker/`)

Internal system workflow worker. Runs Tempiex-owned workflows for platform features.

**Responsibilities:**
- Run system task queue pollers on `tempiex-sys-*` task queues
- Execute built-in system workflows:
  - **Schedule workflow**: implements cron/interval schedules (create/update/delete/trigger/pause)
  - **Archival workflow**: moves completed workflow histories to archival storage
  - **Namespace replication workflow**: XDC cross-DC sync (future)
  - **Scanner workflow**: background consistency checks

```
internal/worker/
├── fx.go
├── worker.go                # starts Tempiex SDK worker on system task queues
├── schedules/               # schedule workflow implementation
│   ├── workflow.go
│   └── activities.go
└── archival/                # archival workflow implementation
    ├── workflow.go
    └── activities.go
```

---

### 5. Persistence Layer (`internal/persistence/`)

Single abstraction over the PostgreSQL stores.

**Interfaces** (`internal/persistence/datainterfaces/`):

```go
type ExecutionStore interface {
    CreateWorkflowExecution(ctx, req) (*CreateWorkflowExecutionResponse, error)
    GetWorkflowExecution(ctx, req) (*GetWorkflowExecutionResponse, error)
    UpdateWorkflowExecution(ctx, req) (*UpdateWorkflowExecutionResponse, error)
    DeleteWorkflowExecution(ctx, req) error
    GetCurrentExecution(ctx, req) (*GetCurrentExecutionResponse, error)

    // History events
    AppendHistoryNodes(ctx, req) (*AppendHistoryNodesResponse, error)
    ReadHistoryBranch(ctx, req) (*ReadHistoryBranchResponse, error)
    DeleteHistoryBranch(ctx, req) error

    // Transfer tasks
    GetTransferTasks(ctx, req) (*GetTransferTasksResponse, error)
    CompleteTransferTask(ctx, req) error

    // Timer tasks
    GetTimerIndexTasks(ctx, req) (*GetTimerIndexTasksResponse, error)
    CompleteTimerTask(ctx, req) error

    // Task queues (for Matching persistence)
    CreateTaskQueue(ctx, req) error
    GetTaskQueue(ctx, req) (*GetTaskQueueResponse, error)
    CreateTasks(ctx, req) (*CreateTasksResponse, error)
    GetTasks(ctx, req) (*GetTasksResponse, error)
    CompleteTasksLessThan(ctx, req) (*CompleteTasksLessThanResponse, error)
}

type VisibilityStore interface {
    RecordWorkflowExecutionStarted(ctx, req) error
    RecordWorkflowExecutionClosed(ctx, req) error
    UpsertWorkflowExecution(ctx, req) error
    ListOpenWorkflowExecutions(ctx, req) (*ListWorkflowExecutionsResponse, error)
    ListClosedWorkflowExecutions(ctx, req) (*ListWorkflowExecutionsResponse, error)
    ListWorkflowExecutions(ctx, req) (*ListWorkflowExecutionsResponse, error) // SQL query
    CountWorkflowExecutions(ctx, req) (*CountWorkflowExecutionsResponse, error)
    GetWorkflowExecution(ctx, req) (*GetVisibilityWorkflowExecutionResponse, error)
}

type NamespaceStore interface {
    CreateNamespace(ctx, req) (*CreateNamespaceResponse, error)
    GetNamespace(ctx, req) (*GetNamespaceResponse, error)
    UpdateNamespace(ctx, req) error
    DeleteNamespace(ctx, req) error
    ListNamespaces(ctx, req) (*ListNamespacesResponse, error)
    GetMetadata(ctx) (*GetMetadataResponse, error)
}
```

**PostgreSQL implementation** (`internal/persistence/postgresql/`):
- Uses `pgx/v5` pool with configurable connection limits
- Shard table layout: each shard is a set of rows keyed by `(shard_id, namespace_id, workflow_id, run_id)`
- History events stored as serialized proto blobs per branch node
- Optimistic concurrency via conditional updates on `db_shard_id` and version columns
- All writes within a shard serialized via per-shard mutex (no distributed transactions needed for single-node)

**Serialization** (`internal/persistence/serialization/`):
- `ProtoSerializer`: marshals/unmarshals protobuf messages to `[]byte` for DB storage
- Supports encoding version in a header byte for future migration

---

### 6. Membership Manager (`internal/membership/`)

Tracks which host owns which shards. Used by Frontend to route workflow requests.

**Single-node mode**: All services on `localhost`; all shards owned by this process. No actual ring needed.

**Multi-node mode (future)**: Consistent hash ring over membership announcements. Each History instance claims ownership of `numHistoryShards / numHistoryHosts` shards. Frontend watches ring changes, updates shard→host routing table.

```
internal/membership/
├── ring.go              # consistent hash ring
├── host_info.go         # host identity (address + role)
└── monitor.go           # membership change watcher
```

---

### 7. Namespace Registry (`internal/namespace/`)

In-memory cache of namespace configuration. Loaded from DB on startup, kept fresh via polling.

- Namespace holds: retention period, archival config, replication config, bad binary checksums, search attribute mapping
- All services call `namespaceRegistry.GetNamespace(id)` — never query DB directly for namespace info
- Refreshed every 10s from PostgreSQL NamespaceStore

---

### 8. Rate Limiter / Quotas (`internal/quotas/`)

Token-bucket rate limiter applied in Frontend interceptor.

- Global rate limit: configurable RPS for the whole cluster
- Per-namespace rate limit: configurable per namespace
- Per-API rate limit: some RPCs (e.g., Admin) have separate limits
- Returns `codes.ResourceExhausted` when limits exceeded

---

### 9. Auth (`internal/auth/`)

gRPC interceptor applied in Frontend.

| Method | Mechanism |
|--------|-----------|
| JWT Bearer | `Authorization: Bearer <token>` header; validated against JWKS URL |
| API Key | `tempiex-namespace-id` + `authorization` header; validated against static key store |
| mTLS | TLS client certificate; CN extracted as identity |
| No-auth | Dev mode; all requests allowed |

RBAC: namespace-level roles (Reader, Writer, Admin). Enforced per RPC via a policy table.

---

## Configuration

### Minimal development config (`config/development.yaml`)

```yaml
log:
  stdout: true
  level: info

persistence:
  defaultStore: postgres-default
  visibilityStore: postgres-visibility
  numHistoryShards: 4
  datastores:
    postgres-default:
      sql:
        pluginName: postgres12
        databaseName: tempiex
        connectAddr: "localhost:5432"
        connectProtocol: tcp
        user: tempiex
        password: tempiex
        maxConns: 20
        maxIdleConns: 20
        maxConnLifetime: 1h
    postgres-visibility:
      sql:
        pluginName: postgres12
        databaseName: tempiex_visibility
        connectAddr: "localhost:5432"
        connectProtocol: tcp
        user: tempiex
        password: tempiex
        maxConns: 10
        maxIdleConns: 10
        maxConnLifetime: 1h

global:
  membership:
    maxJoinDuration: 30s
    broadcastAddress: "127.0.0.1"
  metrics:
    otel:
      endpoint: "localhost:4317"
      insecure: true

services:
  frontend:
    rpc:
      grpcPort: 8133
      httpPort: 8243
      bindOnLocalHost: true
      membershipPort: 6933
  history:
    rpc:
      grpcPort: 8134
      bindOnLocalHost: true
      membershipPort: 6934
  matching:
    rpc:
      grpcPort: 8135
      bindOnLocalHost: true
      membershipPort: 6935
  worker:
    rpc:
      grpcPort: 8139
      bindOnLocalHost: true
      membershipPort: 6939
```

> **Production:** set `numHistoryShards: 512`. Point datastores at production PostgreSQL. Enable TLS. Set `bindOnLocalHost: false`.

---

## API Surface

### gRPC Services (port 8133)

All service names are Tempiex-native.

| Service | gRPC service name |
|---------|-------------------|
| WorkflowService | `tempiex.api.workflowservice.v1.WorkflowService` |
| OperatorService | `tempiex.api.operatorservice.v1.OperatorService` |
| AdminService | `tempiex.api.adminservice.v1.AdminService` |
| HealthService | `grpc.health.v1.Health` |

### Key WorkflowService RPCs

| RPC | Routed To | Description |
|-----|-----------|-------------|
| `StartWorkflowExecution` | History | Create + start new workflow run |
| `PollWorkflowTaskQueue` | Matching | Worker long-poll for workflow tasks |
| `RespondWorkflowTaskCompleted` | History | Apply commands from worker |
| `RespondWorkflowTaskFailed` | History | Mark workflow task as failed |
| `PollActivityTaskQueue` | Matching | Worker long-poll for activity tasks |
| `RespondActivityTaskCompleted` | History | Complete activity |
| `RespondActivityTaskFailed` | History | Fail activity |
| `RecordActivityTaskHeartbeat` | History | Heartbeat + cancellation check |
| `RequestCancelWorkflowExecution` | History | Request graceful cancel |
| `SignalWorkflowExecution` | History | Deliver signal to running workflow |
| `QueryWorkflow` | History | Read workflow state |
| `UpdateWorkflowExecution` | History | Synchronous update |
| `TerminateWorkflowExecution` | History | Hard stop |
| `GetWorkflowExecutionHistory` | History | Retrieve event history |
| `ListWorkflowExecutions` | Visibility | Search/list workflows |
| `CountWorkflowExecutions` | Visibility | Count matching workflows |
| `DescribeWorkflowExecution` | Persistence | Current execution info |
| `RegisterNamespace` | Namespace | Create namespace |
| `DescribeNamespace` | Namespace | Get namespace |
| `ListNamespaces` | Namespace | List all namespaces |
| `UpdateNamespace` | Namespace | Update namespace config |
| `CreateSchedule` | History/Worker | Create a schedule |
| `DescribeSchedule` | History/Worker | Get schedule info |
| `ListSchedules` | Visibility | List schedules |
| `DeleteSchedule` | History/Worker | Delete schedule |
| `PatchSchedule` | History/Worker | Trigger/pause/unpause |

### HTTP (port 8243)

- `GET /health` — liveness
- `GET /ready` — readiness (checks DB connectivity)
- `POST /api/v1/...` — REST gateway via grpc-gateway (mirrors gRPC RPCs)

---

## Workflow Execution State Machine

The core invariant: workflow state changes only by appending events and updating mutable state in a single atomic DB write.

```
CREATED → RUNNING → COMPLETED
                 → FAILED
                 → TIMED_OUT
                 → CANCELLED
                 → TERMINATED
                 → CONTINUED_AS_NEW
```

### Activity lifecycle within a workflow:
```
SCHEDULED → STARTED → COMPLETED
                    → FAILED
                    → TIMED_OUT
                    → CANCEL_REQUESTED → CANCELLED
```

### Commands a worker can return in `RespondWorkflowTaskCompleted`:

| Command | Effect |
|---------|--------|
| `ScheduleActivityTask` | Write ActivityInfo, enqueue transfer task to Matching |
| `RequestCancelActivityTask` | Write cancel request; if STARTED, deliver cancel to heartbeat |
| `StartTimer` | Write TimerInfo, enqueue timer task |
| `CancelTimer` | Remove TimerInfo |
| `CompleteWorkflowExecution` | Append CompleteEvent, set status=Completed |
| `FailWorkflowExecution` | Append FailEvent, set status=Failed |
| `CancelWorkflowExecution` | Append CancelEvent, set status=Cancelled |
| `ContinueAsNewWorkflowExecution` | Close run, open new run atomically |
| `StartChildWorkflowExecution` | Write ChildInfo, enqueue start transfer task |
| `SignalExternalWorkflowExecution` | Enqueue signal transfer task |
| `UpsertWorkflowSearchAttributes` | Update visibility record |
| `ModifyWorkflowProperties` | Update memo |
| `ProtocolMessageCommand` | Update handler invocation (for Workflow Updates) |

---

## Build & Test

```bash
cd tempiex/

# Generate proto stubs (requires buf)
buf generate

# Build server binary
go build -o bin/tempiex-server ./cmd/server

# Build migration tool
go build -o bin/tempiex-tool ./cmd/tools

# Unit tests
go test ./...

# Integration tests (requires PostgreSQL on localhost:5432)
go test -tags integration ./tests/integration/...

# Lint
golangci-lint run ./...

# Schema migration
./bin/tempiex-tool --plugin postgres12 \
  --db-connect "postgres://tempiex:tempiex@localhost:5432/tempiex" \
  create-default-db

./bin/tempiex-tool --plugin postgres12 \
  --db-connect "postgres://tempiex:tempiex@localhost:5432/tempiex_visibility" \
  create-visibility-db
```

---

## SDK Clients & Workers

### Worker vs. Internal Worker Service

| | Internal Worker Service (:8139) | Application Worker (external) |
|-|---------------------------------|-------------------------------|
| What it is | One of the four Tempiex server services | A separate process you build using a Tempiex SDK |
| What it runs | System workflows: schedules, archival | Your workflow and activity code |
| Who manages it | Auto-started with `tempiex-server` | You build, deploy, and scale it |

### Supported SDKs

Tempiex provides its own SDKs. Third-party Temporal SDKs are **not** supported — they speak a different wire protocol.

| Language | Package | Status |
|----------|---------|--------|
| Python | `tempiex` (`sdk-python/` in this repo) | Phase 7 — active |
| Go | `github.com/tempiex/sdk-go` | Planned |
| TypeScript | `@tempiex/sdk` | Planned |

### Key Rule

**You always need at least one external worker process** to execute workflows and activities. The Tempiex server only routes and stores state — it does not execute your workflow or activity code.

---

## Authentication

| Method | Use Case |
|--------|----------|
| mTLS | Service-to-service (internal) |
| JWT / JWKS | SDK clients and workers |
| API keys | Programmatic/scripted access |
| No-auth | Dev mode |

---

## Observability

All metrics exported to OTel Collector via OTLP gRPC (`:4317`). No Prometheus client in the binary.

| Metric | Type | Description |
|--------|------|-------------|
| `workflow_start_count` | Counter | Workflows started |
| `workflow_complete_count` | Counter | Workflows completed |
| `workflow_failed_count` | Counter | Workflows failed |
| `activity_schedule_latency` | Histogram | Time from schedule to start |
| `task_schedule_to_start_latency` | Histogram | Task queue dispatch latency |
| `persistence_latency` | Histogram | DB operation duration |
| `frontend_request_count` | Counter | Total RPC requests by method |
| `frontend_error_count` | Counter | RPC errors by method + code |
| `membership_changed_count` | Counter | Ring membership changes |

Traces: per-RPC spans on Frontend; per-workflow-task spans on History; propagated via gRPC metadata.

---

## Dependencies (key)

```
go.uber.org/fx v1.24+
google.golang.org/grpc v1.80+
google.golang.org/protobuf v1.36+
github.com/grpc-ecosystem/grpc-gateway/v2
github.com/rs/zerolog v1.35+
go.opentelemetry.io/otel v1.44+
go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc
go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc
github.com/jackc/pgx/v5 v5.10+
github.com/golang-jwt/jwt/v4
```

---

## Resource Requirements

| Deployment | CPU | Memory | Storage |
|------------|-----|--------|---------|
| Development | 2 cores | 4 GB | 10 GB |
| Small Prod | 8 cores | 16 GB | 100 GB |
| Large Prod | 32+ cores | 64+ GB | 1+ TB |

---

## Implementation Notes

- **Clean-room**: No `go.temporal.io/server` source is used. The reference source at `~/app/temporal/temporal` is read-only design material — state machine invariants, schema structure, and error semantics.
- Proto definitions are written fresh in `api/proto/tempiex/api/*/v1/` using the Temporal reference as a structural guide. All package names are `tempiex.api.*`. No compatibility with the Temporal wire protocol is required or maintained.
- Tempiex SDKs (`sdk-python/`, and future SDKs) are the only supported clients. Third-party Temporal SDKs are not supported.
- Use `go.uber.org/fx` modules for every service. All wiring in `cmd/server/main.go`.
- OTel SDK initialized once at startup; `TracerProvider` and `MeterProvider` shared via fx.
- Configuration: YAML with `TEMPIEX_*` env-var overrides.
- `numHistoryShards` cannot be changed after first DB write — choose power of 2 (4 for dev, 512+ for large prod).
- History engine acquires a per-shard mutex before any state mutation. No cross-shard transactions.
- gRPC interceptor chain (outermost to innermost): OTel → Auth → RateLimit → Logging → handler
