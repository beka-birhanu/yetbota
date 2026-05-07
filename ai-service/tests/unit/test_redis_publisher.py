from unittest.mock import AsyncMock, patch

import pytest

from infrastructure.config.settings import RedisSettings
from infrastructure.messaging import RedisStreamPublisher


def _settings() -> RedisSettings:
    return RedisSettings(
        url="redis://localhost:6379/0",
        result_stream="ai.results",
        result_routing_key="content.processed",
        result_maxlen=1000,
    )


@pytest.fixture
def patched_client():
    client = AsyncMock()
    client.ping = AsyncMock(return_value=True)
    client.xadd = AsyncMock(return_value=b"100-0")
    client.aclose = AsyncMock()
    with patch(
        "infrastructure.messaging.redis_publisher.redis.from_url",
        return_value=client,
    ):
        yield client


@pytest.mark.asyncio
async def test_connect_pings(patched_client: AsyncMock) -> None:
    publisher = RedisStreamPublisher(_settings())
    await publisher.connect()
    patched_client.ping.assert_awaited_once()


@pytest.mark.asyncio
async def test_publish_xadds_to_result_stream(patched_client: AsyncMock) -> None:
    publisher = RedisStreamPublisher(_settings())
    await publisher.connect()
    await publisher.publish("content.processed", b'{"x":1}')
    patched_client.xadd.assert_awaited_once()
    args, kwargs = patched_client.xadd.call_args
    assert args[0] == "ai.results"
    assert args[1] == {"routing_key": "content.processed", "body": b'{"x":1}'}
    assert kwargs["maxlen"] == 1000
    assert kwargs["approximate"] is True


@pytest.mark.asyncio
async def test_publish_before_connect_raises() -> None:
    publisher = RedisStreamPublisher(_settings())
    with pytest.raises(RuntimeError):
        await publisher.publish("content.processed", b"x")


@pytest.mark.asyncio
async def test_close_after_connect_releases_client(
    patched_client: AsyncMock,
) -> None:
    publisher = RedisStreamPublisher(_settings())
    await publisher.connect()
    await publisher.close()
    patched_client.aclose.assert_awaited_once()


@pytest.mark.asyncio
async def test_close_when_never_connected_is_noop() -> None:
    publisher = RedisStreamPublisher(_settings())
    await publisher.close()
