VENV ?= .venv
PY ?= $(VENV)/bin/python

.PHONY: venv install gen-proto run-rest run-grpc
.PHONY: weaviate-up weaviate-down ingest smoke-rest

venv:
	@test -x "$(PY)" || (python -m venv "$(VENV)" && "$(PY)" -m pip install -U pip)

install:
	@test -x "$(PY)" || (echo "missing $(PY); run: make venv" && exit 2)
	$(PY) -m pip install -r requirements.txt

gen-proto:
	@test -x "$(PY)" || (echo "missing $(PY); run: make venv" && exit 2)
	@mkdir -p ./interfaces/grpc/gen
	$(PY) -m grpc_tools.protoc \
		-I ../common/proto/definition \
		--python_out=./interfaces/grpc/gen \
		--grpc_python_out=./interfaces/grpc/gen \
		../common/proto/definition/ai/v1/rag.proto

run-rest:
	@test -x "$(PY)" || (echo "missing $(PY); run: make venv" && exit 2)
	$(PY) main.py rest --reload

run-grpc:
	@test -x "$(PY)" || (echo "missing $(PY); run: make venv" && exit 2)
	$(PY) main.py grpc

weaviate-up:
	cd .. && docker compose up -d weaviate

weaviate-down:
	cd .. && docker compose down

ingest:
	@test -x "$(PY)" || (echo "missing $(PY); run: make venv" && exit 2)
	$(PY) scripts/ingest_weaviate.py

smoke-rest:
	@echo 'curl -sS -X POST http://localhost:8001/v1/chat -H "Content-Type: application/json" -d "{\"query\":\"What is YetBota?\"}"'
