import grpc
from ai.events.v1 import events_pb2
from ai.screening.v1 import screening_pb2, screening_pb2_grpc
from google.protobuf import wrappers_pb2

from application.screen_text import ScreenText
from domain.entities import ContentKind, ScreeningRequest
from infrastructure.observability import (
    SCREENING_BLOCKS,
    SCREENING_DURATION,
    time_histogram,
)

_PROTO_TO_KIND: dict[int, ContentKind] = {
    events_pb2.CONTENT_KIND_POST: "post",
    events_pb2.CONTENT_KIND_QUESTION: "question",
    events_pb2.CONTENT_KIND_ANSWER: "answer",
}


def _resolve_kind(value: int) -> ContentKind:
    kind = _PROTO_TO_KIND.get(value)
    if kind is None:
        return "post"
    return kind


class ScreeningHandler(screening_pb2_grpc.ScreeningServiceServicer):
    def __init__(self, use_case: ScreenText) -> None:
        self._use_case = use_case

    async def Check(
        self,
        request: screening_pb2.CheckRequest,
        context: grpc.aio.ServicerContext,
    ) -> screening_pb2.CheckResponse:
        domain_request = ScreeningRequest(
            text=request.text,
            kind=_resolve_kind(request.kind),
        )
        async with time_histogram(SCREENING_DURATION):
            result = await self._use_case.execute(domain_request)
        if not result.ok:
            SCREENING_BLOCKS.labels(reason=result.reason or "unknown").inc()

        reason = (
            wrappers_pb2.StringValue(value=result.reason) if result.reason is not None else None
        )
        data = screening_pb2.CheckData(
            ok=result.ok,
            reason=reason,
            categories={k: float(v) for k, v in result.categories.items()},
        )
        return screening_pb2.CheckResponse(
            code="00",
            success=True,
            message="ok",
            data=data,
        )
