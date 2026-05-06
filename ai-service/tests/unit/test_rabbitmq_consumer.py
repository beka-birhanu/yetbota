from collections.abc import AsyncIterator
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from domain.entities import IncomingMessage
from infrastructure.config.settings import RabbitMQSettings
from infrastructure.messaging import RabbitMQConsumer
from infrastructure.messaging.rabbitmq_consumer import _delivery_count


def _settings(max_attempts: int = 3) -> RabbitMQSettings:
    return RabbitMQSettings(
        uri="amqp://guest:guest@localhost:5672/",
        ingest_queue="ai.ingest",
        max_delivery_attempts=max_attempts,
    )


def _raw_message(
    *, body: bytes = b"{}", redelivered: bool = False, x_death_count: int | None = None
) -> AsyncMock:
    msg = AsyncMock()
    msg.body = body
    msg.redelivered = redelivered
    msg.headers = {"x-death": [{"count": x_death_count}]} if x_death_count is not None else {}
    return msg


class _FakeIterator:
    def __init__(self, messages: list[AsyncMock]) -> None:
        self._messages = messages

    async def __aenter__(self) -> "_FakeIterator":
        return self

    async def __aexit__(self, *_: object) -> None:
        return None

    def __aiter__(self) -> AsyncIterator[AsyncMock]:
        return self._async_iter()

    async def _async_iter(self) -> AsyncIterator[AsyncMock]:
        for m in self._messages:
            yield m


def _consumer_with_messages(
    messages: list[AsyncMock], settings: RabbitMQSettings | None = None
) -> RabbitMQConsumer:
    settings = settings or _settings()
    consumer = RabbitMQConsumer(settings)
    queue = MagicMock()
    queue.iterator = MagicMock(return_value=_FakeIterator(messages))
    consumer._queue = queue
    return consumer


def test_delivery_count_first_attempt() -> None:
    msg = _raw_message()
    assert _delivery_count(msg) == 1


def test_delivery_count_redelivered_no_x_death() -> None:
    msg = _raw_message(redelivered=True)
    assert _delivery_count(msg) == 2


def test_delivery_count_uses_x_death() -> None:
    msg = _raw_message(redelivered=True, x_death_count=3)
    assert _delivery_count(msg) == 4


@pytest.mark.asyncio
async def test_handler_success_acks() -> None:
    raw = _raw_message(body=b'{"x":1}')
    handler = AsyncMock()
    consumer = _consumer_with_messages([raw])
    await consumer.consume(handler)
    handler.assert_awaited_once()
    arg = handler.call_args.args[0]
    assert isinstance(arg, IncomingMessage)
    assert arg.body == b'{"x":1}'
    assert arg.delivery_count == 1
    raw.ack.assert_awaited_once()
    raw.nack.assert_not_awaited()


@pytest.mark.asyncio
async def test_handler_exception_nacks_with_requeue_when_under_limit() -> None:
    raw = _raw_message()
    handler = AsyncMock(side_effect=RuntimeError("boom"))
    consumer = _consumer_with_messages([raw], _settings(max_attempts=3))
    await consumer.consume(handler)
    raw.nack.assert_awaited_once_with(requeue=True)
    raw.ack.assert_not_awaited()


@pytest.mark.asyncio
async def test_handler_exception_nacks_without_requeue_at_limit() -> None:
    raw = _raw_message(redelivered=True, x_death_count=2)
    handler = AsyncMock(side_effect=RuntimeError("boom"))
    consumer = _consumer_with_messages([raw], _settings(max_attempts=3))
    await consumer.consume(handler)
    raw.nack.assert_awaited_once_with(requeue=False)


@pytest.mark.asyncio
async def test_consume_before_connect_raises() -> None:
    consumer = RabbitMQConsumer(_settings())
    with pytest.raises(RuntimeError):
        await consumer.consume(AsyncMock())


@pytest.mark.asyncio
async def test_close_when_never_connected_is_noop() -> None:
    consumer = RabbitMQConsumer(_settings())
    await consumer.close()


@pytest.mark.asyncio
async def test_close_releases_connection() -> None:
    consumer = RabbitMQConsumer(_settings())
    connection = AsyncMock()
    connection.close = AsyncMock()
    consumer._connection = connection
    await consumer.close()
    connection.close.assert_awaited_once()


@pytest.mark.asyncio
async def test_connect_attaches_queue_passively() -> None:
    queue = AsyncMock()
    channel = AsyncMock()
    channel.set_qos = AsyncMock()
    channel.get_queue = AsyncMock(return_value=queue)
    connection = AsyncMock()
    connection.channel = AsyncMock(return_value=channel)
    with patch(
        "infrastructure.messaging.rabbitmq_consumer.aio_pika.connect_robust",
        new=AsyncMock(return_value=connection),
    ):
        consumer = RabbitMQConsumer(_settings())
        await consumer.connect()
    channel.set_qos.assert_awaited_once_with(prefetch_count=16)
    channel.get_queue.assert_awaited_once_with("ai.ingest", ensure=True)
