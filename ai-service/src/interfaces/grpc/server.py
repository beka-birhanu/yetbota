import grpc
from grpc_reflection.v1alpha import reflection

from infrastructure.config import Settings
from infrastructure.observability import get_logger

logger = get_logger(__name__)

_REFLECTION_SERVICE_NAME = "grpc.reflection.v1alpha.ServerReflection"


class GrpcServer:
    def __init__(self, settings: Settings) -> None:
        self._settings = settings
        self._server: grpc.aio.Server = grpc.aio.server(
            options=[
                ("grpc.keepalive_time_ms", settings.grpc.keepalive_time_s * 1000),
                ("grpc.max_concurrent_streams", settings.grpc.max_concurrent_streams),
            ]
        )
        self._service_names: list[str] = [_REFLECTION_SERVICE_NAME]
        self._started = False

    def register_servicer(
        self,
        adder: object,
        servicer: object,
        *,
        service_name: str,
    ) -> None:
        if self._started:
            raise RuntimeError("cannot register servicers after start")
        adder(servicer, self._server)  # type: ignore[operator]
        self._service_names.append(service_name)

    async def start(self) -> None:
        if self._settings.grpc.enable_reflection:
            reflection.enable_server_reflection(self._service_names, self._server)

        addr = f"{self._settings.grpc.host}:{self._settings.grpc.port}"
        self._server.add_insecure_port(addr)
        await self._server.start()
        self._started = True
        logger.info("grpc.server.started", address=addr)

    async def wait_for_termination(self) -> None:
        await self._server.wait_for_termination()

    async def stop(self, grace_seconds: float = 5.0) -> None:
        await self._server.stop(grace_seconds)
        logger.info("grpc.server.stopped")
