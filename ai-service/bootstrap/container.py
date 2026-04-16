from __future__ import annotations

from dataclasses import dataclass
from typing import Callable

from application.chat_service import ChatService
from application.interfaces import Embedder, Llm, Retriever
from domain.chat_config import ChatConfig
from infrastructure.llm.embedder import GeminiEmbedder
from infrastructure.llm.generator import GeminiLlm
from infrastructure.settings import Settings, get_settings
from infrastructure.vector_store.retriever import WeaviateRetriever


@dataclass(slots=True)
class AppContainer:
    settings: Settings
    chat_service: ChatService
    close: Callable[[], None]


def build_container() -> AppContainer:
    settings = get_settings()
    chat_config = ChatConfig(rag_top_k=settings.rag_top_k, rag_min_score=settings.rag_min_score)

    class _MissingEmbedder:
        def embed(self, text: str) -> list[float]:
            return []

    class _MissingLlm:
        def generate_json_answer(self, prompt: str) -> str:
            return ""

    embedder: Embedder
    llm: Llm
    if settings.gemini_api_key:
        embedder = GeminiEmbedder(api_key=settings.gemini_api_key, embedding_model=settings.gemini_embedding_model)
        llm = GeminiLlm(api_key=settings.gemini_api_key, model=settings.gemini_model)
    else:
        embedder = _MissingEmbedder()
        llm = _MissingLlm()

    retriever: WeaviateRetriever | None = None
    if settings.weaviate_host and settings.weaviate_collection:
        retriever = WeaviateRetriever(
            host=settings.weaviate_host,
            http_port=settings.weaviate_http_port,
            grpc_port=settings.weaviate_grpc_port,
            secure=settings.weaviate_secure,
            api_key=settings.weaviate_api_key,
            collection=settings.weaviate_collection,
            vector_name=settings.weaviate_vector_name,
            verified_property=settings.weaviate_verified_property,
        )

    class _EmptyRetriever:
        def retrieve_verified(self, query_vector: list[float], limit: int):
            return []

    chat_service = ChatService(
        gemini_api_key=settings.gemini_api_key,
        weaviate_configured=bool(settings.weaviate_host and settings.weaviate_collection),
        config=chat_config,
        embedder=embedder,
        retriever=retriever or _EmptyRetriever(),
        llm=llm,
    )

    def close() -> None:
        if retriever is not None:
            retriever.close()

    return AppContainer(settings=settings, chat_service=chat_service, close=close)
