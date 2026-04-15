from __future__ import annotations

from concurrent import futures

import grpc
from grpc_reflection.v1alpha import reflection

from infra.config import get_settings
from interfaces.grpc import genpath  # noqa: F401

import ai.v1.rag_pb2 as rag_pb2
import ai.v1.rag_pb2_grpc as rag_pb2_grpc


class UnimplementedServicer:
    def __init__(self, method_prefix: str):
        self.method_prefix = method_prefix

    def __getattr__(self, name: str):
        def handler(request, context):
            context.abort(grpc.StatusCode.UNIMPLEMENTED, f"{self.method_prefix}.{name} not implemented (phase 2+)")

        return handler


def serve() -> None:
    settings = get_settings()
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=16))

    rag_pb2_grpc.add_RAGServiceServicer_to_server(UnimplementedServicer("RAGService"), server)

    service_names = (
        rag_pb2.DESCRIPTOR.services_by_name["RAGService"].full_name,
        reflection.SERVICE_NAME,
    )
    reflection.enable_server_reflection(service_names, server)

    addr = f"{settings.grpc_host}:{settings.grpc_port}"
    server.add_insecure_port(addr)
    server.start()
    print(f"gRPC server listening on {addr}")
    server.wait_for_termination()

