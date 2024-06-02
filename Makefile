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
PROTOC_ARGS = --proto_path=./ --go_out=./go/lynkapi --go-grpc_out=./go/lynkapi ./lynkdb/lynkapi/*.proto

FITTER_CMD = ./lynkapi-fitter
FITTER_ARGS = ./go/lynkapi

.PHONY: api

all: api
	@echo ""
	@echo "build complete"
	@echo ""

api:
	##  go install github.com/golang/protobuf/protoc-gen-go
	##  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
	$(PROTOC_CMD) $(PROTOC_ARGS)
	go build -o lynkapi-fitter cmd/lynkapi-fitter/lynkapi-fitter.go
	$(FITTER_CMD) $(FITTER_ARGS)

