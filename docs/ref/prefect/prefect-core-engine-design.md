# Prefect Core Orchestration Engine - Design Analysis

**Author**: Analysis of Prefect 3.x Codebase  
**Date**: 2026-08-11  
**Purpose**: Comprehensive architectural design document for building workflow orchestration engines

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Architectural Overview](#architectural-overview)
3. [Core Components](#core-components)
4. [State Management & Orchestration](#state-management--orchestration)
5. [Execution Engines](#execution-engines)
6. [Concurrency & Parallelism](#concurrency--parallelism)
7. [Result Persistence](#result-persistence)
8. [Critical Design Patterns](#critical-design-patterns)
9. [Key Contracts & Invariants](#key-contracts--invariants)
10. [Deployment & Process Model](#deployment--process-model)
11. [Lessons & Best Practices](#lessons--best-practices)

---

## Executive Summary

Prefect is a **distributed workflow orchestration platform** built around a central principle: **the server is the source of truth for all state transitions**. The architecture separates workflow definition (Python decorators), execution (async engines), and orchestration (server-side state validation).

### Key Architectural Decisions

1. **State-first design**: All workflow state transitions flow through a centralized orchestration API
2. **Async-native with sync wrappers**: Core engines are async; sync functions run in worker threads
3. **Subprocess isolation**: Flow runs execute in isolated subprocesses for fault isolation
4. **Dual execution paths**: Flow and task engines have parallel sync/async implementations
5. **Result decoupling**: Results are persisted separately from state objects
6. **Policy-based orchestration**: State transitions governed by composable rule policies

---

## Architectural Overview

### Three-Layer Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      USER CODE LAYER                         │
│  @flow / @task decorators, Python functions, parameters     │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                   EXECUTION ENGINE LAYER                     │
│  FlowRunEngine, TaskRunEngine (sync/async variants)         │
│  - Parameter resolution                                      │
│  - State transitions (propose/validate/commit)               │
│  - Retries, timeouts, caching                                │
│  - Result persistence                                        │
│  - Hook execution (on_completion, on_failure, etc.)          │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                  ORCHESTRATION LAYER (SERVER)                │
│  State validation via orchestration rules                   │
│  - Concurrency limits                                        │
│  - Retry policies                                            │
│  - Dependency tracking                                       │
│  - Database persistence                                      │
└─────────────────────────────────────────────────────────────┘
```

### File Structure

```
src/prefect/
├── flows.py              # @flow decorator, Flow class
├── tasks.py              # @task decorator, Task class
├── flow_engine.py        # Flow execution engines (2329 lines)
├── task_engine.py        # Task execution engines (1916 lines)
├── engine.py             # Entry point and signal handling (182 lines)
├── states.py             # State objects and transitions
├── results.py            # Result persistence layer
├── futures.py            # PrefectFuture for async task results
├── context.py            # Runtime context management
├── transactions.py       # Transaction support for tasks
└── utilities/engine/     # Shared engine utilities
```

---

## Core Components

### 1. Flow and Task Definitions

**Flows** are the top-level workflows, **Tasks** are atomic units of work.

```python
# User-facing API
@flow(name="my-flow", retries=3, timeout_seconds=600)
def my_workflow(param: str):
    result = my_task(param)
    return result

@task(cache_key_fn=..., retries=2, retry_delay_seconds=[1, 10, 100])
def my_task(input: str):
    return process(input)
```

**Key attributes**:
- **Flows**: name, description, version, timeout, retries, retry_delay, task_runner, result_storage
- **Tasks**: cache_policy, persist_result, timeout, retries, retry_delay, retry_jitter_factor

### 2. State Objects

States represent the **lifecycle stage** of a flow or task run.

```python
class StateType(AutoEnum):
    SCHEDULED = AutoEnum.auto()
    PENDING = AutoEnum.auto()
    RUNNING = AutoEnum.auto()
    COMPLETED = AutoEnum.auto()
    FAILED = AutoEnum.auto()
    CANCELLED = AutoEnum.auto()
    CRASHED = AutoEnum.auto()
    PAUSED = AutoEnum.auto()
    CANCELLING = AutoEnum.auto()

TERMINAL_STATES = {COMPLETED, CANCELLED, FAILED, CRASHED}
```

**State object structure**:
```python
class State:
    type: StateType
    name: str
    message: str
    data: Any  # ResultRecord or direct value
    timestamp: DateTime
    state_details: StateDetails  # Contains scheduled_time, task_run_id, etc.
```

### 3. Execution Engines

Two parallel implementations for **flows** and **tasks**, each with **sync** and **async** variants:

- `FlowRunEngine` (sync) / `AsyncFlowRunEngine` (async)
- `SyncTaskRunEngine` (sync) / `AsyncTaskRunEngine` (async)

All engines inherit from base classes:
- `BaseFlowRunEngine[P, R]`
- `BaseTaskRunEngine[P, R]`

**Engine responsibilities**:
1. Parameter resolution (unwrap futures, validate)
2. State transitions (begin_run, handle_success, handle_failure, handle_retry)
3. Context management (FlowRunContext, TaskRunContext)
4. Result persistence
5. Hook execution
6. Telemetry and logging

### 4. Context System

**Runtime context** is thread-safe via `ContextVar`:

```python
class FlowRunContext(RunContext):
    flow: Flow
    flow_run: FlowRun
    task_runner: TaskRunner
    result_store: ResultStore
    parameters: dict[str, Any]
    # ... suspension requests, telemetry, etc.

class TaskRunContext(RunContext):
    task: Task
    task_run: TaskRun
    # ... parameters, dependencies, etc.
```

Context is serialized for **subprocess/multiprocess execution**:
```python
serialized = serialize_context()
with hydrated_context(serialized):
    # Execute in subprocess with parent context
```

---

## State Management & Orchestration

### State Transition Flow

The **critical contract**: Flow engines **propose** states to the server; the server **accepts or rejects** them via orchestration rules.

```
┌──────────────┐
│ Flow Engine  │
└──────┬───────┘
       │ 1. Propose state (Pending → Running)
       ▼
┌──────────────────────────────────────┐
│ Server Orchestration Layer           │
│ ┌──────────────────────────────────┐ │
│ │ CoreFlowPolicy.priority():       │ │
│ │ 1. PreventDuplicateTransitions   │ │
│ │ 2. ValidateDeploymentConcurrency │ │
│ │ 3. WaitForScheduledTime          │ │
│ │ 4. RetryFailedFlows              │ │
│ │ ... (15+ orchestration rules)    │ │
│ └──────────────────────────────────┘ │
└──────┬───────────────────────────────┘
       │ 2. Accept/Reject/Wait
       ▼
┌──────────────┐
│ Flow Engine  │
│ (updates     │
│  local state)│
└──────────────┘
```

### Orchestration Rules

Rules are **composable, ordered policies** that fire on state transitions:

```python
class BaseOrchestrationRule:
    async def before_transition(
        self, 
        initial_state: State, 
        proposed_state: State, 
        context: OrchestrationContext
    ) -> None:
        """Validate or modify before transition"""
    
    async def after_transition(
        self,
        initial_state: State,
        validated_state: State,
        context: OrchestrationContext
    ) -> None:
        """Cleanup or side-effects after transition"""
```

**Example rule**: `WaitForScheduledTime`
```python
class WaitForScheduledTime(FlowRunOrchestrationRule):
    async def before_transition(self, ...):
        if proposed_state.is_running():
            scheduled_time = initial_state.state_details.scheduled_time
            if scheduled_time and scheduled_time > now():
                # Reject the transition, tell client to wait
                context.response_status = SetStateStatus.WAIT
                context.response_details = StateWaitDetails(
                    delay_seconds=(scheduled_time - now()).total_seconds()
                )
```

### Flow vs Task State Management

**Critical difference**:

- **Flow engines**: Make **blocking API calls** to propose states; server accepts/rejects
- **Task engines**: Manage state transitions **locally** via `set_state()`; emit `prefect.task_run.*` events

This asymmetry exists because:
- Flows are long-running and require server coordination
- Tasks are numerous and short-lived; local state management reduces API load
- Task events are delivered via WebSockets for monitoring

---

## Execution Engines

### Flow Engine Lifecycle

```python
class FlowRunEngine:
    def begin_run(self) -> State:
        """
        1. Resolve parameters (unwrap futures)
        2. Wait for dependencies
        3. Validate parameters
        4. Set custom flow run name
        5. Acquire concurrency slots
        6. Transition to Running state
        """
        self._resolve_parameters()
        self._wait_for_dependencies()
        
        if self.flow.should_validate_parameters:
            self.parameters = self.flow.validate_parameters(self.parameters)
        
        new_state = Running()
        state = self.set_state(new_state)
        return state
    
    def call_flow_function(self) -> Any:
        """
        Execute user code with:
        - Timeout wrapper
        - Transaction context
        - Exception handling
        - Suspension checks
        """
        with timeout_context(self.flow.timeout_seconds):
            with transaction_context():
                result = call_with_parameters(
                    self.flow.fn, 
                    self.parameters
                )
        return result
    
    def handle_success(self, result: R) -> State:
        """
        1. Convert result to state
        2. Persist result to storage
        3. Set terminal state (Completed)
        4. Execute on_completion hooks
        """
        terminal_state = return_value_to_state(
            result,
            result_store=self.result_store
        )
        self.set_state(terminal_state)
        self._return_value = result
        return terminal_state
```

### Task Engine Lifecycle

Similar structure but with **local state management**:

```python
class SyncTaskRunEngine:
    def begin_run(self) -> None:
        """Transition to Running (local, no server proposal)"""
        new_state = Running()
        self.task_run.start_time = new_state.timestamp
        
        state = self.set_state(new_state)
        
        # Poll until server accepts (with backoff)
        while state.is_pending() or state.is_paused():
            interval = clamped_poisson_interval(backoff_count)
            time.sleep(interval)
            state = self.set_state(new_state)
    
    def set_state(self, state: State) -> State:
        """
        Local state update + event emission
        (no server proposal for tasks)
        """
        self.task_run.state = state
        
        # Emit state change event
        emit_task_run_state_change_event(
            task_run=self.task_run,
            initial_state=last_state,
            validated_state=state
        )
        
        return state
```

### Sync/Async Duality

**Critical invariant**: Sync and async engines must stay **in lockstep**.

Both paths implement:
1. Parameter resolution
2. Dependency waiting
3. State transitions
4. Retries with exponential backoff
5. Timeout handling
6. Result persistence
7. Hook execution
8. Transaction support
9. Concurrency slot management

**Pattern**: Sync functions run in worker threads via `run_sync_in_worker_thread()`

```python
# Async engine
async def call_flow_function(self):
    async with timeout_async(self.flow.timeout_seconds):
        if self.flow.isasync:
            result = await self.flow.fn(**parameters)
        else:
            result = await run_sync_in_worker_thread(
                self.flow.fn, **parameters
            )
    return result

# Sync engine
def call_flow_function(self):
    with timeout(self.flow.timeout_seconds):
        if self.flow.isasync:
            result = run_coro_as_sync(self.flow.fn(**parameters))
        else:
            result = self.flow.fn(**parameters)
    return result
```

---

## Concurrency & Parallelism

### Task Runners

Task runners control **how tasks within a flow execute**:

1. **ThreadPoolTaskRunner** (default)
   - Submit tasks to thread pool
   - Returns `PrefectFuture` immediately
   - Tasks execute concurrently

2. **ProcessPoolTaskRunner**
   - Submit tasks to process pool
   - **Picklable data only** (PrefectFuture objects cannot cross subprocess boundaries)
   - `wait_for` futures must be resolved to `State` objects before submission

3. **SequentialTaskRunner**
   - Execute tasks synchronously
   - No parallelism

### Concurrency Limits

Two systems for limiting concurrent execution:

#### 1. Tag-based Concurrency (v1)
```python
@task(tags=["database"])
def query_db():
    pass

# Configure limit: max 5 concurrent tasks with "database" tag
```

#### 2. Global Concurrency Limits (v2)
```python
from prefect.concurrency.v1.asyncio import concurrency

async with concurrency("database-queries", occupy=1):
    await query_database()
```

**Implementation**: Lease-based system with:
- Server-side slot tracking
- Lease renewal (heartbeat every 60s)
- Automatic lease expiry
- Deadlock prevention via timeout

### Flow-level Concurrency

Controlled via **deployment concurrency limits**:
- Limits concurrent runs of the same deployment
- Enforced at flow run startup
- Configurable strategy: ENQUEUE (wait) or CANCEL_NEW (reject)

---

## Result Persistence

### Result Storage Layer

Results are **decoupled from state objects** and stored separately:

```python
class ResultStore:
    """
    Manages result persistence to storage backends
    """
    filesystem: WritableFileSystem  # LocalFileSystem, S3, GCS, etc.
    serializer: Serializer  # Pickle, JSON, compressed pickle
    
    async def persist_result(
        self, 
        obj: R, 
        key: str,
        expiration: DateTime | None
    ) -> ResultRecord[R]:
        """
        1. Serialize object
        2. Write to storage (filesystem/blob store)
        3. Return metadata record
        """
```

### Result Records

**ResultRecord** wraps metadata about persisted results:

```python
class ResultRecord:
    metadata: ResultRecordMetadata  # Storage key, serializer type
    expiration: DateTime | None
    result: R  # Cached in memory after first read
    
    @property
    def is_persisted(self) -> bool:
        """Has this been written to storage?"""
```

**State objects** store `ResultRecordMetadata`, not full results:

```python
state = Completed(
    data=ResultRecordMetadata(
        storage_key="flows/abc-123/results/xyz",
        serializer={"type": "pickle"},
        expiration=None
    )
)
```

### Three-tier Storage Resolution

```python
async def get_default_result_storage():
    """
    Priority order:
    1. PREFECT_DEFAULT_RESULT_STORAGE_BLOCK setting
    2. Server-configured default block (API call)
    3. Local storage path (PREFECT_RESULTS_LOCAL_STORAGE_PATH)
    """
```

---

## Critical Design Patterns

### 1. Orchestration Context Pattern

State transitions flow through a **rich context object**:

```python
class OrchestrationContext:
    session: AsyncSession  # Database session
    initial_state: State | None
    proposed_state: State | None
    validated_state: State | None
    response_status: SetStateStatus
    response_details: StateResponseDetails
    run: FlowRun | TaskRun
    
    # Rules append their signatures for debugging
    rule_signature: list[str]
    finalization_signature: list[str]
```

Rules mutate the context to **accept, reject, wait, or abort** transitions.

### 2. Futures and Lazy Evaluation

Tasks return **PrefectFuture** objects:

```python
@flow
def my_flow():
    future1 = task_a.submit()  # Returns immediately
    future2 = task_b.submit()  # Parallel execution
    
    result1 = future1.result()  # Block until complete
    result2 = future2.result()
    
    return combine(result1, result2)
```

**Implementation**:
- `submit()` creates TaskRun, returns Future
- `result()` blocks on completion, retrieves result
- Futures can be passed as arguments: automatic dependency tracking

### 3. Visit Collection Pattern

**Dependency resolution** via recursive visitation:

```python
def resolve_parameters(parameters: dict[str, Any]):
    """
    Traverse nested structures, resolve:
    - PrefectFuture → final result
    - State objects → result value
    - Nested dicts/lists recursively
    """
    return visit_collection(
        parameters,
        visit_fn=resolve_to_final_result,
        return_data=True,
        max_depth=-1
    )
```

Raises `UpstreamTaskError` if any dependency failed.

### 4. Context Manager Layering

Engines use **nested context managers** for setup/teardown:

```python
async def run_flow():
    with handle_engine_signals(flow_run_id):
        with _send_heartbeats(heartbeat_interval):
            with timeout_async(flow.timeout_seconds):
                with AsyncClientContext() as client_ctx:
                    with FlowRunContext(...) as flow_ctx:
                        with telemetry.capture():
                            result = await flow.fn(**parameters)
```

Each layer handles a specific concern:
- Signal handling (SIGTERM → TerminationSignal)
- Heartbeats (keep-alive events)
- Timeouts
- API client lifecycle
- Runtime context
- Telemetry spans

### 5. Retry with Exponential Backoff

```python
def handle_retry(self, exc: Exception) -> bool:
    if self.retries < self.task.retries:
        delay = self.task.retry_delay_seconds[
            min(self.retries, len(...) - 1)
        ]
        
        new_state = AwaitingRetry(
            scheduled_time=now() + timedelta(seconds=delay)
        )
        
        self.set_state(new_state, force=True)
        self.retries += 1
        return True
    return False
```

**Features**:
- Configurable retry count
- Delay sequence: `[1, 10, 100]` or callable
- Jitter factor to prevent thundering herd
- Conditional retries via `retry_condition_fn`

### 6. Transaction Support

Tasks can opt into **transactional result persistence**:

```python
@task(cache_policy=CachePolicy(...))
def compute_expensive_value():
    with transaction() as txn:
        result = expensive_computation()
        txn.stage(result, on_commit=[...], on_rollback=[...])
    return result
```

**Isolation levels**:
- `READ_COMMITTED`: Allow concurrent reads
- `SERIALIZABLE`: Exclusive lock on cache key

**Commit modes**:
- `EAGER`: Write immediately on success
- `LAZY`: Write on transaction commit
- `OFF`: No persistence

---

## Key Contracts & Invariants

### 1. State Transition Contracts

- **Flow state transitions MUST go through the server**
  - Never bypass orchestration API, even in tests
  - Server is the single source of truth

- **Task state transitions are managed locally**
  - `set_state()` updates in-memory state
  - Events emitted via WebSockets for monitoring
  - No server proposal for individual task transitions

- **Set run metadata BEFORE state transitions**
  - Custom names, labels, etc. must be persisted before `set_state()`
  - Event subscribers see metadata at the moment of transition

### 2. Engine Ordering Invariant

**Features must be applied in the same order** in both engines:

```python
def execute_task():
    # 1. Resolve parameters
    self._resolve_parameters()
    
    # 2. Wait for dependencies
    self._wait_for_dependencies()
    
    # 3. Begin run (state transition)
    self.begin_run()
    
    # 4. Acquire concurrency slots
    with concurrency(...):
        
        # 5. Set up transaction
        with transaction() as txn:
            
            # 6. Apply timeout
            with timeout(self.task.timeout_seconds):
                
                # 7. Execute user code
                result = self.task.fn(**parameters)
                
        # 8. Handle success/failure
        self.handle_success(result, txn)
    
    # 9. Execute hooks
    self.call_hooks()
```

**Changing this order or omitting a feature in one engine path is the most common source of bugs.**

### 3. Subprocess Pickle Constraints

- **PrefectFuture objects are NOT picklable**
  - Cannot pass futures to `ProcessPoolTaskRunner`
  - Must resolve to `State` objects before submission

- **ThreadPoolTaskRunner must drop locks in `__getstate__`**
  - `threading.Lock` is not picklable
  - Flow runners cloudpickle the task runner when dispatching to subprocesses

### 4. Result Persistence Invariant

- **`mark_persisted()` must be called after EVERY storage write**
  - Used by `to_state_create()` to decide if metadata should be sent to server
  - Missing call silently drops result metadata for already-written results

### 5. Signal Handling

- **SIGTERM handling is opt-in via `capture_sigterm()` context**
  - Converts SIGTERM → `TerminationSignal` exception
  - Allows graceful cancellation
  - Runner sets control intent before sending signal

- **All SIGTERM state checks must hold `_prefect_sigterm_bridge_lock`**
  - Prevents TOCTOU races between handler install and ack

---

## Deployment & Process Model

### Subprocess Execution

Flow runs execute in **isolated subprocesses**:

```bash
# Entry point: python -m prefect.flow_engine
# or direct: python src/prefect/engine.py

# Environment variables:
# - PREFECT__FLOW_RUN_ID: UUID of the flow run
# - PREFECT__FLOW_ENTRYPOINT: Module path (my.module:my_flow)
# - PREFECT__CONTROL_PORT: Optional control channel port
# - PREFECT__CONTROL_TOKEN: Optional control channel token
```

**Flow engine entry point**:
```python
# src/prefect/engine.py
if __name__ == "__main__":
    flow_run_id = UUID(os.environ["PREFECT__FLOW_RUN_ID"])
    
    with handle_engine_signals(flow_run_id):
        flow_run = load_flow_run(flow_run_id)
        flow = load_flow(flow_run)
        
        with RunMetrics(flow_run, flow):
            result = run_flow(flow, flow_run=flow_run)
```

### Worker Architecture

**Workers** poll for scheduled flow runs and dispatch them:

```
┌──────────────┐
│ Work Pool    │  (Server-side queue)
└──────┬───────┘
       │
       │ Poll for scheduled runs
       ▼
┌──────────────┐
│ Worker       │
│ - Process    │
│ - Docker     │
│ - Kubernetes │
│ - ECS        │
└──────┬───────┘
       │
       │ Submit flow run to infrastructure
       ▼
┌─────────────────────────┐
│ Infrastructure          │
│ (subprocess, container) │
│                         │
│  Flow Engine runs here  │
└─────────────────────────┘
```

Workers handle:
- Flow run submission
- Infrastructure provisioning
- Log streaming
- Cancellation requests
- Cleanup after completion

### Deployment Model

Deployments define **how flows are scheduled and executed**:

```python
flow.deploy(
    name="my-deployment",
    work_pool_name="kubernetes-pool",
    schedules=[CronSchedule(cron="0 0 * * *")],
    parameters={"default_param": "value"},
    tags=["production"]
)
```

**Scheduler service** (server-side) creates flow runs from deployments.

---

## Lessons & Best Practices

### What Worked Well

1. **State-first design**
   - Clear separation between definition, execution, and orchestration
   - Server as single source of truth prevents race conditions

2. **Async-native architecture**
   - High throughput for I/O-bound workflows
   - Clean async/await syntax for users

3. **Composable orchestration rules**
   - Easy to add new behaviors without modifying engine core
   - Rules are testable in isolation

4. **Result decoupling**
   - Large results don't bloat state objects
   - Storage backend is pluggable

5. **Subprocess isolation**
   - Flow run crashes don't take down the scheduler
   - Each run has clean environment

### Design Challenges

1. **Sync/async duality maintenance burden**
   - Every feature must be implemented twice
   - Easy to drift out of sync
   - **Mitigation**: Comprehensive test suite, code review checklist

2. **Complex dependency on Pydantic v2**
   - Null fields in JSON treated as "explicitly set"
   - Requires workarounds via `model_fields_set`
   - **Mitigation**: Universal transforms to preserve state_details across transitions

3. **ProcessPoolTaskRunner limitations**
   - Picklability constraints
   - Can't pass futures across subprocess boundary
   - **Mitigation**: Clear documentation, detect nested submit deadlocks

4. **Result storage timing issues**
   - Results written asynchronously
   - State may reference result before it's persisted
   - **Mitigation**: Retry with backoff when reading results

5. **Task engine event delivery**
   - WebSocket events may be lost
   - No strong delivery guarantees
   - **Mitigation**: Task run recorder service processes event batches with deduplication

### Key Architectural Tradeoffs

| Decision | Pros | Cons |
|----------|------|------|
| **Server validates all flow state transitions** | Single source of truth, prevents race conditions | Network latency on every transition |
| **Task states managed locally** | Low latency, high throughput | Requires eventual consistency via events |
| **Subprocess per flow run** | Fault isolation, clean environment | Higher resource overhead |
| **Async-first with sync wrappers** | High I/O throughput | Complexity of maintaining dual paths |
| **Result decoupling** | Handles large results, pluggable storage | Additional storage layer to manage |

### Recommended Patterns for New Features

1. **Always implement sync and async paths together**
   - Use same test cases for both
   - Maintain feature parity

2. **State transitions should be idempotent**
   - Handle duplicate requests gracefully
   - Detect and reject duplicate transitions

3. **Hooks should be non-fatal**
   - Log errors but don't fail the run
   - User code errors shouldn't break orchestration

4. **Context should be serializable**
   - Support subprocess/multiprocess execution
   - Avoid storing non-picklable objects

5. **Telemetry and observability first**
   - Emit events for all significant actions
   - Include correlation IDs for distributed tracing

### Anti-Patterns to Avoid

1. **Never bypass state orchestration** 
   - Don't update database directly
   - Always go through `set_state()` or `propose_state()`

2. **Don't assume synchronous execution**
   - Tasks may run concurrently
   - Flows may be canceled at any time

3. **Don't hold locks during user code execution**
   - User code duration is unbounded
   - Locks should wrap only engine operations

4. **Don't log secrets**
   - Parameters may contain credentials
   - Use redaction helpers

5. **Don't block the event loop**
   - Offload CPU work to threads
   - Use async I/O for network calls

---

## Appendix: File Size Reference

```
flow_engine.py:     2,329 lines
task_engine.py:     1,916 lines  
engine.py:            182 lines
states.py:           ~800 lines
results.py:          ~600 lines
futures.py:          ~400 lines
context.py:          ~900 lines
transactions.py:     ~500 lines
```

**Total core engine code**: ~7,600 lines (excluding tests)

---

## Conclusion

Prefect's core engine demonstrates a **mature, production-grade approach** to distributed workflow orchestration. Key insights:

1. **Separation of concerns**: Definition → Execution → Orchestration
2. **Server authority**: Centralized state validation prevents distributed race conditions
3. **Async-first**: Modern Python async/await for high concurrency
4. **Policy composition**: Orchestration rules as composable, testable units
5. **Fault isolation**: Subprocess execution boundaries

The dual sync/async implementation adds complexity but provides flexibility for users. The state-first design, while requiring network calls, ensures correctness in distributed scenarios.

**For designing your own orchestration system**, adopt:
- State machines with server-side validation
- Composable policy rules
- Clear separation of result storage from state metadata
- Subprocess isolation for fault tolerance
- Comprehensive context serialization for distributed execution

Avoid Prefect's challenges by:
- Using code generation to maintain sync/async parity
- Simplifying the process model if subprocess overhead is too high
- Providing stronger event delivery guarantees if task observability is critical
- Considering event sourcing patterns for complete audit trails
