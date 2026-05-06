from unittest.mock import AsyncMock

import grpc
import pytest
from ai.assistant.v1 import assistant_pb2
from ai.events.v1 import events_pb2
from google.protobuf import wrappers_pb2

from application.errors import LLMUnavailable
from domain.entities import ChatResponse, Citation
from interfaces.grpc.handlers.assistant import AssistantHandler


def _ctx() -> AsyncMock:
    ctx = AsyncMock()
    ctx.abort = AsyncMock(side_effect=Exception("aborted"))
    return ctx


@pytest.mark.asyncio
async def test_chat_returns_envelope_with_data() -> None:
    use_case = AsyncMock()
    use_case.execute = AsyncMock(
        return_value=ChatResponse(
            answer="hi",
            citations=[
                Citation(source_id="p1", kind="post", text="t", score=0.9),
                Citation(source_id="q1", kind="question", text="u", score=0.8),
            ],
        )
    )
    handler = AssistantHandler(use_case)
    request = assistant_pb2.ChatRequest(text="hello")
    response = await handler.Chat(request, _ctx())
    assert response.code == "00"
    assert response.success is True
    assert response.data.answer == "hi"
    kinds = [c.kind for c in response.data.citations]
    assert kinds == [events_pb2.CONTENT_KIND_POST, events_pb2.CONTENT_KIND_QUESTION]


@pytest.mark.asyncio
async def test_chat_passes_user_id_when_set() -> None:
    use_case = AsyncMock()
    use_case.execute = AsyncMock(return_value=ChatResponse(answer="x", citations=[]))
    handler = AssistantHandler(use_case)
    request = assistant_pb2.ChatRequest(text="hello", user_id=wrappers_pb2.StringValue(value="u1"))
    await handler.Chat(request, _ctx())
    arg = use_case.execute.call_args.args[0]
    assert arg.user_id == "u1"


@pytest.mark.asyncio
async def test_chat_user_id_empty_becomes_none() -> None:
    use_case = AsyncMock()
    use_case.execute = AsyncMock(return_value=ChatResponse(answer="x", citations=[]))
    handler = AssistantHandler(use_case)
    request = assistant_pb2.ChatRequest(text="hello", user_id=wrappers_pb2.StringValue(value=""))
    await handler.Chat(request, _ctx())
    arg = use_case.execute.call_args.args[0]
    assert arg.user_id is None


@pytest.mark.asyncio
async def test_chat_aborts_on_llm_unavailable() -> None:
    use_case = AsyncMock()
    use_case.execute = AsyncMock(side_effect=LLMUnavailable("down"))
    handler = AssistantHandler(use_case)
    request = assistant_pb2.ChatRequest(text="x")
    ctx = _ctx()
    with pytest.raises(Exception):  # noqa: B017 - context.abort raises in real grpc.aio
        await handler.Chat(request, ctx)
    ctx.abort.assert_awaited_once()
    code, _ = ctx.abort.call_args.args
    assert code == grpc.StatusCode.UNAVAILABLE
