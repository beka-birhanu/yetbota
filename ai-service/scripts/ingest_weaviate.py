from __future__ import annotations

import argparse
import json
import os
from pathlib import Path
from typing import Any

import weaviate
from weaviate.classes.config import DataType, Property, Configure
from weaviate.classes.init import Auth


def _env(name: str, default: str | None = None) -> str | None:
    v = os.getenv(name)
    if v is None or not str(v).strip():
        return default
    return str(v).strip()


def _read_jsonl(path: Path) -> list[dict[str, Any]]:
    out: list[dict[str, Any]] = []
    with path.open("r", encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            out.append(json.loads(line))
    return out


def _embed_rows_with_gemini(rows: list[dict[str, Any]]) -> None:
    api_key = _env("GEMINI_API_KEY")
    model = _env("GEMINI_EMBEDDING_MODEL", "models/text-embedding-004") or "models/text-embedding-004"
    needs = any(
        not (isinstance(r.get("vector"), list) and r["vector"] and all(isinstance(x, (int, float)) for x in r["vector"]))
        for r in rows
        if str(r.get("source_id", "")).strip() and str(r.get("text", "")).strip()
    )
    if not needs:
        return
    if not api_key:
        raise SystemExit(
            "GEMINI_API_KEY is required to compute vectors for ingestion "
            "(or provide a numeric `vector` field per JSONL row)."
        )

    import google.generativeai as genai

    genai.configure(api_key=api_key)
    resolved = model if str(model).startswith("models/") else f"models/{model}"

    for r in rows:
        if isinstance(r.get("vector"), list) and r["vector"]:
            continue
        text = str(r.get("text", "")).strip()
        if not text:
            continue
        result = genai.embed_content(model=resolved, content=text)
        embedding = result.get("embedding")
        if not isinstance(embedding, list):
            raise RuntimeError("unexpected embedding response from Gemini")
        r["vector"] = [float(x) for x in embedding]


def _connect() -> weaviate.WeaviateClient:
    weaviate_url = _env("WEAVIATE_URL")
    host = _env("WEAVIATE_HOST")
    secure = (_env("WEAVIATE_SECURE", "false") or "false").lower() in {"1", "true", "yes"}
    http_port = int(_env("WEAVIATE_HTTP_PORT", "8080") or "8080")
    grpc_port = int(_env("WEAVIATE_GRPC_PORT", "50051") or "50051")
    api_key = _env("WEAVIATE_API_KEY")

    auth = Auth.api_key(api_key) if api_key else None

    if weaviate_url and not host:
        # Basic parse (scheme://host:port). Ports can still be overridden by env.
        from urllib.parse import urlparse

        parsed = urlparse(weaviate_url)
        if parsed.hostname:
            host = parsed.hostname
        if parsed.scheme:
            secure = parsed.scheme.lower() == "https"
        if parsed.port:
            http_port = int(parsed.port)

    if not host:
        raise SystemExit("Missing WEAVIATE_URL or WEAVIATE_HOST")

    return weaviate.connect_to_custom(
        http_host=host,
        http_port=http_port,
        http_secure=bool(secure),
        grpc_host=host,
        grpc_port=int(grpc_port),
        grpc_secure=bool(secure),
        auth_credentials=auth,
    )


def _ensure_collection(client: weaviate.WeaviateClient, name: str) -> None:
    try:
        client.collections.get(name)
        return
    except Exception:
        # Collection doesn't exist yet (or server is old/unreachable).
        pass

    client.collections.create(
        name=name,
        properties=[
            Property(name="source_id", data_type=DataType.TEXT),
            Property(name="text", data_type=DataType.TEXT),
            Property(name="verified", data_type=DataType.BOOL),
        ],
        vectorizer_config=Configure.Vectorizer.none(),
    )


def _ingest(client: weaviate.WeaviateClient, collection: str, rows: list[dict[str, Any]]) -> int:
    col = client.collections.get(collection)
    inserted = 0
    for r in rows:
        source_id = str(r.get("source_id", "")).strip()
        text = str(r.get("text", "")).strip()
        verified = bool(r.get("verified", True))
        vector = r.get("vector")
        if not source_id or not text:
            continue

        props = {"source_id": source_id, "text": text, "verified": verified}
        if isinstance(vector, list) and vector and all(isinstance(x, (int, float)) for x in vector):
            col.data.insert(properties=props, vector=[float(x) for x in vector])
        else:
            col.data.insert(properties=props)
        inserted += 1
    return inserted


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--collection", default=None, help="Overrides WEAVIATE_COLLECTION")
    ap.add_argument("--data", default="data/sample_chunks.jsonl", help="Path to JSONL")
    args = ap.parse_args()

    collection = args.collection or _env("WEAVIATE_COLLECTION")
    if not collection:
        raise SystemExit("Missing WEAVIATE_COLLECTION (or --collection)")

    data_path = Path(args.data)
    if not data_path.is_absolute():
        data_path = Path(__file__).resolve().parents[1] / data_path
    rows = _read_jsonl(data_path)
    _embed_rows_with_gemini(rows)

    client = _connect()
    try:
        _ensure_collection(client, collection)
        inserted = _ingest(client, collection, rows)
        print(f"Inserted {inserted} objects into collection={collection}")
    finally:
        client.close()


if __name__ == "__main__":
    main()

