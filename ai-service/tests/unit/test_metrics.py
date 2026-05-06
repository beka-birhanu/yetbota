from unittest.mock import AsyncMock

import pytest
from fastapi.testclient import TestClient
from prometheus_client.parser import text_string_to_metric_families

from domain.entities import ChatResponse, Citation, ScreeningResult
from infrastructure.config import Settings
from infrastructure.observability import (
    EMBED_CALLS,
    LLM_CALLS,
    SCREENING_BLOCKS,
    WEAVIATE_CALLS,
    observe,
    render_latest,
)
from interfaces.http import create_app


def _client(
    *, assistant: AsyncMock | None = None, screening: AsyncMock | None = None
) -> TestClient:
    app = create_app(Settings(_env_file=None))  # type: ignore[call-arg]
    app.state.assistant = assistant
    app.state.screening = screening
    return TestClient(app)


def test_metrics_endpoint_returns_prometheus_text() -> None:
    client = _client()
    resp = client.get("/metrics")
    assert resp.status_code == 200
    assert resp.headers["content-type"].startswith("text/plain")
    families = {family.name for family in text_string_to_metric_families(resp.text)}
    expected = {
        "ai_service_ingest",
        "ai_service_ingest_duration_seconds",
        "ai_service_embed_calls",
        "ai_service_llm_calls",
        "ai_service_weaviate_calls",
        "ai_service_rag_chat_duration_seconds",
        "ai_service_screening_duration_seconds",
        "ai_service_screening_blocks",
    }
    assert expected <= families


def test_render_latest_returns_bytes_and_content_type() -> None:
    body, content_type = render_latest()
    assert isinstance(body, bytes)
    assert "text/plain" in content_type


def test_assistant_route_records_rag_chat_duration() -> None:
    use_case = AsyncMock()
    use_case.execute = AsyncMock(
        return_value=ChatResponse(
            answer="hi",
            citations=[Citation(source_id="p1", kind="post", text="t", score=0.9)],
        )
    )
    client = _client(assistant=use_case)
    before = _histogram_count("ai_service_rag_chat_duration_seconds")
    resp = client.post("/v1/assistant/chat", json={"text": "hello"})
    assert resp.status_code == 200
    after = _histogram_count("ai_service_rag_chat_duration_seconds")
    assert after == before + 1


def test_screening_route_records_block_when_not_ok() -> None:
    use_case = AsyncMock()
    use_case.execute = AsyncMock(
        return_value=ScreeningResult(
            ok=False,
            reason="profanity",
            categories={"profanity": 0.9},
        )
    )
    client = _client(screening=use_case)
    before = _counter_value(SCREENING_BLOCKS, reason="profanity")
    resp = client.post("/v1/screening/check", json={"text": "bad", "kind": "answer"})
    assert resp.status_code == 200
    after = _counter_value(SCREENING_BLOCKS, reason="profanity")
    assert after == before + 1


def test_screening_route_does_not_record_block_when_ok() -> None:
    use_case = AsyncMock()
    use_case.execute = AsyncMock(
        return_value=ScreeningResult(ok=True, reason=None, categories={"profanity": 0.05})
    )
    client = _client(screening=use_case)
    before = _counter_value(SCREENING_BLOCKS, reason="profanity")
    resp = client.post("/v1/screening/check", json={"text": "hi", "kind": "post"})
    assert resp.status_code == 200
    after = _counter_value(SCREENING_BLOCKS, reason="profanity")
    assert after == before


def _counter_value(counter: object, **labels: str) -> float:
    metric = counter.labels(**labels)  # type: ignore[attr-defined]
    return float(metric._value.get())  # type: ignore[attr-defined]


def _histogram_count(name: str) -> float:
    body, _ = render_latest()
    text = body.decode("utf-8")
    for family in text_string_to_metric_families(text):
        if family.name != name:
            continue
        for sample in family.samples:
            if sample.name.endswith("_count"):
                return float(sample.value)
    return 0.0


@pytest.mark.asyncio
async def test_observe_increments_outcome_label() -> None:
    before_success = _counter_value(EMBED_CALLS, task_type="RETRIEVAL_DOCUMENT", outcome="success")
    async with observe(EMBED_CALLS, task_type="RETRIEVAL_DOCUMENT"):
        pass
    after_success = _counter_value(EMBED_CALLS, task_type="RETRIEVAL_DOCUMENT", outcome="success")
    assert after_success == before_success + 1


@pytest.mark.asyncio
async def test_observe_records_error_outcome_on_exception() -> None:
    before_error = _counter_value(LLM_CALLS, op="generate", outcome="error")
    with pytest.raises(RuntimeError):
        async with observe(LLM_CALLS, op="generate"):
            raise RuntimeError("boom")
    after_error = _counter_value(LLM_CALLS, op="generate", outcome="error")
    assert after_error == before_error + 1


@pytest.mark.asyncio
async def test_weaviate_observe_uses_op_label() -> None:
    before = _counter_value(WEAVIATE_CALLS, op="upsert", outcome="success")
    async with observe(WEAVIATE_CALLS, op="upsert"):
        pass
    after = _counter_value(WEAVIATE_CALLS, op="upsert", outcome="success")
    assert after == before + 1
