"""Activity definition decorators and runtime."""
import contextvars
from typing import Callable, Optional


_current_activity_context: contextvars.ContextVar[Optional["_ActivityContext"]] = contextvars.ContextVar(
    "_current_activity_context", default=None
)


class _ActivityContext:
    def __init__(self, stub, task_token: bytes):
        self.stub = stub
        self.task_token = task_token
        self._heartbeat_details = None


def defn(fn: Callable = None, *, name: str = None):
    """Decorator to define an activity."""
    def wrapper(fn):
        act_name = name or fn.__name__
        fn.__tempiex_activity_name__ = act_name
        fn.__tempiex_activity__ = True
        return fn

    if fn is not None:
        return wrapper(fn)
    return wrapper


def heartbeat(*details):
    """Send a heartbeat from within an activity."""
    ctx = _current_activity_context.get()
    if ctx is None:
        return
    ctx._heartbeat_details = details
