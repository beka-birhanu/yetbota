from __future__ import annotations

import weaviate
from weaviate.classes.init import Auth
from weaviate.classes.query import Filter, MetadataQuery

from domain.models import RetrievedChunk
from application.interfaces import Retriever


class WeaviateRetriever(Retriever):
    def __init__(
        self,
        *,
        host: str,
        http_port: int,
        grpc_port: int,
        secure: bool,
        api_key: str | None,
        collection: str,
        vector_name: str,
        verified_property: str,
    ):
        http_host = host
        http_secure = bool(secure)
        grpc_host_resolved = host
        grpc_port_resolved = int(grpc_port)
        grpc_secure_resolved = bool(secure)

        auth = Auth.api_key(api_key) if api_key else None
        self._client = weaviate.connect_to_custom(
            http_host=http_host,
            http_port=http_port,
            http_secure=http_secure,
            grpc_host=grpc_host_resolved,
            grpc_port=grpc_port_resolved,
            grpc_secure=grpc_secure_resolved,
            auth_credentials=auth,
        )
        self._collection = self._client.collections.get(collection)
        self._vector_name = vector_name
        self._verified_property = verified_property

    def close(self) -> None:
        self._client.close()

    def retrieve_verified(self, query_vector: list[float], limit: int) -> list[RetrievedChunk]:
        flt = Filter.by_property(self._verified_property).equal(True)
        kwargs: dict[str, object] = {
            "near_vector": query_vector,
            "limit": limit,
            "filters": flt,
            "return_metadata": MetadataQuery(distance=True),
        }
        if self._vector_name and self._vector_name != "default":
            kwargs["target_vector"] = self._vector_name
        response = self._collection.query.near_vector(**kwargs)
        chunks: list[RetrievedChunk] = []
        for obj in response.objects:
            props = obj.properties or {}
            source_id = str(props.get("source_id", "")).strip()
            text = str(props.get("text", "")).strip()
            if not source_id or not text:
                continue
            distance = getattr(obj.metadata, "distance", None)
            score = 0.0
            if isinstance(distance, (int, float)):
                score = float(1.0 / (1.0 + float(distance)))
            chunks.append(RetrievedChunk(source_id=source_id, text=text, score=score))
        return chunks
