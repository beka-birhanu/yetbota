from types import SimpleNamespace
from unittest.mock import AsyncMock, MagicMock

import pytest
from weaviate.exceptions import WeaviateBaseError

from application.errors import IndexingFailed, SimilaritySearchFailed
from domain.entities import Chunk
from infrastructure.config.settings import WeaviateSettings
from infrastructure.vector.weaviate_store import (
    WeaviateVectorStore,
    _chunk_uuid,
    _metadata_from_props,
    _properties,
)


def _settings() -> WeaviateSettings:
    return WeaviateSettings(
        url="http://localhost:8081",
        grpc_port=50051,
        api_key="",
        class_name="ContentChunk",
    )


def _store_with_collection(collection: MagicMock) -> WeaviateVectorStore:
    store = WeaviateVectorStore(_settings())
    client = MagicMock()
    client.collections.get = MagicMock(return_value=collection)
    store._client = client
    return store


def test_chunk_uuid_is_deterministic_per_input() -> None:
    a = _chunk_uuid("post", "abc", 0)
    b = _chunk_uuid("post", "abc", 0)
    c = _chunk_uuid("post", "abc", 1)
    d = _chunk_uuid("question", "abc", 0)
    assert a == b
    assert a != c
    assert a != d


def test_properties_maps_metadata_fields() -> None:
    chunk = Chunk(
        source_id="p1",
        kind="post",
        text="hello",
        metadata={
            "category": "cafe",
            "tags": "a, b , c",
            "lat": "10.5",
            "lon": "-20.25",
        },
        segment_idx=2,
    )
    props = _properties(chunk)
    assert props["sourceId"] == "p1"
    assert props["kind"] == "post"
    assert props["segmentIdx"] == 2
    assert props["category"] == "cafe"
    assert props["tags"] == ["a", "b", "c"]
    assert props["coordinate"] == {"latitude": 10.5, "longitude": -20.25}


def test_properties_omits_coordinate_when_missing() -> None:
    chunk = Chunk(source_id="p1", kind="post", text="hi")
    props = _properties(chunk)
    assert "coordinate" not in props


def test_metadata_from_props_round_trip() -> None:
    meta = _metadata_from_props(
        {
            "category": "cafe",
            "parentId": "q1",
            "tags": ["a", "b"],
            "coordinate": {"latitude": 1.5, "longitude": 2.5},
        }
    )
    assert meta["category"] == "cafe"
    assert meta["parent_id"] == "q1"
    assert meta["tags"] == "a,b"
    assert meta["lat"] == "1.5"
    assert meta["lon"] == "2.5"


@pytest.mark.asyncio
async def test_upsert_length_mismatch_raises() -> None:
    store = WeaviateVectorStore(_settings())
    with pytest.raises(ValueError):
        await store.upsert(
            [Chunk(source_id="p1", kind="post", text="x")],
            [[0.1] * 1536, [0.2] * 1536],
        )


@pytest.mark.asyncio
async def test_upsert_empty_no_op() -> None:
    store = WeaviateVectorStore(_settings())
    await store.upsert([], [])


@pytest.mark.asyncio
async def test_upsert_calls_delete_then_insert() -> None:
    collection = MagicMock()
    collection.data.delete_many = AsyncMock()
    collection.data.insert_many = AsyncMock(
        return_value=SimpleNamespace(has_errors=False, errors={})
    )
    store = _store_with_collection(collection)
    chunks = [
        Chunk(source_id="p1", kind="post", text="a", segment_idx=0),
        Chunk(source_id="p1", kind="post", text="b", segment_idx=1),
    ]
    await store.upsert(chunks, [[0.1] * 4, [0.2] * 4])
    collection.data.delete_many.assert_awaited_once()
    collection.data.insert_many.assert_awaited_once()
    args, _ = collection.data.insert_many.call_args
    objs = args[0]
    assert len(objs) == 2
    assert objs[0].uuid == _chunk_uuid("post", "p1", 0)
    assert objs[1].uuid == _chunk_uuid("post", "p1", 1)


@pytest.mark.asyncio
async def test_delete_by_source_calls_delete_many() -> None:
    collection = MagicMock()
    collection.data.delete_many = AsyncMock()
    store = _store_with_collection(collection)
    await store.delete_by_source("p1", "post")
    collection.data.delete_many.assert_awaited_once()


@pytest.mark.asyncio
async def test_delete_by_source_wraps_weaviate_errors() -> None:
    collection = MagicMock()
    collection.data.delete_many = AsyncMock(side_effect=WeaviateBaseError("boom"))
    store = _store_with_collection(collection)
    with pytest.raises(IndexingFailed):
        await store.delete_by_source("p1", "post")


@pytest.mark.asyncio
async def test_upsert_raises_indexing_failed_on_batch_error() -> None:
    collection = MagicMock()
    collection.data.delete_many = AsyncMock()
    collection.data.insert_many = AsyncMock(
        return_value=SimpleNamespace(has_errors=True, errors={"0": "boom"})
    )
    store = _store_with_collection(collection)
    with pytest.raises(IndexingFailed):
        await store.upsert([Chunk(source_id="p1", kind="post", text="a")], [[0.1] * 4])


@pytest.mark.asyncio
async def test_search_returns_scored_chunks() -> None:
    collection = MagicMock()
    obj = SimpleNamespace(
        properties={
            "sourceId": "p1",
            "kind": "post",
            "text": "hello",
            "segmentIdx": 0,
            "category": "cafe",
            "tags": [],
        },
        metadata=SimpleNamespace(distance=0.1),
    )
    collection.query.near_vector = AsyncMock(return_value=SimpleNamespace(objects=[obj]))
    store = _store_with_collection(collection)
    out = await store.search([0.0] * 4, limit=5)
    assert len(out) == 1
    assert out[0].chunk.source_id == "p1"
    assert out[0].chunk.kind == "post"
    assert out[0].score == pytest.approx(0.9)


@pytest.mark.asyncio
async def test_search_dedup_filters_to_post_with_distance() -> None:
    collection = MagicMock()
    collection.query.near_vector = AsyncMock(return_value=SimpleNamespace(objects=[]))
    store = _store_with_collection(collection)
    await store.search_dedup([0.0] * 4, distance_threshold=0.15, limit=1)
    kwargs = collection.query.near_vector.call_args.kwargs
    assert kwargs["distance"] == 0.15
    assert kwargs["limit"] == 1
    assert kwargs["filters"] is not None


@pytest.mark.asyncio
async def test_search_wraps_weaviate_errors() -> None:
    collection = MagicMock()
    collection.query.near_vector = AsyncMock(side_effect=WeaviateBaseError("nope"))
    store = _store_with_collection(collection)
    with pytest.raises(SimilaritySearchFailed):
        await store.search([0.0] * 4, limit=5)


def _post_chunk_obj(source_id: str, distance: float, segment_idx: int = 0) -> SimpleNamespace:
    return SimpleNamespace(
        properties={
            "sourceId": source_id,
            "kind": "post",
            "text": f"text-{source_id}-{segment_idx}",
            "segmentIdx": segment_idx,
        },
        metadata=SimpleNamespace(distance=distance),
    )


@pytest.mark.asyncio
async def test_find_similar_posts_groups_chunks_by_source_keeping_max_score() -> None:
    collection = MagicMock()
    objs = [
        _post_chunk_obj("q1", 0.20, 0),
        _post_chunk_obj("q1", 0.10, 1),
        _post_chunk_obj("q2", 0.40, 0),
        _post_chunk_obj("q3", 0.05, 0),
    ]
    collection.query.near_vector = AsyncMock(return_value=SimpleNamespace(objects=objs))
    store = _store_with_collection(collection)
    similar = await store.find_similar_posts([0.0] * 4, exclude_post_id="p0", top_n=2, oversample=5)
    assert [s.post_id for s in similar] == ["q3", "q1"]
    assert similar[0].score == pytest.approx(0.95)
    assert similar[1].score == pytest.approx(0.90)


@pytest.mark.asyncio
async def test_find_similar_posts_filters_to_kind_post_excluding_self() -> None:
    collection = MagicMock()
    collection.query.near_vector = AsyncMock(return_value=SimpleNamespace(objects=[]))
    store = _store_with_collection(collection)
    await store.find_similar_posts([0.0] * 4, exclude_post_id="p0", top_n=5, oversample=3)
    kwargs = collection.query.near_vector.call_args.kwargs
    assert kwargs["filters"] is not None
    assert kwargs["limit"] == 15


@pytest.mark.asyncio
async def test_find_similar_posts_returns_empty_when_top_n_zero() -> None:
    collection = MagicMock()
    collection.query.near_vector = AsyncMock()
    store = _store_with_collection(collection)
    out = await store.find_similar_posts([0.0] * 4, exclude_post_id="p0", top_n=0, oversample=5)
    assert out == []
    collection.query.near_vector.assert_not_awaited()


@pytest.mark.asyncio
async def test_find_similar_posts_wraps_weaviate_errors() -> None:
    collection = MagicMock()
    collection.query.near_vector = AsyncMock(side_effect=WeaviateBaseError("boom"))
    store = _store_with_collection(collection)
    with pytest.raises(SimilaritySearchFailed):
        await store.find_similar_posts([0.0] * 4, exclude_post_id="p0", top_n=2, oversample=5)


@pytest.mark.asyncio
async def test_get_post_embedding_averages_chunk_vectors() -> None:
    collection = MagicMock()
    objs = [
        SimpleNamespace(properties={}, metadata=None, vector={"default": [0.0, 1.0]}),
        SimpleNamespace(properties={}, metadata=None, vector={"default": [1.0, 0.0]}),
    ]
    collection.query.fetch_objects = AsyncMock(return_value=SimpleNamespace(objects=objs))
    store = _store_with_collection(collection)
    vec = await store.get_post_embedding("p1")
    assert vec == [0.5, 0.5]


@pytest.mark.asyncio
async def test_get_post_embedding_handles_plain_list_vector() -> None:
    collection = MagicMock()
    objs = [
        SimpleNamespace(properties={}, metadata=None, vector=[0.2, 0.4]),
    ]
    collection.query.fetch_objects = AsyncMock(return_value=SimpleNamespace(objects=objs))
    store = _store_with_collection(collection)
    vec = await store.get_post_embedding("p1")
    assert vec == [0.2, 0.4]


@pytest.mark.asyncio
async def test_get_post_embedding_returns_none_when_no_chunks() -> None:
    collection = MagicMock()
    collection.query.fetch_objects = AsyncMock(return_value=SimpleNamespace(objects=[]))
    store = _store_with_collection(collection)
    assert await store.get_post_embedding("p1") is None


@pytest.mark.asyncio
async def test_get_post_embedding_wraps_weaviate_errors() -> None:
    collection = MagicMock()
    collection.query.fetch_objects = AsyncMock(side_effect=WeaviateBaseError("nope"))
    store = _store_with_collection(collection)
    with pytest.raises(SimilaritySearchFailed):
        await store.get_post_embedding("p1")
