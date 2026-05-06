from collections.abc import Awaitable, Callable
from typing import Protocol, runtime_checkable

from domain.entities import IncomingMessage

IngestHandler = Callable[[IncomingMessage], Awaitable[None]]


@runtime_checkable
class MessageConsumer(Protocol):
    async def consume(self, handler: IngestHandler) -> None: ...

    async def close(self) -> None: ...
