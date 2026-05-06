from unittest.mock import AsyncMock, MagicMock, patch

import pytest
from google.genai import errors, types

from application.errors import EmbeddingFailed
from infrastructure.config.settings import GeminiSettings
from infrastructure.embedding import GeminiEmbedder


def _make_settings(dims: int = 1536) -> GeminiSettings:
    return GeminiSettings(api_key="test", embedding_dimensions=dims)


def _response(vectors: list[list[float]]) -> types.EmbedContentResponse:
    return types.EmbedContentResponse(
        embeddings=[types.ContentEmbedding(values=v) for v in vectors]
    )


@pytest.fixture
def patch_client() -> AsyncMock:
    with patch("infrastructure.embedding.gemini_embedder.genai.Client") as client_cls:
        instance = MagicMock()
        instance.aio.models.embed_content = AsyncMock()
        client_cls.return_value = instance
        yield instance.aio.models.embed_content


@pytest.mark.asyncio
async def test_embed_returns_single_vector(patch_client: AsyncMock) -> None:
    patch_client.return_value = _response([[0.1] * 1536])
    embedder = GeminiEmbedder(_make_settings())
    vec = await embedder.embed("hello")
    assert len(vec) == 1536
    patch_client.assert_awaited_once()
    call = patch_client.await_args
    assert call.kwargs["contents"] == ["hello"]
    assert call.kwargs["config"].task_type == "RETRIEVAL_DOCUMENT"
    assert call.kwargs["config"].output_dimensionality == 1536


@pytest.mark.asyncio
async def test_embed_query_uses_query_task_type(patch_client: AsyncMock) -> None:
    patch_client.return_value = _response([[0.2] * 1536])
    embedder = GeminiEmbedder(_make_settings())
    await embedder.embed("hello", task_type="RETRIEVAL_QUERY")
    assert patch_client.await_args.kwargs["config"].task_type == "RETRIEVAL_QUERY"


@pytest.mark.asyncio
async def test_embed_batch_shards_above_limit(patch_client: AsyncMock) -> None:
    patch_client.side_effect = [
        _response([[0.1] * 1536 for _ in range(96)]),
        _response([[0.2] * 1536 for _ in range(4)]),
    ]
    embedder = GeminiEmbedder(_make_settings())
    out = await embedder.embed_batch(["t"] * 100)
    assert len(out) == 100
    assert patch_client.await_count == 2


@pytest.mark.asyncio
async def test_embed_batch_empty_short_circuits(patch_client: AsyncMock) -> None:
    embedder = GeminiEmbedder(_make_settings())
    assert await embedder.embed_batch([]) == []
    patch_client.assert_not_awaited()


@pytest.mark.asyncio
async def test_dimension_mismatch_raises(patch_client: AsyncMock) -> None:
    patch_client.return_value = _response([[0.1] * 768])
    embedder = GeminiEmbedder(_make_settings(dims=1536))
    with pytest.raises(EmbeddingFailed):
        await embedder.embed("x")


@pytest.mark.asyncio
async def test_count_mismatch_raises(patch_client: AsyncMock) -> None:
    patch_client.return_value = _response([[0.1] * 1536])
    embedder = GeminiEmbedder(_make_settings())
    with pytest.raises(EmbeddingFailed):
        await embedder.embed_batch(["a", "b"])


@pytest.mark.asyncio
async def test_429_then_success_retries(patch_client: AsyncMock) -> None:
    transient = errors.ClientError(429, {"error": {"message": "rate limited"}}, None)
    patch_client.side_effect = [transient, _response([[0.3] * 1536])]
    embedder = GeminiEmbedder(_make_settings())
    vec = await embedder.embed("x")
    assert len(vec) == 1536
    assert patch_client.await_count == 2


@pytest.mark.asyncio
async def test_400_does_not_retry(patch_client: AsyncMock) -> None:
    patch_client.side_effect = errors.ClientError(400, {"error": {"message": "bad input"}}, None)
    embedder = GeminiEmbedder(_make_settings())
    with pytest.raises(EmbeddingFailed):
        await embedder.embed("x")
    assert patch_client.await_count == 1
