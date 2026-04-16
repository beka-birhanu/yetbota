from __future__ import annotations

from typing import Protocol

from domain.models import RetrievedChunk


class Embedder(Protocol):
    def embed(self, text: str) -> list[float]:
        raise NotImplementedError


class Retriever(Protocol):
    def retrieve_verified(self, query_vector: list[float], limit: int) -> list["RetrievedChunk"]:
        raise NotImplementedError


class Llm(Protocol):
    def generate_json_answer(self, prompt: str) -> str:
        raise NotImplementedError

