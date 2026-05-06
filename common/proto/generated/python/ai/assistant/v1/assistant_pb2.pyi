from google.protobuf import wrappers_pb2 as _wrappers_pb2
from ai.events.v1 import events_pb2 as _events_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Citation(_message.Message):
    __slots__ = ("source_id", "kind", "text", "score")
    SOURCE_ID_FIELD_NUMBER: _ClassVar[int]
    KIND_FIELD_NUMBER: _ClassVar[int]
    TEXT_FIELD_NUMBER: _ClassVar[int]
    SCORE_FIELD_NUMBER: _ClassVar[int]
    source_id: str
    kind: _events_pb2.ContentKind
    text: str
    score: float
    def __init__(self, source_id: _Optional[str] = ..., kind: _Optional[_Union[_events_pb2.ContentKind, str]] = ..., text: _Optional[str] = ..., score: _Optional[float] = ...) -> None: ...

class ChatRequest(_message.Message):
    __slots__ = ("text", "user_id")
    TEXT_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    text: str
    user_id: _wrappers_pb2.StringValue
    def __init__(self, text: _Optional[str] = ..., user_id: _Optional[_Union[_wrappers_pb2.StringValue, _Mapping]] = ...) -> None: ...

class ChatData(_message.Message):
    __slots__ = ("answer", "citations")
    ANSWER_FIELD_NUMBER: _ClassVar[int]
    CITATIONS_FIELD_NUMBER: _ClassVar[int]
    answer: str
    citations: _containers.RepeatedCompositeFieldContainer[Citation]
    def __init__(self, answer: _Optional[str] = ..., citations: _Optional[_Iterable[_Union[Citation, _Mapping]]] = ...) -> None: ...

class ChatResponse(_message.Message):
    __slots__ = ("code", "success", "message", "data")
    CODE_FIELD_NUMBER: _ClassVar[int]
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    code: str
    success: bool
    message: str
    data: ChatData
    def __init__(self, code: _Optional[str] = ..., success: bool = ..., message: _Optional[str] = ..., data: _Optional[_Union[ChatData, _Mapping]] = ...) -> None: ...
