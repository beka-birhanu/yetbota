from __future__ import annotations

from fastapi import APIRouter, HTTPException, Request, status

from application.errors import InvalidQuery, Misconfigured, UpstreamUnavailable
from interfaces.rest.schemas import ChatRequest, ChatResponse


router = APIRouter(prefix="/v1", tags=["v1"])


@router.post("/chat", response_model=ChatResponse)
def chat(request: Request, body: ChatRequest) -> ChatResponse:
    try:
        result = request.app.state.container.chat_service.chat(body.query)
    except InvalidQuery as e:
        raise HTTPException(status_code=status.HTTP_422_UNPROCESSABLE_ENTITY, detail=str(e))
    except Misconfigured as e:
        raise HTTPException(status_code=status.HTTP_503_SERVICE_UNAVAILABLE, detail=str(e))
    except UpstreamUnavailable as e:
        raise HTTPException(status_code=status.HTTP_503_SERVICE_UNAVAILABLE, detail=str(e))

    return ChatResponse(
        code="00",
        success=result.success,
        message=result.message,
        answer=result.answer,
        citations=[
            {"source_id": c.source_id, "text": c.text, "score": float(c.score)} for c in result.citations
        ],
    )
