"""Pi agent test - workflow that executes shell commands as activities."""
import asyncio
import pytest
import uuid
import subprocess
import signal

from tempiex.client import Client
from tempiex.worker import Worker
from tempiex import workflow, activity


@activity.defn
def run_shell_command(cmd: str) -> str:
    """Execute a shell command and return stdout."""
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True, timeout=30)
    if result.returncode != 0:
        raise RuntimeError(f"Command failed with code {result.returncode}: {result.stderr}")
    return result.stdout


@workflow.defn
class PiWorkflow:
    @workflow.run
    async def run(self) -> str:
        from datetime import timedelta
        result1 = await workflow.execute_activity(
            run_shell_command,
            "echo hello",
            start_to_close_timeout=timedelta(seconds=30),
        )
        result2 = await workflow.execute_activity(
            run_shell_command,
            "echo world",
            start_to_close_timeout=timedelta(seconds=30),
        )
        return result1 + result2


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
    c = await Client.connect("localhost:8133", namespace="test-pi-" + uuid.uuid4().hex[:8])
    yield c
    await c.close()


@pytest.mark.asyncio
async def test_pi_workflow(client):
    """Test the pi workflow that runs shell commands."""
    task_queue = "pi-queue-" + uuid.uuid4().hex[:8]

    worker = Worker(
        client,
        task_queue,
        workflows=[PiWorkflow],
        activities=[run_shell_command],
    )
    worker_task = asyncio.create_task(worker.run())

    try:
        result = await client.execute_workflow(
            PiWorkflow,
            id="pi-workflow-" + uuid.uuid4().hex[:8],
            task_queue=task_queue,
        )
        assert result == "hello\nworld\n"
    finally:
        worker.shutdown()
        await asyncio.sleep(0.5)
        worker_task.cancel()
        try:
            await worker_task
        except asyncio.CancelledError:
            pass
