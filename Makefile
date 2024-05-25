# Copyright 2024 Eryx <evorui at gmail dot com>, All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

PROTOC_CMD = protoc
PROTOC_ARGS = --proto_path=./datax/ --go_opt=paths=source_relative --go_out=./datax --go-grpc_out=./datax ./datax/*.proto

FITTER_CMD = htoml-tag-fix
FITTER_ARGS = datax

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
	go build -o lynkx-fitter cmd/lynkx-fitter/lynkx-fitter.go
	./lynkx-fitter ./datax

