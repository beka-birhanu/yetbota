from unittest.mock import AsyncMock, MagicMock

import pytest

from application.errors import (
    EmbeddingFailed,
    IndexingFailed,
    MetadataError,
    SimilaritySearchFailed,
)
from application.ingest_content import IngestContent, _mean
from domain.entities import Chunk, IngestRequest, ScoredChunk


def _post(**overrides: object) -> IngestRequest:
    base = {
        "content_id": "p1",
        "kind": "post",
        "user_id": "u1",
        "text": "hello world",
        "category": "cafe",
    }
    base.update(overrides)
    return IngestRequest(**base)  # type: ignore[arg-type]


def _question(**overrides: object) -> IngestRequest:
    base = {
        "content_id": "q1",
        "kind": "question",
        "user_id": "u1",
        "text": "where can I get good coffee?",
    }
    base.update(overrides)
    return IngestRequest(**base)  # type: ignore[arg-type]


def _answer(**overrides: object) -> IngestRequest:
    base = {
        "content_id": "a1",
        "kind": "answer",
        "user_id": "u1",
        "text": "try the cafe on main street",
        "parent_id": "q1",
    }
    base.update(overrides)
    return IngestRequest(**base)  # type: ignore[arg-type]


def _build(
    *,
    chunks_per_call: list[Chunk] | None = None,
    embeddings: list[list[float]] | None = None,
    dedup_hits: list[ScoredChunk] | None = None,
    embedder_error: Exception | None = None,
    dedup_error: Exception | None = None,
    upsert_error: Exception | None = None,
    distance_threshold: float = 0.15,
    similarity_graph: AsyncMock | None = None,
    similarity_top_n: int = 0,
    similarity_results: list[list[object]] | None = None,
    neighbor_embeddings: dict[str, list[float] | None] | None = None,
) -> tuple[IngestContent, MagicMock, AsyncMock, AsyncMock]:
    chunker = MagicMock()
    chunker.chunk = MagicMock(return_value=chunks_per_call or [])

    embedder = AsyncMock()
    if embedder_error is not None:
        embedder.embed_batch = AsyncMock(side_effect=embedder_error)
    else:
        embedder.embed_batch = AsyncMock(return_value=embeddings or [])

    store = AsyncMock()
    store.search_dedup = AsyncMock(
        side_effect=dedup_error if dedup_error else None,
        return_value=dedup_hits if dedup_hits is not None else [],
    )
    store.upsert = AsyncMock(side_effect=upsert_error)
    if similarity_results is not None:
        store.find_similar_posts = AsyncMock(side_effect=similarity_results)
    else:
        store.find_similar_posts = AsyncMock(return_value=[])
    if neighbor_embeddings is not None:

        async def _get_post_embedding(post_id: str) -> list[float] | None:
            return neighbor_embeddings.get(post_id)

        store.get_post_embedding = AsyncMock(side_effect=_get_post_embedding)
    else:
        store.get_post_embedding = AsyncMock(return_value=None)

    use_case = IngestContent(
        chunker=chunker,
        embedder=embedder,
        vector_store=store,
        distance_threshold=distance_threshold,
        similarity_graph=similarity_graph,
        similarity_top_n=similarity_top_n,
    )
    return use_case, chunker, embedder, store


def test_mean_handles_empty_input() -> None:
    assert _mean([]) == []


def test_mean_averages_aligned_vectors() -> None:
    assert _mean([[1.0, 2.0], [3.0, 4.0]]) == [2.0, 3.0]


def test_mean_rejects_dim_mismatch() -> None:
    with pytest.raises(ValueError):
        _mean([[1.0, 2.0], [1.0]])


@pytest.mark.asyncio
async def test_post_unique_upserts_and_returns_unique() -> None:
    chunks = [Chunk(source_id="p1", kind="post", text="hello world", segment_idx=0)]
    use_case, _, embedder, store = _build(
        chunks_per_call=chunks,
        embeddings=[[0.1] * 4],
        dedup_hits=[],
    )
    result = await use_case.execute(_post())
    assert result.verdict == "unique"
    assert result.duplicate_of is None
    embedder.embed_batch.assert_awaited_once()
    store.search_dedup.assert_awaited_once()
    store.upsert.assert_awaited_once()


@pytest.mark.asyncio
async def test_post_duplicate_skips_upsert_and_returns_duplicate_of() -> None:
    chunks = [Chunk(source_id="p1", kind="post", text="hello world")]
    hit = ScoredChunk(
        chunk=Chunk(source_id="other", kind="post", text="hi world"),
        score=0.92,
    )
    use_case, _, _, store = _build(
        chunks_per_call=chunks,
        embeddings=[[0.1] * 4],
        dedup_hits=[hit],
    )
    result = await use_case.execute(_post())
    assert result.verdict == "duplicate"
    assert result.duplicate_of == "other"
    store.upsert.assert_not_awaited()


@pytest.mark.asyncio
async def test_question_skips_dedup_and_returns_indexed() -> None:
    chunks = [Chunk(source_id="q1", kind="question", text="where can I get good coffee?")]
    use_case, _, _, store = _build(
        chunks_per_call=chunks,
        embeddings=[[0.1] * 4],
    )
    result = await use_case.execute(_question())
    assert result.verdict == "indexed"
    store.search_dedup.assert_not_awaited()
    store.upsert.assert_awaited_once()


@pytest.mark.asyncio
async def test_answer_skips_dedup_and_returns_indexed() -> None:
    chunks = [Chunk(source_id="a1", kind="answer", text="try the cafe on main street")]
    use_case, _, _, store = _build(
        chunks_per_call=chunks,
        embeddings=[[0.1] * 4],
    )
    result = await use_case.execute(_answer())
    assert result.verdict == "indexed"
    store.search_dedup.assert_not_awaited()


@pytest.mark.asyncio
async def test_empty_text_raises_metadata_error() -> None:
    use_case, _, _, _ = _build()
    with pytest.raises(MetadataError):
        await use_case.execute(_post(text="   "))


@pytest.mark.asyncio
async def test_post_without_category_raises_metadata_error() -> None:
    use_case, _, _, _ = _build()
    with pytest.raises(MetadataError):
        await use_case.execute(_post(category=None))


@pytest.mark.asyncio
async def test_answer_without_parent_id_raises_metadata_error() -> None:
    use_case, _, _, _ = _build()
    with pytest.raises(MetadataError):
        await use_case.execute(_answer(parent_id=None))


@pytest.mark.asyncio
async def test_chunker_returning_empty_raises_metadata_error() -> None:
    use_case, _, _, _ = _build(chunks_per_call=[])
    with pytest.raises(MetadataError):
        await use_case.execute(_post())


@pytest.mark.asyncio
async def test_embedding_failure_propagates() -> None:
    chunks = [Chunk(source_id="p1", kind="post", text="hello")]
    use_case, _, _, _ = _build(
        chunks_per_call=chunks,
        embedder_error=EmbeddingFailed("nope"),
    )
    with pytest.raises(EmbeddingFailed):
        await use_case.execute(_post())


@pytest.mark.asyncio
async def test_similarity_search_failure_propagates() -> None:
    chunks = [Chunk(source_id="p1", kind="post", text="hello")]
    use_case, _, _, _ = _build(
        chunks_per_call=chunks,
        embeddings=[[0.1] * 4],
        dedup_error=SimilaritySearchFailed("nope"),
    )
    with pytest.raises(SimilaritySearchFailed):
        await use_case.execute(_post())


@pytest.mark.asyncio
async def test_indexing_failure_propagates() -> None:
    chunks = [Chunk(source_id="p1", kind="post", text="hello")]
    use_case, _, _, _ = _build(
        chunks_per_call=chunks,
        embeddings=[[0.1] * 4],
        dedup_hits=[],
        upsert_error=IndexingFailed("nope"),
    )
    with pytest.raises(IndexingFailed):
        await use_case.execute(_post())


@pytest.mark.asyncio
async def test_dedup_uses_mean_of_chunk_embeddings() -> None:
    chunks = [
        Chunk(source_id="p1", kind="post", text="a"),
        Chunk(source_id="p1", kind="post", text="b", segment_idx=1),
    ]
    use_case, _, _, store = _build(
        chunks_per_call=chunks,
        embeddings=[[0.0, 1.0], [1.0, 0.0]],
        dedup_hits=[],
    )
    await use_case.execute(_post())
    args, kwargs = store.search_dedup.call_args
    query_vec = args[0] if args else kwargs["query_embedding"]
    assert query_vec == [0.5, 0.5]


@pytest.mark.asyncio
async def test_chunker_receives_metadata_with_tags_and_category() -> None:
    chunks = [Chunk(source_id="p1", kind="post", text="x")]
    use_case, chunker, _, _ = _build(
        chunks_per_call=chunks,
        embeddings=[[0.1] * 4],
        dedup_hits=[],
    )
    await use_case.execute(_post(tags=["cafe", "wifi"]))
    kwargs = chunker.chunk.call_args.kwargs
    assert kwargs["metadata"]["category"] == "cafe"
    assert kwargs["metadata"]["tags"] == "cafe,wifi"
    assert kwargs["kind"] == "post"
    assert kwargs["source_id"] == "p1"


@pytest.mark.asyncio
async def test_post_unique_writes_similarity_edges_and_recomputes_neighbors() -> None:
    from domain.entities import SimilarPost

    chunks = [Chunk(source_id="p1", kind="post", text="hello", segment_idx=0)]
    similarity_graph = AsyncMock()
    primary_neighbors = [
        SimilarPost(post_id="q1", score=0.9),
        SimilarPost(post_id="q2", score=0.8),
    ]
    q1_neighbors = [SimilarPost(post_id="p1", score=0.9)]
    q2_neighbors = [SimilarPost(post_id="p1", score=0.8)]
    use_case, _, _, store = _build(
        chunks_per_call=chunks,
        embeddings=[[0.1, 0.2]],
        dedup_hits=[],
        similarity_graph=similarity_graph,
        similarity_top_n=2,
        similarity_results=[primary_neighbors, q1_neighbors, q2_neighbors],
        neighbor_embeddings={"q1": [0.3, 0.4], "q2": [0.5, 0.6]},
    )
    result = await use_case.execute(_post())
    assert result.verdict == "unique"

    assert store.find_similar_posts.await_count == 3
    first_call = store.find_similar_posts.await_args_list[0]
    assert first_call.kwargs["exclude_post_id"] == "p1"
    second_call = store.find_similar_posts.await_args_list[1]
    assert second_call.kwargs["exclude_post_id"] == "q1"

    assert similarity_graph.update_similar.await_count == 3
    update_calls = similarity_graph.update_similar.await_args_list
    assert update_calls[0].args[0] == "p1"
    assert {call.args[0] for call in update_calls[1:]} == {"q1", "q2"}


@pytest.mark.asyncio
async def test_post_unique_skips_neighbor_when_embedding_missing() -> None:
    from domain.entities import SimilarPost

    chunks = [Chunk(source_id="p1", kind="post", text="hello", segment_idx=0)]
    similarity_graph = AsyncMock()
    use_case, _, _, store = _build(
        chunks_per_call=chunks,
        embeddings=[[0.1, 0.2]],
        dedup_hits=[],
        similarity_graph=similarity_graph,
        similarity_top_n=1,
        similarity_results=[[SimilarPost(post_id="q1", score=0.9)]],
        neighbor_embeddings={"q1": None},
    )
    await use_case.execute(_post())
    assert store.find_similar_posts.await_count == 1
    similarity_graph.update_similar.assert_awaited_once()


@pytest.mark.asyncio
async def test_post_duplicate_records_duplicate_edge_and_skips_similar() -> None:
    chunks = [Chunk(source_id="p1", kind="post", text="hello")]
    hit = ScoredChunk(
        chunk=Chunk(source_id="other", kind="post", text="hi"),
        score=0.95,
    )
    similarity_graph = AsyncMock()
    use_case, _, _, store = _build(
        chunks_per_call=chunks,
        embeddings=[[0.1, 0.2]],
        dedup_hits=[hit],
        similarity_graph=similarity_graph,
        similarity_top_n=5,
    )
    result = await use_case.execute(_post())

    assert result.verdict == "duplicate"
    assert result.duplicate_of == "other"
    similarity_graph.mark_duplicate.assert_awaited_once_with("p1", "other", 0.95)
    store.upsert.assert_not_awaited()
    store.find_similar_posts.assert_not_awaited()
    similarity_graph.update_similar.assert_not_awaited()


@pytest.mark.asyncio
async def test_post_duplicate_swallows_graph_failure() -> None:
    chunks = [Chunk(source_id="p1", kind="post", text="hello")]
    hit = ScoredChunk(
        chunk=Chunk(source_id="other", kind="post", text="hi"),
        score=0.9,
    )
    similarity_graph = AsyncMock()
    similarity_graph.mark_duplicate = AsyncMock(side_effect=IndexingFailed("boom"))
    use_case, _, _, _ = _build(
        chunks_per_call=chunks,
        embeddings=[[0.1, 0.2]],
        dedup_hits=[hit],
        similarity_graph=similarity_graph,
    )
    result = await use_case.execute(_post())
    assert result.verdict == "duplicate"
    assert result.duplicate_of == "other"


@pytest.mark.asyncio
async def test_question_skips_similarity_graph() -> None:
    chunks = [Chunk(source_id="q1", kind="question", text="hi")]
    similarity_graph = AsyncMock()
    use_case, _, _, store = _build(
        chunks_per_call=chunks,
        embeddings=[[0.1, 0.2]],
        similarity_graph=similarity_graph,
        similarity_top_n=5,
    )
    await use_case.execute(_question())
    store.find_similar_posts.assert_not_awaited()
    similarity_graph.update_similar.assert_not_awaited()


@pytest.mark.asyncio
async def test_similarity_step_is_best_effort_on_failure() -> None:
    from application.errors import IndexingFailed

    chunks = [Chunk(source_id="p1", kind="post", text="hello")]
    similarity_graph = AsyncMock()
    similarity_graph.update_similar = AsyncMock(side_effect=IndexingFailed("neo4j down"))
    use_case, _, _, store = _build(
        chunks_per_call=chunks,
        embeddings=[[0.1, 0.2]],
        dedup_hits=[],
        similarity_graph=similarity_graph,
        similarity_top_n=2,
        similarity_results=[[]],
    )
    result = await use_case.execute(_post())
    assert result.verdict == "unique"
    assert result.error_code is None
    store.upsert.assert_awaited_once()


@pytest.mark.asyncio
async def test_similarity_disabled_when_top_n_zero() -> None:
    chunks = [Chunk(source_id="p1", kind="post", text="hello")]
    similarity_graph = AsyncMock()
    use_case, _, _, store = _build(
        chunks_per_call=chunks,
        embeddings=[[0.1, 0.2]],
        dedup_hits=[],
        similarity_graph=similarity_graph,
        similarity_top_n=0,
    )
    await use_case.execute(_post())
    store.find_similar_posts.assert_not_awaited()
    similarity_graph.update_similar.assert_not_awaited()
