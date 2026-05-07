import redis.asyncio as redis

from infrastructure.config.settings import RedisSettings
from infrastructure.observability import get_logger

logger = get_logger(__name__)


class RedisStreamPublisher:
    def __init__(self, settings: RedisSettings) -> None:
        self._settings = settings
        self._client: redis.Redis | None = None

    async def connect(self) -> None:
        self._client = redis.from_url(self._settings.url, decode_responses=False)
        await self._client.ping()
        logger.info("redis.publisher.connected", stream=self._settings.result_stream)

    async def publish(self, routing_key: str, body: bytes) -> None:
        if self._client is None:
            raise RuntimeError("publisher not connected")
        await self._client.xadd(
            self._settings.result_stream,
            {"routing_key": routing_key, "body": body},
            maxlen=self._settings.result_maxlen,
            approximate=True,
        )

    async def close(self) -> None:
        if self._client is not None:
            await self._client.aclose()
            self._client = None
            logger.info("redis.publisher.closed")
