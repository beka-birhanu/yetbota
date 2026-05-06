from fastapi import APIRouter, HTTPException, Request

from application.consult_assistant import ConsultAssistant
from application.errors import LLMUnavailable
from domain.entities import ChatQuery, ChatResponse
from infrastructure.observability import RAG_CHAT_DURATION, time_histogram

router = APIRouter(prefix="/v1/assistant", tags=["assistant"])


@router.post("/chat", response_model=ChatResponse)
async def chat(query: ChatQuery, request: Request) -> ChatResponse:
    use_case: ConsultAssistant = request.app.state.assistant
    try:
        async with time_histogram(RAG_CHAT_DURATION):
            return await use_case.execute(query)
    except LLMUnavailable as exc:
        raise HTTPException(
            status_code=503,
            detail="AI Assistant is currently unavailable.",
        ) from exc
