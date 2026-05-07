import asyncio
import uuid

import redis.asyncio as redis
from redis.exceptions import ResponseError

from domain.entities import IncomingMessage
from domain.ports.message_consumer import IngestHandler
from infrastructure.config.settings import RedisSettings
from infrastructure.observability import get_logger

logger = get_logger(__name__)


def _decode(value: bytes | str) -> str:
    return value.decode() if isinstance(value, bytes) else value


def _body(fields: dict) -> bytes:
    raw = fields.get(b"body") if b"body" in fields else fields.get("body")
    if raw is None:
        return b""
    if isinstance(raw, str):
        return raw.encode("utf-8")
    return raw


class RedisStreamConsumer:
    def __init__(self, settings: RedisSettings) -> None:
        self._settings = settings
        self._client: redis.Redis | None = None
        self._consumer_name = f"ai-service-{uuid.uuid4().hex[:8]}"
        self._stopping = asyncio.Event()
        self._attempts_key = f"{settings.ingest_stream}:attempts"

    async def connect(self) -> None:
        self._client = redis.from_url(self._settings.url, decode_responses=False)
        try:
            await self._client.xgroup_create(
                name=self._settings.ingest_stream,
                groupname=self._settings.consumer_group,
                id="$",
                mkstream=True,
            )
        except ResponseError as exc:
            if "BUSYGROUP" not in str(exc):
                raise
        logger.info(
            "redis.consumer.connected",
            stream=self._settings.ingest_stream,
            group=self._settings.consumer_group,
            consumer=self._consumer_name,
        )

    async def consume(self, handler: IngestHandler) -> None:
        if self._client is None:
            raise RuntimeError("consumer not connected")
        await asyncio.gather(
            self._consume_new(handler),
            self._reclaim_loop(handler),
            return_exceptions=False,
        )

    async def _consume_new(self, handler: IngestHandler) -> None:
        assert self._client is not None
        while not self._stopping.is_set():
            response = await self._client.xreadgroup(
                groupname=self._settings.consumer_group,
                consumername=self._consumer_name,
                streams={self._settings.ingest_stream: ">"},
                count=self._settings.prefetch,
                block=self._settings.block_ms,
            )
            for _, entries in response or []:
                for msg_id, fields in entries:
                    if self._stopping.is_set():
                        return
                    await self._dispatch(msg_id, fields, handler)

    async def _reclaim_loop(self, handler: IngestHandler) -> None:
        await self._reclaim_once(handler)
        while not self._stopping.is_set():
            try:
                await asyncio.wait_for(
                    self._stopping.wait(),
                    timeout=self._settings.claim_idle_ms / 1000.0,
                )
                return
            except asyncio.TimeoutError:
                pass
            await self._reclaim_once(handler)

    async def _reclaim_once(self, handler: IngestHandler) -> None:
        assert self._client is not None
        cursor: bytes | str = "0-0"
        while not self._stopping.is_set():
            cursor, entries, _ = await self._client.xautoclaim(
                name=self._settings.ingest_stream,
                groupname=self._settings.consumer_group,
                consumername=self._consumer_name,
                min_idle_time=self._settings.claim_idle_ms,
                start_id=cursor,
                count=self._settings.prefetch,
            )
            if not entries:
                return
            for msg_id, fields in entries:
                if self._stopping.is_set():
                    return
                await self._dispatch(msg_id, fields, handler)

    async def _dispatch(
        self, msg_id: bytes | str, fields: dict, handler: IngestHandler
    ) -> None:
        assert self._client is not None
        msg_id_s = _decode(msg_id)
        attempt = int(await self._client.hincrby(self._attempts_key, msg_id_s, 1))
        message = IncomingMessage(body=_body(fields), delivery_count=attempt)
        try:
            await handler(message)
        except Exception:
            requeue = attempt < self._settings.max_delivery_attempts
            logger.exception(
                "redis.consumer.handler_failed",
                msg_id=msg_id_s,
                attempt=attempt,
                requeue=requeue,
            )
            if not requeue:
                await self._client.xack(
                    self._settings.ingest_stream,
                    self._settings.consumer_group,
                    msg_id_s,
                )
                await self._client.hdel(self._attempts_key, msg_id_s)
            return
        await self._client.xack(
            self._settings.ingest_stream,
            self._settings.consumer_group,
            msg_id_s,
        )
        await self._client.hdel(self._attempts_key, msg_id_s)

    async def close(self) -> None:
        self._stopping.set()
        if self._client is not None:
            await self._client.aclose()
            self._client = None
            logger.info("redis.consumer.closed")
