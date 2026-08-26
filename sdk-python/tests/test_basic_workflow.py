"""Basic workflow test - simple workflow that returns a value."""
import asyncio
import pytest
import uuid
import subprocess
import signal

from tempiex.client import Client
from tempiex.worker import Worker
from tempiex import workflow, activity


@workflow.defn
class GreetingWorkflow:
    @workflow.run
    async def run(self, name: str) -> str:
        return f"Hello, {name}!"


@pytest.fixture(scope="module")
async def server():
    """Start the Tempiex server."""
    import socket
    # Check if already running
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        already_up = s.connect_ex(("localhost", 8133)) == 0

    if already_up:
        yield None
        return

    server_bin = "/home/ubuntu/app/tempiex/tempiex/bin/tempiex-server"
    proc = subprocess.Popen(
        [server_bin],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        cwd="/home/ubuntu/app/tempiex/tempiex",
    )
    await asyncio.sleep(3)
    yield proc
    proc.send_signal(signal.SIGTERM)
    try:
        proc.wait(timeout=5)
    except subprocess.TimeoutExpired:
        proc.kill()


@pytest.fixture
async def client(server):
    """Create a connected client."""
    c = await Client.connect("localhost:8133", namespace="test-" + uuid.uuid4().hex[:8])
    yield c
    await c.close()


@pytest.mark.asyncio
async def test_simple_workflow(client):
    """Test a simple workflow that returns immediately."""
    task_queue = "test-queue-" + uuid.uuid4().hex[:8]

    worker = Worker(client, task_queue, workflows=[GreetingWorkflow])
    worker_task = asyncio.create_task(worker.run())

    try:
        result = await client.execute_workflow(
            GreetingWorkflow,
            "World",
            id="greeting-" + uuid.uuid4().hex[:8],
            task_queue=task_queue,
        )
        assert result == "Hello, World!"
    finally:
        worker.shutdown()
        await asyncio.sleep(0.5)
        worker_task.cancel()
        try:
            await worker_task
        except asyncio.CancelledError:
            pass


@workflow.defn
class ApprovalWorkflow:
    """Workflow that waits for an approve or reject signal."""

    def __init__(self):
        self._approved = False
        self._rejected = False
        self._reason = ""

    @workflow.signal
    async def approve(self):
        self._approved = True

    @workflow.signal
    async def reject(self, reason: str = ""):
        self._rejected = True
        self._reason = reason

    @workflow.query
    def get_status(self) -> str:
        if self._approved:
            return "approved"
        if self._rejected:
            return f"rejected:{self._reason}"
        return "pending"

    @workflow.run
    async def run(self) -> str:
        await workflow.wait_condition(lambda: self._approved or self._rejected)
        if self._approved:
            return "approved"
        return f"rejected:{self._reason}"


@pytest.mark.asyncio
async def test_workflow_signal(client):
    """Test that a signal is delivered and processed by the workflow."""
    task_queue = "signal-test-" + uuid.uuid4().hex[:8]

    worker = Worker(client, task_queue, workflows=[ApprovalWorkflow])
    worker_task = asyncio.create_task(worker.run())

    try:
        handle = await client.start_workflow(
            ApprovalWorkflow,
            id="approval-" + uuid.uuid4().hex[:8],
            task_queue=task_queue,
        )

        # Give worker time to pick up and start workflow
        await asyncio.sleep(0.5)

        # Send the approve signal via handle
        await handle.signal("approve")

        # Wait for result
        result = await handle.result(timeout=15.0)
        assert result == "approved", f"Expected 'approved', got {result!r}"
    finally:
        worker.shutdown()
        await asyncio.sleep(0.3)
        worker_task.cancel()
        try:
            await worker_task
        except asyncio.CancelledError:
            pass
