# 05 — CLI: Tempiex CLI Specification

## Overview

**Component**: Tempiex CLI  
**Folder**: `cli/`  
**Language**: TypeScript (strict)  
**Runtime**: Node.js 22+  
**TUI Framework**: [Ink v5](https://github.com/vadimdemedes/ink) — React components rendered to terminal  
**Package Manager**: pnpm  
**Implementation Priority**: 5 — Depends on tempiex gRPC API + server HTTP API

## Purpose

`cli/` provides a terminal interface for interacting with Tempiex. It is a TypeScript pnpm package that renders TUI components using Ink (React for CLIs).

Two interaction modes:

- **Command mode** — `tempiex workflow list` runs a command, prints output, exits.
- **Interactive TUI mode** — `tempiex` with no arguments (or `tempiex ui`) launches a full keyboard-navigable dashboard rendered as a React component tree via Ink.

All command output is React components. Table views, status badges, spinners, and error panels are reusable Ink components. This means the interactive TUI and the command output share the exact same UI component library.

### Scope

Only commands relevant to day-to-day Tempiex usage are included. Excluded from this CLI (available via the Tempiex Go admin tool instead):

| Excluded | Reason |
|----------|--------|
| `activity` (standalone) | Not part of core Tempiex workflow model |
| `batch` | Advanced bulk operation management |
| `nexus` | Separate feature surface |
| `worker deployment` / build IDs | Advanced versioning; tooling-only |
| `search-attribute create/remove` | Admin schema operations |
| `workflow fix-history-json` | Internal tooling |
| `workflow update-options` (versioning overrides) | Advanced pinning |

---

## Architecture

```
cli/
│
│  tempiex [command]
│     │
│     ▼
│  Commander (command routing + flag parsing)
│     │
│     ▼
│  Ink render()
│     │
│     ├─ Command component (e.g. <WorkflowList />)
│     │    └─ uses React Query / SWR for data
│     │         └─ gRPC client (via server HTTP API or direct gRPC)
│     │
│     └─ Interactive TUI (<App />) — when no command given
│          ├─ <Sidebar /> (namespace selector, nav)
│          ├─ <WorkflowListPanel />
│          └─ <WorkflowDetailPanel />
│
│  Data layer
│  ├─ gRPC client  → tempiex :8133 (direct) or proxy
│  └─ HTTP client  → server  :8080 (for settings, auth status)
```

### Key design decisions

- **Ink v5** renders React component trees into the terminal. Components use hooks (`useState`, `useEffect`, custom hooks) exactly as in a browser React app.
- **Commander v12** handles command parsing. Each sub-command imports and renders its own Ink component.
- **No code generation** from a YAML spec — commands are written by hand as TypeScript.
- **gRPC-web / connect-rpc** client for browser-style gRPC calls over HTTP/1.1 (avoids native gRPC bindings in Node). Falls back to server HTTP API for all data needs.
- **Config** stored in `~/.tempiex/config.toml`. Profiles supported. Env vars override with `TEMPIEX_` prefix.

---

## Technical Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| Language | TypeScript | 5.x strict |
| Runtime | Node.js | 22+ |
| TUI | Ink | 5.x |
| React | React | 18+ |
| CLI framework | Commander | 12.x |
| gRPC client | @connectrpc/connect + @connectrpc/connect-node | 2.x |
| Config | TOML (`@iarna/toml`) | 3.x |
| Schema validation | Zod | 3.x |
| Date/Time | date-fns | 3.x |
| Build | tsup | 8.x |
| Test | Vitest | 2.x |
| Lint | ESLint + @typescript-eslint | 9.x |
| Formatter | Prettier | 3.x |

---

## Folder Structure

```
cli/
├── src/
│   ├── main.tsx               # entry point: commander root + ink render()
│   ├── app/
│   │   └── App.tsx            # interactive TUI root component
│   │
│   ├── commands/              # one file per command group
│   │   ├── workflow/
│   │   │   ├── index.ts       # registers sub-commands on Commander program
│   │   │   ├── WorkflowList.tsx
│   │   │   ├── WorkflowDescribe.tsx
│   │   │   ├── WorkflowStart.tsx
│   │   │   ├── WorkflowExecute.tsx
│   │   │   ├── WorkflowSignal.tsx
│   │   │   ├── WorkflowQuery.tsx
│   │   │   ├── WorkflowUpdate.tsx
│   │   │   ├── WorkflowTerminate.tsx
│   │   │   ├── WorkflowCancel.tsx
│   │   │   ├── WorkflowReset.tsx
│   │   │   ├── WorkflowShow.tsx
│   │   │   ├── WorkflowTrace.tsx
│   │   │   └── WorkflowCount.tsx
│   │   ├── namespace/
│   │   │   ├── index.ts
│   │   │   ├── NamespaceList.tsx
│   │   │   ├── NamespaceDescribe.tsx
│   │   │   ├── NamespaceCreate.tsx
│   │   │   ├── NamespaceUpdate.tsx
│   │   │   └── NamespaceDelete.tsx
│   │   ├── schedule/
│   │   │   ├── index.ts
│   │   │   ├── ScheduleList.tsx
│   │   │   ├── ScheduleDescribe.tsx
│   │   │   ├── ScheduleCreate.tsx
│   │   │   ├── ScheduleUpdate.tsx
│   │   │   ├── ScheduleDelete.tsx
│   │   │   ├── ScheduleTrigger.tsx
│   │   │   ├── SchedulePause.tsx
│   │   │   └── ScheduleUnpause.tsx
│   │   ├── taskqueue/
│   │   │   ├── index.ts
│   │   │   ├── TaskQueueDescribe.tsx
│   │   │   ├── TaskQueueListPartition.tsx
│   │   │   ├── TaskQueueConfigGet.tsx
│   │   │   └── TaskQueueConfigSet.tsx
│   │   ├── cluster/
│   │   │   ├── index.ts
│   │   │   ├── ClusterDescribe.tsx
│   │   │   └── ClusterHealth.tsx
│   │   ├── server/
│   │   │   ├── index.ts
│   │   │   └── ServerStartDev.tsx
│   │   ├── proxy/
│   │   │   ├── index.ts
│   │   │   ├── ProxyStart.tsx
│   │   │   └── ProxyValidate.tsx
│   │   ├── config/
│   │   │   ├── index.ts
│   │   │   ├── ConfigGet.tsx
│   │   │   ├── ConfigSet.tsx
│   │   │   ├── ConfigList.tsx
│   │   │   └── ConfigDelete.tsx
│   │   └── pi/
│   │       ├── index.ts
│   │       ├── PiStatus.tsx
│   │       ├── PiSessionList.tsx
│   │       ├── PiSessionDescribe.tsx
│   │       ├── PiSessionCancel.tsx
│   │       ├── PiSessionEvents.tsx
│   │       ├── PiToolList.tsx
│   │       └── PiStart.tsx
│   │
│   ├── components/            # shared Ink UI components
│   │   ├── Table.tsx          # data table (columns, rows, borders)
│   │   ├── StatusBadge.tsx    # colored workflow status badge
│   │   ├── Spinner.tsx        # loading indicator
│   │   ├── ErrorPanel.tsx     # formatted error display
│   │   ├── JsonViewer.tsx     # syntax-highlighted JSON
│   │   ├── Confirm.tsx        # y/n confirmation prompt
│   │   ├── Input.tsx          # text input field
│   │   └── Pagination.tsx     # page indicator + nav hint
│   │
│   ├── hooks/                 # data-fetching hooks (connect-rpc + SSE)
│   │   ├── useWorkflows.ts
│   │   ├── useWorkflowDetail.ts
│   │   ├── useEventHistory.ts
│   │   ├── useNamespaces.ts
│   │   ├── useSchedules.ts
│   │   ├── useCluster.ts
│   │   ├── useTaskQueue.ts
│   │   ├── usePiSessions.ts
│   │   ├── usePiSessionEvents.ts  # SSE stream hook
│   │   └── usePiTools.ts
│   │
│   ├── client/
│   │   ├── grpc.ts            # connect-rpc client factory
│   │   ├── http.ts            # server HTTP API client
│   │   └── pi.ts              # pi agent HTTP + SSE client
│   │
│   ├── config/
│   │   ├── loader.ts          # load/save ~/.tempiex/config.toml
│   │   └── types.ts           # Config, Profile, zod schemas
│   │
│   └── utils/
│       ├── output.ts          # --output flag: json | yaml | text
│       ├── datetime.ts        # duration formatting, relative time
│       └── payload.ts         # base64, JSON input helpers
│
├── test/
│   └── unit/
│       ├── config.test.ts
│       ├── output.test.ts
│       └── payload.test.ts
│
├── package.json
├── tsconfig.json
├── tsup.config.ts
├── vitest.config.ts
└── eslint.config.mjs
```

---

## Global Flags

Available on every command.

| Flag | Short | Default | Env Var | Description |
|------|-------|---------|---------|-------------|
| `--address` | | `localhost:8133` | `TEMPIEX_ADDRESS` | Tempiex gRPC endpoint |
| `--namespace` | `-n` | `default` | `TEMPIEX_NAMESPACE` | Namespace |
| `--api-key` | | — | `TEMPIEX_API_KEY` | API key |
| `--tls` | | `false` | `TEMPIEX_TLS` | Enable TLS |
| `--tls-cert-path` | | — | `TEMPIEX_TLS_CLIENT_CERT_PATH` | mTLS client cert |
| `--tls-key-path` | | — | `TEMPIEX_TLS_CLIENT_KEY_PATH` | mTLS client key |
| `--tls-ca-path` | | — | `TEMPIEX_TLS_SERVER_CA_CERT_PATH` | CA cert |
| `--codec-endpoint` | | — | `TEMPIEX_CODEC_ENDPOINT` | Codec server URL |
| `--output` | `-o` | `text` | | Output format: `text` \| `json` \| `yaml` |
| `--color` | | `auto` | | Color output: `always` \| `never` \| `auto` |
| `--profile` | | `default` | `TEMPIEX_PROFILE` | Config profile |
| `--log-level` | | `never` | | `debug` \| `info` \| `warn` \| `error` \| `never` |

Flag resolution order: CLI flag → env var → config profile → built-in default.

---

## Command Reference

### `tempiex workflow`

Manage Workflow Executions.

---

#### `tempiex workflow list`

List Workflow Executions.

```bash
tempiex workflow list
tempiex workflow list --query 'WorkflowType="OrderWorkflow" AND ExecutionStatus="Running"'
tempiex workflow list --limit 50 --output json
```

Interactive TUI: renders a navigable, auto-refreshing table. Arrow keys to move, Enter to open detail.

| Flag | Default | Description |
|------|---------|-------------|
| `--query` / `-q` | — | SQL-like List Filter |
| `--limit` | 50 | Max executions to display |
| `--archived` | false | Show archived executions |

---

#### `tempiex workflow describe`

Show Workflow Execution details.

```bash
tempiex workflow describe --workflow-id my-wf
tempiex workflow describe --workflow-id my-wf --run-id abc123
```

| Flag | Required | Description |
|------|----------|-------------|
| `--workflow-id` / `-w` | Yes | Workflow ID |
| `--run-id` / `-r` | No | Run ID |

---

#### `tempiex workflow start`

Start a Workflow Execution. Prints workflow ID and run ID then exits.

```bash
tempiex workflow start \
  --workflow-id my-wf \
  --type OrderWorkflow \
  --task-queue orders \
  --input '{"orderId": "123"}'
```

| Flag | Required | Description |
|------|----------|-------------|
| `--workflow-id` / `-w` | No | Workflow ID (generated if omitted) |
| `--type` | Yes | Workflow type name |
| `--task-queue` / `-t` | Yes | Task queue name |
| `--input` / `-i` | No | JSON input. Repeatable |
| `--input-file` | No | Path to JSON input file |
| `--execution-timeout` | No | Max Workflow Execution time |
| `--run-timeout` | No | Max single run time |
| `--task-timeout` | No | Max Workflow Task time |
| `--id-reuse-policy` | No | `AllowDuplicate` \| `RejectDuplicate` \| `TerminateIfRunning` |
| `--memo` | No | `KEY=VALUE` memo pairs. Repeatable |
| `--search-attribute` | No | `KEY=VALUE` search attributes. Repeatable |

---

#### `tempiex workflow execute`

Start a Workflow and block until it completes. Prints the result.

```bash
tempiex workflow execute \
  --type OrderWorkflow \
  --task-queue orders \
  --input '{"orderId": "123"}'
```

Same flags as `workflow start`, plus:

| Flag | Default | Description |
|------|---------|-------------|
| `--detailed` | false | Show events as sections rather than a summary table |

Interactive TUI: renders a live progress panel showing event history as the workflow runs.

---

#### `tempiex workflow show`

Display Event History for a Workflow Execution.

```bash
tempiex workflow show --workflow-id my-wf
tempiex workflow show --workflow-id my-wf --follow
tempiex workflow show --workflow-id my-wf --output json
```

| Flag | Default | Description |
|------|---------|-------------|
| `--workflow-id` / `-w` | — | **Required** |
| `--run-id` / `-r` | — | Run ID |
| `--follow` / `-f` | false | Follow execution progress in real time |
| `--detailed` | false | Render events as expanded sections |
| `--reverse` | false | Newest events first |

Interactive TUI: expandable event rows with inline JSON payload viewer.

---

#### `tempiex workflow signal`

Send an asynchronous Signal to a running Workflow.

```bash
tempiex workflow signal \
  --workflow-id my-wf \
  --name approve \
  --input '{"approvedBy": "alice"}'
```

| Flag | Required | Description |
|------|----------|-------------|
| `--workflow-id` / `-w` | Yes | Workflow ID |
| `--run-id` / `-r` | No | Run ID |
| `--name` | Yes | Signal name |
| `--input` / `-i` | No | JSON input. Repeatable |
| `--input-file` | No | Path to JSON input file |

---

#### `tempiex workflow query`

Retrieve Workflow state via a Query.

```bash
tempiex workflow query \
  --workflow-id my-wf \
  --name getStatus
```

| Flag | Required | Description |
|------|----------|-------------|
| `--workflow-id` / `-w` | Yes | Workflow ID |
| `--run-id` / `-r` | No | Run ID |
| `--name` | Yes | Query type name |
| `--input` / `-i` | No | JSON input |
| `--reject-condition` | No | `not_open` \| `not_completed_cleanly` |

---

#### `tempiex workflow update`

Send a synchronous Update to a Workflow and wait for it to complete.

```bash
tempiex workflow update execute \
  --workflow-id my-wf \
  --name applyDiscount \
  --input '{"pct": 10}'
```

Sub-commands:

| Sub-command | Description |
|-------------|-------------|
| `execute` | Send Update, wait for completion result |
| `start` | Send Update, wait for acceptance only |
| `result` | Wait for a previously sent Update to complete |
| `describe` | Show status of a specific Update |

**`execute` / `start` flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--workflow-id` / `-w` | Yes | Workflow ID |
| `--run-id` / `-r` | No | Run ID |
| `--name` | Yes | Update handler name |
| `--input` / `-i` | No | JSON input |
| `--update-id` | No | Idempotency key (generated if omitted) |
| `--wait-for-stage` | `start` only | Must be `accepted` |

**`result` / `describe` flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--workflow-id` / `-w` | Yes | Workflow ID |
| `--run-id` / `-r` | No | Run ID |
| `--update-id` | Yes | Update ID |

---

#### `tempiex workflow terminate`

Forcefully end a Workflow Execution. Workflow code cannot respond.

```bash
tempiex workflow terminate \
  --workflow-id my-wf \
  --reason "stale order"

# Bulk via query:
tempiex workflow terminate \
  --query 'WorkflowType="OrderWorkflow" AND ExecutionStatus="Running"' \
  --reason "batch cleanup" \
  --yes
```

| Flag | Required | Description |
|------|----------|-------------|
| `--workflow-id` / `-w` | Either this or `--query` | Workflow ID |
| `--run-id` / `-r` | No | Run ID |
| `--query` / `-q` | Either this or `--workflow-id` | Bulk filter |
| `--reason` | No | Reason (defaults to current user) |
| `--yes` / `-y` | No | Skip confirmation prompt (bulk only) |

---

#### `tempiex workflow cancel`

Request graceful cancellation (allows cleanup code to run).

```bash
tempiex workflow cancel --workflow-id my-wf
tempiex workflow cancel --query 'ExecutionStatus="Running"' --yes
```

| Flag | Required | Description |
|------|----------|-------------|
| `--workflow-id` / `-w` | Either this or `--query` | Workflow ID |
| `--run-id` / `-r` | No | Run ID |
| `--query` / `-q` | Either this or `--workflow-id` | Bulk filter |
| `--reason` | No | Reason |
| `--yes` / `-y` | No | Skip confirmation (bulk only) |

---

#### `tempiex workflow reset`

Reset Workflow Execution history to an earlier event.

```bash
tempiex workflow reset \
  --workflow-id my-wf \
  --event-id 5 \
  --reason "retry after bug fix"

tempiex workflow reset \
  --workflow-id my-wf \
  --type LastWorkflowTask \
  --reason "retry"
```

| Flag | Required | Description |
|------|----------|-------------|
| `--workflow-id` / `-w` | Yes | Workflow ID |
| `--run-id` / `-r` | No | Run ID |
| `--event-id` / `-e` | Either this or `--type` | Event ID to reset to |
| `--type` / `-t` | Either this or `--event-id` | `FirstWorkflowTask` \| `LastWorkflowTask` \| `LastContinuedAsNew` |
| `--reason` | Yes | Required reason |
| `--reapply-exclude` | No | Event types to skip on re-apply: `Signal` \| `Update` \| `All` |

---

#### `tempiex workflow count`

Count Workflow Executions matching a filter.

```bash
tempiex workflow count
tempiex workflow count --query 'ExecutionStatus="Running"'
```

| Flag | Required | Description |
|------|----------|-------------|
| `--query` / `-q` | No | SQL-like List Filter |

---

#### `tempiex workflow trace`

Display live Workflow Execution tree with child workflows.

```bash
tempiex workflow trace --workflow-id my-wf
```

| Flag | Default | Description |
|------|---------|-------------|
| `--workflow-id` / `-w` | — | **Required** |
| `--run-id` / `-r` | — | Run ID |
| `--depth` | -1 | Child Workflow fetch depth (-1 = unlimited) |
| `--fold` | — | Fold child workflows with these statuses. Repeatable |
| `--no-fold` | false | Disable folding; show all children |

Interactive TUI: renders a live tree view with status indicators, auto-refreshing.

---

### `tempiex namespace`

Manage Namespaces.

---

#### `tempiex namespace list`

```bash
tempiex namespace list
tempiex namespace list --output json
```

---

#### `tempiex namespace describe`

```bash
tempiex namespace describe --namespace default
tempiex namespace describe --namespace-id abc-123
```

| Flag | Description |
|------|-------------|
| `--namespace-id` | Describe by ID instead of name |

---

#### `tempiex namespace create`

```bash
tempiex namespace create \
  --namespace my-ns \
  --retention 7d \
  --description "Production namespace"
```

| Flag | Default | Description |
|------|---------|-------------|
| `--namespace` / `-n` | — | **Required** |
| `--retention` | `72h` | Closed Workflow retention duration |
| `--description` | — | Description |
| `--email` | — | Owner email |
| `--history-archival-state` | `disabled` | `disabled` \| `enabled` |
| `--history-uri` | — | Archival URI |
| `--visibility-archival-state` | `disabled` | `disabled` \| `enabled` |
| `--visibility-uri` | — | Visibility archival URI |

---

#### `tempiex namespace update`

```bash
tempiex namespace update \
  --namespace my-ns \
  --retention 30d
```

Same flags as `create` (all optional).

---

#### `tempiex namespace delete`

```bash
tempiex namespace delete --namespace my-ns
tempiex namespace delete --namespace my-ns --yes
```

| Flag | Default | Description |
|------|---------|-------------|
| `--yes` / `-y` | false | Skip confirmation |

---

### `tempiex schedule`

Manage Schedules (periodic Workflow starts).

---

#### `tempiex schedule list`

```bash
tempiex schedule list
tempiex schedule list --output json
```

---

#### `tempiex schedule describe`

```bash
tempiex schedule describe --schedule-id nightly-report
```

---

#### `tempiex schedule create`

```bash
tempiex schedule create \
  --schedule-id nightly-report \
  --cron "0 2 * * *" \
  --workflow-type ReportWorkflow \
  --task-queue reports \
  --input '{"format": "pdf"}'
```

| Flag | Required | Description |
|------|----------|-------------|
| `--schedule-id` | Yes | Unique schedule ID |
| `--cron` | Yes | Cron expression |
| `--workflow-type` | Yes | Workflow type to start |
| `--task-queue` | Yes | Task queue |
| `--input` / `-i` | No | JSON input |
| `--workflow-id-prefix` | No | Prefix for generated Workflow IDs |
| `--overlap-policy` | `Skip` | `Skip` \| `BufferOne` \| `BufferAll` \| `CancelOther` \| `TerminateOther` \| `AllowAll` |
| `--start-time` | No | Schedule active start (RFC3339) |
| `--end-time` | No | Schedule active end (RFC3339) |
| `--jitter` | No | Max random delay per trigger |
| `--tz` | No | Timezone for cron evaluation |
| `--memo` | No | `KEY=VALUE` memo. Repeatable |
| `--note` | No | Initial schedule note |
| `--paused` | false | Create in paused state |

---

#### `tempiex schedule update`

Update an existing schedule. Only provided flags are changed.

```bash
tempiex schedule update \
  --schedule-id nightly-report \
  --cron "0 3 * * *"
```

Same flags as `create` (all optional, `--schedule-id` required).

---

#### `tempiex schedule delete`

```bash
tempiex schedule delete --schedule-id nightly-report
tempiex schedule delete --schedule-id nightly-report --yes
```

| Flag | Default | Description |
|------|---------|-------------|
| `--yes` / `-y` | false | Skip confirmation |

---

#### `tempiex schedule trigger`

Immediately trigger one run of a schedule.

```bash
tempiex schedule trigger --schedule-id nightly-report
```

| Flag | Default | Description |
|------|---------|-------------|
| `--overlap-policy` | `AllowAll` | Override overlap policy for this trigger |

---

#### `tempiex schedule pause`

```bash
tempiex schedule pause --schedule-id nightly-report --note "maintenance window"
```

| Flag | Description |
|------|-------------|
| `--note` | Reason for pause |

---

#### `tempiex schedule unpause`

```bash
tempiex schedule unpause --schedule-id nightly-report --note "maintenance done"
```

| Flag | Description |
|------|-------------|
| `--note` | Reason for unpause |

---

### `tempiex taskqueue`

Inspect and configure Task Queues.

---

#### `tempiex taskqueue describe`

Show active workers and statistics for a Task Queue.

```bash
tempiex taskqueue describe --task-queue orders
tempiex taskqueue describe --task-queue orders --task-queue-type activity
tempiex taskqueue describe --task-queue orders --report-config
tempiex taskqueue describe --task-queue orders --output json
```

| Flag | Required | Description |
|------|----------|-------------|
| `--task-queue` / `-t` | Yes | Task Queue name |
| `--task-queue-type` | No | Filter by type: `workflow` \| `activity`. All types if omitted |
| `--report-reachability` | No | Include task reachability info per Build ID |
| `--report-config` | No | Include rate limit and fairness config |
| `--disable-stats` | No | Omit task queue statistics (faster) |

---

#### `tempiex taskqueue list-partition`

List Task Queue partitions with their assigned matching nodes. Useful for debugging cluster-level routing.

```bash
tempiex taskqueue list-partition --task-queue orders
```

| Flag | Required | Description |
|------|----------|-------------|
| `--task-queue` / `-t` | Yes | Task Queue name |

---

#### `tempiex taskqueue config get`

Retrieve the current rate limit and fairness configuration for a Task Queue.

```bash
tempiex taskqueue config get --task-queue orders --task-queue-type activity
```

| Flag | Required | Description |
|------|----------|-------------|
| `--task-queue` / `-t` | Yes | Task Queue name |
| `--task-queue-type` | Yes | `workflow` \| `activity` |

---

#### `tempiex taskqueue config set`

Update rate limit and fairness weight configuration for a Task Queue.

```bash
tempiex taskqueue config set \
  --task-queue orders \
  --task-queue-type activity \
  --queue-rps-limit 100.0

# Unset limit:
tempiex taskqueue config set \
  --task-queue orders \
  --task-queue-type activity \
  --queue-rps-limit default

# Fairness weights:
tempiex taskqueue config set \
  --task-queue orders \
  --task-queue-type activity \
  --fairness-key-weight high-priority=2.0 \
  --fairness-key-weight low-priority=0.5
```

| Flag | Required | Description |
|------|----------|-------------|
| `--task-queue` / `-t` | Yes | Task Queue name |
| `--task-queue-type` | Yes | `workflow` \| `activity` |
| `--queue-rps-limit` | No | Rate limit in RPS. Pass `default` to unset |
| `--queue-rps-limit-reason` | No | Reason for the change (audit log) |
| `--fairness-key-weight` | No | `key=weight` weight overrides. Repeatable. Pass `key=default` to unset one |
| `--fairness-key-weight-clear-all` | No | Remove all fairness key weight overrides |

---

### `tempiex cluster`

Inspect the Tempiex cluster.

---

#### `tempiex cluster describe`

```bash
tempiex cluster describe
tempiex cluster describe --detail
```

Shows: cluster name, persistence store type, visibility store, shard count, version.

| Flag | Default | Description |
|------|---------|-------------|
| `--detail` | false | Include shard count and version details |

---

#### `tempiex cluster health`

```bash
tempiex cluster health
```

Exits 0 if healthy, non-zero if any service is unhealthy. Suitable for health checks and scripts.

---

### `tempiex server`

---

#### `tempiex server start-dev`

Start a local single-node Tempiex development server (PostgreSQL-backed). Blocks until stopped.

```bash
tempiex server start-dev
tempiex server start-dev --port 8133 --ui-port 8080
tempiex server start-dev --db-connect "postgres://tempiex:tempiex@localhost:5432/tempiex" --namespace prod-ns --namespace staging-ns
```

The dev server automatically:
- Runs PostgreSQL schema migrations on startup
- Creates the listed namespaces on startup
- Starts the UI server on `--ui-port` (serves the web UI)
- Prints connection instructions to stdout

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `8133` | gRPC port |
| `--ui-port` | `8080` | UI server port |
| `--ui` | true | Start UI server alongside |
| `--namespace` | `["default"]` | Namespaces to create. Repeatable |
| `--db-connect` | `postgres://tempiex:tempiex@localhost:5432/tempiex` | PostgreSQL connection string |
| `--log-level` | `info` | Server log level |
| `--ip` | `127.0.0.1` | Bind address |

Interactive TUI: renders a status panel showing service health, active namespaces, and a live workflow count. `q` or `Ctrl+C` stops the server.

---

### `tempiex proxy`

---

#### `tempiex proxy start`

Start the Tempiex gRPC proxy. Blocks until stopped.

```bash
tempiex proxy start --config ./proxy.yaml
```

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `./tempiex-proxy.yaml` | Path to proxy YAML config |

---

#### `tempiex proxy validate`

Validate a proxy config file without starting.

```bash
tempiex proxy validate --config ./proxy.yaml
```

Exits 0 on valid config, 1 with a descriptive error on invalid config.

---

### `tempiex config`

Manage CLI configuration profiles stored in `~/.tempiex/config.toml`.

Config file location: `$HOME/.config/tempiex/config.toml` (Unix/macOS) / `%AppData%\tempiex\config.toml` (Windows). Override with `TEMPIEX_CONFIG_FILE`.

---

#### `tempiex config list`

List all profile names.

```bash
tempiex config list
```

---

#### `tempiex config get`

Show config properties for a profile.

```bash
tempiex config get
tempiex config get --profile prod
tempiex config get --profile prod --prop address
```

| Flag | Description |
|------|-------------|
| `--prop` / `-p` | Specific property to show |

---

#### `tempiex config set`

Set a config property.

```bash
tempiex config set --profile prod --prop address --value cloud.example.com:8133
tempiex config set --prop namespace --value my-ns
```

| Flag | Required | Description |
|------|----------|-------------|
| `--prop` / `-p` | Yes | Property name |
| `--value` / `-v` | Yes | Property value |

---

#### `tempiex config delete`

Delete a property or an entire profile.

```bash
tempiex config delete --prop tls-cert-path
tempiex config delete --profile staging
```

| Flag | Description |
|------|-------------|
| `--prop` / `-p` | Property to delete. If omitted, deletes the entire profile |

---

### `tempiex pi`

Manage Pi agent harness instances. The `pi` sub-commands talk to the Pi HTTP API (default `http://localhost:8090`).

Global pi flag (applies to all `tempiex pi` sub-commands):

| Flag | Default | Env Var | Description |
|------|---------|---------|-------------|
| `--pi-address` | `http://localhost:8090` | `TEMPIEX_PI_ADDRESS` | Pi agent HTTP API base URL |

---

#### `tempiex pi status`

Show Pi agent health and Tempiex connection status.

```bash
tempiex pi status
tempiex pi status --pi-address http://pi-host:8090
```

Prints:

```
Pi agent: running
Tempiex:  connected (localhost:8133)
Sessions: 3 active
```

Exits non-zero if Pi is unreachable or Tempiex is disconnected.

---

#### `tempiex pi session list`

List tracked sessions (active and recent completed).

```bash
tempiex pi session list
tempiex pi session list --status active
tempiex pi session list --output json
```

| Flag | Default | Description |
|------|---------|-------------|
| `--status` | — | Filter: `active` \| `completed` \| `failed` \| `cancelled` |
| `--limit` | 50 | Max sessions to show |

Interactive TUI: navigable table; Enter opens session detail.

---

#### `tempiex pi session describe`

Show full details for a session: workflow ID, run ID, task queue, tools used, start/end time.

```bash
tempiex pi session describe --session-id abc123
```

| Flag | Required | Description |
|------|----------|-------------|
| `--session-id` | Yes | Session ID |

---

#### `tempiex pi session cancel`

Cancel an active session. Sends a cancel signal to the corresponding Tempiex workflow.

```bash
tempiex pi session cancel --session-id abc123
tempiex pi session cancel --session-id abc123 --yes
```

| Flag | Default | Description |
|------|---------|-------------|
| `--session-id` | — | **Required** |
| `--yes` / `-y` | false | Skip confirmation |

---

#### `tempiex pi session events`

Stream SSE events for a session. Replays historical events then follows live.

```bash
tempiex pi session events --session-id abc123
tempiex pi session events --session-id abc123 --follow
```

| Flag | Default | Description |
|------|---------|-------------|
| `--session-id` | — | **Required** |
| `--follow` / `-f` | true | Keep streaming after history replay |
| `--output` | `text` | `text` \| `json` |

Example output (`text`):

```
[14:02:01] execution_start   shell          {"cmd":"ls -la /tmp"}
[14:02:01] execution_update  shell          {"stdout":"total 12\n"}
[14:02:01] execution_end     shell          {"exitCode":0}
```

Interactive TUI: live scrolling event log panel with color-coded event types.

---

#### `tempiex pi tool list`

List tools registered with the Pi agent.

```bash
tempiex pi tool list
tempiex pi tool list --output json
```

Prints: tool name, description, execution mode.

---

#### `tempiex pi start`

Start a Pi agent daemon (foreground). Convenience wrapper around the `pi` binary with config.

```bash
tempiex pi start
tempiex pi start --config ./pi.yaml
tempiex pi start --task-queue my-queue --pi-port 8090
```

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `./tempiex-pi.yaml` | Pi YAML config file |
| `--task-queue` | `pi-default` | Task queue to poll (overrides config) |
| `--pi-port` | `8090` | Pi HTTP API port |
| `--tempiex-address` | `localhost:8133` | Tempiex gRPC address |
| `--namespace` | `default` | Namespace |
| `--log-level` | `info` | `debug` \| `info` \| `warn` \| `error` |

Interactive TUI: renders a live panel showing:
- Tempiex connection status
- Active task queues and poll counts
- Recent activity executions (tool name, status, duration)
- Current sessions

`Ctrl+C` stops the agent gracefully (drains in-flight activities).

---

## Interactive TUI Mode

Running `tempiex` with no arguments launches the full interactive TUI.

```
tempiex
```

The TUI is a full Ink (React) application:

```
┌─ Tempiex ───────────────────────────────────────────────┐
│  Namespace: default ▾         Cluster: healthy ●        │
├──────────────────────────────────────────────────────────┤
│ > Workflows    Schedules    Namespaces    Cluster    Pi        │
├──────────────────────────────────────────────────────────┤
│ Workflow ID          Type              Status    Started │
│ order-9f2a1         OrderWorkflow     Running   2m ago  │
│ order-7b3c4         OrderWorkflow     Completed 5m ago  │
│ report-daily        ReportWorkflow    Running   1h ago  │
├──────────────────────────────────────────────────────────┤
│ [↑↓] Navigate  [Enter] Detail  [/] Filter  [q] Quit     │
└──────────────────────────────────────────────────────────┘
```

Navigation:
- `Tab` / `Shift+Tab` — switch top-level sections
- `↑` / `↓` — move row focus
- `Enter` — open detail panel
- `/` — open filter / search input
- `r` — refresh
- `Esc` — go back / close panel
- `q` — quit

Each section (Workflows, Schedules, Namespaces, Cluster) is an independent React component that manages its own data-fetching and keyboard focus. The detail panel slides in as an overlay.

---

## Output Formats

All commands support `--output` / `-o`:

| Value | Description |
|-------|-------------|
| `text` (default) | Human-readable Ink-rendered table or structured output |
| `json` | JSON object or array; suitable for `jq` piping |
| `yaml` | YAML |

In `json` / `yaml` mode, Ink renders a plain text serialization and exits without any interactive UI elements.

---

## Configuration File Format

`~/.config/tempiex/config.toml`:

```toml
[default]
address = "localhost:8133"
namespace = "default"
output = "text"

[prod]
address = "prod.example.com:8133"
namespace = "prod-ns"
api-key = ""             # set via TEMPIEX_API_KEY env
tls = true
tls-ca-path = "/etc/tempiex/ca.crt"
```

---

## Build & Test

```bash
cd cli/

# Install
pnpm install

# Dev (watch + run)
pnpm dev

# Build (outputs to dist/)
pnpm build       # tsup: bundles to dist/tempiex.js + shebang

# Typecheck
pnpm typecheck   # tsc --noEmit

# Lint
pnpm lint        # eslint src/

# Tests
pnpm test        # vitest run

# Install globally from local build
npm link         # or: node dist/tempiex.js
```

Compiled output: `dist/tempiex.js` (ESM, shebang `#!/usr/bin/env node`).

---

## package.json (key fields)

```json
{
  "name": "@tempiex/cli",
  "version": "0.1.0",
  "type": "module",
  "bin": {
    "tempiex": "./dist/tempiex.js"
  },
  "scripts": {
    "dev": "tsup --watch",
    "build": "tsup",
    "typecheck": "tsc --noEmit",
    "lint": "eslint src/",
    "test": "vitest run"
  },
  "dependencies": {
    "ink": "^5",
    "react": "^18",
    "commander": "^12",
    "@connectrpc/connect": "^2",
    "@connectrpc/connect-node": "^2",
    "@bufbuild/protobuf": "^2",
    "@iarna/toml": "^3",
    "zod": "^3",
    "date-fns": "^3",
    "chalk": "^5",
    "ink-table": "^3",
    "eventsource": "^2"
  },
  "devDependencies": {
    "typescript": "^5",
    "tsup": "^8",
    "vitest": "^2",
    "@types/react": "^18",
    "eslint": "^9",
    "@typescript-eslint/eslint-plugin": "^7",
    "prettier": "^3"
  }
}
```

---

## Ink Component Design Guidelines

- Every command component accepts a `flags` prop typed with a Zod schema. Validation happens at the command entry point before `render()`.
- Components that fetch data show a `<Spinner />` while loading, `<ErrorPanel />` on failure, and the actual content on success.
- Use `useEffect` with `process.exit(0)` (or `useApp().exit()`) after a non-interactive command finishes rendering its output — this tells Ink to unmount and return control to the shell.
- Interactive components (TUI mode) never call `exit()` automatically; they wait for `q` / `Ctrl+C`.
- All color use `chalk` via the Ink `<Text color="...">` prop; respect `--color never` by checking `chalk.level === 0`.

---

## Dependencies (key)

```
ink ^5
react ^18
commander ^12
@connectrpc/connect ^2
@connectrpc/connect-node ^2
@bufbuild/protobuf ^2
@iarna/toml ^3
zod ^3
date-fns ^3
chalk ^5
ink-table ^3
eventsource ^2
```
