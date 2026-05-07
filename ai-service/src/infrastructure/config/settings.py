from functools import lru_cache
from typing import Literal

from pydantic import BaseModel, Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class AppSettings(BaseModel):
    name: str = "ai-service"
    version: str = "0.0.1"
    debug: bool = False


class HttpSettings(BaseModel):
    host: str = "0.0.0.0"
    port: int = 8080


class GrpcSettings(BaseModel):
    host: str = "0.0.0.0"
    port: int = 9090
    max_concurrent_streams: int = 100
    keepalive_time_s: int = 60
    enable_reflection: bool = True


class RedisSettings(BaseModel):
    url: str = "redis://localhost:6379/0"
    ingest_stream: str = "ai.ingest"
    consumer_group: str = "ai-service"
    result_stream: str = "ai.results"
    result_routing_key: str = "content.processed"
    prefetch: int = 16
    block_ms: int = 5000
    claim_idle_ms: int = 30000
    result_maxlen: int = 10000
    max_delivery_attempts: int = 5


class WeaviateSettings(BaseModel):
    url: str = "http://localhost:8081"
    grpc_host: str = ""
    grpc_port: int = 50051
    api_key: str = ""
    class_name: str = "ContentChunk"


class Neo4jSettings(BaseModel):
    uri: str = "bolt://localhost:7687"
    username: str = "neo4j"
    password: str = ""
    database: str = "neo4j"


class SimilaritySettings(BaseModel):
    enabled: bool = True
    top_n: int = 10
    candidate_oversample: int = 5


class GeminiSettings(BaseModel):
    api_key: str = ""
    llm_model: str = "gemini-2.5-flash"
    embedding_model: str = "gemini-embedding-001"
    embedding_dimensions: int = 1536
    timeout_s: int = 30
    max_tokens: int = 1024
    temperature: float = 0.2


class DedupSettings(BaseModel):
    distance_threshold: float = 0.15


class RagSettings(BaseModel):
    top_k: int = 5
    min_similarity: float = 0.6


class ScreeningSettings(BaseModel):
    block_threshold: float = 0.7
    timeout_s: int = 5
    cache_size: int = 1024
    cache_ttl_s: int = 300


class ChunkerSettings(BaseModel):
    size: int = 512
    overlap: int = 64


class LoggingSettings(BaseModel):
    level: Literal["DEBUG", "INFO", "WARNING", "ERROR"] = "INFO"
    format: Literal["json", "console"] = "json"


class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        env_nested_delimiter="__",
        extra="ignore",
    )

    app: AppSettings = Field(default_factory=AppSettings)
    http: HttpSettings = Field(default_factory=HttpSettings)
    grpc: GrpcSettings = Field(default_factory=GrpcSettings)
    redis: RedisSettings = Field(default_factory=RedisSettings)
    weaviate: WeaviateSettings = Field(default_factory=WeaviateSettings)
    neo4j: Neo4jSettings = Field(default_factory=Neo4jSettings)
    gemini: GeminiSettings = Field(default_factory=GeminiSettings)
    dedup: DedupSettings = Field(default_factory=DedupSettings)
    rag: RagSettings = Field(default_factory=RagSettings)
    similarity: SimilaritySettings = Field(default_factory=SimilaritySettings)
    screening: ScreeningSettings = Field(default_factory=ScreeningSettings)
    chunker: ChunkerSettings = Field(default_factory=ChunkerSettings)
    logging: LoggingSettings = Field(default_factory=LoggingSettings)


@lru_cache(maxsize=1)
def get_settings() -> Settings:
    return Settings()
