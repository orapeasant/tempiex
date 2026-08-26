from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class Failure(_message.Message):
    __slots__ = ("message", "type", "stack_trace")
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    STACK_TRACE_FIELD_NUMBER: _ClassVar[int]
    message: str
    type: str
    stack_trace: str
    def __init__(self, message: _Optional[str] = ..., type: _Optional[str] = ..., stack_trace: _Optional[str] = ...) -> None: ...
