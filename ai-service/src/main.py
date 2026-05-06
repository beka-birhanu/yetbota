import asyncio
import signal
from typing import Any

import uvicorn
from ai.assistant.v1 import assistant_pb2_grpc
from ai.screening.v1 import screening_pb2_grpc

from application.consult_assistant import ConsultAssistant
from application.ingest_content import IngestContent
from application.screen_text import ScreenText
from infrastructure.chunking import RuneWindowChunker
from infrastructure.config import Settings, get_settings
from infrastructure.embedding import GeminiEmbedder
from infrastructure.graph import Neo4jSimilarityGraph
from infrastructure.llm import (
    SCREENING_RESPONSE_SCHEMA,
    AssistantPromptBuilder,
    GeminiLLM,
    ScreeningPromptBuilder,
)
from infrastructure.messaging import RabbitMQConsumer, RabbitMQPublisher
from infrastructure.observability import configure_logging, get_logger
from infrastructure.vector import WeaviateVectorStore
from interfaces.grpc import GrpcServer
from interfaces.grpc.handlers.assistant import AssistantHandler
from interfaces.grpc.handlers.screening import ScreeningHandler
from interfaces.http import create_app
from interfaces.workers.ingest_worker import IngestWorker


async def _serve(settings: Settings) -> None:
    log = get_logger(__name__)
    log.info("service.starting", version=settings.app.version)

    embedder = GeminiEmbedder(settings.gemini)
    llm = GeminiLLM(settings.gemini)
    vector_store = WeaviateVectorStore(settings.weaviate)
    await vector_store.connect()

    similarity_graph: Neo4jSimilarityGraph | None = None
    if settings.similarity.enabled:
        similarity_graph = Neo4jSimilarityGraph(settings.neo4j)
        await similarity_graph.connect()

    consumer = RabbitMQConsumer(settings.rabbitmq)
    publisher = RabbitMQPublisher(settings.rabbitmq)
    await consumer.connect()
    await publisher.connect()

    chunker = RuneWindowChunker(settings.chunker)
    ingest_use_case = IngestContent(
        chunker=chunker,
        embedder=embedder,
        vector_store=vector_store,
        distance_threshold=settings.dedup.distance_threshold,
        similarity_graph=similarity_graph,
        similarity_top_n=settings.similarity.top_n,
        similarity_oversample=settings.similarity.candidate_oversample,
        logger=log,
    )
    worker = IngestWorker(
        consumer=consumer,
        publisher=publisher,
        use_case=ingest_use_case,
        rabbitmq=settings.rabbitmq,
    )

    assistant_use_case = ConsultAssistant(
        embedder=embedder,
        vector_store=vector_store,
        llm=llm,
        prompt_builder=AssistantPromptBuilder(),
        top_k=settings.rag.top_k,
        min_similarity=settings.rag.min_similarity,
        max_tokens=settings.gemini.max_tokens,
        temperature=settings.gemini.temperature,
    )

    screening_use_case = ScreenText(
        llm=llm,
        prompt_builder=ScreeningPromptBuilder(),
        response_schema=SCREENING_RESPONSE_SCHEMA,
        block_threshold=settings.screening.block_threshold,
        timeout_s=settings.screening.timeout_s,
        cache_size=settings.screening.cache_size,
        cache_ttl_s=settings.screening.cache_ttl_s,
    )

    app = create_app(settings)
    app.state.settings = settings
    app.state.assistant = assistant_use_case
    app.state.screening = screening_use_case

    uv_config = uvicorn.Config(
        app,
        host=settings.http.host,
        port=settings.http.port,
        log_config=None,
        access_log=False,
        lifespan="on",
    )
    uv_server = uvicorn.Server(uv_config)

    grpc_server = GrpcServer(settings)
    grpc_server.register_servicer(
        assistant_pb2_grpc.add_AssistantServiceServicer_to_server,
        AssistantHandler(assistant_use_case),
        service_name="ai.v1.AssistantService",
    )
    grpc_server.register_servicer(
        screening_pb2_grpc.add_ScreeningServiceServicer_to_server,
        ScreeningHandler(screening_use_case),
        service_name="ai.v1.ScreeningService",
    )
    await grpc_server.start()

    stop_event = asyncio.Event()

    def _request_shutdown(*_: Any) -> None:
        if not stop_event.is_set():
            log.info("service.shutdown_requested")
            stop_event.set()

    loop = asyncio.get_running_loop()
    for sig in (signal.SIGINT, signal.SIGTERM):
        try:
            loop.add_signal_handler(sig, _request_shutdown)
        except NotImplementedError:
            signal.signal(sig, _request_shutdown)

    async def _watch_stop() -> None:
        await stop_event.wait()
        uv_server.should_exit = True
        await consumer.close()
        await publisher.close()
        if similarity_graph is not None:
            await similarity_graph.close()
        await vector_store.close()
        await grpc_server.stop(grace_seconds=5.0)

    log.info(
        "service.started",
        http_port=settings.http.port,
        grpc_port=settings.grpc.port,
    )

    try:
        await asyncio.gather(
            uv_server.serve(),
            grpc_server.wait_for_termination(),
            worker.run(),
            _watch_stop(),
            return_exceptions=False,
        )
    except asyncio.CancelledError:
        pass
    finally:
        log.info("service.stopped")


def run() -> None:
    settings = get_settings()
    configure_logging(settings.logging)
    asyncio.run(_serve(settings))


if __name__ == "__main__":
    run()
