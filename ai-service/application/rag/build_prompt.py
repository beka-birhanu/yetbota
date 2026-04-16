from __future__ import annotations

import json

from domain.models import RetrievedChunk


def build_prompt(query: str, chunks: list[RetrievedChunk]) -> str:
    blocks: list[str] = []
    for idx, chunk in enumerate(chunks, start=1):
        blocks.append(
            f"[{idx}] source_id={chunk.source_id}\nscore={chunk.score}\n{chunk.text}".strip()
        )
    context = "\n\n".join(blocks)
    schema = {
        "answer": "string",
        "citations": [{"source_id": "string", "text": "string", "score": "number"}],
    }
    return (
        "You are YetBota's assistant. Answer ONLY using the numbered context blocks.\n"
        "If the context does not contain enough verified information, set answer to exactly:\n"
        "\"I do not have verified information on that topic in YetBota.\"\n"
        "Return ONLY valid JSON matching this schema (no markdown, no prose):\n"
        f"{json.dumps(schema, ensure_ascii=False)}\n\n"
        f"user_query: {query}\n\n"
        f"context_blocks:\n{context}"
    )


def build_fallback_answer(chunks: list[RetrievedChunk]) -> tuple[str, list[dict[str, object]]]:
    if not chunks:
        return (
            "I do not have verified information on that topic in YetBota.",
            [],
        )
    top = chunks[0]
    citations = [
        {
            "source_id": top.source_id,
            "text": top.text,
            "score": float(top.score),
        }
    ]
    return (
        "I can only confirm what appears in the retrieved verified context; see citation.",
        citations,
    )
