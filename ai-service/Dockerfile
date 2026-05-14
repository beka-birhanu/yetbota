# Build context is the repo root (yetbota/), not ai-service/. The Makefile
# `docker-build` target sets it correctly.

FROM python:3.12-slim AS proto-builder
WORKDIR /repo
RUN apt-get update \
 && apt-get install -y --no-install-recommends make \
 && rm -rf /var/lib/apt/lists/*
COPY ai-service/requirements-build.txt ./ai-service/
RUN pip install --no-cache-dir -r ai-service/requirements-build.txt
COPY common/ ./common/
RUN make -C common genpy-ai


FROM python:3.12-slim AS deps
WORKDIR /build
COPY ai-service/requirements.txt ./
RUN pip install --no-cache-dir --prefix=/install -r requirements.txt


FROM python:3.12-slim
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    APP__NAME=ai-service \
    APP__VERSION=0.0.1 \
    APP__DEBUG=false \
    HTTP__HOST=0.0.0.0 \
    HTTP__PORT=8080 \
    GRPC__HOST=0.0.0.0 \
    GRPC__PORT=9090 \
    GRPC__MAX_CONCURRENT_STREAMS=100 \
    GRPC__KEEPALIVE_TIME_S=60 \
    GRPC__ENABLE_REFLECTION=false \
    WEAVIATE__URL=https://shgw7jmaqbcwnpxpfpfwvw.c0.us-west3.gcp.weaviate.cloud \
    WEAVIATE__CLASS_NAME=ContentChunk \
    NEO4J__URI=bolt://neo4j.yetbota-infra.svc.cluster.local:7687 \
    NEO4J__USERNAME=neo4j \
    NEO4J__DATABASE=neo4j \
    SIMILARITY__ENABLED=true \
    SIMILARITY__TOP_N=10 \
    SIMILARITY__CANDIDATE_OVERSAMPLE=5 \
    GEMINI__LLM_MODEL=gemini-2.5-flash \
    GEMINI__EMBEDDING_MODEL=gemini-embedding-001 \
    GEMINI__EMBEDDING_DIMENSIONS=1536 \
    GEMINI__TIMEOUT_S=30 \
    GEMINI__MAX_TOKENS=1024 \
    GEMINI__TEMPERATURE=0.2 \
    DEDUP__DISTANCE_THRESHOLD=0.15 \
    RAG__TOP_K=5 \
    RAG__MIN_SIMILARITY=0.6 \
    SCREENING__BLOCK_THRESHOLD=0.7 \
    SCREENING__TIMEOUT_S=5 \
    SCREENING__CACHE_SIZE=1024 \
    SCREENING__CACHE_TTL_S=300 \
    CHUNKER__SIZE=512 \
    CHUNKER__OVERLAP=64 \
    LOGGING__LEVEL=INFO \
    LOGGING__FORMAT=json
RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates \
 && rm -rf /var/lib/apt/lists/* \
 && useradd --create-home --uid 1000 appuser

WORKDIR /app
COPY --from=deps /install /usr/local
COPY --from=proto-builder /repo/common/proto/generated/python /app/protos
COPY ai-service/src /app/src

ENV PYTHONPATH=/app/src:/app/protos
EXPOSE 8080 9090
USER appuser
CMD ["python", "-m", "main"]
