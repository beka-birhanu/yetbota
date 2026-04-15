from __future__ import annotations

from fastapi import APIRouter

from infra.config import get_settings


router = APIRouter(tags=["meta"])


@router.get("/version")
def version() -> dict[str, object]:
    s = get_settings()
    return {"name": s.app_name, "version": s.app_version}

