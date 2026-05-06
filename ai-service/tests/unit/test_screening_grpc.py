from unittest.mock import AsyncMock

import pytest
from ai.events.v1 import events_pb2
from ai.screening.v1 import screening_pb2

from domain.entities import ScreeningResult
from interfaces.grpc.handlers.screening import ScreeningHandler


@pytest.mark.asyncio
async def test_check_returns_envelope_with_categories() -> None:
    use_case = AsyncMock()
    use_case.execute = AsyncMock(
        return_value=ScreeningResult(
            ok=True,
            reason=None,
            categories={"profanity": 0.05, "hate": 0.0},
        )
    )
    handler = ScreeningHandler(use_case)
    request = screening_pb2.CheckRequest(text="hi", kind=events_pb2.CONTENT_KIND_POST)
    response = await handler.Check(request, AsyncMock())
    assert response.code == "00"
    assert response.success is True
    assert response.data.ok is True
    assert response.data.HasField("reason") is False
    assert response.data.categories["profanity"] == pytest.approx(0.05)


@pytest.mark.asyncio
async def test_check_passes_kind_to_use_case() -> None:
    use_case = AsyncMock()
    use_case.execute = AsyncMock(return_value=ScreeningResult(ok=True, reason=None, categories={}))
    handler = ScreeningHandler(use_case)
    request = screening_pb2.CheckRequest(text="hi", kind=events_pb2.CONTENT_KIND_ANSWER)
    await handler.Check(request, AsyncMock())
    arg = use_case.execute.call_args.args[0]
    assert arg.kind == "answer"


@pytest.mark.asyncio
async def test_check_blocked_sets_reason_string_value() -> None:
    use_case = AsyncMock()
    use_case.execute = AsyncMock(
        return_value=ScreeningResult(
            ok=False,
            reason="profanity",
            categories={"profanity": 0.9},
        )
    )
    handler = ScreeningHandler(use_case)
    request = screening_pb2.CheckRequest(text="bad", kind=events_pb2.CONTENT_KIND_QUESTION)
    response = await handler.Check(request, AsyncMock())
    assert response.data.ok is False
    assert response.data.HasField("reason") is True
    assert response.data.reason.value == "profanity"


@pytest.mark.asyncio
async def test_unspecified_kind_defaults_to_post() -> None:
    use_case = AsyncMock()
    use_case.execute = AsyncMock(return_value=ScreeningResult(ok=True, reason=None, categories={}))
    handler = ScreeningHandler(use_case)
    request = screening_pb2.CheckRequest(text="hi", kind=events_pb2.CONTENT_KIND_UNSPECIFIED)
    await handler.Check(request, AsyncMock())
    arg = use_case.execute.call_args.args[0]
    assert arg.kind == "post"
