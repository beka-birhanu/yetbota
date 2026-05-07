import json
from datetime import UTC, datetime

from pydantic import ValidationError

from application.errors import AIServiceError, MessageMalformed
from application.ingest_content import IngestContent
from domain.entities import IncomingMessage, IngestRequest, IngestResult
from domain.ports.message_consumer import IngestHandler
from infrastructure.config.settings import RedisSettings
from infrastructure.messaging import RedisStreamConsumer, RedisStreamPublisher
from infrastructure.observability import (
    INGEST_DURATION,
    INGEST_TOTAL,
    get_logger,
    time_histogram,
)

logger = get_logger(__name__)

_STATUS_BY_VERDICT = {
    "unique": "unique",
    "duplicate": "duplicate",
    "indexed": "indexed",
    "error": "error",
}


def _serialize_result(result: IngestResult) -> bytes:
    payload = {
        "content_id": result.content_id,
        "kind": result.kind,
        "status": _STATUS_BY_VERDICT.get(result.verdict, "error"),
        "duplicate_of": result.duplicate_of,
        "error_code": result.error_code,
        "processed_at": result.processed_at.isoformat(),
    }
    return json.dumps(payload).encode("utf-8")


def _error_result(*, content_id: str, kind: str, error_code: str) -> IngestResult:
    return IngestResult(
        content_id=content_id,
        kind=kind,  # type: ignore[arg-type]
        verdict="error",
        duplicate_of=None,
        error_code=error_code,
        processed_at=datetime.now(UTC),
    )


def _parse_request(body: bytes) -> IngestRequest:
    try:
        return IngestRequest.model_validate_json(body)
    except ValidationError as exc:
        raise MessageMalformed(f"invalid ingest payload: {exc}", cause=exc) from exc


class IngestWorker:
    def __init__(
        self,
        *,
        consumer: RedisStreamConsumer,
        publisher: RedisStreamPublisher,
        use_case: IngestContent,
        redis: RedisSettings,
    ) -> None:
        self._consumer = consumer
        self._publisher = publisher
        self._use_case = use_case
        self._redis = redis

    async def run(self) -> None:
        handler: IngestHandler = self._handle
        await self._consumer.consume(handler)

    async def _handle(self, message: IncomingMessage) -> None:
        try:
            request = _parse_request(message.body)
        except MessageMalformed as exc:
            logger.warning(
                "ingest.message_malformed",
                error=exc.message,
                delivery_count=message.delivery_count,
            )
            await self._publish(_error_result(content_id="", kind="post", error_code=exc.code))
            INGEST_TOTAL.labels(kind="unknown", status="malformed").inc()
            return

        try:
            async with time_histogram(INGEST_DURATION, kind=request.kind):
                result = await self._use_case.execute(request)
        except AIServiceError as exc:
            await self._maybe_recover(message, request, exc)
            return

        await self._publish(result)
        INGEST_TOTAL.labels(kind=request.kind, status=result.verdict).inc()

    async def _maybe_recover(
        self,
        message: IncomingMessage,
        request: IngestRequest,
        exc: AIServiceError,
    ) -> None:
        is_final = message.delivery_count >= self._redis.max_delivery_attempts
        if exc.transient and not is_final:
            logger.info(
                "ingest.retrying",
                content_id=request.content_id,
                kind=request.kind,
                error_code=exc.code,
                attempt=message.delivery_count,
            )
            raise exc

        logger.warning(
            "ingest.failed",
            content_id=request.content_id,
            kind=request.kind,
            error_code=exc.code,
            attempt=message.delivery_count,
            transient=exc.transient,
        )
        error_code = "MANUAL_REVIEW" if exc.code == "SIMILARITY_FAILED" and is_final else exc.code
        await self._publish(
            _error_result(
                content_id=request.content_id,
                kind=request.kind,
                error_code=error_code,
            )
        )
        INGEST_TOTAL.labels(kind=request.kind, status="error").inc()

    async def _publish(self, result: IngestResult) -> None:
        body = _serialize_result(result)
        await self._publisher.publish(self._redis.result_routing_key, body)
        logger.info(
            "ingest.result_published",
            content_id=result.content_id,
            kind=result.kind,
            verdict=result.verdict,
            error_code=result.error_code,
        )
