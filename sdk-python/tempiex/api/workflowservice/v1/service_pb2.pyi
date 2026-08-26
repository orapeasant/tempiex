import datetime

from tempiex.api.common.v1 import message_pb2 as _message_pb2
from tempiex.api.history.v1 import message_pb2 as _message_pb2_1
from tempiex.api.enums.v1 import workflow_pb2 as _workflow_pb2
from tempiex.api.failure.v1 import message_pb2 as _message_pb2_1_1
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class RegisterNamespaceRequest(_message.Message):
    __slots__ = ("name", "description", "retention_days")
    NAME_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    RETENTION_DAYS_FIELD_NUMBER: _ClassVar[int]
    name: str
    description: str
    retention_days: int
    def __init__(self, name: _Optional[str] = ..., description: _Optional[str] = ..., retention_days: _Optional[int] = ...) -> None: ...

class RegisterNamespaceResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class DescribeNamespaceRequest(_message.Message):
    __slots__ = ("name",)
    NAME_FIELD_NUMBER: _ClassVar[int]
    name: str
    def __init__(self, name: _Optional[str] = ...) -> None: ...

class DescribeNamespaceResponse(_message.Message):
    __slots__ = ("id", "name", "description", "retention_days")
    ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    RETENTION_DAYS_FIELD_NUMBER: _ClassVar[int]
    id: str
    name: str
    description: str
    retention_days: int
    def __init__(self, id: _Optional[str] = ..., name: _Optional[str] = ..., description: _Optional[str] = ..., retention_days: _Optional[int] = ...) -> None: ...

class StartWorkflowExecutionRequest(_message.Message):
    __slots__ = ("namespace", "workflow_id", "workflow_type", "task_queue", "input", "workflow_execution_timeout", "workflow_run_timeout", "workflow_task_timeout", "request_id")
    NAMESPACE_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_ID_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_TYPE_FIELD_NUMBER: _ClassVar[int]
    TASK_QUEUE_FIELD_NUMBER: _ClassVar[int]
    INPUT_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_EXECUTION_TIMEOUT_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_RUN_TIMEOUT_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_TASK_TIMEOUT_FIELD_NUMBER: _ClassVar[int]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    namespace: str
    workflow_id: str
    workflow_type: _message_pb2.WorkflowType
    task_queue: str
    input: _message_pb2.Payloads
    workflow_execution_timeout: _duration_pb2.Duration
    workflow_run_timeout: _duration_pb2.Duration
    workflow_task_timeout: _duration_pb2.Duration
    request_id: str
    def __init__(self, namespace: _Optional[str] = ..., workflow_id: _Optional[str] = ..., workflow_type: _Optional[_Union[_message_pb2.WorkflowType, _Mapping]] = ..., task_queue: _Optional[str] = ..., input: _Optional[_Union[_message_pb2.Payloads, _Mapping]] = ..., workflow_execution_timeout: _Optional[_Union[datetime.timedelta, _duration_pb2.Duration, _Mapping]] = ..., workflow_run_timeout: _Optional[_Union[datetime.timedelta, _duration_pb2.Duration, _Mapping]] = ..., workflow_task_timeout: _Optional[_Union[datetime.timedelta, _duration_pb2.Duration, _Mapping]] = ..., request_id: _Optional[str] = ...) -> None: ...

class StartWorkflowExecutionResponse(_message.Message):
    __slots__ = ("run_id",)
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    run_id: str
    def __init__(self, run_id: _Optional[str] = ...) -> None: ...

class DescribeWorkflowExecutionRequest(_message.Message):
    __slots__ = ("namespace", "execution")
    NAMESPACE_FIELD_NUMBER: _ClassVar[int]
    EXECUTION_FIELD_NUMBER: _ClassVar[int]
    namespace: str
    execution: _message_pb2.WorkflowExecution
    def __init__(self, namespace: _Optional[str] = ..., execution: _Optional[_Union[_message_pb2.WorkflowExecution, _Mapping]] = ...) -> None: ...

class DescribeWorkflowExecutionResponse(_message.Message):
    __slots__ = ("execution", "status", "workflow_type", "start_time", "close_time", "task_queue")
    EXECUTION_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_TYPE_FIELD_NUMBER: _ClassVar[int]
    START_TIME_FIELD_NUMBER: _ClassVar[int]
    CLOSE_TIME_FIELD_NUMBER: _ClassVar[int]
    TASK_QUEUE_FIELD_NUMBER: _ClassVar[int]
    execution: _message_pb2.WorkflowExecution
    status: _workflow_pb2.WorkflowExecutionStatus
    workflow_type: _message_pb2.WorkflowType
    start_time: _timestamp_pb2.Timestamp
    close_time: _timestamp_pb2.Timestamp
    task_queue: str
    def __init__(self, execution: _Optional[_Union[_message_pb2.WorkflowExecution, _Mapping]] = ..., status: _Optional[_Union[_workflow_pb2.WorkflowExecutionStatus, str]] = ..., workflow_type: _Optional[_Union[_message_pb2.WorkflowType, _Mapping]] = ..., start_time: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., close_time: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., task_queue: _Optional[str] = ...) -> None: ...

class GetWorkflowExecutionHistoryRequest(_message.Message):
    __slots__ = ("namespace", "execution")
    NAMESPACE_FIELD_NUMBER: _ClassVar[int]
    EXECUTION_FIELD_NUMBER: _ClassVar[int]
    namespace: str
    execution: _message_pb2.WorkflowExecution
    def __init__(self, namespace: _Optional[str] = ..., execution: _Optional[_Union[_message_pb2.WorkflowExecution, _Mapping]] = ...) -> None: ...

class GetWorkflowExecutionHistoryResponse(_message.Message):
    __slots__ = ("history",)
    HISTORY_FIELD_NUMBER: _ClassVar[int]
    history: _message_pb2_1.History
    def __init__(self, history: _Optional[_Union[_message_pb2_1.History, _Mapping]] = ...) -> None: ...

class PollWorkflowTaskQueueRequest(_message.Message):
    __slots__ = ("namespace", "task_queue")
    NAMESPACE_FIELD_NUMBER: _ClassVar[int]
    TASK_QUEUE_FIELD_NUMBER: _ClassVar[int]
    namespace: str
    task_queue: str
    def __init__(self, namespace: _Optional[str] = ..., task_queue: _Optional[str] = ...) -> None: ...

class PollWorkflowTaskQueueResponse(_message.Message):
    __slots__ = ("task_token", "workflow_execution", "workflow_type", "started_event_id", "previous_started_event_id", "history")
    TASK_TOKEN_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_EXECUTION_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_TYPE_FIELD_NUMBER: _ClassVar[int]
    STARTED_EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    PREVIOUS_STARTED_EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    HISTORY_FIELD_NUMBER: _ClassVar[int]
    task_token: bytes
    workflow_execution: _message_pb2.WorkflowExecution
    workflow_type: _message_pb2.WorkflowType
    started_event_id: int
    previous_started_event_id: int
    history: _message_pb2_1.History
    def __init__(self, task_token: _Optional[bytes] = ..., workflow_execution: _Optional[_Union[_message_pb2.WorkflowExecution, _Mapping]] = ..., workflow_type: _Optional[_Union[_message_pb2.WorkflowType, _Mapping]] = ..., started_event_id: _Optional[int] = ..., previous_started_event_id: _Optional[int] = ..., history: _Optional[_Union[_message_pb2_1.History, _Mapping]] = ...) -> None: ...

class Command(_message.Message):
    __slots__ = ("command_type", "schedule_activity_task_command_attributes", "complete_workflow_execution_command_attributes", "fail_workflow_execution_command_attributes")
    COMMAND_TYPE_FIELD_NUMBER: _ClassVar[int]
    SCHEDULE_ACTIVITY_TASK_COMMAND_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    COMPLETE_WORKFLOW_EXECUTION_COMMAND_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    FAIL_WORKFLOW_EXECUTION_COMMAND_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    command_type: _workflow_pb2.CommandType
    schedule_activity_task_command_attributes: ScheduleActivityTaskCommandAttributes
    complete_workflow_execution_command_attributes: CompleteWorkflowExecutionCommandAttributes
    fail_workflow_execution_command_attributes: FailWorkflowExecutionCommandAttributes
    def __init__(self, command_type: _Optional[_Union[_workflow_pb2.CommandType, str]] = ..., schedule_activity_task_command_attributes: _Optional[_Union[ScheduleActivityTaskCommandAttributes, _Mapping]] = ..., complete_workflow_execution_command_attributes: _Optional[_Union[CompleteWorkflowExecutionCommandAttributes, _Mapping]] = ..., fail_workflow_execution_command_attributes: _Optional[_Union[FailWorkflowExecutionCommandAttributes, _Mapping]] = ...) -> None: ...

class ScheduleActivityTaskCommandAttributes(_message.Message):
    __slots__ = ("activity_id", "activity_type", "task_queue", "input", "start_to_close_timeout")
    ACTIVITY_ID_FIELD_NUMBER: _ClassVar[int]
    ACTIVITY_TYPE_FIELD_NUMBER: _ClassVar[int]
    TASK_QUEUE_FIELD_NUMBER: _ClassVar[int]
    INPUT_FIELD_NUMBER: _ClassVar[int]
    START_TO_CLOSE_TIMEOUT_FIELD_NUMBER: _ClassVar[int]
    activity_id: str
    activity_type: _message_pb2.ActivityType
    task_queue: str
    input: _message_pb2.Payloads
    start_to_close_timeout: _duration_pb2.Duration
    def __init__(self, activity_id: _Optional[str] = ..., activity_type: _Optional[_Union[_message_pb2.ActivityType, _Mapping]] = ..., task_queue: _Optional[str] = ..., input: _Optional[_Union[_message_pb2.Payloads, _Mapping]] = ..., start_to_close_timeout: _Optional[_Union[datetime.timedelta, _duration_pb2.Duration, _Mapping]] = ...) -> None: ...

class CompleteWorkflowExecutionCommandAttributes(_message.Message):
    __slots__ = ("result",)
    RESULT_FIELD_NUMBER: _ClassVar[int]
    result: _message_pb2.Payloads
    def __init__(self, result: _Optional[_Union[_message_pb2.Payloads, _Mapping]] = ...) -> None: ...

class FailWorkflowExecutionCommandAttributes(_message.Message):
    __slots__ = ("failure",)
    FAILURE_FIELD_NUMBER: _ClassVar[int]
    failure: _message_pb2_1_1.Failure
    def __init__(self, failure: _Optional[_Union[_message_pb2_1_1.Failure, _Mapping]] = ...) -> None: ...

class RespondWorkflowTaskCompletedRequest(_message.Message):
    __slots__ = ("task_token", "commands")
    TASK_TOKEN_FIELD_NUMBER: _ClassVar[int]
    COMMANDS_FIELD_NUMBER: _ClassVar[int]
    task_token: bytes
    commands: _containers.RepeatedCompositeFieldContainer[Command]
    def __init__(self, task_token: _Optional[bytes] = ..., commands: _Optional[_Iterable[_Union[Command, _Mapping]]] = ...) -> None: ...

class RespondWorkflowTaskCompletedResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class PollActivityTaskQueueRequest(_message.Message):
    __slots__ = ("namespace", "task_queue")
    NAMESPACE_FIELD_NUMBER: _ClassVar[int]
    TASK_QUEUE_FIELD_NUMBER: _ClassVar[int]
    namespace: str
    task_queue: str
    def __init__(self, namespace: _Optional[str] = ..., task_queue: _Optional[str] = ...) -> None: ...

class PollActivityTaskQueueResponse(_message.Message):
    __slots__ = ("task_token", "workflow_execution", "activity_type", "activity_id", "input", "scheduled_event_id", "started_event_id", "start_to_close_timeout", "heartbeat_details")
    TASK_TOKEN_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_EXECUTION_FIELD_NUMBER: _ClassVar[int]
    ACTIVITY_TYPE_FIELD_NUMBER: _ClassVar[int]
    ACTIVITY_ID_FIELD_NUMBER: _ClassVar[int]
    INPUT_FIELD_NUMBER: _ClassVar[int]
    SCHEDULED_EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    STARTED_EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    START_TO_CLOSE_TIMEOUT_FIELD_NUMBER: _ClassVar[int]
    HEARTBEAT_DETAILS_FIELD_NUMBER: _ClassVar[int]
    task_token: bytes
    workflow_execution: _message_pb2.WorkflowExecution
    activity_type: _message_pb2.ActivityType
    activity_id: str
    input: _message_pb2.Payloads
    scheduled_event_id: int
    started_event_id: int
    start_to_close_timeout: _duration_pb2.Duration
    heartbeat_details: _message_pb2.Payloads
    def __init__(self, task_token: _Optional[bytes] = ..., workflow_execution: _Optional[_Union[_message_pb2.WorkflowExecution, _Mapping]] = ..., activity_type: _Optional[_Union[_message_pb2.ActivityType, _Mapping]] = ..., activity_id: _Optional[str] = ..., input: _Optional[_Union[_message_pb2.Payloads, _Mapping]] = ..., scheduled_event_id: _Optional[int] = ..., started_event_id: _Optional[int] = ..., start_to_close_timeout: _Optional[_Union[datetime.timedelta, _duration_pb2.Duration, _Mapping]] = ..., heartbeat_details: _Optional[_Union[_message_pb2.Payloads, _Mapping]] = ...) -> None: ...

class RespondActivityTaskCompletedRequest(_message.Message):
    __slots__ = ("task_token", "result")
    TASK_TOKEN_FIELD_NUMBER: _ClassVar[int]
    RESULT_FIELD_NUMBER: _ClassVar[int]
    task_token: bytes
    result: _message_pb2.Payloads
    def __init__(self, task_token: _Optional[bytes] = ..., result: _Optional[_Union[_message_pb2.Payloads, _Mapping]] = ...) -> None: ...

class RespondActivityTaskCompletedResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class RespondActivityTaskFailedRequest(_message.Message):
    __slots__ = ("task_token", "failure")
    TASK_TOKEN_FIELD_NUMBER: _ClassVar[int]
    FAILURE_FIELD_NUMBER: _ClassVar[int]
    task_token: bytes
    failure: _message_pb2_1_1.Failure
    def __init__(self, task_token: _Optional[bytes] = ..., failure: _Optional[_Union[_message_pb2_1_1.Failure, _Mapping]] = ...) -> None: ...

class RespondActivityTaskFailedResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class RecordActivityTaskHeartbeatRequest(_message.Message):
    __slots__ = ("task_token", "details")
    TASK_TOKEN_FIELD_NUMBER: _ClassVar[int]
    DETAILS_FIELD_NUMBER: _ClassVar[int]
    task_token: bytes
    details: _message_pb2.Payloads
    def __init__(self, task_token: _Optional[bytes] = ..., details: _Optional[_Union[_message_pb2.Payloads, _Mapping]] = ...) -> None: ...

class RecordActivityTaskHeartbeatResponse(_message.Message):
    __slots__ = ("cancel_requested",)
    CANCEL_REQUESTED_FIELD_NUMBER: _ClassVar[int]
    cancel_requested: bool
    def __init__(self, cancel_requested: _Optional[bool] = ...) -> None: ...
