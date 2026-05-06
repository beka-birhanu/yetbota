import grpc
from ai.assistant.v1 import assistant_pb2, assistant_pb2_grpc
from ai.events.v1 import events_pb2

from application.consult_assistant import ConsultAssistant
from application.errors import LLMUnavailable
from domain.entities import ChatQuery, ContentKind
from infrastructure.observability import RAG_CHAT_DURATION, time_histogram

_KIND_TO_PROTO: dict[ContentKind, int] = {
    "post": events_pb2.CONTENT_KIND_POST,
    "question": events_pb2.CONTENT_KIND_QUESTION,
    "answer": events_pb2.CONTENT_KIND_ANSWER,
}


def _user_id(request: assistant_pb2.ChatRequest) -> str | None:
    if not request.HasField("user_id"):
        return None
    value = request.user_id.value
    return value or None


class AssistantHandler(assistant_pb2_grpc.AssistantServiceServicer):
    def __init__(self, use_case: ConsultAssistant) -> None:
        self._use_case = use_case

    async def Chat(
        self,
        request: assistant_pb2.ChatRequest,
        context: grpc.aio.ServicerContext,
    ) -> assistant_pb2.ChatResponse:
        query = ChatQuery(text=request.text, user_id=_user_id(request))
        try:
            async with time_histogram(RAG_CHAT_DURATION):
                result = await self._use_case.execute(query)
        except LLMUnavailable:
            await context.abort(
                grpc.StatusCode.UNAVAILABLE,
                "AI Assistant is currently unavailable.",
            )
            raise

        citations = [
            assistant_pb2.Citation(
                source_id=c.source_id,
                kind=_KIND_TO_PROTO[c.kind],
                text=c.text,
                score=c.score,
            )
            for c in result.citations
        ]
        return assistant_pb2.ChatResponse(
            code="00",
            success=True,
            message="ok",
            data=assistant_pb2.ChatData(answer=result.answer, citations=citations),
        )
