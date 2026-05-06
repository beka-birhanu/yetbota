from unittest.mock import AsyncMock

import pytest

from application.delete_content import DeleteContent
from application.errors import IndexingFailed, MetadataError
from domain.entities import DeleteRequest


def _build(
    *,
    similarity_graph: AsyncMock | None = None,
    delete_error: Exception | None = None,
) -> tuple[DeleteContent, AsyncMock]:
    store = AsyncMock()
    store.delete_by_source = AsyncMock(side_effect=delete_error)
    use_case = DeleteContent(
        vector_store=store,
        similarity_graph=similarity_graph,
    )
    return use_case, store


@pytest.mark.asyncio
async def test_delete_post_removes_vectors_and_graph_node() -> None:
    similarity_graph = AsyncMock()
    use_case, store = _build(similarity_graph=similarity_graph)

    result = await use_case.execute(DeleteRequest(content_id="p1", kind="post"))

    assert result.deleted is True
    assert result.content_id == "p1"
    assert result.kind == "post"
    store.delete_by_source.assert_awaited_once_with("p1", "post")
    similarity_graph.delete_post.assert_awaited_once_with("p1")


@pytest.mark.asyncio
async def test_delete_question_skips_graph() -> None:
    similarity_graph = AsyncMock()
    use_case, store = _build(similarity_graph=similarity_graph)

    result = await use_case.execute(DeleteRequest(content_id="q1", kind="question"))

    assert result.deleted is True
    store.delete_by_source.assert_awaited_once_with("q1", "question")
    similarity_graph.delete_post.assert_not_awaited()


@pytest.mark.asyncio
async def test_delete_answer_skips_graph() -> None:
    similarity_graph = AsyncMock()
    use_case, store = _build(similarity_graph=similarity_graph)

    await use_case.execute(DeleteRequest(content_id="a1", kind="answer"))

    similarity_graph.delete_post.assert_not_awaited()


@pytest.mark.asyncio
async def test_delete_post_without_graph_still_deletes_vectors() -> None:
    use_case, store = _build(similarity_graph=None)
    result = await use_case.execute(DeleteRequest(content_id="p1", kind="post"))
    assert result.deleted is True
    store.delete_by_source.assert_awaited_once()


@pytest.mark.asyncio
async def test_delete_swallows_graph_failure() -> None:
    similarity_graph = AsyncMock()
    similarity_graph.delete_post = AsyncMock(side_effect=IndexingFailed("boom"))
    use_case, _ = _build(similarity_graph=similarity_graph)

    result = await use_case.execute(DeleteRequest(content_id="p1", kind="post"))
    assert result.deleted is True


@pytest.mark.asyncio
async def test_delete_propagates_vector_store_failure() -> None:
    similarity_graph = AsyncMock()
    use_case, _ = _build(
        similarity_graph=similarity_graph,
        delete_error=IndexingFailed("boom"),
    )
    with pytest.raises(IndexingFailed):
        await use_case.execute(DeleteRequest(content_id="p1", kind="post"))
    similarity_graph.delete_post.assert_not_awaited()


@pytest.mark.asyncio
async def test_delete_rejects_empty_content_id() -> None:
    use_case, _ = _build()
    with pytest.raises(MetadataError):
        await use_case.execute(DeleteRequest(content_id="", kind="post"))
