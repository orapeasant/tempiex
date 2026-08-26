from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from typing import ClassVar as _ClassVar

DESCRIPTOR: _descriptor.FileDescriptor

class WorkflowExecutionStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    WORKFLOW_EXECUTION_STATUS_UNSPECIFIED: _ClassVar[WorkflowExecutionStatus]
    WORKFLOW_EXECUTION_STATUS_RUNNING: _ClassVar[WorkflowExecutionStatus]
    WORKFLOW_EXECUTION_STATUS_COMPLETED: _ClassVar[WorkflowExecutionStatus]
    WORKFLOW_EXECUTION_STATUS_FAILED: _ClassVar[WorkflowExecutionStatus]
    WORKFLOW_EXECUTION_STATUS_CANCELLED: _ClassVar[WorkflowExecutionStatus]
    WORKFLOW_EXECUTION_STATUS_TIMED_OUT: _ClassVar[WorkflowExecutionStatus]
    WORKFLOW_EXECUTION_STATUS_TERMINATED: _ClassVar[WorkflowExecutionStatus]

class EventType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    EVENT_TYPE_UNSPECIFIED: _ClassVar[EventType]
    EVENT_TYPE_WORKFLOW_EXECUTION_STARTED: _ClassVar[EventType]
    EVENT_TYPE_WORKFLOW_TASK_SCHEDULED: _ClassVar[EventType]
    EVENT_TYPE_WORKFLOW_TASK_STARTED: _ClassVar[EventType]
    EVENT_TYPE_WORKFLOW_TASK_COMPLETED: _ClassVar[EventType]
    EVENT_TYPE_ACTIVITY_TASK_SCHEDULED: _ClassVar[EventType]
    EVENT_TYPE_ACTIVITY_TASK_STARTED: _ClassVar[EventType]
    EVENT_TYPE_ACTIVITY_TASK_COMPLETED: _ClassVar[EventType]
    EVENT_TYPE_ACTIVITY_TASK_FAILED: _ClassVar[EventType]
    EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED: _ClassVar[EventType]
    EVENT_TYPE_WORKFLOW_EXECUTION_FAILED: _ClassVar[EventType]

class CommandType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    COMMAND_TYPE_UNSPECIFIED: _ClassVar[CommandType]
    COMMAND_TYPE_SCHEDULE_ACTIVITY_TASK: _ClassVar[CommandType]
    COMMAND_TYPE_COMPLETE_WORKFLOW_EXECUTION: _ClassVar[CommandType]
    COMMAND_TYPE_FAIL_WORKFLOW_EXECUTION: _ClassVar[CommandType]
WORKFLOW_EXECUTION_STATUS_UNSPECIFIED: WorkflowExecutionStatus
WORKFLOW_EXECUTION_STATUS_RUNNING: WorkflowExecutionStatus
WORKFLOW_EXECUTION_STATUS_COMPLETED: WorkflowExecutionStatus
WORKFLOW_EXECUTION_STATUS_FAILED: WorkflowExecutionStatus
WORKFLOW_EXECUTION_STATUS_CANCELLED: WorkflowExecutionStatus
WORKFLOW_EXECUTION_STATUS_TIMED_OUT: WorkflowExecutionStatus
WORKFLOW_EXECUTION_STATUS_TERMINATED: WorkflowExecutionStatus
EVENT_TYPE_UNSPECIFIED: EventType
EVENT_TYPE_WORKFLOW_EXECUTION_STARTED: EventType
EVENT_TYPE_WORKFLOW_TASK_SCHEDULED: EventType
EVENT_TYPE_WORKFLOW_TASK_STARTED: EventType
EVENT_TYPE_WORKFLOW_TASK_COMPLETED: EventType
EVENT_TYPE_ACTIVITY_TASK_SCHEDULED: EventType
EVENT_TYPE_ACTIVITY_TASK_STARTED: EventType
EVENT_TYPE_ACTIVITY_TASK_COMPLETED: EventType
EVENT_TYPE_ACTIVITY_TASK_FAILED: EventType
EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED: EventType
EVENT_TYPE_WORKFLOW_EXECUTION_FAILED: EventType
COMMAND_TYPE_UNSPECIFIED: CommandType
COMMAND_TYPE_SCHEDULE_ACTIVITY_TASK: CommandType
COMMAND_TYPE_COMPLETE_WORKFLOW_EXECUTION: CommandType
COMMAND_TYPE_FAIL_WORKFLOW_EXECUTION: CommandType
