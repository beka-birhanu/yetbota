from __future__ import annotations

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
    weaviate_url: str | None = None
    weaviate_api_key: str | None = None


def get_settings() -> Settings:
    return Settings()

