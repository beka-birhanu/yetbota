import json
from datetime import UTC, datetime
from unittest.mock import AsyncMock

import pytest

from application.errors import (
    EmbeddingFailed,
    IndexingFailed,
    MetadataError,
    SimilaritySearchFailed,
)
from domain.entities import IncomingMessage, IngestRequest, IngestResult
from infrastructure.config.settings import RedisSettings
from interfaces.workers.ingest_worker import IngestWorker, _serialize_result


def _settings(max_attempts: int = 3) -> RedisSettings:
    return RedisSettings(
        result_routing_key="content.processed",
        max_delivery_attempts=max_attempts,
    )


def _post_request_payload() -> bytes:
    return json.dumps(
        {
            "content_id": "p1",
            "kind": "post",
            "user_id": "u1",
            "text": "hello world",
            "category": "cafe",
            "tags": ["coffee"],
        }
    ).encode("utf-8")


def _build(
    *,
    use_case_result: IngestResult | None = None,
    use_case_error: Exception | None = None,
    settings: RedisSettings | None = None,
) -> tuple[IngestWorker, AsyncMock, AsyncMock]:
    consumer = AsyncMock()
    publisher = AsyncMock()
    use_case = AsyncMock()
    if use_case_error is not None:
        use_case.execute = AsyncMock(side_effect=use_case_error)
    else:
        use_case.execute = AsyncMock(return_value=use_case_result)
    worker = IngestWorker(
        consumer=consumer,
        publisher=publisher,
        use_case=use_case,
        redis=settings or _settings(),
    )
    return worker, publisher, use_case


def _published_payload(publisher: AsyncMock) -> dict[str, object]:
    publisher.publish.assert_awaited_once()
    args = publisher.publish.call_args.args
    return json.loads(args[1])


def test_serialize_result_uses_status_field() -> None:
    result = IngestResult(
        content_id="p1",
        kind="post",
        verdict="duplicate",
        duplicate_of="p2",
        error_code=None,
        processed_at=datetime(2026, 5, 6, 12, 0, tzinfo=UTC),
    )
    payload = json.loads(_serialize_result(result))
    assert payload == {
        "content_id": "p1",
        "kind": "post",
        "status": "duplicate",
        "duplicate_of": "p2",
        "error_code": None,
        "processed_at": "2026-05-06T12:00:00+00:00",
    }


@pytest.mark.asyncio
async def test_unique_post_publishes_unique_status() -> None:
    result = IngestResult(
        content_id="p1",
        kind="post",
        verdict="unique",
        processed_at=datetime.now(UTC),
    )
    worker, publisher, _ = _build(use_case_result=result)
    await worker._handle(IncomingMessage(body=_post_request_payload(), delivery_count=1))
    payload = _published_payload(publisher)
    assert payload["status"] == "unique"
    assert payload["error_code"] is None


@pytest.mark.asyncio
async def test_duplicate_publishes_duplicate_of() -> None:
    result = IngestResult(
        content_id="p1",
        kind="post",
        verdict="duplicate",
        duplicate_of="p2",
        processed_at=datetime.now(UTC),
    )
    worker, publisher, _ = _build(use_case_result=result)
    await worker._handle(IncomingMessage(body=_post_request_payload(), delivery_count=1))
    payload = _published_payload(publisher)
    assert payload["status"] == "duplicate"
    assert payload["duplicate_of"] == "p2"


@pytest.mark.asyncio
async def test_question_indexed_publishes_indexed() -> None:
    result = IngestResult(
        content_id="q1",
        kind="question",
        verdict="indexed",
        processed_at=datetime.now(UTC),
    )
    worker, publisher, _ = _build(use_case_result=result)
    body = json.dumps(
        {
            "content_id": "q1",
            "kind": "question",
            "user_id": "u1",
            "text": "where can I get coffee?",
        }
    ).encode("utf-8")
    await worker._handle(IncomingMessage(body=body, delivery_count=1))
    payload = _published_payload(publisher)
    assert payload["status"] == "indexed"
    assert payload["kind"] == "question"


@pytest.mark.asyncio
async def test_malformed_payload_publishes_message_malformed_and_skips_use_case() -> None:
    worker, publisher, use_case = _build()
    await worker._handle(IncomingMessage(body=b"not json", delivery_count=1))
    payload = _published_payload(publisher)
    assert payload["status"] == "error"
    assert payload["error_code"] == "MESSAGE_MALFORMED"
    use_case.execute.assert_not_awaited()


@pytest.mark.asyncio
async def test_metadata_error_publishes_terminal_error_no_reraise() -> None:
    worker, publisher, _ = _build(use_case_error=MetadataError("bad"))
    await worker._handle(IncomingMessage(body=_post_request_payload(), delivery_count=1))
    payload = _published_payload(publisher)
    assert payload["error_code"] == "METADATA_ERROR"


@pytest.mark.asyncio
async def test_indexing_failed_publishes_terminal_error_no_reraise() -> None:
    worker, publisher, _ = _build(use_case_error=IndexingFailed("boom"))
    await worker._handle(IncomingMessage(body=_post_request_payload(), delivery_count=1))
    payload = _published_payload(publisher)
    assert payload["error_code"] == "INDEXING_GAP"


@pytest.mark.asyncio
async def test_transient_under_max_reraises() -> None:
    worker, publisher, _ = _build(
        use_case_error=EmbeddingFailed("rate limited"),
        settings=_settings(max_attempts=3),
    )
    with pytest.raises(EmbeddingFailed):
        await worker._handle(IncomingMessage(body=_post_request_payload(), delivery_count=1))
    publisher.publish.assert_not_awaited()


@pytest.mark.asyncio
async def test_transient_at_max_publishes_error() -> None:
    worker, publisher, _ = _build(
        use_case_error=EmbeddingFailed("rate limited"),
        settings=_settings(max_attempts=3),
    )
    await worker._handle(IncomingMessage(body=_post_request_payload(), delivery_count=3))
    payload = _published_payload(publisher)
    assert payload["error_code"] == "EMBED_FAILED"


@pytest.mark.asyncio
async def test_similarity_failed_at_max_publishes_manual_review() -> None:
    worker, publisher, _ = _build(
        use_case_error=SimilaritySearchFailed("nope"),
        settings=_settings(max_attempts=3),
    )
    await worker._handle(IncomingMessage(body=_post_request_payload(), delivery_count=3))
    payload = _published_payload(publisher)
    assert payload["error_code"] == "MANUAL_REVIEW"


@pytest.mark.asyncio
async def test_run_delegates_to_consumer() -> None:
    worker, _, _ = _build()
    await worker.run()
    consumer = worker._consumer  # type: ignore[attr-defined]
    consumer.consume.assert_awaited_once()
    handler = consumer.consume.call_args.args[0]
    assert callable(handler)


@pytest.mark.asyncio
async def test_use_case_receives_parsed_request() -> None:
    result = IngestResult(
        content_id="p1",
        kind="post",
        verdict="unique",
        processed_at=datetime.now(UTC),
    )
    worker, _, use_case = _build(use_case_result=result)
    await worker._handle(IncomingMessage(body=_post_request_payload(), delivery_count=1))
    use_case.execute.assert_awaited_once()
    arg = use_case.execute.call_args.args[0]
    assert isinstance(arg, IngestRequest)
    assert arg.content_id == "p1"
    assert arg.tags == ["coffee"]
