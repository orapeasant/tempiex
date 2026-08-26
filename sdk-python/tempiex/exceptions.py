class TempiexError(Exception):
    """Base exception for Tempiex SDK."""
    pass


class WorkflowFailureError(TempiexError):
    """Raised when a workflow execution fails."""
    def __init__(self, message: str, failure_type: str = "", stack_trace: str = ""):
        super().__init__(message)
        self.failure_type = failure_type
        self.stack_trace = stack_trace
