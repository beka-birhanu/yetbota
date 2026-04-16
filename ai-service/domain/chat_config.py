from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True, slots=True)
class ChatConfig:
    rag_top_k: int
    rag_min_score: float | None
