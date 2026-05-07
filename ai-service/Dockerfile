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
    PYTHONDONTWRITEBYTECODE=1
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
