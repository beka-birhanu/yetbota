from __future__ import annotations

import json
from typing import Any

from pydantic import BaseModel, Field, ValidationError


class ParsedCitation(BaseModel):
    source_id: str = Field(min_length=1)
    text: str = Field(min_length=1)
    score: float


class ParsedAnswer(BaseModel):
    answer: str = Field(min_length=1)
    citations: list[ParsedCitation] = Field(default_factory=list)


def parse_llm_json(text: str) -> ParsedAnswer | None:
    cleaned = text.strip()
    if not cleaned:
        return None
    try:
        data: Any = json.loads(cleaned)
    except json.JSONDecodeError:
        return None
    try:
        return ParsedAnswer.model_validate(data)
    except ValidationError:
        return None
