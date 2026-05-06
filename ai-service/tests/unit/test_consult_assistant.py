from unittest.mock import AsyncMock, MagicMock

import pytest

from application.consult_assistant import ConsultAssistant
from application.errors import LLMUnavailable
from domain.entities import ChatQuery, Chunk, ScoredChunk


def _hit(source_id: str, kind: str = "post", score: float = 0.8) -> ScoredChunk:
    return ScoredChunk(
        chunk=Chunk(source_id=source_id, kind=kind, text=f"text-{source_id}"),  # type: ignore[arg-type]
        score=score,
    )


def _build(
    *,
    query_vec: list[float] | None = None,
    hits: list[ScoredChunk] | None = None,
    llm_text: str = "answer",
    llm_error: Exception | None = None,
    min_similarity: float = 0.6,
    top_k: int = 5,
) -> tuple[ConsultAssistant, AsyncMock, AsyncMock, AsyncMock, MagicMock]:
    embedder = AsyncMock()
    embedder.embed = AsyncMock(return_value=query_vec or [0.1] * 4)
    store = AsyncMock()
    store.search = AsyncMock(return_value=hits if hits is not None else [])
    llm = AsyncMock()
    if llm_error is not None:
        llm.generate = AsyncMock(side_effect=llm_error)
    else:
        llm.generate = AsyncMock(return_value=llm_text)
    prompt_builder = MagicMock(return_value="rendered prompt")
    use_case = ConsultAssistant(
        embedder=embedder,
        vector_store=store,
        llm=llm,
        prompt_builder=prompt_builder,
        top_k=top_k,
        min_similarity=min_similarity,
        max_tokens=512,
        temperature=0.2,
    )
    return use_case, embedder, store, llm, prompt_builder


@pytest.mark.asyncio
async def test_high_score_hits_invoke_llm_and_return_citations() -> None:
    hits = [_hit("p1", score=0.9), _hit("q1", kind="question", score=0.7)]
    use_case, _, _, llm, prompt_builder = _build(hits=hits, llm_text="The answer is X.")
    resp = await use_case.execute(ChatQuery(text="where can I get coffee?"))
    assert resp.answer == "The answer is X."
    assert [c.source_id for c in resp.citations] == ["p1", "q1"]
    assert [c.kind for c in resp.citations] == ["post", "question"]
    llm.generate.assert_awaited_once()
    prompt_builder.assert_called_once()


@pytest.mark.asyncio
async def test_no_hits_returns_no_info_without_llm_call() -> None:
    use_case, _, _, llm, prompt_builder = _build(hits=[])
    resp = await use_case.execute(ChatQuery(text="anything"))
    assert resp.citations == []
    assert "do not have verified information" in resp.answer
    llm.generate.assert_not_awaited()
    prompt_builder.assert_not_called()


@pytest.mark.asyncio
async def test_below_min_similarity_returns_no_info() -> None:
    hits = [_hit("p1", score=0.4)]
    use_case, _, _, llm, _ = _build(hits=hits, min_similarity=0.6)
    resp = await use_case.execute(ChatQuery(text="x"))
    assert resp.citations == []
    assert "do not have verified information" in resp.answer
    llm.generate.assert_not_awaited()


@pytest.mark.asyncio
async def test_llm_failure_propagates() -> None:
    hits = [_hit("p1", score=0.9)]
    use_case, _, _, _, _ = _build(hits=hits, llm_error=LLMUnavailable("down"))
    with pytest.raises(LLMUnavailable):
        await use_case.execute(ChatQuery(text="x"))


@pytest.mark.asyncio
async def test_query_embedded_with_query_task_type() -> None:
    use_case, embedder, store, _, _ = _build(hits=[_hit("p1", score=0.8)])
    await use_case.execute(ChatQuery(text="hello"))
    embedder.embed.assert_awaited_once()
    kwargs = embedder.embed.call_args.kwargs
    assert kwargs["task_type"] == "RETRIEVAL_QUERY"
    store.search.assert_awaited_once()
    search_kwargs = store.search.call_args.kwargs
    assert search_kwargs["kinds"] is None
    assert search_kwargs["limit"] == 5


@pytest.mark.asyncio
async def test_answer_is_stripped() -> None:
    hits = [_hit("p1", score=0.9)]
    use_case, _, _, _, _ = _build(hits=hits, llm_text="  hello  \n")
    resp = await use_case.execute(ChatQuery(text="x"))
    assert resp.answer == "hello"


@pytest.mark.asyncio
async def test_prompt_builder_receives_hits_and_query_text() -> None:
    hits = [_hit("p1", score=0.9), _hit("a1", kind="answer", score=0.8)]
    use_case, _, _, _, prompt_builder = _build(hits=hits)
    await use_case.execute(ChatQuery(text="why?"))
    args, _ = prompt_builder.call_args
    passed_hits, query_text = args
    assert passed_hits == hits
    assert query_text == "why?"
