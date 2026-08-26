"""Tempiex client for connecting to the server."""
import asyncio
import grpc
import uuid

from tempiex.api.workflowservice.v1 import service_pb2, service_pb2_grpc
from tempiex.api.common.v1 import message_pb2
from tempiex.api.enums.v1 import workflow_pb2
from tempiex.converter import encode_payloads, decode_payloads, decode_single
from tempiex.exceptions import WorkflowFailureError


class WorkflowHandle:
    """Handle to a running or completed workflow execution."""

    def __init__(self, client: "Client", workflow_id: str, run_id: str):
        self._client = client
        self.workflow_id = workflow_id
        self.run_id = run_id

    async def result(self, timeout: float = 60.0):
        """Wait for the workflow to complete and return its result."""
        deadline = asyncio.get_event_loop().time() + timeout
        while True:
            desc = await self.describe()
            status = desc.status
            if status == workflow_pb2.WORKFLOW_EXECUTION_STATUS_COMPLETED:
                hist_resp = await self._client._stub.GetWorkflowExecutionHistory(
                    service_pb2.GetWorkflowExecutionHistoryRequest(
                        namespace=self._client._namespace,
                        execution=message_pb2.WorkflowExecution(
                            workflow_id=self.workflow_id,
                            run_id=self.run_id,
                        ),
                    )
                )
                for event in reversed(hist_resp.history.events):
                    if event.event_type == workflow_pb2.EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED:
                        attrs = event.workflow_execution_completed_event_attributes
                        return decode_single(attrs.result)
                return None
            elif status == workflow_pb2.WORKFLOW_EXECUTION_STATUS_FAILED:
                hist_resp = await self._client._stub.GetWorkflowExecutionHistory(
                    service_pb2.GetWorkflowExecutionHistoryRequest(
                        namespace=self._client._namespace,
                        execution=message_pb2.WorkflowExecution(
                            workflow_id=self.workflow_id,
                            run_id=self.run_id,
                        ),
                    )
                )
                for event in reversed(hist_resp.history.events):
                    if event.event_type == workflow_pb2.EVENT_TYPE_WORKFLOW_EXECUTION_FAILED:
                        attrs = event.workflow_execution_failed_event_attributes
                        raise WorkflowFailureError(
                            attrs.failure.message,
                            attrs.failure.type,
                            attrs.failure.stack_trace,
                        )
                raise WorkflowFailureError("Workflow failed (unknown reason)")

            if asyncio.get_event_loop().time() > deadline:
                raise TimeoutError(f"Timed out waiting for workflow {self.workflow_id}")
            await asyncio.sleep(0.5)

    async def describe(self):
        """Describe the workflow execution."""
        return await self._client._stub.DescribeWorkflowExecution(
            service_pb2.DescribeWorkflowExecutionRequest(
                namespace=self._client._namespace,
                execution=message_pb2.WorkflowExecution(
                    workflow_id=self.workflow_id,
                    run_id=self.run_id,
                ),
            )
        )

    async def signal(self, signal_name: str, *args) -> None:
        """Send a signal to this workflow execution."""
        from tempiex.converter import encode_payloads
        await self._client._stub.SignalWorkflowExecution(
            service_pb2.SignalWorkflowExecutionRequest(
                namespace=self._client._namespace,
                workflow_execution=message_pb2.WorkflowExecution(
                    workflow_id=self.workflow_id,
                    run_id=self.run_id,
                ),
                signal_name=signal_name,
                input=encode_payloads(list(args)),
            )
        )

    async def query(self, query_type: str, *args):
        """Send a synchronous query to this workflow execution."""
        from tempiex.converter import encode_payloads, decode_single
        resp = await self._client._stub.QueryWorkflow(
            service_pb2.QueryWorkflowRequest(
                namespace=self._client._namespace,
                execution=message_pb2.WorkflowExecution(
                    workflow_id=self.workflow_id,
                    run_id=self.run_id,
                ),
                query=service_pb2.WorkflowQuery(
                    query_type=query_type,
                    query_args=encode_payloads(list(args)),
                ),
            )
        )
        return decode_single(resp.query_result)

    async def cancel(self, reason: str = "") -> None:
        """Request graceful cancellation of this workflow execution."""
        await self._client._stub.RequestCancelWorkflowExecution(
            service_pb2.RequestCancelWorkflowExecutionRequest(
                namespace=self._client._namespace,
                workflow_execution=message_pb2.WorkflowExecution(
                    workflow_id=self.workflow_id,
                    run_id=self.run_id,
                ),
                reason=reason,
            )
        )

    async def terminate(self, reason: str = "") -> None:
        """Forcefully terminate this workflow execution."""
        await self._client._stub.TerminateWorkflowExecution(
            service_pb2.TerminateWorkflowExecutionRequest(
                namespace=self._client._namespace,
                workflow_execution=message_pb2.WorkflowExecution(
                    workflow_id=self.workflow_id,
                    run_id=self.run_id,
                ),
                reason=reason,
            )
        )


class Client:
    """Tempiex client."""

    def __init__(self, stub, namespace: str, channel):
        self._stub = stub
        self._namespace = namespace
        self._channel = channel

    @staticmethod
    async def connect(target: str, namespace: str = "default") -> "Client":
        """Connect to a Tempiex server."""
        channel = grpc.aio.insecure_channel(target)
        stub = service_pb2_grpc.WorkflowServiceStub(channel)
        client = Client(stub, namespace, channel)

        try:
            await stub.RegisterNamespace(
                service_pb2.RegisterNamespaceRequest(
                    name=namespace,
                    retention_days=7,
                )
            )
        except grpc.aio.AioRpcError:
            pass  # Already exists

        return client

    async def start_workflow(self, workflow, *args, id: str, task_queue: str) -> WorkflowHandle:
        """Start a workflow execution."""
        wf_name = _get_workflow_name(workflow)
        input_payloads = encode_payloads(list(args)) if args else None

        resp = await self._stub.StartWorkflowExecution(
            service_pb2.StartWorkflowExecutionRequest(
                namespace=self._namespace,
                workflow_id=id,
                workflow_type=message_pb2.WorkflowType(name=wf_name),
                task_queue=task_queue,
                input=input_payloads,
                request_id=str(uuid.uuid4()),
            )
        )
        return WorkflowHandle(self, id, resp.run_id)

    async def execute_workflow(self, workflow, *args, id: str, task_queue: str):
        """Start a workflow and wait for its result."""
        handle = await self.start_workflow(workflow, *args, id=id, task_queue=task_queue)
        return await handle.result()

    def get_workflow_handle(self, workflow_id: str, run_id: str = "") -> WorkflowHandle:
        """Get a handle to an existing workflow."""
        return WorkflowHandle(self, workflow_id, run_id)

    async def close(self):
        """Close the client connection."""
        await self._channel.close()


def _get_workflow_name(workflow) -> str:
    """Get the workflow type name from a decorated class."""
    if hasattr(workflow, "__tempiex_workflow_name__"):
        return workflow.__tempiex_workflow_name__
    if hasattr(workflow, "__name__"):
        return workflow.__name__
    return type(workflow).__name__
