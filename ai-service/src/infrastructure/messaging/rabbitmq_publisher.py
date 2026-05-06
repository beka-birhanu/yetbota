import aio_pika
from aio_pika import DeliveryMode, Message
from aio_pika.abc import AbstractRobustChannel, AbstractRobustConnection, AbstractRobustExchange

from infrastructure.config.settings import RabbitMQSettings
from infrastructure.observability import get_logger

logger = get_logger(__name__)


class RabbitMQPublisher:
    def __init__(self, settings: RabbitMQSettings) -> None:
        self._settings = settings
        self._connection: AbstractRobustConnection | None = None
        self._channel: AbstractRobustChannel | None = None
        self._exchange: AbstractRobustExchange | None = None

    async def connect(self) -> None:
        self._connection = await aio_pika.connect_robust(self._settings.uri)
        self._channel = await self._connection.channel()
        self._exchange = await self._channel.get_exchange(
            self._settings.result_exchange, ensure=True
        )
        logger.info("rabbitmq.publisher.connected", exchange=self._settings.result_exchange)

    async def publish(self, routing_key: str, body: bytes) -> None:
        if self._exchange is None:
            raise RuntimeError("publisher not connected")
        message = Message(body=body, delivery_mode=DeliveryMode.PERSISTENT)
        await self._exchange.publish(message, routing_key=routing_key)

    async def close(self) -> None:
        if self._connection is not None:
            await self._connection.close()
            self._connection = None
            self._channel = None
            self._exchange = None
            logger.info("rabbitmq.publisher.closed")
