from google.protobuf import wrappers_pb2 as _wrappers_pb2
from ai.events.v1 import events_pb2 as _events_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CheckRequest(_message.Message):
    __slots__ = ("text", "kind")
    TEXT_FIELD_NUMBER: _ClassVar[int]
    KIND_FIELD_NUMBER: _ClassVar[int]
    text: str
    kind: _events_pb2.ContentKind
    def __init__(self, text: _Optional[str] = ..., kind: _Optional[_Union[_events_pb2.ContentKind, str]] = ...) -> None: ...

class CheckData(_message.Message):
    __slots__ = ("ok", "reason", "categories")
    class CategoriesEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: float
        def __init__(self, key: _Optional[str] = ..., value: _Optional[float] = ...) -> None: ...
    OK_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    CATEGORIES_FIELD_NUMBER: _ClassVar[int]
    ok: bool
    reason: _wrappers_pb2.StringValue
    categories: _containers.ScalarMap[str, float]
    def __init__(self, ok: bool = ..., reason: _Optional[_Union[_wrappers_pb2.StringValue, _Mapping]] = ..., categories: _Optional[_Mapping[str, float]] = ...) -> None: ...

class CheckResponse(_message.Message):
    __slots__ = ("code", "success", "message", "data")
    CODE_FIELD_NUMBER: _ClassVar[int]
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    code: str
    success: bool
    message: str
    data: CheckData
    def __init__(self, code: _Optional[str] = ..., success: bool = ..., message: _Optional[str] = ..., data: _Optional[_Union[CheckData, _Mapping]] = ...) -> None: ...
