"""Worker that polls for and executes workflow and activity tasks."""
import asyncio
import logging
import traceback
from typing import Any, Dict, List, Type

import grpc

from tempiex.api.workflowservice.v1 import service_pb2, service_pb2_grpc
from tempiex.api.common.v1 import message_pb2
from tempiex.api.enums.v1 import workflow_pb2
from tempiex.api.failure.v1 import message_pb2 as failure_pb2
from tempiex.converter import encode_payloads, decode_payloads, decode_single
from tempiex import workflow as wf_module
from tempiex import activity as act_module

logger = logging.getLogger("tempiex.worker")


class Worker:
    """Polls the server for tasks and executes workflows and activities."""

    def __init__(self, client, task_queue: str, workflows: list = None, activities: list = None):
        self._client = client
        self._task_queue = task_queue
        self._workflows: Dict[str, Type] = {}
        self._activities: Dict[str, Any] = {}
        self._stopping = False
        self._tasks: List[asyncio.Task] = []

        for wf in (workflows or []):
            name = getattr(wf, "__tempiex_workflow_name__", wf.__name__)
            self._workflows[name] = wf

        for act in (activities or []):
            name = getattr(act, "__tempiex_activity_name__", act.__name__)
            self._activities[name] = act

    async def run(self):
        """Run the worker until stopped."""
        self._stopping = False
        self._tasks = [
            asyncio.create_task(self._poll_workflow_tasks()),
            asyncio.create_task(self._poll_activity_tasks()),
        ]
        try:
            await asyncio.gather(*self._tasks, return_exceptions=True)
        except asyncio.CancelledError:
            pass

    def shutdown(self):
        """Signal the worker to stop."""
        self._stopping = True
        for t in self._tasks:
            t.cancel()

    async def _poll_workflow_tasks(self):
        """Poll for workflow tasks."""
        while not self._stopping:
            try:
                resp = await self._client._stub.PollWorkflowTaskQueue(
                    service_pb2.PollWorkflowTaskQueueRequest(
                        namespace=self._client._namespace,
                        task_queue=self._task_queue,
                    ),
                    timeout=35,
                )
                if resp.task_token:
                    await self._execute_workflow_task(resp)
            except grpc.aio.AioRpcError as e:
                if e.code() == grpc.StatusCode.DEADLINE_EXCEEDED:
                    continue
                if e.code() == grpc.StatusCode.CANCELLED:
                    break
                logger.error(f"Workflow poll error: {e}")
                await asyncio.sleep(1)
            except asyncio.CancelledError:
                break
            except Exception as e:
                logger.error(f"Workflow poll error: {e}")
                await asyncio.sleep(1)

    async def _execute_workflow_task(self, task):
        """Execute a single workflow task by replaying history."""
        try:
            wf_type = task.workflow_type.name
            wf_class = self._workflows.get(wf_type)
            if not wf_class:
                logger.error(f"Unknown workflow type: {wf_type}")
                return

            events = list(task.history.events) if task.history else []

            # Build replay context
            ctx = wf_module._WorkflowContext()
            ctx.task_queue = self._task_queue

            # Find workflow input from WorkflowExecutionStarted event
            wf_input = []
            for event in events:
                if event.event_type == workflow_pb2.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED:
                    attrs = event.workflow_execution_started_event_attributes
                    wf_input = decode_payloads(attrs.input)
                    break

            # Build completed activities map and signal queue from history
            scheduled_to_activity_id = {}
            for event in events:
                if event.event_type == workflow_pb2.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED:
                    attrs = event.activity_task_scheduled_event_attributes
                    scheduled_to_activity_id[event.event_id] = attrs.activity_id
                elif event.event_type == workflow_pb2.EVENT_TYPE_ACTIVITY_TASK_COMPLETED:
                    attrs = event.activity_task_completed_event_attributes
                    sched_id = attrs.scheduled_event_id
                    act_id = scheduled_to_activity_id.get(sched_id, str(sched_id))
                    result = decode_single(attrs.result)
                    ctx.completed_activities[act_id] = result
                elif event.event_type == workflow_pb2.EVENT_TYPE_ACTIVITY_TASK_FAILED:
                    attrs = event.activity_task_failed_event_attributes
                    sched_id = attrs.scheduled_event_id
                    act_id = scheduled_to_activity_id.get(sched_id, str(sched_id))
                    ctx.completed_activities[act_id] = Exception(
                        attrs.failure.message if attrs.failure else "Activity failed"
                    )
                elif event.event_type == workflow_pb2.EVENT_TYPE_WORKFLOW_EXECUTION_SIGNALED:
                    attrs = event.workflow_execution_signaled_event_attributes
                    sig_args = decode_payloads(attrs.input) if attrs.input else []
                    ctx.deliver_signal(attrs.signal_name, sig_args)

            token = wf_module._current_workflow_context.set(ctx)
            try:
                wf_instance = wf_class()

                # Collect signal/query/update handlers from annotations
                signal_handlers = wf_module._get_signal_handlers(wf_instance)
                query_handlers = wf_module._get_query_handlers(wf_instance)

                # Deliver queued signals to handlers before running
                for sig_name, handler in signal_handlers.items():
                    while True:
                        args = ctx.drain_signal(sig_name)
                        if args is None:
                            break
                        if asyncio.iscoroutinefunction(handler):
                            await handler(*args)
                        else:
                            handler(*args)

                run_method = None
                for attr_name in dir(wf_instance):
                    attr = getattr(wf_instance, attr_name)
                    if callable(attr) and getattr(attr, "__tempiex_workflow_run__", False):
                        run_method = attr
                        break

                if not run_method:
                    raise RuntimeError(f"No @workflow.run method found on {wf_type}")

                try:
                    if wf_input:
                        result = await asyncio.wait_for(run_method(*wf_input), timeout=5.0)
                    else:
                        result = await asyncio.wait_for(run_method(), timeout=5.0)

                    # Workflow completed
                    cmd = service_pb2.Command(
                        command_type=workflow_pb2.COMMAND_TYPE_COMPLETE_WORKFLOW_EXECUTION,
                        complete_workflow_execution_command_attributes=service_pb2.CompleteWorkflowExecutionCommandAttributes(
                            result=encode_payloads([result]),
                        ),
                    )
                    ctx.commands.append(cmd)
                except asyncio.TimeoutError:
                    # Workflow is blocked waiting for activities - commands collected so far will be sent
                    pass
                except Exception as e:
                    cmd = service_pb2.Command(
                        command_type=workflow_pb2.COMMAND_TYPE_FAIL_WORKFLOW_EXECUTION,
                        fail_workflow_execution_command_attributes=service_pb2.FailWorkflowExecutionCommandAttributes(
                            failure=failure_pb2.Failure(
                                message=str(e),
                                type=type(e).__name__,
                                stack_trace=traceback.format_exc(),
                            ),
                        ),
                    )
                    ctx.commands.append(cmd)
            finally:
                wf_module._current_workflow_context.reset(token)

            await self._client._stub.RespondWorkflowTaskCompleted(
                service_pb2.RespondWorkflowTaskCompletedRequest(
                    task_token=task.task_token,
                    commands=ctx.commands,
                ),
            )
        except Exception as e:
            logger.error(f"Error executing workflow task: {e}", exc_info=True)

    async def _poll_activity_tasks(self):
        """Poll for activity tasks."""
        while not self._stopping:
            try:
                resp = await self._client._stub.PollActivityTaskQueue(
                    service_pb2.PollActivityTaskQueueRequest(
                        namespace=self._client._namespace,
                        task_queue=self._task_queue,
                    ),
                    timeout=35,
                )
                if resp.task_token:
                    await self._execute_activity_task(resp)
            except grpc.aio.AioRpcError as e:
                if e.code() == grpc.StatusCode.DEADLINE_EXCEEDED:
                    continue
                if e.code() == grpc.StatusCode.CANCELLED:
                    break
                logger.error(f"Activity poll error: {e}")
                await asyncio.sleep(1)
            except asyncio.CancelledError:
                break
            except Exception as e:
                logger.error(f"Activity poll error: {e}")
                await asyncio.sleep(1)

    async def _execute_activity_task(self, task):
        """Execute a single activity task."""
        try:
            act_type = task.activity_type.name
            act_fn = self._activities.get(act_type)
            if not act_fn:
                logger.error(f"Unknown activity type: {act_type}")
                await self._fail_activity(task.task_token, f"Unknown activity: {act_type}")
                return

            act_ctx = act_module._ActivityContext(self._client._stub, task.task_token)
            token = act_module._current_activity_context.set(act_ctx)

            try:
                args = decode_payloads(task.input)

                if asyncio.iscoroutinefunction(act_fn):
                    result = await act_fn(*args)
                else:
                    result = await asyncio.get_event_loop().run_in_executor(None, lambda: act_fn(*args))

                await self._client._stub.RespondActivityTaskCompleted(
                    service_pb2.RespondActivityTaskCompletedRequest(
                        task_token=task.task_token,
                        result=encode_payloads([result]),
                    ),
                )
            except Exception as e:
                await self._fail_activity(task.task_token, str(e))
            finally:
                act_module._current_activity_context.reset(token)
        except Exception as e:
            logger.error(f"Error executing activity task: {e}", exc_info=True)

    async def _fail_activity(self, task_token: bytes, message: str):
        """Report activity failure."""
        try:
            await self._client._stub.RespondActivityTaskFailed(
                service_pb2.RespondActivityTaskFailedRequest(
                    task_token=task_token,
                    failure=failure_pb2.Failure(
                        message=message,
                        type="ActivityError",
                    ),
                ),
            )
        except Exception as e:
            logger.error(f"Error failing activity: {e}", exc_info=True)
