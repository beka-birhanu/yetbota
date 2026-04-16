from __future__ import annotations

from urllib.parse import urlparse

from pydantic import model_validator
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", extra="ignore")

    app_name: str = "yetbota-ai-service"
    app_version: str = "0.1.0"

    rest_host: str = "0.0.0.0"
    rest_port: int = 8080

    grpc_host: str = "0.0.0.0"
    grpc_port: int = 9090

    cors_hosts: str = "localhost,127.0.0.1"

    gemini_api_key: str | None = None
    gemini_model: str = "gemini-2.0-flash"
    gemini_embedding_model: str = "models/text-embedding-004"

    # If provided, this will be parsed into host/ports/secure defaults.
    # Example: http://localhost:8080
    weaviate_url: str | None = None
    weaviate_host: str | None = None
    weaviate_http_port: int = 8080
    weaviate_grpc_port: int = 50051
    weaviate_secure: bool = False
    weaviate_api_key: str | None = None
    weaviate_collection: str | None = None
    weaviate_vector_name: str = "default"
    weaviate_verified_property: str = "verified"

    rag_top_k: int = 8
    rag_min_score: float | None = None

    @model_validator(mode="after")
    def _parse_weaviate_url(self) -> Settings:
        # Back-compat with `.env.example` which uses WEAVIATE_URL.
        if not self.weaviate_url or self.weaviate_host:
            return self

        parsed = urlparse(self.weaviate_url)
        if parsed.hostname:
            self.weaviate_host = parsed.hostname
        if parsed.port:
            self.weaviate_http_port = int(parsed.port)
        if parsed.scheme:
            self.weaviate_secure = parsed.scheme.lower() == "https"
        return self


def get_settings() -> Settings:
    return Settings()

