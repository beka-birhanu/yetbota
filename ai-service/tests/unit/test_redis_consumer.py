from unittest.mock import AsyncMock, MagicMock, patch

import pytest
from redis.exceptions import ResponseError

from domain.entities import IncomingMessage
from infrastructure.config.settings import RedisSettings
from infrastructure.messaging import RedisStreamConsumer


def _settings(max_attempts: int = 3) -> RedisSettings:
    return RedisSettings(
        url="redis://localhost:6379/0",
        ingest_stream="ai.ingest",
        consumer_group="ai-service",
        max_delivery_attempts=max_attempts,
    )


def _consumer_with_client(
    client: AsyncMock, settings: RedisSettings | None = None
) -> RedisStreamConsumer:
    consumer = RedisStreamConsumer(settings or _settings())
    consumer._client = client
    return consumer


@pytest.mark.asyncio
async def test_dispatch_acks_and_clears_attempts_on_success() -> None:
    client = AsyncMock()
    client.hincrby = AsyncMock(return_value=1)
    client.xack = AsyncMock()
    client.hdel = AsyncMock()
    handler = AsyncMock()
    consumer = _consumer_with_client(client)

    await consumer._dispatch(b"100-0", {b"body": b'{"x":1}'}, handler)

    handler.assert_awaited_once()
    arg = handler.call_args.args[0]
    assert isinstance(arg, IncomingMessage)
    assert arg.body == b'{"x":1}'
    assert arg.delivery_count == 1
    client.xack.assert_awaited_once_with("ai.ingest", "ai-service", "100-0")
    client.hdel.assert_awaited_once_with("ai.ingest:attempts", "100-0")


@pytest.mark.asyncio
async def test_dispatch_leaves_in_pel_under_max_attempts() -> None:
    client = AsyncMock()
    client.hincrby = AsyncMock(return_value=1)
    client.xack = AsyncMock()
    client.hdel = AsyncMock()
    handler = AsyncMock(side_effect=RuntimeError("boom"))
    consumer = _consumer_with_client(client, _settings(max_attempts=3))

    await consumer._dispatch(b"100-0", {b"body": b"{}"}, handler)

    client.xack.assert_not_awaited()
    client.hdel.assert_not_awaited()


@pytest.mark.asyncio
async def test_dispatch_drops_at_max_attempts() -> None:
    client = AsyncMock()
    client.hincrby = AsyncMock(return_value=3)
    client.xack = AsyncMock()
    client.hdel = AsyncMock()
    handler = AsyncMock(side_effect=RuntimeError("boom"))
    consumer = _consumer_with_client(client, _settings(max_attempts=3))

    await consumer._dispatch(b"100-0", {b"body": b"{}"}, handler)

    client.xack.assert_awaited_once_with("ai.ingest", "ai-service", "100-0")
    client.hdel.assert_awaited_once_with("ai.ingest:attempts", "100-0")


@pytest.mark.asyncio
async def test_dispatch_passes_delivery_count_to_handler() -> None:
    client = AsyncMock()
    client.hincrby = AsyncMock(return_value=4)
    client.xack = AsyncMock()
    client.hdel = AsyncMock()
    handler = AsyncMock()
    consumer = _consumer_with_client(client)

    await consumer._dispatch(b"200-0", {b"body": b"x"}, handler)

    arg = handler.call_args.args[0]
    assert arg.delivery_count == 4


@pytest.mark.asyncio
async def test_consume_before_connect_raises() -> None:
    consumer = RedisStreamConsumer(_settings())
    with pytest.raises(RuntimeError):
        await consumer.consume(AsyncMock())


@pytest.mark.asyncio
async def test_close_when_never_connected_is_noop() -> None:
    consumer = RedisStreamConsumer(_settings())
    await consumer.close()


@pytest.mark.asyncio
async def test_close_releases_client() -> None:
    client = AsyncMock()
    client.aclose = AsyncMock()
    consumer = _consumer_with_client(client)
    await consumer.close()
    client.aclose.assert_awaited_once()


@pytest.mark.asyncio
async def test_connect_creates_consumer_group_idempotently() -> None:
    client = AsyncMock()
    client.xgroup_create = AsyncMock(side_effect=ResponseError("BUSYGROUP exists"))
    with patch(
        "infrastructure.messaging.redis_consumer.redis.from_url",
        return_value=client,
    ):
        consumer = RedisStreamConsumer(_settings())
        await consumer.connect()
    client.xgroup_create.assert_awaited_once()


@pytest.mark.asyncio
async def test_connect_propagates_non_busygroup_errors() -> None:
    client = AsyncMock()
    client.xgroup_create = AsyncMock(side_effect=ResponseError("WRONGTYPE"))
    with patch(
        "infrastructure.messaging.redis_consumer.redis.from_url",
        return_value=client,
    ):
        consumer = RedisStreamConsumer(_settings())
        with pytest.raises(ResponseError):
            await consumer.connect()


@pytest.mark.asyncio
async def test_dispatch_decodes_string_body() -> None:
    client = AsyncMock()
    client.hincrby = AsyncMock(return_value=1)
    client.xack = AsyncMock()
    client.hdel = AsyncMock()
    handler = AsyncMock()
    consumer = _consumer_with_client(client)

    await consumer._dispatch("100-0", {"body": "hello"}, handler)

    arg = handler.call_args.args[0]
    assert arg.body == b"hello"
