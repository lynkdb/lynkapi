# Copyright 2024 Eryx <evorui at gmail dot com>, All rights reserved.

PROTOC_CMD = protoc
PROTOC_ARGS = --proto_path=./datax/ --go_opt=paths=source_relative --go_out=./datax --go-grpc_out=./datax ./datax/*.proto

HTOML_TAG_FIX_CMD = htoml-tag-fix
HTOML_TAG_FIX_ARGS = datax

.PHONY: api

all: api
	@echo ""
	@echo "build complete"
	@echo ""

api:
	##  go install github.com/golang/protobuf/protoc-gen-go
	##  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
	##  go install github.com/hooto/htoml4g/cmd/htoml-tag-fix
	$(PROTOC_CMD) $(PROTOC_ARGS)
	$(HTOML_TAG_FIX_CMD) $(HTOML_TAG_FIX_ARGS)


