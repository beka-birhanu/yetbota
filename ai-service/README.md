# YetBota AI Service (FastAPI + gRPC)

This service implements **RAG chat** over:

- REST: `POST /v1/chat`
- gRPC: `ai.v1.RAGService.Chat`

Proto source lives at `../common/proto/definition/ai/v1/rag.proto`.

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

## Configuration
See `.env.example` for the full list. The minimum to get real answers:

- `GEMINI_API_KEY`
- `WEAVIATE_URL` (example: `http://localhost:8080`)
- `WEAVIATE_COLLECTION`

Optional tuning:

- `GEMINI_MODEL`, `GEMINI_EMBEDDING_MODEL`
- `WEAVIATE_VECTOR_NAME` (defaults to `default`)
- `WEAVIATE_VERIFIED_PROPERTY` (defaults to `verified` boolean filter)
- `WEAVIATE_GRPC_HOST`, `WEAVIATE_GRPC_PORT`, `WEAVIATE_GRPC_SECURE` (only if your deployment needs non-default gRPC settings)
- `RAG_TOP_K`, `RAG_MIN_SCORE`

## Weaviate schema expectations (Phase 2)
Your Weaviate collection should include at least:

- `source_id` (text)
- `text` (text)
- `verified` (boolean) — retrieval only returns `verified == true`
- a named vector configured as `WEAVIATE_VECTOR_NAME` (defaults to `default`)

## Endpoints
- REST:
  - `GET /healthz`
  - `GET /readyz`
  - `GET /version`
  - `POST /v1/chat`
- gRPC:
  - `ai.v1.RAGService/Chat`

## Tests
```bash
cd ai-service
.venv/bin/python -m pytest -q
```
