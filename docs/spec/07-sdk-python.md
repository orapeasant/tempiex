# 07 — sdk-python: Tempiex Python SDK Specification

## Overview

**Component**: Tempiex Python SDK  
**Folder**: `sdk-python/`  
**Language**: Python 3.10+  
**Package name**: `tempiex`  
**Package manager**: uv  
**Implementation Priority**: 7 — Depends on tempiex server (gRPC :8133)

## Purpose

`sdk-python/` is a pure Python SDK for building Tempiex workers and clients. It provides:

- **`@workflow.defn`** — decorator to define durable workflow classes
- **`@activity.defn`** — decorator to define activity functions
- **`Worker`** — polls Tempiex task queues, executes workflows and activities
- **`Client`** — connects to Tempiex, starts workflows, sends signals/queries/updates
- **`DataConverter`** — payload serialization/deserialization (JSON by default, pluggable)
- **`WorkflowEnvironment`** — in-process test environment

**No Rust bridge.** The ref SDK (`temporalio`) compiles a Rust extension via Maturin for the workflow execution core. This SDK is pure Python + asyncio + `grpcio`. It is simpler, easier to build and distribute, and sufficient for the Tempiex use case.

### What "pure Python" means for the workflow engine

The deterministic workflow execution loop is implemented in Python:
- Workflows run as Python coroutines.
- On each task, the worker replays history events through the coroutine to restore state, then executes new commands.
- I/O inside workflows is intercepted — `execute_activity`, `sleep`, `wait_condition` return awaitables backed by the event history, not real I/O.
- Activities run freely in asyncio tasks (or a thread pool for sync activities) with no replay constraints.

---

## Architecture

```
Your code                         Tempiex Server (:8133)
────────────────────────────────  ─────────────────────

@workflow.defn                     ┌─────────────────────┐
class MyWorkflow:                  │  Frontend Service   │
    async def run(self, x):        │  gRPC API           │
        return await               └────────┬────────────┘
            workflow.execute_activity(...)  │
                                           gRPC (PollWorkflowTaskQueue,
@activity.defn                             PollActivityTaskQueue,
async def process(x):                      RespondWorkflowTaskCompleted, ...)
    ...

client = await Client.connect("localhost:8133")
worker = Worker(client, task_queue="orders",
                workflows=[MyWorkflow],
                activities=[process])
await worker.run()
```

### Component Roles

| Component | Role |
|-----------|------|
| `tempiex.client.Client` | Connect to server; start/cancel/signal/query workflows; manage schedules |
| `tempiex.worker.Worker` | Poll task queues; dispatch workflow tasks and activity tasks |
| `tempiex.workflow` | Decorators and APIs usable inside workflow code |
| `tempiex.activity` | Decorators and APIs usable inside activity code |
| `tempiex.converter` | Serialize/deserialize payloads to/from protobuf `Payload` |
| `tempiex.service` | Low-level gRPC `WorkflowService` client (thin generated wrapper) |
| `tempiex.testing` | `WorkflowEnvironment` for unit and integration tests |
| `tempiex.exceptions` | `ApplicationError`, `CancelledError`, `ActivityError`, etc. |

---

## Folder Structure

```
sdk-python/
├── tempiex/
│   ├── __init__.py            # re-exports: Client, Worker, workflow, activity
│   ├── client.py              # Client, ClientConfig, WorkflowHandle, ScheduleHandle
│   ├── worker.py              # Worker, WorkerConfig
│   ├── workflow.py            # @workflow.defn + in-workflow APIs
│   ├── activity.py            # @activity.defn + in-activity APIs
│   ├── converter.py           # DataConverter, JSONPayloadConverter, PayloadCodec
│   ├── exceptions.py          # ApplicationError, CancelledError, ActivityError, etc.
│   ├── service.py             # ServiceClient (raw gRPC), TLSConfig, ConnectConfig
│   ├── testing.py             # WorkflowEnvironment (in-process + external)
│   ├── common.py              # shared dataclasses: RetryPolicy, WorkflowIDReusePolicy, etc.
│   ├── py.typed               # PEP 561 marker
│   └── _internal/
│       ├── _workflow_instance.py   # coroutine-based workflow execution + replay engine
│       ├── _workflow_task.py       # handles a single WFT: replay + new commands
│       ├── _activity_task.py       # executes one activity (async or thread)
│       ├── _poll_loop.py           # asyncio polling loop for workflow + activity tasks
│       ├── _command.py             # builds gRPC Command protos from Python operations
│       └── _proto_helpers.py       # Payload ↔ Python object conversion helpers
│
├── tempiex/api/               # generated protobuf stubs (grpc tools output)
│   └── tempiex/
│       ├── api/
│       │   ├── workflowservice/v1/
│       │   ├── common/v1/
│       │   ├── history/v1/
│       │   └── ...
│       └── bridge/            # (not used — pure gRPC only)
│
├── tests/
│   ├── test_client.py
│   ├── test_workflow.py
│   ├── test_activity.py
│   ├── test_converter.py
│   ├── test_worker.py
│   └── test_replay.py         # deterministic replay correctness tests
│
├── scripts/
│   └── gen_protos.py          # regenerate grpc stubs from proto files
│
├── pyproject.toml
├── uv.lock
└── AGENTS.md
```

---

## Public API

### `tempiex.client`

```python
from tempiex.client import Client, ClientConfig

# Connect
client = await Client.connect(
    "localhost:8133",
    namespace="default",
    tls=None,            # TLSConfig or None
    api_key=None,        # str or None
    data_converter=DataConverter.default(),
)

# Start workflow (fire-and-forget)
handle = await client.start_workflow(
    MyWorkflow.run,
    "input-arg",
    id="my-wf-1",
    task_queue="orders",
    execution_timeout=timedelta(hours=1),
    run_timeout=timedelta(minutes=30),
    id_reuse_policy=WorkflowIDReusePolicy.ALLOW_DUPLICATE,
    memo={"key": "value"},
    search_attributes={"CustomAttr": "value"},
)

# Start and wait for result
result = await client.execute_workflow(
    MyWorkflow.run, "input",
    id="my-wf-1", task_queue="orders"
)

# Get handle to existing workflow
handle = client.get_workflow_handle("my-wf-1")

# WorkflowHandle operations
await handle.result()
await handle.cancel()
await handle.terminate(reason="stale")
await handle.signal(MyWorkflow.approve, "alice")
result = await handle.query(MyWorkflow.get_status)
update = await handle.start_update(MyWorkflow.apply_discount, 10)
await handle.describe()

# Workflow list
async for wf in client.list_workflows('WorkflowType="OrderWorkflow"'):
    print(wf.id, wf.status)

# Count
count = await client.count_workflows('ExecutionStatus="Running"')

# Schedules
schedule_handle = await client.create_schedule(
    "daily-report",
    Schedule(
        action=ScheduleActionStartWorkflow(
            workflow=ReportWorkflow.run,
            task_queue="reports",
        ),
        spec=ScheduleSpec(cron_expressions=["0 2 * * *"]),
    ),
)
await schedule_handle.trigger()
await schedule_handle.pause(note="maintenance")
await schedule_handle.unpause()
await schedule_handle.delete()
```

### `tempiex.workflow`

APIs safe to call **inside** workflow code only. Using them outside a workflow raises `RuntimeError`.

```python
from tempiex import workflow
from datetime import timedelta

@workflow.defn
class OrderWorkflow:
    def __init__(self) -> None:
        self._status = "pending"
        self._approved = False

    @workflow.run
    async def run(self, order_id: str) -> str:
        # Execute an activity
        result = await workflow.execute_activity(
            process_order,
            order_id,
            start_to_close_timeout=timedelta(seconds=30),
            retry_policy=RetryPolicy(maximum_attempts=3),
        )

        # Wait with timeout
        await workflow.wait_condition(
            lambda: self._approved,
            timeout=timedelta(hours=24),
        )

        # Sleep
        await workflow.sleep(timedelta(seconds=5))

        # Start child workflow
        child = await workflow.start_child_workflow(
            NotifyWorkflow.run, order_id,
            id=f"notify-{order_id}",
            task_queue="notifications",
        )
        await child

        return f"done:{result}"

    @workflow.signal
    async def approve(self, approved_by: str) -> None:
        self._approved = True

    @workflow.query
    def get_status(self) -> str:
        return self._status

    @workflow.update
    async def apply_discount(self, pct: int) -> str:
        # synchronous validation
        if pct < 0 or pct > 100:
            raise workflow.ApplicationError("invalid pct")
        self._discount = pct
        return f"applied:{pct}"

# In-workflow context APIs
workflow.info()           # WorkflowInfo (id, run_id, task_queue, etc.)
workflow.now()            # datetime (deterministic — from history)
workflow.memo()           # dict[str, Any]
workflow.upsert_memo({"key": "val"})
workflow.in_workflow()    # bool — True if inside a workflow
workflow.logger           # workflow-safe logger (no real I/O)
workflow.random()         # workflow-safe random (seeded from history)
workflow.patch("my-patch")           # non-breaking versioning
workflow.deprecate_patch("old-patch")
workflow.continue_as_new(MyWorkflow.run, args)  # trim history
workflow.get_external_workflow_handle(wf_id)     # signal/cancel external wf
```

### `tempiex.activity`

APIs safe to call **inside** activity code only.

```python
from tempiex import activity

@activity.defn
async def process_order(order_id: str) -> str:
    info = activity.info()   # ActivityInfo
    # Heartbeat — keeps activity alive, accepts cancellation
    activity.heartbeat(f"processing {order_id}")
    ...
    return f"processed:{order_id}"

# Synchronous activities (run in thread pool)
@activity.defn
def sync_process(order_id: str) -> str:
    activity.heartbeat("step1")
    return f"done:{order_id}"

# ActivityInfo fields
info.activity_id
info.activity_type
info.workflow_id
info.run_id
info.task_queue
info.attempt          # retry attempt number (starts at 1)
info.heartbeat_details  # details from previous heartbeat (for resume)
info.scheduled_time
info.started_time
info.deadline
info.is_local
```

### `tempiex.worker`

```python
from tempiex.worker import Worker, WorkerConfig

worker = Worker(
    client,
    task_queue="orders",
    workflows=[OrderWorkflow, NotifyWorkflow],
    activities=[process_order, sync_process],
    # Optional tuning
    max_concurrent_workflow_tasks=20,
    max_concurrent_activities=100,
    max_concurrent_local_activities=20,
    # Optional interceptors
    interceptors=[MyInterceptor()],
    # Optional data converter override (overrides client's)
    data_converter=DataConverter.default(),
)

# Run until cancelled
await worker.run()

# Or use as async context manager
async with Worker(client, task_queue="orders", workflows=[...], activities=[...]):
    await asyncio.Event().wait()
```

### `tempiex.converter`

```python
from tempiex.converter import (
    DataConverter,
    JSONPayloadConverter,
    PayloadCodec,
)

# Default: JSON encoding for all types
dc = DataConverter.default()

# Custom codec (e.g. encryption)
class EncryptCodec(PayloadCodec):
    async def encode(self, payloads: list[Payload]) -> list[Payload]: ...
    async def decode(self, payloads: list[Payload]) -> list[Payload]: ...

dc = DataConverter(payload_codec=EncryptCodec())

# Use in client
client = await Client.connect("localhost:8133", data_converter=dc)
```

### `tempiex.exceptions`

```python
from tempiex.exceptions import (
    ApplicationError,      # user-defined application error (retryable or not)
    CancelledError,        # workflow/activity cancelled
    ActivityError,         # wraps activity failure seen from workflow
    ChildWorkflowError,    # wraps child workflow failure
    TimeoutError,          # schedule-to-close, start-to-close, etc.
    TerminatedError,       # workflow was terminated
    RPCError,              # gRPC transport error
    RPCStatusCode,         # gRPC status codes enum
)

raise ApplicationError("order not found", "ORDER_NOT_FOUND", non_retryable=True)
```

### `tempiex.testing`

```python
from tempiex.testing import WorkflowEnvironment

# Spin up a test server backed by the real tempiex binary.
# Requires a PostgreSQL instance; reads TEMPIEX_TEST_DB_DSN env var,
# or pass postgres_dsn explicitly.
async with await WorkflowEnvironment.start_local(
    postgres_dsn="postgres://tempiex:tempiex@localhost:5432/tempiex_test"
) as env:
    client = env.client
    # ... run tests ...

# Or connect to an existing running server
async with await WorkflowEnvironment.start_time_skipping() as env:
    # time-skipping: sleep(timedelta(days=1)) completes instantly
    ...

# Typical pytest usage
@pytest.fixture
async def env():
    async with await WorkflowEnvironment.start_local() as e:
        yield e

async def test_order_workflow(env):
    result = await env.client.execute_workflow(
        OrderWorkflow.run, "order-1",
        id="test-order-1", task_queue="test"
    )
    assert result == "done:processed:order-1"
```

---

## Workflow Execution Engine (Internal)

This section documents the `_internal/` implementation — not part of the public API.

### Replay-based execution

Problem: workflow code must be deterministic — the same sequence of events must always produce the same commands. The worker replays all historical events on every workflow task to re-drive the coroutine to the current position, then executes new work from there.

Concrete trace:

```
Server sends WFT with history: [WorkflowExecutionStarted, ActivityTaskScheduled,
                                 ActivityTaskStarted, ActivityTaskCompleted]

Worker:
1. Create fresh workflow coroutine: coro = OrderWorkflow().run("order-1")
2. Replay event 1 (WorkflowExecutionStarted): advance coro to first await
3. The coro hits `await workflow.execute_activity(process_order, ...)`:
   - This creates a ScheduleActivityTask command candidate
   - But history shows ActivityTaskScheduled already — mark as replayed, do not emit
4. Replay ActivityTaskStarted: no coro action needed
5. Replay ActivityTaskCompleted: resolve the execute_activity awaitable with result
6. Coro advances to `await workflow.wait_condition(lambda: self._approved, ...)`
7. This is past the end of history — emit WaitCondition as a new command
8. Send RespondWorkflowTaskCompleted with [new commands]
```

### Determinism constraints

Inside `@workflow.defn` methods:
- No direct `asyncio` I/O (no `await asyncio.sleep`, no file/network I/O)
- No `datetime.now()` — use `workflow.now()`
- No `random.random()` — use `workflow.random()`
- No threading

The `_internal/_workflow_instance.py` sandboxes this by overriding `asyncio`'s event loop for the workflow coroutine to intercept forbidden operations and raise `workflow.SandboxViolationError`.

### Threading model

```
asyncio event loop (main thread)
  │
  ├─ WFT poll task       → long-polls PollWorkflowTaskQueue
  ├─ AT poll task        → long-polls PollActivityTaskQueue
  │
  ├─ per WFT:            → _WorkflowTask.execute() (sync replay + coroutine drive)
  │
  └─ per activity:
       async activity    → asyncio.Task (same loop)
       sync activity     → loop.run_in_executor(thread_pool)
```

---

## common.py — Shared Dataclasses

```python
@dataclass
class RetryPolicy:
    initial_interval: timedelta = timedelta(seconds=1)
    backoff_coefficient: float = 2.0
    maximum_interval: timedelta | None = None
    maximum_attempts: int = 0          # 0 = unlimited
    non_retryable_error_types: list[str] = field(default_factory=list)

@dataclass
class WorkflowIDReusePolicy(Enum):
    ALLOW_DUPLICATE = 0
    ALLOW_DUPLICATE_FAILED_ONLY = 1
    REJECT_DUPLICATE = 2
    TERMINATE_IF_RUNNING = 3

@dataclass
class WorkflowInfo:
    id: str
    run_id: str
    workflow_type: str
    task_queue: str
    namespace: str
    attempt: int
    start_time: datetime
    execution_timeout: timedelta | None
    run_timeout: timedelta | None
    task_timeout: timedelta

@dataclass
class ActivityInfo:
    activity_id: str
    activity_type: str
    workflow_id: str
    run_id: str
    task_queue: str
    attempt: int
    heartbeat_details: list[Any]
    scheduled_time: datetime
    started_time: datetime
    deadline: datetime | None
    is_local: bool
```

---

## Configuration (`pyproject.toml`)

```toml
[project]
name = "tempiex"
version = "0.1.0"
description = "Tempiex Python SDK"
requires-python = ">=3.10"
license = "MIT"
dependencies = [
  "grpcio>=1.60,<2",
  "grpcio-status>=1.60,<2",
  "protobuf>=4.25,<6",
  "python-dateutil>=2.8",
  "typing-extensions>=4.5",
]

[project.optional-dependencies]
opentelemetry = [
  "opentelemetry-api>=1.26,<2",
  "opentelemetry-sdk>=1.26,<2",
  "opentelemetry-exporter-otlp-proto-grpc>=1.11,<2",
]
pydantic = ["pydantic>=2.0,<3"]

[dependency-groups]
dev = [
  "pytest>=8",
  "pytest-asyncio>=0.23",
  "pytest-timeout>=2.2",
  "ruff>=0.5",
  "pyright>=1.1.380",
  "grpcio-tools>=1.60,<2",
]

[tool.pytest.ini_options]
asyncio_mode = "auto"
testpaths = ["tests"]
timeout = 60

[tool.ruff]
target-version = "py310"

[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"
```

---

## Development Commands

```bash
cd sdk-python/

# Install with dev extras
uv sync --all-extras

# Regenerate protobuf stubs from Tempiex proto files
uv run python scripts/gen_protos.py

# Lint (import order + format + type check)
uv run ruff check .
uv run ruff format --check .
uv run pyright

# Format
uv run ruff check --select I --fix && uv run ruff format

# Run all tests (requires tempiex server on :8133)
uv run pytest

# Run specific test
uv run pytest tests/test_workflow.py::test_order_workflow -s

# Run without a server (unit tests only)
uv run pytest tests/test_converter.py tests/test_replay.py -s

# Build wheel
uv build
```

---

## Interceptors

Interceptors allow middleware-style wrapping of client, workflow, and activity calls.

```python
from tempiex.worker import Interceptor, ActivityInboundInterceptor

class LoggingInterceptor(Interceptor):
    def intercept_activity(
        self, next: ActivityInboundInterceptor
    ) -> ActivityInboundInterceptor:
        return LoggingActivityInterceptor(next)

class LoggingActivityInterceptor(ActivityInboundInterceptor):
    async def execute_activity(self, input: ExecuteActivityInput) -> Any:
        print(f"start: {input.fn.__name__}")
        result = await self.next.execute_activity(input)
        print(f"end: {input.fn.__name__}")
        return result
```

---

## OpenTelemetry Integration

Install the optional extra:

```bash
uv sync --extra opentelemetry
```

```python
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from tempiex.opentelemetry import OpenTelemetryInterceptor

provider = TracerProvider()
provider.add_span_processor(
    BatchSpanProcessor(OTLPSpanExporter(endpoint="http://localhost:4317"))
)
trace.set_tracer_provider(provider)

worker = Worker(
    client,
    task_queue="orders",
    workflows=[OrderWorkflow],
    activities=[process_order],
    interceptors=[OpenTelemetryInterceptor()],
)
```

Each workflow task and each activity execution creates a span. Workflow spans carry the Tempiex workflow trace context so activity spans appear as children.

---

## Deployment Example

A complete worker running in the same monorepo:

```python
# sdk-python/examples/order_worker.py
import asyncio
from tempiex.client import Client
from tempiex.worker import Worker
from my_workflows import OrderWorkflow, process_order

async def main():
    client = await Client.connect(
        "localhost:8133",
        namespace="default",
    )
    worker = Worker(
        client,
        task_queue="orders",
        workflows=[OrderWorkflow],
        activities=[process_order],
    )
    print("Worker started. Ctrl+C to stop.")
    await worker.run()

asyncio.run(main())
```

```bash
uv run python examples/order_worker.py
```

---

## Missing vs. Reference SDK

The following features from `temporalio` are intentionally excluded:

| Feature | Reason excluded |
|---------|----------------|
| Rust bridge / Maturin | Pure Python; simpler to build and distribute |
| Nexus operations | Out of Tempiex scope |
| Cloud Operations client | Out of Tempiex scope |
| Worker deployment / build IDs | Advanced versioning; not needed initially |
| FIPS mode | Not required |
| Local activities (in-process) | Can be added later |
| Workflow sandbox (import restrictions) | Lighter enforcement; SandboxViolationError only for direct event loop ops |
| AI agent integrations (openai-agents, langgraph, etc.) | Out of scope |
| `@workflow.update` validator | Included in spec but implementation is Phase 2 |

Features included from the start:

| Feature |
|---------|
| `@workflow.defn`, `@workflow.run`, `@workflow.signal`, `@workflow.query`, `@workflow.update` |
| `@activity.defn` (async + sync/threaded) |
| Heartbeat, cancellation |
| `RetryPolicy`, timeouts |
| Child workflows, continue-as-new |
| `workflow.patch` / `workflow.deprecate_patch` for versioning |
| `DataConverter` with pluggable codec (for encryption) |
| `WorkflowEnvironment` for testing |
| OpenTelemetry interceptor |
| Interceptor framework |
| `Client.list_workflows`, `count_workflows` |
| Schedule management |
