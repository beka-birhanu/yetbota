import datetime

from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf import wrappers_pb2 as _wrappers_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ContentKind(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    CONTENT_KIND_UNSPECIFIED: _ClassVar[ContentKind]
    CONTENT_KIND_POST: _ClassVar[ContentKind]
    CONTENT_KIND_QUESTION: _ClassVar[ContentKind]
    CONTENT_KIND_ANSWER: _ClassVar[ContentKind]
CONTENT_KIND_UNSPECIFIED: ContentKind
CONTENT_KIND_POST: ContentKind
CONTENT_KIND_QUESTION: ContentKind
CONTENT_KIND_ANSWER: ContentKind

class IngestEvent(_message.Message):
    __slots__ = ("content_id", "kind", "user_id", "text", "parent_id", "attached_post_id", "tags", "category")
    CONTENT_ID_FIELD_NUMBER: _ClassVar[int]
    KIND_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    TEXT_FIELD_NUMBER: _ClassVar[int]
    PARENT_ID_FIELD_NUMBER: _ClassVar[int]
    ATTACHED_POST_ID_FIELD_NUMBER: _ClassVar[int]
    TAGS_FIELD_NUMBER: _ClassVar[int]
    CATEGORY_FIELD_NUMBER: _ClassVar[int]
    content_id: str
    kind: ContentKind
    user_id: str
    text: str
    parent_id: _wrappers_pb2.StringValue
    attached_post_id: _wrappers_pb2.StringValue
    tags: _containers.RepeatedScalarFieldContainer[str]
    category: _wrappers_pb2.StringValue
    def __init__(self, content_id: _Optional[str] = ..., kind: _Optional[_Union[ContentKind, str]] = ..., user_id: _Optional[str] = ..., text: _Optional[str] = ..., parent_id: _Optional[_Union[_wrappers_pb2.StringValue, _Mapping]] = ..., attached_post_id: _Optional[_Union[_wrappers_pb2.StringValue, _Mapping]] = ..., tags: _Optional[_Iterable[str]] = ..., category: _Optional[_Union[_wrappers_pb2.StringValue, _Mapping]] = ...) -> None: ...

class ContentProcessed(_message.Message):
    __slots__ = ("content_id", "kind", "status", "duplicate_of", "error_code", "processed_at")
    class Status(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        STATUS_UNSPECIFIED: _ClassVar[ContentProcessed.Status]
        UNIQUE: _ClassVar[ContentProcessed.Status]
        DUPLICATE: _ClassVar[ContentProcessed.Status]
        INDEXED: _ClassVar[ContentProcessed.Status]
        ERROR: _ClassVar[ContentProcessed.Status]
    STATUS_UNSPECIFIED: ContentProcessed.Status
    UNIQUE: ContentProcessed.Status
    DUPLICATE: ContentProcessed.Status
    INDEXED: ContentProcessed.Status
    ERROR: ContentProcessed.Status
    CONTENT_ID_FIELD_NUMBER: _ClassVar[int]
    KIND_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    DUPLICATE_OF_FIELD_NUMBER: _ClassVar[int]
    ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    PROCESSED_AT_FIELD_NUMBER: _ClassVar[int]
    content_id: str
    kind: ContentKind
    status: ContentProcessed.Status
    duplicate_of: _wrappers_pb2.StringValue
    error_code: _wrappers_pb2.StringValue
    processed_at: _timestamp_pb2.Timestamp
    def __init__(self, content_id: _Optional[str] = ..., kind: _Optional[_Union[ContentKind, str]] = ..., status: _Optional[_Union[ContentProcessed.Status, str]] = ..., duplicate_of: _Optional[_Union[_wrappers_pb2.StringValue, _Mapping]] = ..., error_code: _Optional[_Union[_wrappers_pb2.StringValue, _Mapping]] = ..., processed_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...
