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

package lynkapi_test

import (
	"encoding/json"
	"testing"

	"github.com/lynkdb/lynkapi/go/lynkapi"
	// "google.golang.org/protobuf/types/known/structpb"
)

func Test_DataMerge(t *testing.T) {

	type Obj1 struct {
		Name string `json:"name" toml:"name"`
	}

	type Obj struct {
		Name       string    `json:"name" toml:"name"`
		Int        int64     `json:"int" toml:"int" lynkapi_value_limits:"10,20,30"`
		Obj1       Obj1      `json:"obj1" toml:"obj1"`
		Obj2       *Obj1     `json:"obj2" toml:"obj2"`
		Array      []string  `json:"array" toml:"array"`
		ArrayFloat []float64 `json:"array_float" toml:"array_float"`
	}

	spec, _, err := lynkapi.NewSpecFromStruct(Obj{})
	if err != nil {
		t.Fatal(err)
	}
	js, _ := json.MarshalIndent(spec, "", "  ")
	t.Logf("spec %v", string(js))

	base := &Obj{
		Array: []string{"111"},
	}

	update := &Obj{
		Name:       "test",
		Int:        123,
		Array:      []string{"a", "b"},
		ArrayFloat: []float64{1.1, 2.2},
		Obj1: Obj1{
			Name: "test1",
		},
		Obj2: &Obj1{
			Name: "test2",
		},
	}

	if _, err := spec.DataMerge(base, update); err != nil {
		t.Fatal(err)
	}

	js, _ = json.MarshalIndent(base, "", "  ")
	t.Logf("map %v", string(js))
}

/**
func Test_DataUpdate(t *testing.T) {

	type Obj1 struct {
		Name string `json:"name" toml:"name"`
	}

	type Obj struct {
		Name       string    `json:"name" toml:"name"`
		Int        int64     `json:"int" toml:"int"`
		Obj1       Obj1      `json:"obj1" toml:"obj1"`
		Obj2       *Obj1     `json:"obj2" toml:"obj2"`
		Array      []string  `json:"array" toml:"array"`
		ArrayFloat []float64 `json:"array_float" toml:"array_float"`
	}

	spec, _, err := lynkapi.NewSpecFromStruct(Obj{})
	if err != nil {
		t.Fatal(err)
	}
	js, _ := json.MarshalIndent(spec, "", "  ")
	t.Logf("spec %v", string(js))

	base := &Obj{}

	lv, _ := structpb.NewList([]interface{}{"a", "b"})
	lf, _ := structpb.NewList([]interface{}{1.1, 2.2})

	update := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"name":        structpb.NewStringValue("test"),
			"int":         structpb.NewNumberValue(123),
			"array":       structpb.NewListValue(lv),
			"array_float": structpb.NewListValue(lf),
		},
	}

	update.Fields["obj1"] = structpb.NewStructValue(&structpb.Struct{
		Fields: map[string]*structpb.Value{
			"name": structpb.NewStringValue("test1"),
		},
	})

	update.Fields["obj2"] = structpb.NewStructValue(&structpb.Struct{
		Fields: map[string]*structpb.Value{
			"name": structpb.NewStringValue("test2"),
		},
	})

	if err := lynkapi.DataUpdate(spec, base, update); err != nil {
		t.Fatal(err)
	}

	js, _ = json.MarshalIndent(base, "", "  ")
	t.Logf("map %v", string(js))
}
*/
