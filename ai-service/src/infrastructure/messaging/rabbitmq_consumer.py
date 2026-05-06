import asyncio

import aio_pika
from aio_pika.abc import (
    AbstractIncomingMessage,
    AbstractRobustChannel,
    AbstractRobustConnection,
    AbstractRobustQueue,
)

from domain.entities import IncomingMessage
from domain.ports.message_consumer import IngestHandler
from infrastructure.config.settings import RabbitMQSettings
from infrastructure.observability import get_logger

logger = get_logger(__name__)


def _delivery_count(message: AbstractIncomingMessage) -> int:
    headers = message.headers or {}
    x_death = headers.get("x-death")
    if isinstance(x_death, list) and x_death:
        first = x_death[0]
        if isinstance(first, dict):
            count = first.get("count")
            if isinstance(count, int):
                return count + 1
    return 2 if message.redelivered else 1


class RabbitMQConsumer:
    def __init__(self, settings: RabbitMQSettings) -> None:
        self._settings = settings
        self._connection: AbstractRobustConnection | None = None
        self._channel: AbstractRobustChannel | None = None
        self._queue: AbstractRobustQueue | None = None
        self._stopping = asyncio.Event()

    async def connect(self) -> None:
        self._connection = await aio_pika.connect_robust(self._settings.uri)
        self._channel = await self._connection.channel()
        await self._channel.set_qos(prefetch_count=self._settings.prefetch)
        self._queue = await self._channel.get_queue(self._settings.ingest_queue, ensure=True)
        logger.info("rabbitmq.consumer.connected", queue=self._settings.ingest_queue)

    async def consume(self, handler: IngestHandler) -> None:
        if self._queue is None:
            raise RuntimeError("consumer not connected")

        async with self._queue.iterator() as stream:
            async for raw in stream:
                if self._stopping.is_set():
                    return
                await self._dispatch(raw, handler)

    async def _dispatch(self, raw: AbstractIncomingMessage, handler: IngestHandler) -> None:
        attempt = _delivery_count(raw)
        message = IncomingMessage(body=raw.body, delivery_count=attempt)
        try:
            await handler(message)
        except Exception:
            requeue = attempt < self._settings.max_delivery_attempts
            logger.exception(
                "rabbitmq.consumer.handler_failed",
                attempt=attempt,
                requeue=requeue,
            )
            await raw.nack(requeue=requeue)
            return
        await raw.ack()

    async def close(self) -> None:
        self._stopping.set()
        if self._connection is not None:
            await self._connection.close()
            self._connection = None
            self._channel = None
            self._queue = None
            logger.info("rabbitmq.consumer.closed")
