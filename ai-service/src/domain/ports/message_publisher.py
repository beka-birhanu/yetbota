from typing import Protocol, runtime_checkable


@runtime_checkable
class MessagePublisher(Protocol):
    async def publish(self, routing_key: str, body: bytes) -> None: ...

    async def close(self) -> None: ...
