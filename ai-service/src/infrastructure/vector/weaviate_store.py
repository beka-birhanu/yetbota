import uuid as uuid_lib
from typing import Any

import weaviate
from weaviate.classes.data import DataObject
from weaviate.classes.init import AdditionalConfig, Auth, Timeout
from weaviate.classes.query import Filter, MetadataQuery
from weaviate.client import WeaviateAsyncClient
from weaviate.exceptions import WeaviateBaseError

from application.errors import IndexingFailed, SimilaritySearchFailed
from domain.entities import (
    Chunk,
    ContentKind,
    Embedding,
    ScoredChunk,
    SimilarPost,
)
from infrastructure.config.settings import WeaviateSettings
from infrastructure.observability import WEAVIATE_CALLS, get_logger, observe
from infrastructure.vector.schema import (
    content_chunk_properties,
    content_chunk_vector_config,
)

logger = get_logger(__name__)

_NAMESPACE = uuid_lib.UUID("6f1a4c2a-aaaa-4b21-9b07-9a8c2f1c5e21")


def _chunk_uuid(kind: ContentKind, source_id: str, segment_idx: int) -> str:
    return str(uuid_lib.uuid5(_NAMESPACE, f"{kind}:{source_id}:{segment_idx}"))


def _properties(chunk: Chunk) -> dict[str, Any]:
    meta = chunk.metadata or {}
    props: dict[str, Any] = {
        "sourceId": chunk.source_id,
        "kind": chunk.kind,
        "text": chunk.text,
        "segmentIdx": chunk.segment_idx,
        "category": meta.get("category", ""),
        "tags": _split_tags(meta.get("tags", "")),
        "parentId": meta.get("parent_id", ""),
        "attachedPostId": meta.get("attached_post_id", ""),
    }
    coord = _coordinate(meta)
    if coord is not None:
        props["coordinate"] = coord
    return props


def _split_tags(raw: str) -> list[str]:
    if not raw:
        return []
    return [t.strip() for t in raw.split(",") if t.strip()]


def _coordinate(meta: dict[str, str]) -> dict[str, float] | None:
    lat = meta.get("lat")
    lon = meta.get("lon")
    if lat is None or lon is None:
        return None
    try:
        return {"latitude": float(lat), "longitude": float(lon)}
    except (TypeError, ValueError):
        return None


class WeaviateVectorStore:
    def __init__(self, settings: WeaviateSettings) -> None:
        self._settings = settings
        self._client: WeaviateAsyncClient | None = None

    async def connect(self) -> None:
        auth = Auth.api_key(self._settings.api_key) if self._settings.api_key else None
        self._client = weaviate.use_async_with_weaviate_cloud(
            cluster_url=self._settings.url,
            auth_credentials=auth,
            additional_config=AdditionalConfig(
                timeout=Timeout(init=30, query=60, insert=120),
            ),
        )
        await self._client.connect()
        await self._ensure_class()

    async def close(self) -> None:
        if self._client is not None:
            await self._client.close()
            self._client = None

    async def _ensure_class(self) -> None:
        assert self._client is not None
        name = self._settings.class_name
        if await self._client.collections.exists(name):
            return
        await self._client.collections.create(
            name=name,
            properties=content_chunk_properties(),
            vector_config=content_chunk_vector_config(),
        )
        logger.info("weaviate.class_created", name=name)

    async def upsert(self, chunks: list[Chunk], embeddings: list[Embedding]) -> None:
        if len(chunks) != len(embeddings):
            raise ValueError("chunks and embeddings length mismatch")
        if not chunks:
            return
        assert self._client is not None
        collection = self._client.collections.get(self._settings.class_name)
        first = chunks[0]
        async with observe(WEAVIATE_CALLS, op="upsert"):
            try:
                await collection.data.delete_many(
                    where=Filter.all_of(
                        [
                            Filter.by_property("sourceId").equal(first.source_id),
                            Filter.by_property("kind").equal(first.kind),
                        ]
                    )
                )
                objs = [
                    DataObject(
                        properties=_properties(chunk),
                        uuid=_chunk_uuid(chunk.kind, chunk.source_id, chunk.segment_idx),
                        vector=list(vector),
                    )
                    for chunk, vector in zip(chunks, embeddings, strict=True)
                ]
                result = await collection.data.insert_many(objs)
                if result.has_errors:
                    raise IndexingFailed(f"weaviate insert_many errors: {result.errors}")
            except WeaviateBaseError as exc:
                raise IndexingFailed(f"weaviate upsert failed: {exc}", cause=exc) from exc

    async def delete_by_source(self, source_id: str, kind: ContentKind) -> None:
        assert self._client is not None
        collection = self._client.collections.get(self._settings.class_name)
        async with observe(WEAVIATE_CALLS, op="delete_by_source"):
            try:
                await collection.data.delete_many(
                    where=Filter.all_of(
                        [
                            Filter.by_property("sourceId").equal(source_id),
                            Filter.by_property("kind").equal(kind),
                        ]
                    )
                )
            except WeaviateBaseError as exc:
                raise IndexingFailed(
                    f"weaviate delete_by_source failed: {exc}", cause=exc
                ) from exc

    async def search(
        self,
        query_embedding: Embedding,
        limit: int,
        *,
        kinds: list[ContentKind] | None = None,
    ) -> list[ScoredChunk]:
        return await self._search(query_embedding, limit=limit, kinds=kinds)

    async def search_dedup(
        self,
        query_embedding: Embedding,
        distance_threshold: float,
        limit: int,
    ) -> list[ScoredChunk]:
        return await self._search(
            query_embedding,
            limit=limit,
            kinds=["post"],
            distance=distance_threshold,
        )

    async def find_similar_posts(
        self,
        query_embedding: Embedding,
        *,
        exclude_post_id: str,
        top_n: int,
        oversample: int,
    ) -> list[SimilarPost]:
        assert self._client is not None
        if top_n <= 0:
            return []
        collection = self._client.collections.get(self._settings.class_name)
        filters = Filter.all_of(
            [
                Filter.by_property("kind").equal("post"),
                Filter.by_property("sourceId").not_equal(exclude_post_id),
            ]
        )
        candidate_limit = max(top_n * max(oversample, 1), top_n)
        async with observe(WEAVIATE_CALLS, op="find_similar_posts"):
            try:
                result = await collection.query.near_vector(
                    near_vector=list(query_embedding),
                    limit=candidate_limit,
                    filters=filters,
                    return_metadata=MetadataQuery(distance=True),
                )
            except WeaviateBaseError as exc:
                raise SimilaritySearchFailed(
                    f"weaviate find_similar_posts failed: {exc}", cause=exc
                ) from exc

        best: dict[str, float] = {}
        for obj in result.objects:
            props = obj.properties or {}
            source_id = str(props.get("sourceId", ""))
            if not source_id:
                continue
            distance = obj.metadata.distance if obj.metadata is not None else None
            if distance is None:
                continue
            score = 1.0 - float(distance)
            current = best.get(source_id)
            if current is None or score > current:
                best[source_id] = score
        ranked = sorted(best.items(), key=lambda kv: kv[1], reverse=True)
        return [SimilarPost(post_id=pid, score=score) for pid, score in ranked[:top_n]]

    async def get_post_embedding(self, post_id: str) -> Embedding | None:
        assert self._client is not None
        collection = self._client.collections.get(self._settings.class_name)
        filters = Filter.all_of(
            [
                Filter.by_property("kind").equal("post"),
                Filter.by_property("sourceId").equal(post_id),
            ]
        )
        async with observe(WEAVIATE_CALLS, op="get_post_embedding"):
            try:
                result = await collection.query.fetch_objects(
                    filters=filters,
                    include_vector=True,
                    limit=200,
                )
            except WeaviateBaseError as exc:
                raise SimilaritySearchFailed(
                    f"weaviate get_post_embedding failed: {exc}", cause=exc
                ) from exc

        vectors: list[list[float]] = []
        for obj in result.objects:
            vec = _extract_vector(obj.vector)
            if vec:
                vectors.append(vec)
        if not vectors:
            return None
        width = len(vectors[0])
        summed = [0.0] * width
        for vec in vectors:
            if len(vec) != width:
                continue
            for i, value in enumerate(vec):
                summed[i] += value
        n = float(len(vectors))
        return [v / n for v in summed]

    async def _search(
        self,
        query_embedding: Embedding,
        *,
        limit: int,
        kinds: list[ContentKind] | None = None,
        distance: float | None = None,
    ) -> list[ScoredChunk]:
        assert self._client is not None
        collection = self._client.collections.get(self._settings.class_name)
        filters = None
        if kinds:
            filters = Filter.by_property("kind").contains_any(list(kinds))
        op = "search_dedup" if distance is not None else "search"
        async with observe(WEAVIATE_CALLS, op=op):
            try:
                result = await collection.query.near_vector(
                    near_vector=list(query_embedding),
                    limit=limit,
                    distance=distance,
                    filters=filters,
                    return_metadata=MetadataQuery(distance=True),
                )
            except WeaviateBaseError as exc:
                raise SimilaritySearchFailed(
                    f"weaviate near_vector failed: {exc}", cause=exc
                ) from exc

        out: list[ScoredChunk] = []
        for obj in result.objects:
            props = obj.properties or {}
            meta_distance = obj.metadata.distance if obj.metadata is not None else None
            score = 1.0 - meta_distance if meta_distance is not None else 0.0
            chunk = Chunk(
                source_id=str(props.get("sourceId", "")),
                kind=str(props.get("kind", "post")),  # type: ignore[arg-type]
                text=str(props.get("text", "")),
                segment_idx=int(props.get("segmentIdx", 0) or 0),
                metadata=_metadata_from_props(props),
            )
            out.append(ScoredChunk(chunk=chunk, score=score))
        return out


def _extract_vector(stored: Any) -> list[float]:
    if not stored:
        return []
    if isinstance(stored, list):
        return [float(v) for v in stored]
    if isinstance(stored, dict):
        for value in stored.values():
            if isinstance(value, list) and value and isinstance(value[0], (int, float)):
                return [float(v) for v in value]
    return []


def _metadata_from_props(props: dict[str, Any]) -> dict[str, str]:
    meta: dict[str, str] = {}
    for src, dst in (
        ("category", "category"),
        ("parentId", "parent_id"),
        ("attachedPostId", "attached_post_id"),
    ):
        value = props.get(src)
        if value:
            meta[dst] = str(value)
    tags = props.get("tags") or []
    if tags:
        meta["tags"] = ",".join(str(t) for t in tags)
    coord = props.get("coordinate")
    if isinstance(coord, dict):
        if "latitude" in coord:
            meta["lat"] = str(coord["latitude"])
        if "longitude" in coord:
            meta["lon"] = str(coord["longitude"])
    return meta
