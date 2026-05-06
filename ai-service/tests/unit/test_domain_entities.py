from datetime import UTC, datetime

import pytest
from pydantic import ValidationError

from domain.entities import (
    Chunk,
    Citation,
    IngestRequest,
    IngestResult,
    ScoredChunk,
    ScreeningRequest,
    ScreeningResult,
)


def test_chunk_is_frozen() -> None:
    c = Chunk(source_id="p1", kind="post", text="hello", segment_idx=0)
    with pytest.raises(ValidationError):
        c.text = "mutated"  # type: ignore[misc]


def test_ingest_request_validates_kind() -> None:
    with pytest.raises(ValidationError):
        IngestRequest(
            content_id="abc",
            kind="bogus",  # type: ignore[arg-type]
            user_id="u1",
            text="hi",
        )


def test_ingest_result_carries_verdict_and_optional_fields() -> None:
    r = IngestResult(
        content_id="abc",
        kind="post",
        verdict="duplicate",
        duplicate_of="other-id",
        processed_at=datetime.now(UTC),
    )
    assert r.error_code is None
    assert r.duplicate_of == "other-id"


def test_scored_chunk_pairs_chunk_and_score() -> None:
    c = Chunk(source_id="p1", kind="post", text="abc")
    sc = ScoredChunk(chunk=c, score=0.9)
    assert sc.score == 0.9


def test_citation_requires_kind() -> None:
    cit = Citation(source_id="p1", kind="answer", text="x", score=0.8)
    assert cit.kind == "answer"


def test_screening_result_optional_reason() -> None:
    ok = ScreeningResult(ok=True)
    blocked = ScreeningResult(ok=False, reason="profanity", categories={"profanity": 0.9})
    assert ok.reason is None
    assert blocked.reason == "profanity"
    assert blocked.categories["profanity"] == 0.9


def test_screening_request_kind_is_validated() -> None:
    req = ScreeningRequest(text="hi", kind="question")
    assert req.kind == "question"
