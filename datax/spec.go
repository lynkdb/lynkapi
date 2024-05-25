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
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"

	"google.golang.org/protobuf/types/known/structpb"
)

const (
	SpecField_Bool   = "bool"
	SpecField_Int    = "int"
	SpecField_Uint   = "uint"
	SpecField_Float  = "float"
	SpecField_String = "string"
	SpecField_Struct = "struct"
	SpecField_Bytes  = "bytes"
)

var specFieldAttrs = map[string]bool{
	"primary_key":     true,
	"create_required": true,
	"update_required": true,
	"rows":            true,
	"data_rows":       true,
}

type SpecSet struct {
	mu    sync.RWMutex
	specs map[string]*RegSpec
}

type RegSpec struct {
	Type reflect.Type
	Spec *Spec
}

var specSet SpecSet

func (it *Spec) Field(name string) *SpecField {
	for _, field := range it.Fields {
		if field.TagName == name || field.Name == name {
			return field
		}
	}
	return nil
}

func (it *Spec) Rows(data *structpb.Struct) (*SpecField, *structpb.ListValue) {
	if len(it.Fields) > 0 && len(data.Fields) > 0 {
		var field *SpecField
		for _, specField := range it.Fields {
			if slices.Contains(specField.Attrs, "rows") &&
				strings.HasPrefix(specField.Type, "array:") {
				field = specField
				break
			}
		}
		if field != nil {
			if rows, ok := data.Fields[field.TagName]; ok {
				if lv := rows.GetListValue(); lv != nil {
					return field, lv
				}
			}
		}
	}
	return nil, nil
}

func (it *Spec) DataMerge(dstObject, srcObject interface{}, opts ...interface{}) (bool, error) {
	return specDataMerge(it, dstObject, srcObject)
}

func (it *SpecSet) Register(o interface{}) error {
	spec, rtype, err := NewSpecFromStruct(o)
	if err != nil {
		return err
	}

	it.mu.Lock()
	defer it.mu.Unlock()

	if it.specs == nil {
		it.specs = map[string]*RegSpec{}
	}

	if _, ok := it.specs[spec.Kind]; !ok {
		it.specs[spec.Kind] = &RegSpec{
			Type: rtype,
			Spec: spec,
		}
	}

	return nil
}

func RegisterSpec(o interface{}) error {
	return specSet.Register(o)
}

func (it *SpecSet) Get(kind string) *RegSpec {
	it.mu.Lock()
	defer it.mu.Unlock()
	if spec, ok := it.specs[kind]; ok {
		return spec
	}
	return nil
}

// SpecField_Array_Value `array:{value type}`
func specArrayType(t string) string {
	return "array:" + t
}

// SpecField_Map_Key_Value `{key type}:{value type}`
func specMapType(keyType, valueType string) string {
	return keyType + ":" + valueType
}

func NewSpecFromStruct(obj interface{}) (*Spec, reflect.Type, error) {

	if obj == nil {
		return nil, nil, fmt.Errorf("empty object")
	}

	rt := reflect.TypeOf(obj)
	if rt.Kind() != reflect.Struct {
		return nil, nil, fmt.Errorf("invalid object type")
	}

	return parseSpecByStructType(rt)
}

func parseSpecByStructType(rt reflect.Type) (*Spec, reflect.Type, error) {

	var (
		maxDepth = 10
		baseSpec = &SpecField{
			Kind: refTypeKind(rt),
			Name: rt.Name(),
		}
		parseStruct func(depth int, pField *SpecField, rt reflect.Type)
		parseArray  func(depth int, pField *SpecField, rt reflect.Type)
		parseSet    = map[string]bool{}
	)

	parseTagValue := func(v string, t string) *structpb.Value {
		if v == "" {
			return nil
		}
		switch t {
		case SpecField_String:
			return structpb.NewStringValue(v)

		case SpecField_Int:
			if i, err := strconv.ParseInt(v, 10, 64); err == nil && i != 0 {
				return structpb.NewNumberValue(float64(i))
			}

		case SpecField_Uint:
			if i, err := strconv.ParseUint(v, 10, 64); err == nil && i != 0 {
				return structpb.NewNumberValue(float64(i))
			}

		case SpecField_Float:
			if f, err := strconv.ParseFloat(v, 64); err == nil && f != 0 {
				return structpb.NewNumberValue(f)
			}

		case SpecField_Bool:
			if b, err := strconv.ParseBool(v); err == nil && b == true {
				return structpb.NewBoolValue(b)
			}
		}

		return nil
	}

	parseValueLimits := func(field *SpecField, fd reflect.StructField) {

		if field.Type != SpecField_Int &&
			field.Type != SpecField_Uint &&
			field.Type != SpecField_Float {
			return
		}

		vlimits := strings.TrimSpace(fd.Tag.Get("x_value_limits"))
		if vlimits == "" {
			return
		}
		ar := strings.Split(vlimits, ",")
		if len(ar) == 1 {
			defValue := parseTagValue(ar[1], field.Type)
			if defValue == nil {
				return
			}

			if field.Opts == nil {
				field.Opts = map[string]*structpb.Value{}
			}

			field.Opts["def_value"] = defValue

		} else if len(ar) == 3 {

			var (
				minValue = parseTagValue(ar[0], field.Type)
				defValue = parseTagValue(ar[1], field.Type)
				maxValue = parseTagValue(ar[2], field.Type)
			)

			if minValue == nil || defValue == nil || maxValue == nil {
				return
			}

			if minValue.GetNumberValue() > maxValue.GetNumberValue() ||
				defValue.GetNumberValue() < minValue.GetNumberValue() ||
				defValue.GetNumberValue() > maxValue.GetNumberValue() {
				return
			}

			if field.Opts == nil {
				field.Opts = map[string]*structpb.Value{}
			}

			field.Opts["min_value"] = minValue
			field.Opts["def_value"] = defValue
			field.Opts["max_value"] = maxValue
		}
	}

	parseStruct = func(depth int, pField *SpecField, rt reflect.Type) {

		if depth > maxDepth && rt.Kind() != reflect.Struct {
			return
		}

		if pField.Kind != "" {
			if _, ok := parseSet[pField.Kind]; ok {
				return
			}
			parseSet[pField.Kind] = true
			defer func() {
				delete(parseSet, pField.Kind)
			}()
		}

		var (
			fields = reflect.VisibleFields(rt)
			names  []string
		)

		for _, fd := range fields {

			if fd.Anonymous || fd.Name == "" ||
				fd.Name[0] < 'A' || fd.Name[0] > 'Z' {
				continue
			}

			var field = &SpecField{
				Name:    fd.Name,
				TagName: fd.Name,
			}

			if s := fd.Tag.Get("json"); s != "" {
				sa := strings.Split(s, ",")
				if sa[0] != "" && sa[0] != "omitempty" {
					field.TagName = sa[0]
				}
			}

			if slices.Contains(names, field.Name) {
				continue
			}
			names = append(names, field.Name)

			switch fd.Type.Kind() {

			case reflect.String:
				field.Type = SpecField_String
				if es := fd.Tag.Get("x_enums"); es != "" {
					ar := strings.Split(es, ",")
					m := map[string]bool{}
					for _, v := range ar {
						if _, ok := m[v]; !ok {
							field.Enums = append(field.Enums, v)
							m[v] = true
						}
					}
				}

			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				field.Type = SpecField_Int

			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				field.Type = SpecField_Uint

			case reflect.Float32, reflect.Float64:
				field.Type = SpecField_Float

			case reflect.Bool:
				field.Type = SpecField_Bool

			case reflect.Struct:
				field.Type = SpecField_Struct
				field.Kind = refTypeKind(fd.Type)
				parseStruct(depth+1, field, fd.Type)

			case reflect.Pointer:
				switch fd.Type.Elem().Kind() {

				case reflect.Struct:
					field.Type = SpecField_Struct
					field.Kind = refTypeKind(fd.Type.Elem())
					parseStruct(depth+1, field, fd.Type.Elem())
				}

			case reflect.Slice:
				if fd.Type.Elem().Kind() == reflect.Uint8 {
					field.Type = SpecField_Bytes
				} else {
					parseArray(depth+1, field, fd.Type.Elem())
				}

			case reflect.Map:

				var (
					mapKeyType   string
					mapValueType = fd.Type.Elem()
				)

				switch fd.Type.Key().Kind() {
				case reflect.String:
					mapKeyType = SpecField_String

				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					mapKeyType = SpecField_Int

				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					mapKeyType = SpecField_Uint
				}

				if mapKeyType != "" {

					switch mapValueType.Kind() {
					case reflect.String:
						field.Type = specMapType(mapKeyType, SpecField_String)

					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						field.Type = specMapType(mapKeyType, SpecField_Int)

					case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
						field.Type = specMapType(mapKeyType, SpecField_Uint)

					case reflect.Float32, reflect.Float64:
						field.Type = specMapType(mapKeyType, SpecField_Float)

					case reflect.Bool:
						field.Type = specMapType(mapKeyType, SpecField_Bool)

					case reflect.Pointer:
						mapValueType = mapValueType.Elem()
						switch mapValueType.Kind() {
						case reflect.Struct:
							field.Type = specMapType(mapKeyType, SpecField_Struct)
							field.Kind = refTypeKind(mapValueType)
							parseStruct(depth+1, field, mapValueType)
						}
					}
				}
			}

			if field.Type == "" {
				continue
			}

			parseValueLimits(field, fd)

			if fd.Tag.Get("x_attrs") != "" {
				attrs := strings.Split(fd.Tag.Get("x_attrs"), ",")
				for _, attr := range attrs {
					if _, ok := specFieldAttrs[attr]; ok {
						field.Attrs = append(field.Attrs, attr)
					}
				}
			}
			pField.Fields = append(pField.Fields, field)
		}
	}

	parseArray = func(depth int, cField *SpecField, rt reflect.Type) {
		if depth > maxDepth {
			return
		}

		switch rt.Kind() {

		case reflect.String:
			cField.Type = specArrayType(SpecField_String)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			cField.Type = specArrayType(SpecField_Int)

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			cField.Type = specArrayType(SpecField_Uint)

		case reflect.Float32, reflect.Float64:
			cField.Type = specArrayType(SpecField_Float)

		case reflect.Bool:
			cField.Type = specArrayType(SpecField_Bool)

		case reflect.Struct:
			cField.Type = specArrayType(SpecField_Struct)
			cField.Kind = refTypeKind(rt)
			parseStruct(depth+1, cField, rt)

		case reflect.Pointer:
			if rt.Elem().Kind() == reflect.Struct {
				cField.Type = specArrayType(SpecField_Struct)
				cField.Kind = refTypeKind(rt.Elem())
				parseStruct(depth+1, cField, rt.Elem())
			}
		}
	}

	parseStruct(0, baseSpec, rt)

	return &Spec{
		Name:   baseSpec.Name,
		Kind:   baseSpec.Kind,
		Fields: baseSpec.Fields,
	}, rt, nil
}

// func (it *SpecField) ReflectKindEqual(kind reflect.Kind) bool {
// 	switch kind {
// 	case reflect.Bool:
// 		return it.Type == SpecField_Bool
//
// 	case reflect.String:
// 		return it.Type == SpecField_String
// 	}
// 	return false
// }

func ParseStructValid(kind string, data *structpb.Struct) (interface{}, error) {
	spec := specSet.Get(kind)
	if spec == nil {
		return nil, fmt.Errorf("spec kind (%s) not found", kind)
	}
	js, _ := json.Marshal(data)
	fmt.Println(string(js))
	fmt.Println(spec.Type)

	obj := reflect.New(spec.Type).Interface()
	fmt.Println(obj)
	if err := json.Unmarshal(js, obj); err != nil {
		return nil, err
	}
	return obj, nil
}
