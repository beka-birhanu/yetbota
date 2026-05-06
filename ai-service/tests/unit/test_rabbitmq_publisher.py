from unittest.mock import AsyncMock, patch

import pytest
from aio_pika import DeliveryMode

from infrastructure.config.settings import RabbitMQSettings
from infrastructure.messaging import RabbitMQPublisher


def _settings() -> RabbitMQSettings:
    return RabbitMQSettings(
        uri="amqp://guest:guest@localhost:5672/",
        result_exchange="ai.results",
        result_routing_key="content.processed",
    )


@pytest.fixture
def patched_connect() -> tuple[AsyncMock, AsyncMock, AsyncMock]:
    exchange = AsyncMock()
    channel = AsyncMock()
    channel.get_exchange = AsyncMock(return_value=exchange)
    connection = AsyncMock()
    connection.channel = AsyncMock(return_value=channel)
    connection.close = AsyncMock()
    with patch(
        "infrastructure.messaging.rabbitmq_publisher.aio_pika.connect_robust",
        new=AsyncMock(return_value=connection),
    ):
        yield connection, channel, exchange


@pytest.mark.asyncio
async def test_publish_uses_get_exchange_with_ensure(
    patched_connect: tuple[AsyncMock, AsyncMock, AsyncMock],
) -> None:
    _, channel, _ = patched_connect
    publisher = RabbitMQPublisher(_settings())
    await publisher.connect()
    channel.get_exchange.assert_awaited_once_with("ai.results", ensure=True)


@pytest.mark.asyncio
async def test_publish_sends_persistent_message(
    patched_connect: tuple[AsyncMock, AsyncMock, AsyncMock],
) -> None:
    _, _, exchange = patched_connect
    publisher = RabbitMQPublisher(_settings())
    await publisher.connect()
    await publisher.publish("content.processed", b'{"x":1}')
    exchange.publish.assert_awaited_once()
    args, kwargs = exchange.publish.call_args
    msg = args[0]
    assert msg.body == b'{"x":1}'
    assert msg.delivery_mode == DeliveryMode.PERSISTENT
    assert kwargs["routing_key"] == "content.processed"


@pytest.mark.asyncio
async def test_publish_before_connect_raises() -> None:
    publisher = RabbitMQPublisher(_settings())
    with pytest.raises(RuntimeError):
        await publisher.publish("content.processed", b"x")


@pytest.mark.asyncio
async def test_close_after_connect_releases_connection(
    patched_connect: tuple[AsyncMock, AsyncMock, AsyncMock],
) -> None:
    connection, _, _ = patched_connect
    publisher = RabbitMQPublisher(_settings())
    await publisher.connect()
    await publisher.close()
    connection.close.assert_awaited_once()


@pytest.mark.asyncio
async def test_close_when_never_connected_is_noop() -> None:
    publisher = RabbitMQPublisher(_settings())
    await publisher.close()
