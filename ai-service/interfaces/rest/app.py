from __future__ import annotations

from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from bootstrap.container import build_container
from infrastructure.settings import get_settings
from interfaces.rest.chat_router import router as chat_router
from interfaces.rest.health_router import router as health_router
from interfaces.rest.version_router import router as version_router


def create_app() -> FastAPI:
    settings = get_settings()

    @asynccontextmanager
    async def lifespan(app: FastAPI):
        container = build_container()
        app.state.container = container
        try:
            yield
        finally:
            container.close()

    app = FastAPI(
        title=settings.app_name,
        version=settings.app_version,
        lifespan=lifespan,
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

