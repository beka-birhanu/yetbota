VENV ?= .venv
PY ?= $(VENV)/bin/python

.PHONY: venv install gen-proto run-rest run-grpc

venv:
	@test -x "$(PY)" || (python -m venv "$(VENV)" && "$(PY)" -m pip install -U pip)

install:
	@test -x "$(PY)" || (echo "missing $(PY); run: make venv" && exit 2)
	$(PY) -m pip install -r requirements.txt

gen-proto:
	@test -x "$(PY)" || (echo "missing $(PY); run: make venv" && exit 2)
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
