import os

from infrastructure.config import Settings


def test_defaults_load_without_env_file() -> None:
    s = Settings(_env_file=None)  # type: ignore[call-arg]
    assert s.app.name == "ai-service"
    assert s.gemini.llm_model == "gemini-2.5-flash"
    assert s.gemini.embedding_dimensions == 1536
    assert s.dedup.distance_threshold == 0.15
    assert s.rag.top_k == 5
    assert s.chunker.size == 512


def test_nested_env_overrides(monkeypatch: object) -> None:
    mp = monkeypatch  # type: ignore[assignment]
    mp.setenv("HTTP__PORT", "9999")  # type: ignore[attr-defined]
    mp.setenv("GEMINI__EMBEDDING_DIMENSIONS", "768")  # type: ignore[attr-defined]
    mp.setenv("RABBITMQ__INGEST_QUEUE", "custom.queue")  # type: ignore[attr-defined]
    try:
        s = Settings(_env_file=None)  # type: ignore[call-arg]
        assert s.http.port == 9999
        assert s.gemini.embedding_dimensions == 768
        assert s.rabbitmq.ingest_queue == "custom.queue"
    finally:
        for k in ("HTTP__PORT", "GEMINI__EMBEDDING_DIMENSIONS", "RABBITMQ__INGEST_QUEUE"):
            os.environ.pop(k, None)
