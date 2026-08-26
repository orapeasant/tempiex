# tempiex (Python SDK)

Python SDK for Tempiex. Provides async workflow and activity primitives backed by the Tempiex gRPC API.

## Quick Start

### Prerequisites

- Python 3.10+
- [uv](https://github.com/astral-sh/uv) package manager
- A running `tempiex` instance on `localhost:8133`

### Installation

```bash
# Inside your project with uv:
uv add tempiex

# Or from source (monorepo):
cd sdk-python/
uv sync
```

### Hello World

```python
import asyncio
from tempiex import WorkflowClient, workflow, activity

@activity.defn
async def say_hello(name: str) -> str:
    return f"Hello, {name}!"

@workflow.defn
class GreetWorkflow:
    @workflow.run
    async def run(self, name: str) -> str:
        return await workflow.execute_activity(
            say_hello,
            name,
            start_to_close_timeout=30,
        )

async def main():
    client = await WorkflowClient.connect("localhost:8133")

    result = await client.execute_workflow(
        GreetWorkflow.run,
        "World",
        id="greet-1",
        task_queue="my-queue",
    )
    print(result)  # Hello, World!

if __name__ == "__main__":
    asyncio.run(main())
```

### Running a Worker

```python
import asyncio
from tempiex import WorkflowClient, Worker

async def main():
    client = await WorkflowClient.connect("localhost:8133")

    worker = Worker(
        client,
        task_queue="my-queue",
        workflows=[GreetWorkflow],
        activities=[say_hello],
    )
    await worker.run()

if __name__ == "__main__":
    asyncio.run(main())
```

## Installation (development)

```bash
cd sdk-python/

# Install all dependencies including dev
uv sync --extra dev

# Run tests
uv run pytest

# Type check
uv run pyright

# Lint and format check
uv run ruff check .
uv run ruff format --check .
```

## SDK Reference

### `WorkflowClient`

```python
client = await WorkflowClient.connect(
    target="localhost:8133",   # tempiex gRPC address
    namespace="default",
)
```

| Method | Description |
|---|---|
| `execute_workflow(fn, *args, id, task_queue)` | Start and await a workflow |
| `start_workflow(fn, *args, id, task_queue)` | Start without awaiting |
| `get_workflow_handle(workflow_id, run_id)` | Get a handle to an existing run |

### `Worker`

```python
worker = Worker(
    client,
    task_queue="my-queue",
    workflows=[MyWorkflow],
    activities=[my_activity],
    max_concurrent_activities=10,
)
await worker.run()
```

### Decorators

| Decorator | Usage |
|---|---|
| `@workflow.defn` | Mark a class as a workflow definition |
| `@workflow.run` | Mark the workflow entry method |
| `@activity.defn` | Mark an async function as an activity |

### Inside a workflow

```python
# Execute an activity
result = await workflow.execute_activity(
    my_activity,
    arg,
    start_to_close_timeout=60,  # seconds
    retry_policy=RetryPolicy(maximum_attempts=3),
)
```

## Samples

See [`samples/`](samples/) for runnable examples:

- `samples/order_workflow.py` — a multi-step order processing workflow

## Project Structure

```
sdk-python/
├── tempiex/
│   ├── client.py     # WorkflowClient
│   ├── worker.py     # Worker
│   ├── workflow.py   # @workflow.defn, workflow context
│   ├── activity.py   # @activity.defn
│   ├── converter.py  # Payload serialization
│   └── exceptions.py # SDK exceptions
├── tests/            # pytest test suite
├── samples/          # Runnable examples
└── pyproject.toml
```
