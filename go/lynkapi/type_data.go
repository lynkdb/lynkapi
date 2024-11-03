package lynkapi

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"slices"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"
)

func specDataMerge(spec *TypeSpec, dstObject, srcObject any, opts ...any) (bool, error) {

	var (
		chg        = false
		dataMerge  func(spec *FieldSpec, dstValue, srcValue reflect.Value) error
		arrayMerge func(spec *FieldSpec, dstValue, srcValue reflect.Value) error
		mapMerge   func(spec *FieldSpec, dstValue, srcValue reflect.Value) error
		mergeType  DataMerge_Type
	)

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		switch opt.(type) {
		case DataMerge_Type:
			mergeType = opt.(DataMerge_Type)
		}
	}

	arrayMerge = func(fieldSpec *FieldSpec, dstValue, srcValue reflect.Value) error {

		if !dstValue.IsValid() {
			return nil
		}

		if !srcValue.IsValid() || srcValue.Kind() != reflect.Slice || srcValue.Len() == 0 {
			return nil
		}

		var (
			subType = fieldSpec.Type[len("array:"):]
		)

		for i := 0; i < srcValue.Len(); i++ {

			switch srcValue.Index(i).Kind() {
			case reflect.Bool:
				if subType != FieldSpec_Bool {
					return fmt.Errorf("invalid array:bool type")
				}

			case reflect.String:
				if subType != FieldSpec_String {
					return fmt.Errorf("invalid array:string type")
				}

			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if subType != FieldSpec_Int {
					return fmt.Errorf("invalid array:int type")
				}

			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if subType != FieldSpec_Uint {
					return fmt.Errorf("invalid array:uint type")
				}

			case reflect.Float32, reflect.Float64:
				if subType != FieldSpec_Float {
					return fmt.Errorf("invalid array:float type")
				}

			default:
				return fmt.Errorf("invalid array: type")
			}
		}

		chg = true
		dstValue.Set(srcValue)

		return nil
	}

	mapMerge = func(fieldSpec *FieldSpec, dstValue, srcValue reflect.Value) error {

		if !dstValue.IsValid() {
			return nil
		}

		if !srcValue.IsValid() || srcValue.Kind() != reflect.Map {
			return nil
		}

		subType := strings.Split(fieldSpec.Type, ":")
		if len(subType) != 2 || subType[0] != FieldSpec_String {
			return nil
		}

		for iter := srcValue.MapRange(); iter != nil && iter.Next(); {

			key := iter.Key()
			val := iter.Value()

			if key.Kind() != reflect.String {
				continue
			}

			if val.Kind() == reflect.Pointer {
				val = val.Elem()
			}

			switch val.Kind() {
			case reflect.Bool:
				if subType[1] != FieldSpec_Bool {
					return fmt.Errorf("invalid map:bool type")
				}

			case reflect.String:
				if subType[1] != FieldSpec_String {
					return fmt.Errorf("invalid map:string type")
				}

			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if subType[1] != FieldSpec_Int {
					return fmt.Errorf("invalid map:int type")
				}

			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if subType[1] != FieldSpec_Uint {
					return fmt.Errorf("invalid map:uint type")
				}

			case reflect.Float32, reflect.Float64:
				if subType[1] != FieldSpec_Float {
					return fmt.Errorf("invalid map:float type")
				}

			case reflect.Struct:
				if subType[1] != fieldSpec_Any &&
					subType[1] != FieldSpec_Struct {
					return fmt.Errorf("invalid map:any|struct type")
				}

			default:
				return fmt.Errorf("invalid map: type %v", val.Kind())
			}
		}

		chg = true
		dstValue.Set(srcValue)

		return nil
	}

	fieldValueLimitsInt := func(field *FieldSpec) (int64, int64, int64) {

		var (
			minValue int64 = math.MinInt64
			defValue int64 = 0
			maxValue int64 = math.MaxInt64
		)

		if len(field.Opts) > 0 && field.Type == FieldSpec_Int {

			if v, ok := field.Opts["def_value"]; ok {
				defValue = int64(v.GetNumberValue())
			}

			if v, ok := field.Opts["min_value"]; ok {
				minValue = int64(v.GetNumberValue())
			}

			if v, ok := field.Opts["max_value"]; ok {
				maxValue = int64(v.GetNumberValue())
			}
		}

		return minValue, defValue, maxValue
	}

	fieldValueLimitsUint := func(field *FieldSpec) (uint64, uint64, uint64) {

		var (
			minValue uint64 = 0
			defValue uint64 = 0
			maxValue uint64 = math.MaxUint64
		)

		if len(field.Opts) > 0 && field.Type == FieldSpec_Uint {

			if v, ok := field.Opts["def_value"]; ok {
				defValue = uint64(v.GetNumberValue())
			}

			if v, ok := field.Opts["min_value"]; ok {
				minValue = uint64(v.GetNumberValue())
			}

			if v, ok := field.Opts["max_value"]; ok {
				maxValue = uint64(v.GetNumberValue())
			}
		}

		return minValue, defValue, maxValue
	}

	fieldValueLimitsFloat := func(field *FieldSpec) (float64, float64, float64) {

		var (
			minValue float64 = math.SmallestNonzeroFloat64
			defValue float64 = 0
			maxValue float64 = math.MaxFloat64
		)

		if len(field.Opts) > 0 && field.Type == FieldSpec_Float {

			if v, ok := field.Opts["def_value"]; ok {
				defValue = v.GetNumberValue()
			}

			if v, ok := field.Opts["min_value"]; ok {
				minValue = v.GetNumberValue()
			}

			if v, ok := field.Opts["max_value"]; ok {
				maxValue = v.GetNumberValue()
			}
		}

		return minValue, defValue, maxValue
	}

	fieldValueLimitsString := func(field *FieldSpec) string {
		if len(field.Opts) > 0 && field.Type == FieldSpec_String {
			if v, ok := field.Opts["def_value"]; ok {
				return v.GetStringValue()
			}
		}
		return ""
	}

	requiredCheck := func(fieldSpec *FieldSpec) error {

		if mergeType != DataMerge_Create &&
			mergeType != DataMerge_Update {
			return nil
		}

		switch mergeType {
		case DataMerge_Create:
			if fieldSpec.HasAttr("create_required") {
				return fmt.Errorf("field %s update_required", fieldSpec.Name)
			}

		case DataMerge_Update:
			if fieldSpec.HasAttr("update_required") {
				return fmt.Errorf("field %s update_required", fieldSpec.Name)
			}
		}
		return nil
	}

	dataMerge = func(spec *FieldSpec, dstValue, srcValue reflect.Value) error {

		if spec == nil || len(spec.Fields) == 0 ||
			!dstValue.IsValid() ||
			!srcValue.IsValid() {
			return nil
		}

		if dstValue.Kind() == reflect.Pointer {
			dstValue = dstValue.Elem()
		}
		if dstValue.Kind() != reflect.Struct {
			return nil
		}

		if srcValue.Kind() == reflect.Pointer {
			srcValue = srcValue.Elem()
		}
		if srcValue.Kind() != reflect.Struct {
			return nil
		}

		for _, fieldSpec := range spec.Fields {

			var (
				value = srcValue.FieldByName(fieldSpec.Name)
			)

			if !value.IsValid() {
				value = srcValue.FieldByName(fieldSpec.TagName)
			}

			if !value.IsValid() {
				if err := requiredCheck(fieldSpec); err != nil {
					return err
				}
				continue
			}

			dstField := dstValue.FieldByName(fieldSpec.Name)
			if !dstField.CanSet() {
				continue
			}

			switch value.Kind() {
			case reflect.Bool:
				if fieldSpec.Type != FieldSpec_Bool {
					return fmt.Errorf("invalid field (%s) type (bool:%s)", fieldSpec.Name, fieldSpec.Type)
				}
				if dstField.Kind() != reflect.Bool {
					return fmt.Errorf("invalid field (%s) type (string:%v)", fieldSpec.Name, dstField.Kind())
				}
				if value.Bool() != dstField.Bool() {
					chg = true
					dstField.SetBool(value.Bool())
				}

			case reflect.String:
				if fieldSpec.Type != FieldSpec_String {
					return fmt.Errorf("invalid field (%s) type (string:%s)", fieldSpec.Name, fieldSpec.Type)
				}
				if dstField.Kind() != reflect.String {
					return fmt.Errorf("invalid field (%s) type (string:%v)", fieldSpec.Name, dstField.Kind())
				}
				defValue := fieldValueLimitsString(fieldSpec)
				if value.String() != "" {
					if len(fieldSpec.Enums) > 0 && !slices.Contains(fieldSpec.Enums, value.String()) {
						return fmt.Errorf("field (%s), deny by enums", fieldSpec.Name)
					}
					if dstField.String() != value.String() {
						chg = true
						dstField.SetString(value.String())
					}
				} else if dstField.String() == "" && defValue != "" {
					chg = true
					dstField.SetString(defValue)
				} else {
					if err := requiredCheck(fieldSpec); err != nil {
						return err
					}
				}

			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if fieldSpec.Type != FieldSpec_Int {
					return fmt.Errorf("invalid field (%s) type (int:%v)", fieldSpec.Name, dstField.Kind())
				}
				minValue, defValue, maxValue := fieldValueLimitsInt(fieldSpec)
				if value.Int() != 0 {
					if value.Int() < minValue || value.Int() > maxValue {
						return fmt.Errorf("field (%s) deny value limits [%d ~ %d]", fieldSpec.Name, minValue, maxValue)
					}
					if dstField.Int() != value.Int() {
						chg = true
						dstField.SetInt(value.Int())
					}
				} else if dstField.Int() == 0 && defValue != 0 && defValue >= minValue && defValue <= maxValue {
					chg = true
					dstField.SetInt(defValue)
				} else {
					if err := requiredCheck(fieldSpec); err != nil {
						return err
					}
				}

			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if fieldSpec.Type != FieldSpec_Uint {
					return fmt.Errorf("invalid field (%s) type (uint:%v)", fieldSpec.Name, dstField.Kind())
				}
				minValue, defValue, maxValue := fieldValueLimitsUint(fieldSpec)
				if value.Uint() != 0 {
					if value.Uint() < minValue || value.Uint() > maxValue {
						return fmt.Errorf("field (%s) deny value limits [%d ~ %d]", fieldSpec.Name, minValue, maxValue)
					}
					if dstField.Uint() != value.Uint() {
						chg = true
						dstField.SetUint(value.Uint())
					}
				} else if dstField.Uint() == 0 && defValue != 0 && defValue >= minValue && defValue <= maxValue {
					chg = true
					dstField.SetUint(defValue)
				} else {
					if err := requiredCheck(fieldSpec); err != nil {
						return err
					}
				}

			case reflect.Float32, reflect.Float64:
				if fieldSpec.Type != FieldSpec_Float {
					return fmt.Errorf("invalid field (%s) type (float:%v)", fieldSpec.Name, dstField.Kind())
				}
				minValue, defValue, maxValue := fieldValueLimitsFloat(fieldSpec)
				if value.Float() != 0 {
					if value.Float() < minValue || value.Float() > maxValue {
						return fmt.Errorf("field (%s) deny value limits [%f ~ %f]", fieldSpec.Name, minValue, maxValue)
					}
					if dstField.Float() != value.Float() {
						chg = true
						dstField.SetFloat(value.Float())
					}
				} else if dstField.Float() == 0 && defValue != 0 && defValue >= minValue && defValue <= maxValue {
					chg = true
					dstField.SetFloat(defValue)
				} else {
					if err := requiredCheck(fieldSpec); err != nil {
						return err
					}
				}

			case reflect.Slice:
				if !strings.HasPrefix(fieldSpec.Type, "array:") {
					return fmt.Errorf("invalid field (%s) type (array:%s)", fieldSpec.Name, fieldSpec.Type)
				}
				if dstField.Kind() != reflect.Slice {
					return fmt.Errorf("invalid field (%s) type (array:%s)", fieldSpec.Name, fieldSpec.Type)
				}
				if err := arrayMerge(fieldSpec, dstField, value); err != nil {
					return err
				}

			case reflect.Map:
				if strings.Count(fieldSpec.Type, ":") != 1 {
					return fmt.Errorf("invalid field (%s) type (map:%s)", fieldSpec.Name, fieldSpec.Type)
				}
				if dstField.Kind() != reflect.Map {
					return fmt.Errorf("invalid field (%s) type (map:%s)", fieldSpec.Name, fieldSpec.Type)
				}
				if err := mapMerge(fieldSpec, dstField, value); err != nil {
					return err
				}

			case reflect.Pointer, reflect.Struct:
				if fieldSpec.Type != FieldSpec_Struct {
					return fmt.Errorf("invalid field (%s) type (struct:%s)", fieldSpec.Name, fieldSpec.Type)
				}
				if dstField.Kind() != reflect.Pointer && dstField.Kind() != reflect.Struct {
					return fmt.Errorf("invalid field (%s) type (struct:%s)", fieldSpec.Name, fieldSpec.Type)
				}
				if dstField.Kind() == reflect.Pointer {
					if dstField.IsNil() {
						dstField.Set(reflect.New(dstField.Type().Elem()))
					}

					if err := dataMerge(fieldSpec, dstField, value); err != nil {
						return err
					}
				} else if dstField.Kind() == reflect.Struct {
					if err := dataMerge(fieldSpec, dstField, value); err != nil {
						return err
					}
				}
			}
		}

		return nil
	}

	err := dataMerge(&FieldSpec{
		Fields: spec.Fields,
	}, reflect.ValueOf(dstObject), reflect.ValueOf(srcObject))

	return chg, err
}

func NewRequestFromObject(serviceName, methodName string, obj any) (*Request, error) {
	js, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	var data structpb.Struct
	if err = json.Unmarshal(js, &data); err != nil {
		return nil, err
	}
	return &Request{
		ServiceName: serviceName,
		MethodName:  methodName,
		Data:        &data,
	}, nil
}

func ConvertReflectValueToApiValue(value reflect.Value) (*structpb.Value, error) {
	js, _ := json.Marshal(value.Interface())
	var pbStruct structpb.Struct
	if err := json.Unmarshal(js, &pbStruct); err != nil {
		return nil, err
	}
	return structpb.NewStructValue(&pbStruct), nil
}

func ConvertReflectValueToMapValue(value reflect.Value) (map[string]*structpb.Value, error) {
	js, _ := json.Marshal(value.Interface())
	var fields = map[string]*structpb.Value{}
	if err := json.Unmarshal(js, &fields); err != nil {
		return nil, err
	}
	return fields, nil
}

func DecodeStruct(data *structpb.Struct, obj any) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(js, &obj)
}

func dataUpdate(spec *TypeSpec, baseObject any, updateData *structpb.Struct) error {

	var (
		dataUpdate  func(spec *TypeSpec, baseValue reflect.Value, updateData *structpb.Struct) error
		arrayUpdate func(fieldSpec *FieldSpec, baseValue reflect.Value, updateData *structpb.ListValue) error
	)

	arrayUpdate = func(fieldSpec *FieldSpec, baseValue reflect.Value, updateData *structpb.ListValue) error {
		if updateData == nil || len(updateData.Values) == 0 {
			return nil
		}
		subType := fieldSpec.Type[6:]

		ls := []reflect.Value{}
		switch subType {

		case FieldSpec_String:
			for _, v := range updateData.Values {
				switch v.Kind.(type) {
				case *structpb.Value_StringValue:
					ls = append(ls, reflect.ValueOf(v.GetStringValue()))
				default:
					return fmt.Errorf("invalid array type")
				}
			}

		case FieldSpec_Int:
			for _, v := range updateData.Values {
				switch v.Kind.(type) {
				case *structpb.Value_NumberValue:
					ls = append(ls, reflect.ValueOf(int64(v.GetNumberValue())))
				default:
					return fmt.Errorf("invalid int type")
				}
			}

		case FieldSpec_Uint:
			for _, v := range updateData.Values {
				switch v.Kind.(type) {
				case *structpb.Value_NumberValue:
					ls = append(ls, reflect.ValueOf(uint64(v.GetNumberValue())))
				default:
					return fmt.Errorf("invalid int type")
				}
			}

		case FieldSpec_Float:
			for _, v := range updateData.Values {
				switch v.Kind.(type) {
				case *structpb.Value_NumberValue:
					ls = append(ls, reflect.ValueOf(v.GetNumberValue()))
				default:
					return fmt.Errorf("invalid float type")
				}
			}
		}

		if len(ls) > 0 {
			baseValue.Clear()
			baseValue.Set(reflect.Append(baseValue, ls...))
		}

		return nil
	}

	dataUpdate = func(spec *TypeSpec, baseValue reflect.Value, updateData *structpb.Struct) error {

		if spec == nil || len(spec.Fields) == 0 ||
			!baseValue.IsValid() ||
			updateData == nil || len(updateData.Fields) == 0 {
			return nil
		}

		if baseValue.Kind() == reflect.Pointer {
			baseValue = baseValue.Elem()
		}

		if baseValue.Kind() != reflect.Struct {
			return nil
		}

		for name, value := range updateData.Fields {

			if value == nil {
				continue
			}

			fieldSpec := spec.Field(name)
			if fieldSpec == nil {
				continue
			}

			dstField := baseValue.FieldByName(fieldSpec.Name)
			if !dstField.CanSet() {
				continue
			}

			switch value.Kind.(type) {
			case *structpb.Value_BoolValue:
				if fieldSpec.Type != FieldSpec_Bool {
					return fmt.Errorf("invalid field (%s) type (bool:%s)", name, fieldSpec.Type)
				}
				if dstField.Kind() != reflect.Bool {
					return fmt.Errorf("invalid field (%s) type (string:%v)", name, dstField.Kind())
				}
				dstField.SetBool(value.GetBoolValue())

			case *structpb.Value_StringValue:
				if fieldSpec.Type != FieldSpec_String {
					return fmt.Errorf("invalid field (%s) type (string:%s)", name, fieldSpec.Type)
				}
				if value.GetStringValue() != "" {
					if dstField.Kind() != reflect.String {
						return fmt.Errorf("invalid field (%s) type (string:%v)", name, dstField.Kind())
					}
					dstField.SetString(value.GetStringValue())
				}

			case *structpb.Value_NumberValue:
				switch fieldSpec.Type {
				case FieldSpec_Int, FieldSpec_Uint, FieldSpec_Float:
					if value.GetNumberValue() != 0 {
						switch dstField.Kind() {
						case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
							dstField.SetInt(int64(value.GetNumberValue()))
						case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
							dstField.SetUint(uint64(value.GetNumberValue()))
						case reflect.Float32, reflect.Float64:
							dstField.SetFloat(value.GetNumberValue())
						default:
							return fmt.Errorf("invalid field (%s) type (string:%v)", name, dstField.Kind())
						}
					}

				default:
					return fmt.Errorf("invalid field (%s) type (number:%s)", name, fieldSpec.Type)
				}

			case *structpb.Value_ListValue:
				if !strings.HasPrefix(fieldSpec.Type, "array:") {
					return fmt.Errorf("invalid field (%s) type (struct:%s)", name, fieldSpec.Type)
				}
				if dstField.Kind() != reflect.Slice {
					return fmt.Errorf("invalid field (%s) type (struct:%s)", name, fieldSpec.Type)
				}

				if err := arrayUpdate(fieldSpec, dstField, value.GetListValue()); err != nil {
					return err
				}

			case *structpb.Value_StructValue:
				if fieldSpec.Type != FieldSpec_Struct {
					return fmt.Errorf("invalid field (%s) type (struct:%s)", name, fieldSpec.Type)
				}
				if dstField.Kind() != reflect.Pointer && dstField.Kind() != reflect.Struct {
					return fmt.Errorf("invalid field (%s) type (struct:%s)", name, fieldSpec.Type)
				}
				if dstField.Kind() == reflect.Pointer {
					if dstField.IsNil() {
						dstField.Set(reflect.New(dstField.Type().Elem()))
					}

					if err := dataUpdate(&TypeSpec{
						Fields: fieldSpec.Fields,
					}, dstField, value.GetStructValue()); err != nil {
						return err
					}
				} else if dstField.Kind() == reflect.Struct {
					if err := dataUpdate(&TypeSpec{
						Fields: fieldSpec.Fields,
					}, dstField, value.GetStructValue()); err != nil {
						return err
					}
				}

			default:
				//
			}
		}

		return nil
	}

	return dataUpdate(spec, reflect.ValueOf(baseObject), updateData)
}
