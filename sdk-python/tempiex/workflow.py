"""Workflow definition decorators and runtime."""
import asyncio
import contextvars
from typing import Any, Callable, Optional, Type
from datetime import timedelta

from tempiex.converter import encode_payloads, decode_payloads, decode_single
from tempiex.api.enums.v1 import workflow_pb2


_current_workflow_context: contextvars.ContextVar[Optional["_WorkflowContext"]] = contextvars.ContextVar(
    "_current_workflow_context", default=None
)


class _WorkflowContext:
    """Runtime context for a workflow execution."""

    def __init__(self):
        self.commands = []
        self.pending_activities = {}
        self.completed_activities = {}
        self.activity_counter = 0
        self.task_queue = ""
        # Signal queue: name -> list of (args, kwargs)
        self.signal_queue: dict[str, list] = {}
        # Query handlers resolved synchronously (no await)
        self.query_results: dict[str, Any] = {}

    def next_activity_id(self) -> str:
        self.activity_counter += 1
        return str(self.activity_counter)

    def deliver_signal(self, name: str, args: list):
        self.signal_queue.setdefault(name, []).append(args)

    def drain_signal(self, name: str) -> list | None:
        queue = self.signal_queue.get(name)
        if queue:
            return queue.pop(0)
        return None


def defn(cls: Type = None, *, name: str = None):
    """Decorator to mark a class as a Workflow definition.

    Usage:
        @workflow.defn
        class MyWorkflow: ...

        @workflow.defn(name="custom-name")
        class MyWorkflow: ...
    """
    def wrapper(c: Type) -> Type:
        wf_name = name or c.__name__
        c.__tempiex_workflow_name__ = wf_name
        c.__tempiex_workflow__ = True
        return c

    if cls is not None:
        return wrapper(cls)
    return wrapper


def run(fn: Callable) -> Callable:
    """Decorator to mark the workflow entry-point method (must be async)."""
    fn.__tempiex_workflow_run__ = True
    return fn


def signal(fn: Callable = None, *, name: str = None) -> Callable:
    """Decorator to mark a workflow method as a Signal handler.

    Usage:
        @workflow.signal
        async def approve(self, reason: str): ...

        @workflow.signal(name="custom-signal")
        async def approve(self, reason: str): ...
    """
    def wrapper(f: Callable) -> Callable:
        sig_name = name or f.__name__
        f.__tempiex_signal__ = True
        f.__tempiex_signal_name__ = sig_name
        return f

    if fn is not None:
        return wrapper(fn)
    return wrapper


def query(fn: Callable = None, *, name: str = None) -> Callable:
    """Decorator to mark a workflow method as a Query handler (sync, no side effects).

    Usage:
        @workflow.query
        def get_status(self) -> str: ...

        @workflow.query(name="custom-query")
        def get_status(self) -> str: ...
    """
    def wrapper(f: Callable) -> Callable:
        q_name = name or f.__name__
        f.__tempiex_query__ = True
        f.__tempiex_query_name__ = q_name
        return f

    if fn is not None:
        return wrapper(fn)
    return wrapper


def update(fn: Callable = None, *, name: str = None) -> Callable:
    """Decorator to mark a workflow method as an Update handler (async, returns value).

    Usage:
        @workflow.update
        async def add_item(self, item: str) -> int: ...
    """
    def wrapper(f: Callable) -> Callable:
        upd_name = name or f.__name__
        f.__tempiex_update__ = True
        f.__tempiex_update_name__ = upd_name
        return f

    if fn is not None:
        return wrapper(fn)
    return wrapper


async def execute_activity(activity, *args, start_to_close_timeout: timedelta = None, task_queue: str = None) -> Any:
    """Execute an activity from within a workflow."""
    ctx = _current_workflow_context.get()
    if ctx is None:
        raise RuntimeError("execute_activity must be called from within a workflow")

    activity_name = _get_activity_name(activity)
    activity_id = ctx.next_activity_id()

    if activity_id in ctx.completed_activities:
        result_or_exc = ctx.completed_activities[activity_id]
        if isinstance(result_or_exc, Exception):
            raise result_or_exc
        return result_or_exc

    from tempiex.api.workflowservice.v1 import service_pb2
    from tempiex.api.common.v1 import message_pb2
    from google.protobuf.duration_pb2 import Duration

    timeout_duration = None
    if start_to_close_timeout:
        timeout_duration = Duration()
        timeout_duration.FromTimedelta(start_to_close_timeout)

    tq = task_queue or ctx.task_queue or ""

    cmd = service_pb2.Command(
        command_type=workflow_pb2.COMMAND_TYPE_SCHEDULE_ACTIVITY_TASK,
        schedule_activity_task_command_attributes=service_pb2.ScheduleActivityTaskCommandAttributes(
            activity_id=activity_id,
            activity_type=message_pb2.ActivityType(name=activity_name),
            task_queue=tq,
            input=encode_payloads(list(args)),
            start_to_close_timeout=timeout_duration,
        ),
    )
    ctx.commands.append(cmd)

    future = asyncio.get_event_loop().create_future()
    ctx.pending_activities[activity_id] = future
    return await future


async def wait_condition(condition: Callable[[], bool], *, timeout: timedelta = None) -> bool:
    """Suspend the workflow until condition() returns True.

    Returns True when condition is met, False on timeout (if timeout provided).

    Usage:
        await workflow.wait_condition(lambda: self._approved)
        timed_out = not await workflow.wait_condition(lambda: self._done, timeout=timedelta(hours=1))
    """
    import time
    deadline = None
    if timeout is not None:
        deadline = asyncio.get_event_loop().time() + timeout.total_seconds()

    while not condition():
        if deadline is not None and asyncio.get_event_loop().time() >= deadline:
            return False
        await asyncio.sleep(0)  # yield to event loop
    return True


def _get_activity_name(activity) -> str:
    if hasattr(activity, "__tempiex_activity_name__"):
        return activity.__tempiex_activity_name__
    if hasattr(activity, "__name__"):
        return activity.__name__
    return str(activity)


def _get_signal_handlers(instance) -> dict[str, Callable]:
    """Collect all @workflow.signal methods from a workflow instance."""
    handlers = {}
    for attr_name in dir(type(instance)):
        method = getattr(type(instance), attr_name, None)
        if method and getattr(method, "__tempiex_signal__", False):
            handlers[method.__tempiex_signal_name__] = getattr(instance, attr_name)
    return handlers


def _get_query_handlers(instance) -> dict[str, Callable]:
    """Collect all @workflow.query methods from a workflow instance."""
    handlers = {}
    for attr_name in dir(type(instance)):
        method = getattr(type(instance), attr_name, None)
        if method and getattr(method, "__tempiex_query__", False):
            handlers[method.__tempiex_query_name__] = getattr(instance, attr_name)
    return handlers


def _get_update_handlers(instance) -> dict[str, Callable]:
    """Collect all @workflow.update methods from a workflow instance."""
    handlers = {}
    for attr_name in dir(type(instance)):
        method = getattr(type(instance), attr_name, None)
        if method and getattr(method, "__tempiex_update__", False):
            handlers[method.__tempiex_update_name__] = getattr(instance, attr_name)
    return handlers



_current_workflow_context: contextvars.ContextVar[Optional["_WorkflowContext"]] = contextvars.ContextVar(
    "_current_workflow_context", default=None
)

