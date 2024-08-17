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

package lynkapi

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
	FieldSpec_Bool   = "bool"
	FieldSpec_Int    = "int"
	FieldSpec_Uint   = "uint"
	FieldSpec_Float  = "float"
	FieldSpec_String = "string"
	FieldSpec_Struct = "struct"
	FieldSpec_Bytes  = "bytes"

	fieldSpec_Array = "array:"

	// todo
	FieldSpec_StringTerm = "string_term"
	FieldSpec_StringText = "string_text"

	fieldSpec_Any        = "any"
	fieldSpec_AnyTypeUri = "google.golang.org/protobuf/types/known/structpb.Value"
)

var fieldSpecAttrs = map[string]int{
	"primary_key": 1,
	"unique_key":  1,
	"unique_keys": 2,

	"string_text": 1,

	"create_required": 1,
	"update_required": 1,

	"rows":      1,
	"data_rows": 1,

	"rand_hex":  2,
	"object_id": 2,
}

// var (
// 	FieldSpec_FuncAttrMatcher = regexp.MustCompile("^rand_hex\\(([0-9]+)\\)$")
// )

type FieldSpec_FuncAttr struct {
	name    string
	intArgs []int
}

type SpecSet struct {
	mu    sync.RWMutex
	specs map[string]*RegSpec
}

type RegSpec struct {
	Type reflect.Type
	Spec *TypeSpec
}

var specSet SpecSet

// FieldSpec_Array_Value `array:{value type}`
func specArrayType(t string) string {
	return fieldSpec_Array + t
}

// FieldSpec_Map_Key_Value `{key type}:{value type}`
func specMapType(keyType, valueType string) string {
	return keyType + ":" + valueType
}

func (it *TypeSpec) Field(name string) *FieldSpec {
	for _, field := range it.Fields {
		if field.TagName == name || field.Name == name {
			return field
		}
	}
	return nil
}

func (it *TypeSpec) Rows(data *structpb.Struct) (*FieldSpec, *structpb.ListValue) {
	if len(it.Fields) > 0 && len(data.Fields) > 0 {
		for _, fieldSpec := range it.Fields {
			if slices.Contains(fieldSpec.Attrs, "rows") &&
				strings.HasPrefix(fieldSpec.Type, fieldSpec_Array) {

				if rows, ok := data.Fields[fieldSpec.TagName]; ok {
					if lv := rows.GetListValue(); lv != nil {
						return fieldSpec, lv
					}
				}
			}
		}
	}
	return nil, nil
}

func (it *TypeSpec) DataMerge(dstObject, srcObject interface{}, opts ...interface{}) (bool, error) {
	return specDataMerge(it, dstObject, srcObject, opts...)
}

func (it *FieldSpec) Field(name string) *FieldSpec {
	for _, field := range it.Fields {
		if field.TagName == name || field.Name == name {
			return field
		}
	}
	return nil
}

func (it *FieldSpec) HasAttr(attr string) bool {
	t, ok := fieldSpecAttrs[attr]
	if !ok {
		return false
	}
	if t == 1 {
		return slices.ContainsFunc(it.Attrs, func(v string) bool {
			return strings.HasPrefix(v, attr)
		})
	}
	return slices.Contains(it.Attrs, attr)
}

func fieldSpecFuncAttrMatcher(attr string) *FieldSpec_FuncAttr {

	if len(attr) <= 3 {
		return nil
	}

	n := strings.Index(attr, "(")
	if n < 0 || attr[len(attr)-1] != ')' {
		return nil
	}

	var name = attr[:n]

	if t, ok := fieldSpecAttrs[name]; !ok || t != 2 {
		return nil
	}

	var args = strings.Split(attr[n+1:len(attr)-1], ",")

	switch name {
	case "rand_hex", "object_id":
		if len(args) != 1 {
			return nil
		}
		if argv, err := strconv.Atoi(args[0]); err == nil && argv >= 8 && argv <= 32 {
			return &FieldSpec_FuncAttr{
				name:    name,
				intArgs: []int{argv},
			}
		}
	}

	return nil
}

func (it *FieldSpec) FuncAttr(names ...string) *FieldSpec_FuncAttr {

	for _, name := range names {
		if t, ok := fieldSpecAttrs[name]; !ok || t != 2 {
			return nil
		}

		for _, attr := range it.Attrs {

			if !strings.HasPrefix(attr, name) {
				continue
			}

			if fa := fieldSpecFuncAttrMatcher(attr); fa != nil {
				return fa
			}
		}
	}

	return nil
}

func (it *FieldSpec) PrimaryKeys() ([]string, map[string]*FieldSpec, map[string]*FieldSpec) {
	if it.Type == specArrayType(FieldSpec_Struct) {
		var (
			ukeys = map[string]*FieldSpec{}
			pkeys []*FieldSpec
		)
		for _, field := range it.Fields {
			if field.HasAttr("primary_key") {
				pkeys = append(pkeys, field)
				ukeys[field.TagName] = field
			} else if field.HasAttr("unique_key") {
				ukeys[field.TagName] = field
			}
		}
		if len(pkeys) == 1 { // TODO multi primary-key
			return []string{pkeys[0].TagName}, map[string]*FieldSpec{
				pkeys[0].TagName: pkeys[0],
			}, ukeys
		}
	}
	return nil, nil, nil
}

func (it *TableSpec) PrimaryId(fields map[string]*structpb.Value) string {
	pkeys := []string{}
	for _, field := range it.Fields {
		if field.HasAttr("primary_key") {
			if v, ok := fields[field.TagName]; ok {
				pkeys = append(pkeys, v.GetStringValue())
			}
		}
	}
	if len(pkeys) == 1 { // TODO multi primary-key
		return pkeys[0]
	}
	return ""
}

func (it *FieldSpec) DataMerge(dstObject, srcObject interface{}, opts ...interface{}) (bool, error) {
	return specDataMerge(&TypeSpec{
		Fields: it.Fields,
	}, dstObject, srcObject, opts...)
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

func NewSpecFromStruct(obj interface{}) (*TypeSpec, reflect.Type, error) {

	if obj == nil {
		return nil, nil, fmt.Errorf("empty object")
	}

	rt := reflect.TypeOf(obj)
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}
	if rt.Kind() != reflect.Struct {
		return nil, nil, fmt.Errorf("invalid object type")
	}

	return parseSpecByStructType(rt)
}

func parseSpecByStructType(rt reflect.Type) (*TypeSpec, reflect.Type, error) {

	var (
		maxDepth = 10
		baseSpec = &FieldSpec{
			Kind: RefTypeKind(rt),
			Name: rt.Name(),
		}
		parseStruct func(depth int, pField *FieldSpec, rt reflect.Type)
		parseArray  func(depth int, pField *FieldSpec, rt reflect.Type)
		parseSet    = map[string]bool{}
	)

	parseTagValue := func(v string, t string) *structpb.Value {
		if v == "" {
			return nil
		}
		switch t {
		case FieldSpec_String:
			return structpb.NewStringValue(v)

		case FieldSpec_Int:
			if i, err := strconv.ParseInt(v, 10, 64); err == nil && i != 0 {
				return structpb.NewNumberValue(float64(i))
			}

		case FieldSpec_Uint:
			if i, err := strconv.ParseUint(v, 10, 64); err == nil && i != 0 {
				return structpb.NewNumberValue(float64(i))
			}

		case FieldSpec_Float:
			if f, err := strconv.ParseFloat(v, 64); err == nil && f != 0 {
				return structpb.NewNumberValue(f)
			}

		case FieldSpec_Bool:
			if b, err := strconv.ParseBool(v); err == nil && b == true {
				return structpb.NewBoolValue(b)
			}
		}

		return nil
	}

	parseValueLimits := func(field *FieldSpec, fd reflect.StructField) {

		if field.Type != FieldSpec_Int &&
			field.Type != FieldSpec_Uint &&
			field.Type != FieldSpec_Float {
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

	parseStruct = func(depth int, pField *FieldSpec, rt reflect.Type) {

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

			var field = &FieldSpec{
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
				field.Type = FieldSpec_String
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
				field.Type = FieldSpec_Int

			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				field.Type = FieldSpec_Uint

			case reflect.Float32, reflect.Float64:
				field.Type = FieldSpec_Float

			case reflect.Bool:
				field.Type = FieldSpec_Bool

			case reflect.Struct:
				field.Type = FieldSpec_Struct
				field.Kind = RefTypeKind(fd.Type)
				parseStruct(depth+1, field, fd.Type)

			case reflect.Pointer:
				switch fd.Type.Elem().Kind() {

				case reflect.Struct:
					field.Type = FieldSpec_Struct
					field.Kind = RefTypeKind(fd.Type.Elem())
					parseStruct(depth+1, field, fd.Type.Elem())
				}

			case reflect.Slice:
				if fd.Type.Elem().Kind() == reflect.Uint8 {
					field.Type = FieldSpec_Bytes
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
					mapKeyType = FieldSpec_String

				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					mapKeyType = FieldSpec_Int

				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					mapKeyType = FieldSpec_Uint
				}

				if mapKeyType != "" {

					switch mapValueType.Kind() {
					case reflect.String:
						field.Type = specMapType(mapKeyType, FieldSpec_String)

					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						field.Type = specMapType(mapKeyType, FieldSpec_Int)

					case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
						field.Type = specMapType(mapKeyType, FieldSpec_Uint)

					case reflect.Float32, reflect.Float64:
						field.Type = specMapType(mapKeyType, FieldSpec_Float)

					case reflect.Bool:
						field.Type = specMapType(mapKeyType, FieldSpec_Bool)

					case reflect.Pointer:
						mapValueType = mapValueType.Elem()
						switch mapValueType.Kind() {
						case reflect.Struct:
							if RefTypeKind(mapValueType) == fieldSpec_AnyTypeUri {
								field.Type = specMapType(mapKeyType, fieldSpec_Any)
							} else {
								field.Type = specMapType(mapKeyType, FieldSpec_Struct)
							}
							field.Kind = RefTypeKind(mapValueType)
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
					if fieldSpecAttrs[attr] == 1 ||
						fieldSpecFuncAttrMatcher(attr) != nil {
						field.Attrs = append(field.Attrs, attr)
					}
				}
			}
			pField.Fields = append(pField.Fields, field)
		}
	}

	parseArray = func(depth int, cField *FieldSpec, rt reflect.Type) {
		if depth > maxDepth {
			return
		}

		switch rt.Kind() {

		case reflect.String:
			cField.Type = specArrayType(FieldSpec_String)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			cField.Type = specArrayType(FieldSpec_Int)

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			cField.Type = specArrayType(FieldSpec_Uint)

		case reflect.Float32, reflect.Float64:
			cField.Type = specArrayType(FieldSpec_Float)

		case reflect.Bool:
			cField.Type = specArrayType(FieldSpec_Bool)

		case reflect.Struct:
			cField.Type = specArrayType(FieldSpec_Struct)
			cField.Kind = RefTypeKind(rt)
			parseStruct(depth+1, cField, rt)

		case reflect.Pointer:
			if rt.Elem().Kind() == reflect.Struct {
				cField.Type = specArrayType(FieldSpec_Struct)
				cField.Kind = RefTypeKind(rt.Elem())
				parseStruct(depth+1, cField, rt.Elem())
			}
		}
	}

	parseStruct(0, baseSpec, rt)

	return &TypeSpec{
		Name:   baseSpec.Name,
		Kind:   baseSpec.Kind,
		Fields: baseSpec.Fields,
	}, rt, nil
}

// func (it *FieldSpec) ReflectKindEqual(kind reflect.Kind) bool {
// 	switch kind {
// 	case reflect.Bool:
// 		return it.Type == FieldSpec_Bool
//
// 	case reflect.String:
// 		return it.Type == FieldSpec_String
// 	}
// 	return false
// }

func ParseStructValid(kind string, data *structpb.Struct) (interface{}, error) {
	spec := specSet.Get(kind)
	if spec == nil {
		return nil, fmt.Errorf("spec kind (%s) not found", kind)
	}
	js, _ := json.Marshal(data)

	obj := reflect.New(spec.Type).Interface()
	if err := json.Unmarshal(js, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (it *FieldSpec_FuncAttr) GenId() string {
	if len(it.intArgs) > 0 && it.intArgs[0] >= 8 && it.intArgs[0] <= 32 {
		switch it.name {
		case "rand_hex":
			return RandHexString(it.intArgs[0])
		case "object_id":
			return RandObjectId(it.intArgs[0])
		}
	}
	return RandHexString(16)
}
