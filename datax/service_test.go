// Copyright 2024 Eryx <evorui at gmail dot com>, All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package datax_test

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/lynkdb/lynkx/datax"
)

type TestService struct{}

type TestRequest struct {
	Name string `json:"name" toml:"name"`
}

type TestResponse struct {
	Content string `json:"content" toml:"content"`
}

func (it *TestService) TestCall(ctx context.Context, req *TestRequest) (*TestResponse, error) {
	return &TestResponse{
		Content: "hello grpc " + req.Name,
	}, nil
}

func (it *TestService) TestStdCall(ctx datax.Context, req *TestRequest) (*TestResponse, error) {
	return &TestResponse{
		Content: "hello std " + req.Name,
	}, nil
}

func Test_Service(t *testing.T) {

	s := datax.NewService()

	err := s.RegisterService(&TestService{})
	if err != nil {
		t.Fatal(err)
	}

	//
	{
		rs, err := s.ApiList(nil, &datax.ApiListRequest{})
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(rs)
	}

	//
	{
		req := &datax.Request{
			ServiceName: "TestService",
			MethodName:  "TestCall",
			Data: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"name": structpb.NewStringValue("test"),
				},
			},
		}

		rs, err := s.Exec(nil, req)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(rs)
	}
}
