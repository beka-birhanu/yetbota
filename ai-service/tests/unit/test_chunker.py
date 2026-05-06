import pytest

from infrastructure.chunking import RuneWindowChunker
from infrastructure.config.settings import ChunkerSettings


def _chunker(size: int = 10, overlap: int = 2) -> RuneWindowChunker:
    return RuneWindowChunker(ChunkerSettings(size=size, overlap=overlap))


def test_short_text_yields_single_chunk() -> None:
    chunks = _chunker().chunk(source_id="p1", kind="post", text="hello")
    assert len(chunks) == 1
    assert chunks[0].text == "hello"
    assert chunks[0].segment_idx == 0
    assert chunks[0].source_id == "p1"


def test_exact_size_yields_single_chunk() -> None:
    text = "a" * 10
    chunks = _chunker(size=10, overlap=2).chunk(source_id="p1", kind="post", text=text)
    assert len(chunks) == 1
    assert chunks[0].text == text


def test_multi_window_with_overlap() -> None:
    text = "0123456789ABCDEFGHIJ"
    chunks = _chunker(size=10, overlap=3).chunk(source_id="p1", kind="post", text=text)
    assert [c.text for c in chunks] == ["0123456789", "789ABCDEFG", "EFGHIJ"]
    assert [c.segment_idx for c in chunks] == [0, 1, 2]


def test_empty_after_normalization_returns_no_chunks() -> None:
    chunks = _chunker().chunk(source_id="p1", kind="post", text="   \n\t   ")
    assert chunks == []


def test_normalization_collapses_whitespace() -> None:
    chunks = _chunker(size=100).chunk(source_id="p1", kind="post", text="hello   world\n\nfoo\tbar")
    assert chunks[0].text == "hello world foo bar"


def test_metadata_passes_through() -> None:
    chunks = _chunker(size=100).chunk(
        source_id="p1",
        kind="question",
        text="hi",
        metadata={"category": "cafe", "name": "Joe's"},
    )
    assert chunks[0].metadata == {"category": "cafe", "name": "Joe's"}
    assert chunks[0].kind == "question"


def test_unicode_preserved() -> None:
    text = "café 🚀 naïve résumé"
    chunks = _chunker(size=100).chunk(source_id="p1", kind="post", text=text)
    assert chunks[0].text == text


def test_invalid_overlap_raises() -> None:
    with pytest.raises(ValueError):
        RuneWindowChunker(ChunkerSettings(size=10, overlap=10))
    with pytest.raises(ValueError):
        RuneWindowChunker(ChunkerSettings(size=10, overlap=-1))


def test_invalid_size_raises() -> None:
    with pytest.raises(ValueError):
        RuneWindowChunker(ChunkerSettings(size=0, overlap=0))
