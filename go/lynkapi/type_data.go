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

func specDataMerge(spec *Spec, dstObject, srcObject interface{}, opts ...interface{}) (bool, error) {

	var (
		chg        = false
		dataMerge  func(spec *SpecField, dstValue, srcValue reflect.Value) error
		arrayMerge func(spec *SpecField, dstValue, srcValue reflect.Value) error
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

	arrayMerge = func(specField *SpecField, dstValue, srcValue reflect.Value) error {

		if !dstValue.IsValid() {
			return nil
		}

		if !srcValue.IsValid() || srcValue.Kind() != reflect.Slice || srcValue.Len() == 0 {
			return nil
		}

		var (
			subType = specField.Type[len("array:"):]
		)

		for i := 0; i < srcValue.Len(); i++ {

			switch srcValue.Index(i).Kind() {
			case reflect.Bool:
				if subType != SpecField_Bool {
					return fmt.Errorf("invalid array:bool type")
				}

			case reflect.String:
				if subType != SpecField_String {
					return fmt.Errorf("invalid array:string type")
				}

			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if subType != SpecField_Int {
					return fmt.Errorf("invalid array:int type")
				}

			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if subType != SpecField_Uint {
					return fmt.Errorf("invalid array:uint type")
				}

			case reflect.Float32, reflect.Float64:
				if subType != SpecField_Float {
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

	fieldValueLimitsInt := func(field *SpecField) (int64, int64, int64) {

		var (
			minValue int64 = math.MinInt64
			defValue int64 = 0
			maxValue int64 = math.MaxInt64
		)

		if len(field.Opts) > 0 && field.Type == SpecField_Int {

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

	fieldValueLimitsUint := func(field *SpecField) (uint64, uint64, uint64) {

		var (
			minValue uint64 = 0
			defValue uint64 = 0
			maxValue uint64 = math.MaxUint64
		)

		if len(field.Opts) > 0 && field.Type == SpecField_Uint {

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

	fieldValueLimitsFloat := func(field *SpecField) (float64, float64, float64) {

		var (
			minValue float64 = math.SmallestNonzeroFloat64
			defValue float64 = 0
			maxValue float64 = math.MaxFloat64
		)

		if len(field.Opts) > 0 && field.Type == SpecField_Float {

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

	fieldValueLimitsString := func(field *SpecField) string {
		if len(field.Opts) > 0 && field.Type == SpecField_String {
			if v, ok := field.Opts["def_value"]; ok {
				return v.GetStringValue()
			}
		}
		return ""
	}

	requiredCheck := func(specField *SpecField) error {

		if mergeType != DataMerge_Create &&
			mergeType != DataMerge_Update {
			return nil
		}

		switch mergeType {
		case DataMerge_Create:
			if specField.HasAttr("create_required") {
				return fmt.Errorf("field %s update_required", specField.Name)
			}

		case DataMerge_Update:
			if specField.HasAttr("update_required") {
				return fmt.Errorf("field %s update_required", specField.Name)
			}
		}
		return nil
	}

	dataMerge = func(spec *SpecField, dstValue, srcValue reflect.Value) error {

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

		for _, specField := range spec.Fields {

			var (
				value = srcValue.FieldByName(specField.Name)
			)

			if !value.IsValid() {
				if err := requiredCheck(specField); err != nil {
					return err
				}
				continue
			}

			dstField := dstValue.FieldByName(specField.Name)
			if !dstField.CanSet() {
				continue
			}

			// fmt.Println("field", specField.Name, "dstField", dstField, value)

			switch value.Kind() {
			case reflect.Bool:
				if specField.Type != SpecField_Bool {
					return fmt.Errorf("invalid field (%s) type (bool:%s)", specField.Name, specField.Type)
				}
				if dstField.Kind() != reflect.Bool {
					return fmt.Errorf("invalid field (%s) type (string:%v)", specField.Name, dstField.Kind())
				}
				if value.Bool() != dstField.Bool() {
					chg = true
					dstField.SetBool(value.Bool())
				}

			case reflect.String:
				if specField.Type != SpecField_String {
					return fmt.Errorf("invalid field (%s) type (string:%s)", specField.Name, specField.Type)
				}
				if dstField.Kind() != reflect.String {
					return fmt.Errorf("invalid field (%s) type (string:%v)", specField.Name, dstField.Kind())
				}
				defValue := fieldValueLimitsString(specField)
				if value.String() != "" {
					if len(specField.Enums) > 0 && !slices.Contains(specField.Enums, value.String()) {
						return fmt.Errorf("field (%s), deny by enums", specField.Name)
					}
					if dstField.String() != value.String() {
						chg = true
						dstField.SetString(value.String())
					}
				} else if dstField.String() == "" && defValue != "" {
					chg = true
					dstField.SetString(defValue)
				} else {
					if err := requiredCheck(specField); err != nil {
						return err
					}
				}

			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if specField.Type != SpecField_Int {
					return fmt.Errorf("invalid field (%s) type (int:%v)", specField.Name, dstField.Kind())
				}
				minValue, defValue, maxValue := fieldValueLimitsInt(specField)
				if value.Int() != 0 {
					if value.Int() < minValue || value.Int() > maxValue {
						return fmt.Errorf("field (%s) deny value limits [%d ~ %d]", specField.Name, minValue, maxValue)
					}
					if dstField.Int() != value.Int() {
						chg = true
						dstField.SetInt(value.Int())
					}
				} else if dstField.Int() == 0 && defValue != 0 && defValue >= minValue && defValue <= maxValue {
					chg = true
					dstField.SetInt(defValue)
				} else {
					if err := requiredCheck(specField); err != nil {
						return err
					}
				}

			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if specField.Type != SpecField_Uint {
					return fmt.Errorf("invalid field (%s) type (uint:%v)", specField.Name, dstField.Kind())
				}
				minValue, defValue, maxValue := fieldValueLimitsUint(specField)
				if value.Uint() != 0 {
					if value.Uint() < minValue || value.Uint() > maxValue {
						return fmt.Errorf("field (%s) deny value limits [%d ~ %d]", specField.Name, minValue, maxValue)
					}
					if dstField.Uint() != value.Uint() {
						chg = true
						dstField.SetUint(value.Uint())
					}
				} else if dstField.Uint() == 0 && defValue != 0 && defValue >= minValue && defValue <= maxValue {
					chg = true
					dstField.SetUint(defValue)
				} else {
					if err := requiredCheck(specField); err != nil {
						return err
					}
				}

			case reflect.Float32, reflect.Float64:
				if specField.Type != SpecField_Float {
					return fmt.Errorf("invalid field (%s) type (float:%v)", specField.Name, dstField.Kind())
				}
				minValue, defValue, maxValue := fieldValueLimitsFloat(specField)
				if value.Float() != 0 {
					if value.Float() < minValue || value.Float() > maxValue {
						return fmt.Errorf("field (%s) deny value limits [%f ~ %f]", specField.Name, minValue, maxValue)
					}
					if dstField.Float() != value.Float() {
						chg = true
						dstField.SetFloat(value.Float())
					}
				} else if dstField.Float() == 0 && defValue != 0 && defValue >= minValue && defValue <= maxValue {
					chg = true
					dstField.SetFloat(defValue)
				} else {
					if err := requiredCheck(specField); err != nil {
						return err
					}
				}

			case reflect.Slice:
				if !strings.HasPrefix(specField.Type, "array:") {
					return fmt.Errorf("invalid field (%s) type (array:%s)", specField.Name, specField.Type)
				}
				if dstField.Kind() != reflect.Slice {
					return fmt.Errorf("invalid field (%s) type (array:%s)", specField.Name, specField.Type)
				}

				if err := arrayMerge(specField, dstField, value); err != nil {
					return err
				}

			case reflect.Pointer, reflect.Struct:
				if specField.Type != SpecField_Struct {
					return fmt.Errorf("invalid field (%s) type (struct:%s)", specField.Name, specField.Type)
				}
				if dstField.Kind() != reflect.Pointer && dstField.Kind() != reflect.Struct {
					return fmt.Errorf("invalid field (%s) type (struct:%s)", specField.Name, specField.Type)
				}
				if dstField.Kind() == reflect.Pointer {
					if dstField.IsNil() {
						dstField.Set(reflect.New(dstField.Type().Elem()))
					}

					if err := dataMerge(specField, dstField, value); err != nil {
						return err
					}
				} else if dstField.Kind() == reflect.Struct {
					if err := dataMerge(specField, dstField, value); err != nil {
						return err
					}
				}
			}
		}

		return nil
	}

	err := dataMerge(&SpecField{
		Fields: spec.Fields,
	}, reflect.ValueOf(dstObject), reflect.ValueOf(srcObject))

	return chg, err
}

func NewRequestFromObject(serviceName, methodName string, obj interface{}) (*Request, error) {
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

func DecodeStruct(data *structpb.Struct, obj interface{}) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(js, &obj)
}

func dataUpdate(spec *Spec, baseObject interface{}, updateData *structpb.Struct) error {

	var (
		dataUpdate  func(spec *Spec, baseValue reflect.Value, updateData *structpb.Struct) error
		arrayUpdate func(specField *SpecField, baseValue reflect.Value, updateData *structpb.ListValue) error
	)

	arrayUpdate = func(specField *SpecField, baseValue reflect.Value, updateData *structpb.ListValue) error {
		if updateData == nil || len(updateData.Values) == 0 {
			return nil
		}
		subType := specField.Type[6:]

		ls := []reflect.Value{}
		switch subType {

		case SpecField_String:
			for _, v := range updateData.Values {
				switch v.Kind.(type) {
				case *structpb.Value_StringValue:
					ls = append(ls, reflect.ValueOf(v.GetStringValue()))
				default:
					return fmt.Errorf("invalid array type")
				}
			}

		case SpecField_Int:
			for _, v := range updateData.Values {
				switch v.Kind.(type) {
				case *structpb.Value_NumberValue:
					ls = append(ls, reflect.ValueOf(int64(v.GetNumberValue())))
				default:
					return fmt.Errorf("invalid int type")
				}
			}

		case SpecField_Uint:
			for _, v := range updateData.Values {
				switch v.Kind.(type) {
				case *structpb.Value_NumberValue:
					ls = append(ls, reflect.ValueOf(uint64(v.GetNumberValue())))
				default:
					return fmt.Errorf("invalid int type")
				}
			}

		case SpecField_Float:
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

	dataUpdate = func(spec *Spec, baseValue reflect.Value, updateData *structpb.Struct) error {

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
		fmt.Println("baseValue", baseValue)

		for name, value := range updateData.Fields {

			if value == nil {
				continue
			}

			specField := spec.Field(name)
			if specField == nil {
				continue
			}

			dstField := baseValue.FieldByName(specField.Name)
			if !dstField.CanSet() {
				continue
			}
			// fmt.Println("field", specField.Name, "dstField", dstField, value)

			switch value.Kind.(type) {
			case *structpb.Value_BoolValue:
				if specField.Type != SpecField_Bool {
					return fmt.Errorf("invalid field (%s) type (bool:%s)", name, specField.Type)
				}
				if dstField.Kind() != reflect.Bool {
					return fmt.Errorf("invalid field (%s) type (string:%v)", name, dstField.Kind())
				}
				dstField.SetBool(value.GetBoolValue())

			case *structpb.Value_StringValue:
				if specField.Type != SpecField_String {
					return fmt.Errorf("invalid field (%s) type (string:%s)", name, specField.Type)
				}
				if value.GetStringValue() != "" {
					if dstField.Kind() != reflect.String {
						return fmt.Errorf("invalid field (%s) type (string:%v)", name, dstField.Kind())
					}
					dstField.SetString(value.GetStringValue())
				}

			case *structpb.Value_NumberValue:
				switch specField.Type {
				case SpecField_Int, SpecField_Uint, SpecField_Float:
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
					return fmt.Errorf("invalid field (%s) type (number:%s)", name, specField.Type)
				}

			case *structpb.Value_ListValue:
				if !strings.HasPrefix(specField.Type, "array:") {
					return fmt.Errorf("invalid field (%s) type (struct:%s)", name, specField.Type)
				}
				if dstField.Kind() != reflect.Slice {
					return fmt.Errorf("invalid field (%s) type (struct:%s)", name, specField.Type)
				}

				if err := arrayUpdate(specField, dstField, value.GetListValue()); err != nil {
					return err
				}

			case *structpb.Value_StructValue:
				if specField.Type != SpecField_Struct {
					return fmt.Errorf("invalid field (%s) type (struct:%s)", name, specField.Type)
				}
				if dstField.Kind() != reflect.Pointer && dstField.Kind() != reflect.Struct {
					return fmt.Errorf("invalid field (%s) type (struct:%s)", name, specField.Type)
				}
				if dstField.Kind() == reflect.Pointer {
					if dstField.IsNil() {
						dstField.Set(reflect.New(dstField.Type().Elem()))
					}

					if err := dataUpdate(&Spec{
						Fields: specField.Fields,
					}, dstField, value.GetStructValue()); err != nil {
						return err
					}
				} else if dstField.Kind() == reflect.Struct {
					fmt.Println("dst", dstField)
					if err := dataUpdate(&Spec{
						Fields: specField.Fields,
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
