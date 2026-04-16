from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True, slots=True)
class RetrievedChunk:
    source_id: str
    text: str
    score: float


@dataclass(frozen=True, slots=True)
class Citation:
    source_id: str
    text: str
    score: float


@dataclass(frozen=True, slots=True)
class ChatResult:
    success: bool
    message: str
    answer: str
    citations: list[Citation]
