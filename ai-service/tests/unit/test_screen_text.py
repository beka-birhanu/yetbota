import asyncio
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from application.errors import LLMUnavailable
from application.screen_text import ScreenText, _cache_key
from domain.entities import ScreeningRequest

_SCHEMA = {"type": "object"}


def _build(
    *,
    classify_return: dict[str, object] | None = None,
    classify_error: Exception | None = None,
    block_threshold: float = 0.7,
    timeout_s: float = 5.0,
    cache_size: int = 16,
    cache_ttl_s: int = 60,
) -> tuple[ScreenText, AsyncMock, MagicMock]:
    llm = AsyncMock()
    if classify_error is not None:
        llm.classify = AsyncMock(side_effect=classify_error)
    else:
        llm.classify = AsyncMock(return_value=classify_return or {})
    prompt_builder = MagicMock(return_value="rendered prompt")
    use_case = ScreenText(
        llm=llm,
        prompt_builder=prompt_builder,
        response_schema=_SCHEMA,
        block_threshold=block_threshold,
        timeout_s=timeout_s,
        cache_size=cache_size,
        cache_ttl_s=cache_ttl_s,
    )
    return use_case, llm, prompt_builder


@pytest.mark.asyncio
async def test_clean_text_returns_ok_true() -> None:
    scores = {
        "profanity": 0.05,
        "hate": 0.0,
        "sexual": 0.1,
        "violence": 0.0,
        "harassment": 0.2,
    }
    use_case, _, _ = _build(classify_return=scores)
    result = await use_case.execute(ScreeningRequest(text="hello world", kind="question"))
    assert result.ok is True
    assert result.reason is None
    assert result.categories["profanity"] == pytest.approx(0.05)


@pytest.mark.asyncio
async def test_blocked_text_picks_highest_category_as_reason() -> None:
    scores = {
        "profanity": 0.95,
        "hate": 0.4,
        "sexual": 0.1,
        "violence": 0.2,
        "harassment": 0.5,
    }
    use_case, _, _ = _build(classify_return=scores)
    result = await use_case.execute(ScreeningRequest(text="bad words", kind="answer"))
    assert result.ok is False
    assert result.reason == "profanity"


@pytest.mark.asyncio
async def test_threshold_boundary_strict_less_than() -> None:
    scores = {
        "profanity": 0.7,
        "hate": 0.0,
        "sexual": 0.0,
        "violence": 0.0,
        "harassment": 0.0,
    }
    use_case, _, _ = _build(classify_return=scores, block_threshold=0.7)
    result = await use_case.execute(ScreeningRequest(text="x", kind="post"))
    assert result.ok is False
    assert result.reason == "profanity"


@pytest.mark.asyncio
async def test_missing_categories_default_to_zero() -> None:
    use_case, _, _ = _build(classify_return={"profanity": 0.1})
    result = await use_case.execute(ScreeningRequest(text="x", kind="post"))
    assert result.ok is True
    for cat in ("hate", "sexual", "violence", "harassment"):
        assert result.categories[cat] == 0.0


@pytest.mark.asyncio
async def test_timeout_returns_screening_unavailable() -> None:
    async def slow(*_: object, **__: object) -> dict[str, object]:
        await asyncio.sleep(10)
        return {}

    llm = AsyncMock()
    llm.classify = AsyncMock(side_effect=slow)
    use_case = ScreenText(
        llm=llm,
        prompt_builder=MagicMock(return_value="x"),
        response_schema=_SCHEMA,
        block_threshold=0.7,
        timeout_s=0.05,
        cache_size=16,
        cache_ttl_s=60,
    )
    result = await use_case.execute(ScreeningRequest(text="hi", kind="post"))
    assert result.ok is False
    assert result.reason == "screening_unavailable"
    assert result.categories == {}


@pytest.mark.asyncio
async def test_llm_unavailable_returns_screening_unavailable() -> None:
    use_case, _, _ = _build(classify_error=LLMUnavailable("down"))
    result = await use_case.execute(ScreeningRequest(text="hi", kind="post"))
    assert result.ok is False
    assert result.reason == "screening_unavailable"


@pytest.mark.asyncio
async def test_cache_hit_skips_llm_on_repeat() -> None:
    use_case, llm, _ = _build(
        classify_return={
            "profanity": 0.1,
            "hate": 0.0,
            "sexual": 0.0,
            "violence": 0.0,
            "harassment": 0.0,
        }
    )
    req = ScreeningRequest(text="hello", kind="question")
    a = await use_case.execute(req)
    b = await use_case.execute(req)
    assert a == b
    assert llm.classify.await_count == 1


@pytest.mark.asyncio
async def test_cache_distinguishes_kind() -> None:
    use_case, llm, _ = _build(
        classify_return={
            "profanity": 0.1,
            "hate": 0.0,
            "sexual": 0.0,
            "violence": 0.0,
            "harassment": 0.0,
        }
    )
    await use_case.execute(ScreeningRequest(text="hello", kind="question"))
    await use_case.execute(ScreeningRequest(text="hello", kind="answer"))
    assert llm.classify.await_count == 2


@pytest.mark.asyncio
async def test_cache_ttl_expiration_refetches() -> None:
    use_case, llm, _ = _build(
        classify_return={
            "profanity": 0.1,
            "hate": 0.0,
            "sexual": 0.0,
            "violence": 0.0,
            "harassment": 0.0,
        },
        cache_ttl_s=10,
    )
    req = ScreeningRequest(text="hello", kind="question")
    with patch("application.screen_text.time.monotonic", return_value=100.0):
        await use_case.execute(req)
    with patch("application.screen_text.time.monotonic", return_value=200.0):
        await use_case.execute(req)
    assert llm.classify.await_count == 2


def test_cache_key_changes_with_text_or_kind() -> None:
    a = _cache_key("hi", "post")
    b = _cache_key("hi", "post")
    c = _cache_key("hi", "question")
    d = _cache_key("hii", "post")
    assert a == b
    assert a != c
    assert a != d
