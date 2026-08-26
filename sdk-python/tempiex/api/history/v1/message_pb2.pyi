import datetime

from tempiex.api.enums.v1 import workflow_pb2 as _workflow_pb2
from tempiex.api.common.v1 import message_pb2 as _message_pb2
from tempiex.api.failure.v1 import message_pb2 as _message_pb2_1
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class HistoryEvent(_message.Message):
    __slots__ = ("event_id", "event_time", "event_type", "workflow_execution_started_event_attributes", "workflow_task_scheduled_event_attributes", "workflow_task_started_event_attributes", "workflow_task_completed_event_attributes", "activity_task_scheduled_event_attributes", "activity_task_started_event_attributes", "activity_task_completed_event_attributes", "activity_task_failed_event_attributes", "workflow_execution_completed_event_attributes", "workflow_execution_failed_event_attributes")
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    EVENT_TIME_FIELD_NUMBER: _ClassVar[int]
    EVENT_TYPE_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_EXECUTION_STARTED_EVENT_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_TASK_SCHEDULED_EVENT_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_TASK_STARTED_EVENT_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_TASK_COMPLETED_EVENT_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    ACTIVITY_TASK_SCHEDULED_EVENT_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    ACTIVITY_TASK_STARTED_EVENT_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    ACTIVITY_TASK_COMPLETED_EVENT_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    ACTIVITY_TASK_FAILED_EVENT_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_EXECUTION_COMPLETED_EVENT_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_EXECUTION_FAILED_EVENT_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    event_id: int
    event_time: _timestamp_pb2.Timestamp
    event_type: _workflow_pb2.EventType
    workflow_execution_started_event_attributes: WorkflowExecutionStartedEventAttributes
    workflow_task_scheduled_event_attributes: WorkflowTaskScheduledEventAttributes
    workflow_task_started_event_attributes: WorkflowTaskStartedEventAttributes
    workflow_task_completed_event_attributes: WorkflowTaskCompletedEventAttributes
    activity_task_scheduled_event_attributes: ActivityTaskScheduledEventAttributes
    activity_task_started_event_attributes: ActivityTaskStartedEventAttributes
    activity_task_completed_event_attributes: ActivityTaskCompletedEventAttributes
    activity_task_failed_event_attributes: ActivityTaskFailedEventAttributes
    workflow_execution_completed_event_attributes: WorkflowExecutionCompletedEventAttributes
    workflow_execution_failed_event_attributes: WorkflowExecutionFailedEventAttributes
    def __init__(self, event_id: _Optional[int] = ..., event_time: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., event_type: _Optional[_Union[_workflow_pb2.EventType, str]] = ..., workflow_execution_started_event_attributes: _Optional[_Union[WorkflowExecutionStartedEventAttributes, _Mapping]] = ..., workflow_task_scheduled_event_attributes: _Optional[_Union[WorkflowTaskScheduledEventAttributes, _Mapping]] = ..., workflow_task_started_event_attributes: _Optional[_Union[WorkflowTaskStartedEventAttributes, _Mapping]] = ..., workflow_task_completed_event_attributes: _Optional[_Union[WorkflowTaskCompletedEventAttributes, _Mapping]] = ..., activity_task_scheduled_event_attributes: _Optional[_Union[ActivityTaskScheduledEventAttributes, _Mapping]] = ..., activity_task_started_event_attributes: _Optional[_Union[ActivityTaskStartedEventAttributes, _Mapping]] = ..., activity_task_completed_event_attributes: _Optional[_Union[ActivityTaskCompletedEventAttributes, _Mapping]] = ..., activity_task_failed_event_attributes: _Optional[_Union[ActivityTaskFailedEventAttributes, _Mapping]] = ..., workflow_execution_completed_event_attributes: _Optional[_Union[WorkflowExecutionCompletedEventAttributes, _Mapping]] = ..., workflow_execution_failed_event_attributes: _Optional[_Union[WorkflowExecutionFailedEventAttributes, _Mapping]] = ...) -> None: ...

class History(_message.Message):
    __slots__ = ("events",)
    EVENTS_FIELD_NUMBER: _ClassVar[int]
    events: _containers.RepeatedCompositeFieldContainer[HistoryEvent]
    def __init__(self, events: _Optional[_Iterable[_Union[HistoryEvent, _Mapping]]] = ...) -> None: ...

class WorkflowExecutionStartedEventAttributes(_message.Message):
    __slots__ = ("workflow_type", "task_queue", "input", "workflow_id")
    WORKFLOW_TYPE_FIELD_NUMBER: _ClassVar[int]
    TASK_QUEUE_FIELD_NUMBER: _ClassVar[int]
    INPUT_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_ID_FIELD_NUMBER: _ClassVar[int]
    workflow_type: _message_pb2.WorkflowType
    task_queue: str
    input: _message_pb2.Payloads
    workflow_id: str
    def __init__(self, workflow_type: _Optional[_Union[_message_pb2.WorkflowType, _Mapping]] = ..., task_queue: _Optional[str] = ..., input: _Optional[_Union[_message_pb2.Payloads, _Mapping]] = ..., workflow_id: _Optional[str] = ...) -> None: ...

class WorkflowTaskScheduledEventAttributes(_message.Message):
    __slots__ = ("task_queue",)
    TASK_QUEUE_FIELD_NUMBER: _ClassVar[int]
    task_queue: str
    def __init__(self, task_queue: _Optional[str] = ...) -> None: ...

class WorkflowTaskStartedEventAttributes(_message.Message):
    __slots__ = ("scheduled_event_id",)
    SCHEDULED_EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    scheduled_event_id: int
    def __init__(self, scheduled_event_id: _Optional[int] = ...) -> None: ...

class WorkflowTaskCompletedEventAttributes(_message.Message):
    __slots__ = ("scheduled_event_id", "started_event_id")
    SCHEDULED_EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    STARTED_EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    scheduled_event_id: int
    started_event_id: int
    def __init__(self, scheduled_event_id: _Optional[int] = ..., started_event_id: _Optional[int] = ...) -> None: ...

class ActivityTaskScheduledEventAttributes(_message.Message):
    __slots__ = ("activity_id", "activity_type", "task_queue", "input", "start_to_close_timeout", "workflow_task_completed_event_id")
    ACTIVITY_ID_FIELD_NUMBER: _ClassVar[int]
    ACTIVITY_TYPE_FIELD_NUMBER: _ClassVar[int]
    TASK_QUEUE_FIELD_NUMBER: _ClassVar[int]
    INPUT_FIELD_NUMBER: _ClassVar[int]
    START_TO_CLOSE_TIMEOUT_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_TASK_COMPLETED_EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    activity_id: str
    activity_type: _message_pb2.ActivityType
    task_queue: str
    input: _message_pb2.Payloads
    start_to_close_timeout: _duration_pb2.Duration
    workflow_task_completed_event_id: int
    def __init__(self, activity_id: _Optional[str] = ..., activity_type: _Optional[_Union[_message_pb2.ActivityType, _Mapping]] = ..., task_queue: _Optional[str] = ..., input: _Optional[_Union[_message_pb2.Payloads, _Mapping]] = ..., start_to_close_timeout: _Optional[_Union[datetime.timedelta, _duration_pb2.Duration, _Mapping]] = ..., workflow_task_completed_event_id: _Optional[int] = ...) -> None: ...

class ActivityTaskStartedEventAttributes(_message.Message):
    __slots__ = ("scheduled_event_id",)
    SCHEDULED_EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    scheduled_event_id: int
    def __init__(self, scheduled_event_id: _Optional[int] = ...) -> None: ...

class ActivityTaskCompletedEventAttributes(_message.Message):
    __slots__ = ("scheduled_event_id", "started_event_id", "result")
    SCHEDULED_EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    STARTED_EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    RESULT_FIELD_NUMBER: _ClassVar[int]
    scheduled_event_id: int
    started_event_id: int
    result: _message_pb2.Payloads
    def __init__(self, scheduled_event_id: _Optional[int] = ..., started_event_id: _Optional[int] = ..., result: _Optional[_Union[_message_pb2.Payloads, _Mapping]] = ...) -> None: ...

class ActivityTaskFailedEventAttributes(_message.Message):
    __slots__ = ("scheduled_event_id", "started_event_id", "failure")
    SCHEDULED_EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    STARTED_EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    FAILURE_FIELD_NUMBER: _ClassVar[int]
    scheduled_event_id: int
    started_event_id: int
    failure: _message_pb2_1.Failure
    def __init__(self, scheduled_event_id: _Optional[int] = ..., started_event_id: _Optional[int] = ..., failure: _Optional[_Union[_message_pb2_1.Failure, _Mapping]] = ...) -> None: ...

class WorkflowExecutionCompletedEventAttributes(_message.Message):
    __slots__ = ("result", "workflow_task_completed_event_id")
    RESULT_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_TASK_COMPLETED_EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    result: _message_pb2.Payloads
    workflow_task_completed_event_id: int
    def __init__(self, result: _Optional[_Union[_message_pb2.Payloads, _Mapping]] = ..., workflow_task_completed_event_id: _Optional[int] = ...) -> None: ...

class WorkflowExecutionFailedEventAttributes(_message.Message):
    __slots__ = ("failure", "workflow_task_completed_event_id")
    FAILURE_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_TASK_COMPLETED_EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    failure: _message_pb2_1.Failure
    workflow_task_completed_event_id: int
    def __init__(self, failure: _Optional[_Union[_message_pb2_1.Failure, _Mapping]] = ..., workflow_task_completed_event_id: _Optional[int] = ...) -> None: ...
