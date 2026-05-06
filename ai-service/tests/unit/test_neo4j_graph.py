from unittest.mock import AsyncMock, MagicMock, patch

import pytest
from neo4j.exceptions import Neo4jError

from application.errors import IndexingFailed
from domain.entities import SimilarPost
from infrastructure.config.settings import Neo4jSettings
from infrastructure.graph import Neo4jSimilarityGraph


def _settings() -> Neo4jSettings:
    return Neo4jSettings(
        uri="bolt://localhost:7687",
        username="neo4j",
        password="secret",
        database="neo4j",
    )


@pytest.fixture
def patched_driver() -> tuple[AsyncMock, AsyncMock]:
    driver = AsyncMock()
    driver.verify_connectivity = AsyncMock()
    driver.close = AsyncMock()

    session = AsyncMock()
    session.__aenter__ = AsyncMock(return_value=session)
    session.__aexit__ = AsyncMock(return_value=None)
    session.execute_write = AsyncMock()
    driver.session = MagicMock(return_value=session)

    with patch(
        "infrastructure.graph.neo4j_graph.AsyncGraphDatabase.driver",
        return_value=driver,
    ):
        yield driver, session


@pytest.mark.asyncio
async def test_connect_verifies_and_logs(patched_driver: tuple[AsyncMock, AsyncMock]) -> None:
    driver, _ = patched_driver
    graph = Neo4jSimilarityGraph(_settings())
    await graph.connect()
    driver.verify_connectivity.assert_awaited_once()


@pytest.mark.asyncio
async def test_update_similar_writes_rows(patched_driver: tuple[AsyncMock, AsyncMock]) -> None:
    _, session = patched_driver
    graph = Neo4jSimilarityGraph(_settings())
    await graph.connect()

    similar = [
        SimilarPost(post_id="q1", score=0.92),
        SimilarPost(post_id="q2", score=0.81),
    ]
    await graph.update_similar("p1", similar)

    session.execute_write.assert_awaited_once()
    runner = session.execute_write.call_args.args[0]
    assert callable(runner)


@pytest.mark.asyncio
async def test_update_similar_without_connect_raises() -> None:
    graph = Neo4jSimilarityGraph(_settings())
    with pytest.raises(RuntimeError):
        await graph.update_similar("p1", [])


@pytest.mark.asyncio
async def test_update_similar_wraps_neo4j_errors(
    patched_driver: tuple[AsyncMock, AsyncMock],
) -> None:
    _, session = patched_driver
    session.execute_write = AsyncMock(side_effect=Neo4jError("boom"))
    graph = Neo4jSimilarityGraph(_settings())
    await graph.connect()
    with pytest.raises(IndexingFailed):
        await graph.update_similar("p1", [SimilarPost(post_id="q1", score=0.5)])


@pytest.mark.asyncio
async def test_mark_duplicate_writes_edge(patched_driver: tuple[AsyncMock, AsyncMock]) -> None:
    _, session = patched_driver
    graph = Neo4jSimilarityGraph(_settings())
    await graph.connect()

    await graph.mark_duplicate("dup1", "orig1", 0.93)

    session.execute_write.assert_awaited_once()
    runner = session.execute_write.call_args.args[0]
    assert callable(runner)


@pytest.mark.asyncio
async def test_mark_duplicate_without_connect_raises() -> None:
    graph = Neo4jSimilarityGraph(_settings())
    with pytest.raises(RuntimeError):
        await graph.mark_duplicate("dup1", "orig1", 0.5)


@pytest.mark.asyncio
async def test_mark_duplicate_wraps_neo4j_errors(
    patched_driver: tuple[AsyncMock, AsyncMock],
) -> None:
    _, session = patched_driver
    session.execute_write = AsyncMock(side_effect=Neo4jError("boom"))
    graph = Neo4jSimilarityGraph(_settings())
    await graph.connect()
    with pytest.raises(IndexingFailed):
        await graph.mark_duplicate("dup1", "orig1", 0.5)


@pytest.mark.asyncio
async def test_delete_post_runs_detach_delete(
    patched_driver: tuple[AsyncMock, AsyncMock],
) -> None:
    _, session = patched_driver
    graph = Neo4jSimilarityGraph(_settings())
    await graph.connect()

    await graph.delete_post("p1")

    session.execute_write.assert_awaited_once()


@pytest.mark.asyncio
async def test_delete_post_without_connect_raises() -> None:
    graph = Neo4jSimilarityGraph(_settings())
    with pytest.raises(RuntimeError):
        await graph.delete_post("p1")


@pytest.mark.asyncio
async def test_delete_post_wraps_neo4j_errors(
    patched_driver: tuple[AsyncMock, AsyncMock],
) -> None:
    _, session = patched_driver
    session.execute_write = AsyncMock(side_effect=Neo4jError("boom"))
    graph = Neo4jSimilarityGraph(_settings())
    await graph.connect()
    with pytest.raises(IndexingFailed):
        await graph.delete_post("p1")


@pytest.mark.asyncio
async def test_close_releases_driver(patched_driver: tuple[AsyncMock, AsyncMock]) -> None:
    driver, _ = patched_driver
    graph = Neo4jSimilarityGraph(_settings())
    await graph.connect()
    await graph.close()
    driver.close.assert_awaited_once()


@pytest.mark.asyncio
async def test_close_when_never_connected_is_noop() -> None:
    graph = Neo4jSimilarityGraph(_settings())
    await graph.close()
