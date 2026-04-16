from __future__ import annotations

import google.generativeai as genai

from application.interfaces import Embedder


class GeminiEmbedder(Embedder):
    def __init__(self, *, api_key: str, embedding_model: str):
        genai.configure(api_key=api_key)
        self._model = embedding_model if embedding_model.startswith("models/") else f"models/{embedding_model}"

    def embed(self, text: str) -> list[float]:
        result = genai.embed_content(model=self._model, content=text)
        embedding = result.get("embedding")
        if not isinstance(embedding, list):
            raise RuntimeError("unexpected embedding response")
        return [float(x) for x in embedding]
