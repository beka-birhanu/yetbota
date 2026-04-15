from __future__ import annotations

from pydantic import BaseModel, Field


class ChatRequest(BaseModel):
    query: str = Field(min_length=1)


class Citation(BaseModel):
    source_id: str
    text: str
    score: float


class ChatResponse(BaseModel):
    code: str
    success: bool
    message: str
    answer: str
    citations: list[Citation]

