from __future__ import annotations

from typing import Any

from fastapi import APIRouter, status

from interfaces.rest.schemas import ChatRequest, ChatResponse


router = APIRouter(prefix="/v1", tags=["v1"])


@router.post("/chat", status_code=status.HTTP_501_NOT_IMPLEMENTED, response_model=ChatResponse)
def chat(_: ChatRequest) -> Any:
    return {
        "code": "99",
        "success": False,
        "message": "not implemented (phase 2+)",
        "answer": "",
        "citations": [],
    }

