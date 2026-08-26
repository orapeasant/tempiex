from google.protobuf import any_pb2 as _any_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Payload(_message.Message):
    __slots__ = ("encoding", "data")
    ENCODING_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    encoding: str
    data: bytes
    def __init__(self, encoding: _Optional[str] = ..., data: _Optional[bytes] = ...) -> None: ...

class Payloads(_message.Message):
    __slots__ = ("payloads",)
    PAYLOADS_FIELD_NUMBER: _ClassVar[int]
    payloads: _containers.RepeatedCompositeFieldContainer[Payload]
    def __init__(self, payloads: _Optional[_Iterable[_Union[Payload, _Mapping]]] = ...) -> None: ...

class WorkflowExecution(_message.Message):
    __slots__ = ("workflow_id", "run_id")
    WORKFLOW_ID_FIELD_NUMBER: _ClassVar[int]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    workflow_id: str
    run_id: str
    def __init__(self, workflow_id: _Optional[str] = ..., run_id: _Optional[str] = ...) -> None: ...

class WorkflowType(_message.Message):
    __slots__ = ("name",)
    NAME_FIELD_NUMBER: _ClassVar[int]
    name: str
    def __init__(self, name: _Optional[str] = ...) -> None: ...

class ActivityType(_message.Message):
    __slots__ = ("name",)
    NAME_FIELD_NUMBER: _ClassVar[int]
    name: str
    def __init__(self, name: _Optional[str] = ...) -> None: ...

class Header(_message.Message):
    __slots__ = ("fields",)
    class FieldsEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: bytes
        def __init__(self, key: _Optional[str] = ..., value: _Optional[bytes] = ...) -> None: ...
    FIELDS_FIELD_NUMBER: _ClassVar[int]
    fields: _containers.ScalarMap[str, bytes]
    def __init__(self, fields: _Optional[_Mapping[str, bytes]] = ...) -> None: ...
