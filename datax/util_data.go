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

package datax

import (
	"encoding/base64"
	"reflect"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"
)

func ParseStruct(obj interface{}) *structpb.Struct {

	m := &structpb.Struct{
		Fields: map[string]*structpb.Value{},
	}

	if obj == nil || reflect.TypeOf(obj).Kind() != reflect.Struct {
		return m
	}

	var (
		parseStruct func(up *structpb.Struct, rv reflect.Value)
		parseArray  func(up *structpb.ListValue, rv reflect.Value)
	)

	parseStruct = func(up *structpb.Struct, rv reflect.Value) {

		if !rv.IsValid() || rv.Type().Kind() != reflect.Struct {
			return
		}

		fields := reflect.VisibleFields(rv.Type())

		for _, fd := range fields {

			if fd.Name == "" || fd.Name[0] < 'A' || fd.Name[0] > 'Z' {
				continue
			}

			var (
				name = fd.Name
				v1   = rv.FieldByIndex(fd.Index)
				v2   *structpb.Value
			)

			if sa := strings.Split(fd.Tag.Get("json"), ","); len(sa) > 0 &&
				sa[0] != "" && sa[0] != "omitempty" {
				name = sa[0]
			}

			switch v1.Kind() {

			case reflect.String:
				v2 = structpb.NewStringValue(v1.String())

			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				v2 = structpb.NewNumberValue(float64(v1.Int()))

			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				v2 = structpb.NewNumberValue(float64(v1.Uint()))

			case reflect.Float32, reflect.Float64:
				v2 = structpb.NewNumberValue(v1.Float())

			case reflect.Bool:
				v2 = structpb.NewBoolValue(v1.Bool())

			case reflect.Struct:
				st := &structpb.Struct{}
				v2 = structpb.NewStructValue(st)
				parseStruct(st, v1)

			case reflect.Pointer:
				st := &structpb.Struct{}
				v2 = structpb.NewStructValue(st)
				parseStruct(st, v1)

			case reflect.Slice:
				if v1.Len() > 0 {
					if v1.Index(0).Kind() == reflect.Uint8 {
						b64 := base64.StdEncoding.EncodeToString(v1.Bytes())
						v2 = structpb.NewStringValue(b64)
					} else {
						ls := &structpb.ListValue{}
						v2 = structpb.NewListValue(ls)
						parseArray(ls, v1)
					}
				}
			}

			if v2 != nil {
				if up.Fields == nil {
					up.Fields = map[string]*structpb.Value{}
				}
				up.Fields[name] = v2
			}
		}
	}

	parseArray = func(up *structpb.ListValue, rv reflect.Value) {

		if !rv.IsValid() || rv.Type().Kind() != reflect.Slice {
			return
		}

		n := rv.Cap()
		for i := 0; i < n; i++ {

			var (
				v1 = rv.Index(i)
				v2 *structpb.Value
			)

			switch v1.Kind() {

			case reflect.String:
				v2 = structpb.NewStringValue(v1.String())

			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				v2 = structpb.NewNumberValue(float64(v1.Int()))

			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				v2 = structpb.NewNumberValue(float64(v1.Uint()))

			case reflect.Float32, reflect.Float64:
				v2 = structpb.NewNumberValue(v1.Float())

			case reflect.Bool:
				v2 = structpb.NewBoolValue(v1.Bool())

			case reflect.Struct:
				st := &structpb.Struct{}
				v2 = structpb.NewStructValue(st)
				parseStruct(st, v1)

			case reflect.Pointer:
				st := &structpb.Struct{}
				v2 = structpb.NewStructValue(st)
				parseStruct(st, v1)

			case reflect.Slice:
				ls := &structpb.ListValue{}
				v2 = structpb.NewListValue(ls)
				parseArray(ls, v1)
			}

			if v2 != nil {
				up.Values = append(up.Values, v2)
			}
		}
	}

	parseStruct(m, reflect.ValueOf(obj))

	return m
}
