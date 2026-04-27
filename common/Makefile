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

PROTOC_GO_OPTS = --go_opt=paths=source_relative
PROTOC_GRPC_OPTS = --go-grpc_opt=paths=source_relative --go-grpc_opt=require_unimplemented_servers=false

# =====================
# identity service (proto/definition/identity/<VERSION>)
# =====================

IDENTITY_PROTO_DIR := $(PROTO_DIR)/identity/$(VERSION)
CONTENT_PROTO_DIR := $(PROTO_DIR)/content/$(VERSION)
CONTENT_COMMENT_PROTO_DIR := $(PROTO_DIR)/content/comment/$(VERSION)

.PHONY: gengo-identity-auth gengo-identity-user gengo-identity gengo-all

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
	@echo " > Generate content/$(VERSION)/post.proto"
	@mkdir -p $(OUT_DIR)
	@protoc --proto_path=$(PROTO_DIR) \
		--go_out=$(OUT_DIR) \
		--go-grpc_out=$(OUT_DIR) \
		$(PROTOC_GO_OPTS) $(PROTOC_GRPC_OPTS) \
		$(CONTENT_PROTO_DIR)/post.proto

gengo-content-comment:
	@echo " > Generate content/comment/$(VERSION)/comment.proto"
	@mkdir -p $(OUT_DIR)
	@protoc --proto_path=$(PROTO_DIR) \
		--go_out=$(OUT_DIR) \
		--go-grpc_out=$(OUT_DIR) \
		$(PROTOC_GO_OPTS) $(PROTOC_GRPC_OPTS) \
		$(CONTENT_COMMENT_PROTO_DIR)/comment.proto

# Service-level alias: generate all identity protos for the given VERSION
gengo-identity: gengo-identity-auth gengo-identity-user
gengo-content: gengo-content-post gengo-content-comment

# Global alias: run all services' proto generators
gengo-all: gengo-identity gengo-content
