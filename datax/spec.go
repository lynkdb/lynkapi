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
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"
)

const (
	DataSpec_Bool   = "bool"
	DataSpec_Int    = "int"
	DataSpec_Uint   = "uint"
	DataSpec_Float  = "float"
	DataSpec_String = "string"
	DataSpec_Struct = "struct"

	DataSpec_Array = "array"
	DataSpec_Bytes = "bytes"

	DataSpec_ArrayBool   = "array/bool"
	DataSpec_ArrayInt    = "array/int"
	DataSpec_ArrayUint   = "array/uint"
	DataSpec_ArrayFloat  = "array/float"
	DataSpec_ArrayString = "array/string"
	DataSpec_ArrayStruct = "array/struct"
)

func TryParseMap(obj interface{}) map[string]*structpb.Value {

	m := &structpb.Struct{
		Fields: map[string]*structpb.Value{},
	}

	if obj == nil {
		return m.Fields
	}

	if reflect.TypeOf(obj).Kind() != reflect.Struct {
		return m.Fields
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

	return m.Fields
}

func ParseSpec(obj interface{}) (*DataSpec, error) {

	if obj == nil {
		return nil, fmt.Errorf("empty object")
	}

	rt := reflect.TypeOf(obj)
	if rt.Kind() != reflect.Struct {
		return nil, fmt.Errorf("invalid object type")
	}

	var (
		spec        = &DataSpec{}
		parseStruct func(spec *DataSpec, upOpt *DataSpec_Field, rt reflect.Type)
		parseArray  func(upOpt *DataSpec_Field, rt reflect.Type)
	)

	parseDefaultValue := func(v string, t string) *structpb.Value {
		if v == "" {
			return nil
		}
		switch t {
		case DataSpec_String:
			return structpb.NewStringValue(v)

		case DataSpec_Int:
			if i, err := strconv.ParseInt(v, 10, 64); err == nil && i != 0 {
				return structpb.NewNumberValue(float64(i))
			}

		case DataSpec_Uint:
			if i, err := strconv.ParseUint(v, 10, 64); err == nil && i != 0 {
				return structpb.NewNumberValue(float64(i))
			}

		case DataSpec_Float:
			if f, err := strconv.ParseFloat(v, 64); err == nil && f != 0 {
				return structpb.NewNumberValue(f)
			}

		case DataSpec_Bool:
			if b, err := strconv.ParseBool(v); err == nil && b == true {
				return structpb.NewBoolValue(b)
			}
		}

		return nil
	}

	parseStruct = func(spec *DataSpec, upOpt *DataSpec_Field, rt reflect.Type) {

		if rt.Kind() != reflect.Struct {
			return
		}

		fields := reflect.VisibleFields(rt)

		for _, fd := range fields {

			if fd.Name == "" || fd.Name[0] < 'A' || fd.Name[0] > 'Z' {
				continue
			}

			var opt = &DataSpec_Field{
				Name: fd.Name,
			}

			if s := fd.Tag.Get("json"); s != "" {
				sa := strings.Split(s, ",")
				if sa[0] != "" && sa[0] != "omitempty" {
					opt.Name = sa[0]
				}
				if sa[len(sa)-1] != "omitempty" {
					opt.Required = true
				}
			}

			switch fd.Type.Kind() {

			case reflect.String:
				opt.Type = DataSpec_String

			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				opt.Type = DataSpec_Int

			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				opt.Type = DataSpec_Uint

			case reflect.Float32, reflect.Float64:
				opt.Type = DataSpec_Float

			case reflect.Bool:
				opt.Type = DataSpec_Bool

			case reflect.Struct:
				opt.Type = DataSpec_Struct
				parseStruct(nil, opt, fd.Type)

			case reflect.Pointer:
				opt.Type = DataSpec_Struct
				parseStruct(nil, opt, fd.Type.Elem())

			case reflect.Slice:
				if fd.Type.Elem().Kind() == reflect.Uint8 {
					opt.Type = DataSpec_Bytes
				} else {
					parseArray(opt, fd.Type.Elem())
				}
			}

			if opt.Type != "" {
				opt.DefaultValue = parseDefaultValue(fd.Tag.Get("default_value"), opt.Type)
				if spec != nil {
					spec.Fields = append(spec.Fields, opt)
				} else if upOpt != nil {
					upOpt.Fields = append(upOpt.Fields, opt)
				}
			}
		}
	}

	parseArray = func(opt *DataSpec_Field, rt reflect.Type) {

		switch rt.Kind() {

		case reflect.String:
			opt.Type = DataSpec_ArrayString

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			opt.Type = DataSpec_ArrayInt

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			opt.Type = DataSpec_ArrayUint

		case reflect.Float32, reflect.Float64:
			opt.Type = DataSpec_ArrayFloat

		case reflect.Bool:
			opt.Type = DataSpec_ArrayBool

		case reflect.Struct:
			opt.Type = DataSpec_ArrayStruct
			parseStruct(nil, opt, rt)

		case reflect.Pointer:
			if rt.Elem().Kind() == reflect.Struct {
				opt.Type = DataSpec_ArrayStruct
				parseStruct(nil, opt, rt.Elem())
			}
		}
	}

	parseStruct(spec, nil, rt)

	return spec, nil
}
