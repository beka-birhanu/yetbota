from __future__ import annotations

import grpc

import ai.v1.rag_pb2 as rag_pb2
import ai.v1.rag_pb2_grpc as rag_pb2_grpc

from application.chat_service import ChatService
from application.errors import InvalidQuery, Misconfigured, UpstreamUnavailable


class RAGServicer(rag_pb2_grpc.RAGServiceServicer):
    def __init__(self, chat_service: ChatService):
        self._chat_service = chat_service

    def Chat(self, request: rag_pb2.ChatRequest, context: grpc.ServicerContext) -> rag_pb2.ChatResponse:
        try:
            result = self._chat_service.chat(request.query)
        except InvalidQuery as e:
            context.abort(grpc.StatusCode.INVALID_ARGUMENT, str(e))
        except Misconfigured as e:
            context.abort(grpc.StatusCode.FAILED_PRECONDITION, str(e))
        except UpstreamUnavailable as e:
            context.abort(grpc.StatusCode.UNAVAILABLE, str(e))
        resp = rag_pb2.ChatResponse(
            code="00",
            success=result.success,
            message=result.message,
            answer=result.answer,
        )
        for c in result.citations:
            resp.citations.add(source_id=c.source_id, text=c.text, score=float(c.score))
        return resp
