# YetBota AI Service (FastAPI + gRPC)

Phase 1 delivers a Clean Architecture skeleton with REST + gRPC servers booting and proto codegen wired to the existing proto sources in `../common/proto/definition/ai/v1/*`.

## Requirements
- Python 3.11+

## Quickstart
From repo root:

```bash
cd ai-service
cp .env.example .env
make venv
make install
make gen-proto
make run-rest
```

In another terminal:

```bash
cd ai-service
make run-grpc
```

## Endpoints (Phase 1)
- REST: `GET /healthz`, `GET /readyz`, `GET /version` (plus stubbed `/v1/*`)
- gRPC: registers `EmbeddingService`, `VectorService`, `DuplicateService`, `RAGService` (methods return UNIMPLEMENTED in Phase 1)
