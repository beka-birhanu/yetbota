PYTHON ?= python3
VENV := .venv
BIN := $(VENV)/bin
COMMON_DIR := ../common
PROTO_PY_DIR := $(COMMON_DIR)/proto/generated/python
PYTHONPATH_FOR_RUN := src:$(PROTO_PY_DIR)

.PHONY: install proto lint format typecheck contracts test test-cov run docker-build clean venv

venv:
	@test -d $(VENV) || $(PYTHON) -m venv $(VENV)
	@$(BIN)/pip install --upgrade pip wheel >/dev/null

install: venv
	$(BIN)/pip install -r requirements-dev.txt

proto:
	$(MAKE) -C $(COMMON_DIR) genpy-ai PYTHON=$(abspath $(BIN)/python)

lint:
	$(BIN)/ruff check src tests
	$(BIN)/ruff format --check src tests

format:
	$(BIN)/ruff format src tests
	$(BIN)/ruff check --fix src tests

typecheck:
	PYTHONPATH=$(PYTHONPATH_FOR_RUN) $(BIN)/mypy

contracts:
	PYTHONPATH=$(PYTHONPATH_FOR_RUN) $(BIN)/lint-imports

test:
	PYTHONPATH=$(PYTHONPATH_FOR_RUN) $(BIN)/pytest

test-cov:
	PYTHONPATH=$(PYTHONPATH_FOR_RUN) $(BIN)/pytest --cov=src --cov-report=term-missing

run:
	PYTHONPATH=$(PYTHONPATH_FOR_RUN) $(BIN)/python -m main

docker-build:
	docker build -t yetbota/ai-service:dev -f Dockerfile ..

clean:
	rm -rf .pytest_cache .mypy_cache .ruff_cache dist build *.egg-info
	find . -type d -name __pycache__ -exec rm -rf {} +
