from __future__ import annotations

from concurrent import futures

import grpc
from grpc_reflection.v1alpha import reflection

from bootstrap.container import build_container
from infrastructure.settings import get_settings
from interfaces.grpc import genpath  # noqa: F401
from interfaces.grpc.rag_service import RAGServicer

import ai.v1.rag_pb2 as rag_pb2
import ai.v1.rag_pb2_grpc as rag_pb2_grpc


def serve() -> None:
    settings = get_settings()
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=16))

    container = build_container()
    rag_pb2_grpc.add_RAGServiceServicer_to_server(RAGServicer(container.chat_service), server)

    service_names = (
        rag_pb2.DESCRIPTOR.services_by_name["RAGService"].full_name,
        reflection.SERVICE_NAME,
    )
    reflection.enable_server_reflection(service_names, server)

    addr = f"{settings.grpc_host}:{settings.grpc_port}"
    server.add_insecure_port(addr)

    server.start()
    print(f"gRPC server listening on {addr}")
    try:
        server.wait_for_termination()
    finally:
        container.close()

