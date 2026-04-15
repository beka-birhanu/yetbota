from __future__ import annotations

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from infra.config import get_settings
from interfaces.rest.v1.health import router as health_router
from interfaces.rest.v1.chat import router as chat_router
from interfaces.rest.v1.version import router as version_router


def create_app() -> FastAPI:
    settings = get_settings()

    app = FastAPI(
        title=settings.app_name,
        version=settings.app_version,
    )
    app.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )
    app.include_router(health_router)
    app.include_router(version_router)
    app.include_router(chat_router)
    return app

