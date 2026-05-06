#
# Proto generation:
# - Each proto file has a target gengo-<service>-<name>.
# - Service-level alias: gengo-<service> runs all protos for that service.
# - Run all: make gengo-all
# - All commands accept optional VERSION arg (default: v1), e.g.:
#     make gengo-identity VERSION=v1
#     make gengo-identity-user VERSION=v2

# Global options
VERSION ?= v1
PROTO_DIR := proto/definition
OUT_DIR := proto/generated/go
PY_OUT_DIR := proto/generated/python

PROTOC_GO_OPTS = --go_opt=paths=source_relative
PROTOC_GRPC_OPTS = --go-grpc_opt=paths=source_relative --go-grpc_opt=require_unimplemented_servers=false

PYTHON ?= python3
PROTOC_PY := $(PYTHON) -m grpc_tools.protoc

# =====================
# identity service (proto/definition/identity/<VERSION>)
# =====================

IDENTITY_PROTO_DIR := $(PROTO_DIR)/identity/$(VERSION)
CONTENT_POST_PROTO_DIR := $(PROTO_DIR)/content/post/$(VERSION)
CONTENT_COMMENT_PROTO_DIR := $(PROTO_DIR)/content/comment/$(VERSION)
AI_ASSISTANT_PROTO_DIR := $(PROTO_DIR)/ai/assistant/$(VERSION)
AI_SCREENING_PROTO_DIR := $(PROTO_DIR)/ai/screening/$(VERSION)
AI_EVENTS_PROTO_DIR := $(PROTO_DIR)/ai/events/$(VERSION)

.PHONY: gengo-identity-auth gengo-identity-user gengo-identity \
        gengo-content-post gengo-content-comment gengo-content \
        gengo-ai-assistant gengo-ai-screening gengo-ai-events gengo-ai \
        gengo-all \
        genpy-ai-assistant genpy-ai-screening genpy-ai-events genpy-ai \
        genpy-all

gengo-identity-auth:
	@echo " > Generate identity/$(VERSION)/auth.proto"
	@mkdir -p $(OUT_DIR)
	@protoc --proto_path=$(PROTO_DIR) \
		--go_out=$(OUT_DIR) \
		--go-grpc_out=$(OUT_DIR) \
		$(PROTOC_GO_OPTS) $(PROTOC_GRPC_OPTS) \
		$(IDENTITY_PROTO_DIR)/auth.proto

gengo-identity-user:
	@echo " > Generate identity/$(VERSION)/user.proto"
	@mkdir -p $(OUT_DIR)
	@protoc --proto_path=$(PROTO_DIR) \
		--go_out=$(OUT_DIR) \
		--go-grpc_out=$(OUT_DIR) \
		$(PROTOC_GO_OPTS) $(PROTOC_GRPC_OPTS) \
		$(IDENTITY_PROTO_DIR)/user.proto

gengo-content-post:
	@echo " > Generate content/post/$(VERSION)/post.proto"
	@mkdir -p $(OUT_DIR)
	@protoc --proto_path=$(PROTO_DIR) \
		--go_out=$(OUT_DIR) \
		--go-grpc_out=$(OUT_DIR) \
		$(PROTOC_GO_OPTS) $(PROTOC_GRPC_OPTS) \
		$(CONTENT_POST_PROTO_DIR)/post.proto

gengo-content-comment:
	@echo " > Generate content/comment/$(VERSION)/comment.proto"
	@mkdir -p $(OUT_DIR)
	@protoc --proto_path=$(PROTO_DIR) \
		--go_out=$(OUT_DIR) \
		--go-grpc_out=$(OUT_DIR) \
		$(PROTOC_GO_OPTS) $(PROTOC_GRPC_OPTS) \
		$(CONTENT_COMMENT_PROTO_DIR)/comment.proto

# =====================
# ai service (proto/definition/ai/<resource>/<VERSION>) — Go
# =====================

gengo-ai-events:
	@echo " > Generate ai/events/$(VERSION)/events.proto"
	@mkdir -p $(OUT_DIR)
	@protoc --proto_path=$(PROTO_DIR) \
		--go_out=$(OUT_DIR) \
		--go-grpc_out=$(OUT_DIR) \
		$(PROTOC_GO_OPTS) $(PROTOC_GRPC_OPTS) \
		$(AI_EVENTS_PROTO_DIR)/events.proto

gengo-ai-assistant: gengo-ai-events
	@echo " > Generate ai/assistant/$(VERSION)/assistant.proto"
	@mkdir -p $(OUT_DIR)
	@protoc --proto_path=$(PROTO_DIR) \
		--go_out=$(OUT_DIR) \
		--go-grpc_out=$(OUT_DIR) \
		$(PROTOC_GO_OPTS) $(PROTOC_GRPC_OPTS) \
		$(AI_ASSISTANT_PROTO_DIR)/assistant.proto

gengo-ai-screening: gengo-ai-events
	@echo " > Generate ai/screening/$(VERSION)/screening.proto"
	@mkdir -p $(OUT_DIR)
	@protoc --proto_path=$(PROTO_DIR) \
		--go_out=$(OUT_DIR) \
		--go-grpc_out=$(OUT_DIR) \
		$(PROTOC_GO_OPTS) $(PROTOC_GRPC_OPTS) \
		$(AI_SCREENING_PROTO_DIR)/screening.proto

# =====================
# ai service — Python
# =====================

genpy-ai-events:
	@echo " > Generate (py) ai/events/$(VERSION)/events.proto"
	@mkdir -p $(PY_OUT_DIR)
	@$(PROTOC_PY) --proto_path=$(PROTO_DIR) \
		--python_out=$(PY_OUT_DIR) \
		--grpc_python_out=$(PY_OUT_DIR) \
		--pyi_out=$(PY_OUT_DIR) \
		$(AI_EVENTS_PROTO_DIR)/events.proto
	@find $(PY_OUT_DIR) -type d -exec touch {}/__init__.py \;

genpy-ai-assistant: genpy-ai-events
	@echo " > Generate (py) ai/assistant/$(VERSION)/assistant.proto"
	@mkdir -p $(PY_OUT_DIR)
	@$(PROTOC_PY) --proto_path=$(PROTO_DIR) \
		--python_out=$(PY_OUT_DIR) \
		--grpc_python_out=$(PY_OUT_DIR) \
		--pyi_out=$(PY_OUT_DIR) \
		$(AI_ASSISTANT_PROTO_DIR)/assistant.proto
	@find $(PY_OUT_DIR) -type d -exec touch {}/__init__.py \;

genpy-ai-screening: genpy-ai-events
	@echo " > Generate (py) ai/screening/$(VERSION)/screening.proto"
	@mkdir -p $(PY_OUT_DIR)
	@$(PROTOC_PY) --proto_path=$(PROTO_DIR) \
		--python_out=$(PY_OUT_DIR) \
		--grpc_python_out=$(PY_OUT_DIR) \
		--pyi_out=$(PY_OUT_DIR) \
		$(AI_SCREENING_PROTO_DIR)/screening.proto
	@find $(PY_OUT_DIR) -type d -exec touch {}/__init__.py \;

gengo-identity: gengo-identity-auth gengo-identity-user
gengo-content: gengo-content-post gengo-content-comment
gengo-ai: gengo-ai-events gengo-ai-assistant gengo-ai-screening
genpy-ai: genpy-ai-events genpy-ai-assistant genpy-ai-screening

gengo-all: gengo-identity gengo-content gengo-ai
genpy-all: genpy-ai
