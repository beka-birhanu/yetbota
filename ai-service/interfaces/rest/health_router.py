from __future__ import annotations

from fastapi import APIRouter


router = APIRouter(tags=["health"])


@router.get("/healthz")
def healthz() -> str:
    return "ok"


@router.get("/readyz")
def readyz() -> str:
    return "ready"

