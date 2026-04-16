from __future__ import annotations

from application.errors import InvalidQuery, Misconfigured
from application.interfaces import Embedder, Llm, Retriever
from domain.chat_config import ChatConfig
from domain.models import Citation, ChatResult, RetrievedChunk
from application.rag.build_prompt import build_fallback_answer, build_prompt
from application.rag.parse_answer_json import parse_llm_json


class ChatService:
    def __init__(
        self,
        *,
        gemini_api_key: str | None,
        weaviate_configured: bool,
        config: ChatConfig,
        embedder: Embedder,
        retriever: Retriever,
        llm: Llm,
    ):
        self._gemini_api_key = gemini_api_key
        self._weaviate_configured = weaviate_configured
        self._config = config
        self._embedder = embedder
        self._retriever = retriever
        self._llm = llm

    def chat(self, query: str) -> ChatResult:
        normalized = " ".join(query.split()).strip()
        if not normalized:
            raise InvalidQuery("query is required")

        if not self._gemini_api_key:
            raise Misconfigured("GEMINI_API_KEY is not configured")

        if not self._weaviate_configured:
            raise Misconfigured("Weaviate is not configured")

        vector = self._embedder.embed(normalized)
        chunks = self._retriever.retrieve_verified(vector, self._config.rag_top_k)
        chunks = self._apply_min_score(chunks)

        if not chunks:
            return ChatResult(
                success=True,
                message="ok",
                answer="I do not have verified information on that topic in YetBota.",
                citations=[],
            )

        prompt = build_prompt(normalized, chunks)
        raw = self._llm.generate_json_answer(prompt)
        parsed = parse_llm_json(raw)
        if parsed is None:
            answer, citation_dicts = build_fallback_answer(chunks)
            citations = [
                Citation(
                    source_id=str(c["source_id"]),
                    text=str(c["text"]),
                    score=float(c["score"]),
                )
                for c in citation_dicts
            ]
            return ChatResult(success=True, message="ok", answer=answer, citations=citations)

        citations_out = [
            Citation(source_id=c.source_id, text=c.text, score=float(c.score)) for c in parsed.citations
        ]
        return ChatResult(
            success=True,
            message="ok",
            answer=parsed.answer,
            citations=citations_out,
        )

    def _apply_min_score(self, chunks: list[RetrievedChunk]) -> list[RetrievedChunk]:
        if self._config.rag_min_score is None:
            return chunks
        return [c for c in chunks if c.score >= float(self._config.rag_min_score)]
