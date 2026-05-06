from neo4j import AsyncDriver, AsyncGraphDatabase
from neo4j.exceptions import Neo4jError

from application.errors import IndexingFailed
from domain.entities import SimilarPost
from infrastructure.config.settings import Neo4jSettings
from infrastructure.observability import get_logger

logger = get_logger(__name__)


_REPLACE_SIMILAR_CYPHER = """
MERGE (p:Post {id: $post_id})
WITH p
OPTIONAL MATCH (p)-[r:SIMILAR_TO]->()
DELETE r
WITH p
UNWIND $similar AS row
MERGE (q:Post {id: row.post_id})
MERGE (p)-[s:SIMILAR_TO]->(q)
SET s.score = row.score
"""

_MARK_DUPLICATE_CYPHER = """
MERGE (d:Post {id: $duplicate_id})
MERGE (o:Post {id: $original_id})
MERGE (d)-[r:DUPLICATE_OF]->(o)
SET r.score = $score
"""

_DELETE_POST_CYPHER = """
MATCH (p:Post {id: $post_id})
DETACH DELETE p
"""


class Neo4jSimilarityGraph:
    def __init__(self, settings: Neo4jSettings) -> None:
        self._settings = settings
        self._driver: AsyncDriver | None = None

    async def connect(self) -> None:
        self._driver = AsyncGraphDatabase.driver(
            self._settings.uri,
            auth=(self._settings.username, self._settings.password),
        )
        await self._driver.verify_connectivity()
        logger.info("neo4j.connected", uri=self._settings.uri)

    async def update_similar(self, post_id: str, similar: list[SimilarPost]) -> None:
        if self._driver is None:
            raise RuntimeError("neo4j driver not connected")
        rows = [{"post_id": s.post_id, "score": float(s.score)} for s in similar]
        try:
            async with self._driver.session(database=self._settings.database) as session:
                await session.execute_write(
                    lambda tx: tx.run(
                        _REPLACE_SIMILAR_CYPHER,
                        post_id=post_id,
                        similar=rows,
                    ).consume()
                )
        except Neo4jError as exc:
            raise IndexingFailed(f"neo4j update_similar failed: {exc}", cause=exc) from exc

    async def mark_duplicate(
        self, duplicate_post_id: str, original_post_id: str, score: float
    ) -> None:
        if self._driver is None:
            raise RuntimeError("neo4j driver not connected")
        try:
            async with self._driver.session(database=self._settings.database) as session:
                await session.execute_write(
                    lambda tx: tx.run(
                        _MARK_DUPLICATE_CYPHER,
                        duplicate_id=duplicate_post_id,
                        original_id=original_post_id,
                        score=float(score),
                    ).consume()
                )
        except Neo4jError as exc:
            raise IndexingFailed(f"neo4j mark_duplicate failed: {exc}", cause=exc) from exc

    async def delete_post(self, post_id: str) -> None:
        if self._driver is None:
            raise RuntimeError("neo4j driver not connected")
        try:
            async with self._driver.session(database=self._settings.database) as session:
                await session.execute_write(
                    lambda tx: tx.run(_DELETE_POST_CYPHER, post_id=post_id).consume()
                )
        except Neo4jError as exc:
            raise IndexingFailed(f"neo4j delete_post failed: {exc}", cause=exc) from exc

    async def close(self) -> None:
        if self._driver is not None:
            await self._driver.close()
            self._driver = None
            logger.info("neo4j.closed")
