package oneobject

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/lynkdb/lynkapi/go/lynkapi"
)

type Instance struct {
	mu     sync.Mutex
	name   string
	spec   *lynkapi.Spec
	object any
	tables map[string]*table
}

type table struct {
	path  []string
	name  string
	spec  *lynkapi.TableSpec
	field *lynkapi.SpecField
}

func (it *Instance) Instance() *lynkapi.DataInstance {
	it.mu.Lock()
	defer it.mu.Unlock()

	di := &lynkapi.DataInstance{
		Name: it.name,
		Spec: &lynkapi.DataSpec{
			Driver: "oneobject",
		},
	}
	for _, tbl := range it.tables {
		di.Spec.Tables = append(di.Spec.Tables, tbl.spec)
	}
	return di
}

func findValue(path []string, upValue reflect.Value) (reflect.Value, error) {

	if len(path) == 0 || !upValue.IsValid() {
		return upValue, errors.New("data not found")
	}

	value := upValue

	if value.Kind() == reflect.Pointer {
		value = value.Elem()
	}

	fieldValue := value.FieldByName(path[0])
	if !fieldValue.IsValid() {
		return fieldValue, errors.New("data not found")
	}

	if len(path) > 1 {
		if fieldValue.Kind() != reflect.Struct {
			return fieldValue, errors.New("data not found (!struct)")
		}
		return findValue(path[1:], fieldValue)
	}

	if fieldValue.Kind() == reflect.Slice {
		return fieldValue, nil
	}

	return fieldValue, errors.New("data not found (!slice)")
}

func NewInstance(name string, obj any) (*Instance, error) {
	spec, _, err := lynkapi.NewSpecFromStruct(obj)
	if err != nil {
		return nil, err
	}
	return &Instance{
		name:   name,
		spec:   spec,
		object: obj,
		tables: map[string]*table{},
	}, nil
}

func (it *Instance) Query(q *lynkapi.DataQuery) (*lynkapi.DataResult, error) {

	it.mu.Lock()
	defer it.mu.Unlock()

	tbl, ok := it.tables[q.TableName]
	if !ok {
		return nil, errors.New("table not found")
	}

	hit, err := findValue(tbl.path, reflect.ValueOf(it.object))
	if err != nil {
		return nil, err
	}

	indexFields := map[string]*lynkapi.SpecField{}
	for _, fd := range tbl.field.Fields {
		indexFields[fd.TagName] = fd
	}

	if q.Limit == 0 {
		q.Limit = 10
	}

	var (
		rs = &lynkapi.DataResult{
			Spec: tbl.spec,
		}
		filters = map[string]*structpb.Value{}
	)

	if q.Filter != nil {
		for _, fr := range q.Filter.Inner {
			tp, ok := indexFields[lowerName(fr.Name)]
			if !ok {
				return nil, errors.New("filter/field not found")
			}
			switch tp.Type {
			case lynkapi.SpecField_String:
				filters[tp.Name] = fr.Value
			}
		}
	}

	for i := 0; i < hit.Len() && len(rs.Rows) < int(q.Limit); i++ {
		v := hit.Index(i)
		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}
		if !v.IsValid() || v.Kind() != reflect.Struct {
			continue
		}
		frHit := 0
		if len(filters) > 0 {
			for name, frValue := range filters {
				fv := v.FieldByName(name)
				if !fv.IsValid() {
					continue
				}
				switch frValue.Kind.(type) {
				case *structpb.Value_StringValue:
					if fv.String() == frValue.GetStringValue() {
						frHit += 1
					}
				}
			}
		}
		if frHit != len(filters) {
			continue
		}

		// anyValue, err := lynkapi.ConvertReflectValueToApiValue(v)
		// if err != nil {
		// 	fmt.Println("err", err)
		// 	continue
		// }
		// rs.Items = append(rs.Items, anyValue)

		fieldValues, err := lynkapi.ConvertReflectValueToMapValue(v)
		if err != nil {
			continue
		}
		rs.Rows = append(rs.Rows, &lynkapi.TableSpec_Row{
			Fields: fieldValues,
		})
	}

	return rs, nil
}

func (it *Instance) Upsert(q *lynkapi.DataUpsert) (*lynkapi.DataResult, error) {

	if len(q.Fields) == 0 || len(q.Fields) != len(q.Values) {
		return nil, errors.New("invalid request (fields != values)")
	}

	it.mu.Lock()
	defer it.mu.Unlock()

	tbl, ok := it.tables[q.TableName]
	if !ok {
		return nil, errors.New("table not found")
	}

	hit, err := findValue(tbl.path, reflect.ValueOf(it.object))
	if err != nil {
		return nil, err
	}

	var (
		idx       = map[string]int{}
		pks, pkms = tbl.field.PrimaryKeys()
		data      = map[string]*structpb.Value{}
	)
	for i, fieldName := range q.Fields {
		data[fieldName] = q.Values[i]
		specField, ok := pkms[fieldName]
		if !ok {
			continue
		}
		switch specField.Type {
		case "string":
			if s := strings.TrimSpace(q.Values[i].String()); s == "" {
				return nil, errors.New("primary-key not be null")
			}
			idx[specField.TagName] = i
		default:
			return nil, errors.New("un-impl")
		}
	}
	if len(idx) != len(pks) {
		return nil, errors.New("primary-key not found")
	}

	tp := hit.Type().Elem()
	if tp.Kind() == reflect.Pointer {
		tp = tp.Elem()
	}

	reqData := reflect.New(tp)

	js, _ := json.Marshal(data)
	if err := json.Unmarshal(js, reqData.Interface()); err != nil {
		return nil, err
	}

	rs := &lynkapi.DataResult{}

	for i := 0; i < hit.Len(); i++ {
		v := hit.Index(i)
		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}
		if !v.IsValid() || v.Kind() != reflect.Struct {
			continue
		}
		pkhit := 0
		for _, pk := range pkms {
			fv := v.FieldByName(pk.Name)
			if !fv.IsValid() {
				continue
			}
			switch pk.Type {
			case "string":
				if fv.String() == q.Values[idx[pk.TagName]].GetStringValue() {
					pkhit += 1
				}
			}
		}
		if pkhit != len(pks) {
			continue
		}

		if _, err := tbl.field.DataMerge(hit.Index(i).Interface(), reqData.Interface()); err != nil {
			return nil, err
		}

		return rs, nil
	}

	dst := reflect.New(tp)
	_, err = tbl.field.DataMerge(dst.Interface(), reqData.Interface())
	if err != nil {
		return nil, err
	}
	hit.Set(reflect.Append(hit, dst))

	return rs, nil
}

func (it *Instance) Delete(q *lynkapi.DataDelete) (*lynkapi.DataResult, error) {

	it.mu.Lock()
	defer it.mu.Unlock()

	tbl, ok := it.tables[q.TableName]
	if !ok {
		return nil, errors.New("table not found")
	}

	hit, err := findValue(tbl.path, reflect.ValueOf(it.object))
	if err != nil {
		return nil, err
	}

	if q.Filter == nil || len(q.Filter.Inner) == 0 {
		return nil, errors.New("filter not found")
	}

	var (
		idx       = map[string]*structpb.Value{}
		pks, pkms = tbl.field.PrimaryKeys()
	)
	for _, fr := range q.Filter.Inner {
		if fr.Value == nil {
			continue
		}
		specField, ok := pkms[fr.Name]
		if !ok {
			continue
		}
		switch specField.Type {
		case "string":
			if s := strings.TrimSpace(fr.Value.GetStringValue()); s == "" {
				return nil, errors.New("primary-key not be null")
			} else {
				idx[specField.TagName] = fr.Value
			}
		default:
			return nil, errors.New("un-impl")
		}
	}
	if len(idx) != len(pks) {
		return nil, errors.New("primary-key not found")
	}

	rs := &lynkapi.DataResult{}

	for i := 0; i < hit.Len(); i++ {
		v := hit.Index(i)
		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}
		if !v.IsValid() || v.Kind() != reflect.Struct {
			continue
		}
		pkhit := 0
		for _, pk := range pkms {
			fv := v.FieldByName(pk.Name)
			if !fv.IsValid() {
				continue
			}
			switch pk.Type {
			case "string":
				if fv.String() == idx[pk.TagName].GetStringValue() {
					pkhit += 1
				}
			}
		}
		if pkhit != len(pks) {
			continue
		}

		ls := reflect.New(hit.Type()).Elem()
		for j := 0; j < i; j++ {
			ls.Set(reflect.Append(ls, hit.Index(j)))
		}
		for j := i + 1; j < hit.Len(); j++ {
			ls.Set(reflect.Append(ls, hit.Index(j)))
		}
		hit.Set(ls)

		return rs, nil
	}

	return rs, nil
}

func (it *Instance) TableSetup(path string) error {

	var (
		fields = strings.Split(strings.TrimSpace(path), "__")
		find   func(fields, hitPath []string, specField *lynkapi.SpecField) (*lynkapi.SpecField, []string, error)
	)

	for i, fd := range fields {
		fields[i] = lowerName(fd)
	}

	tableName := strings.Join(fields, "__")

	it.mu.Lock()
	defer it.mu.Unlock()

	if _, ok := it.tables[tableName]; ok {
		return nil
	}

	find = func(fields, hitPath []string, specField *lynkapi.SpecField) (*lynkapi.SpecField, []string, error) {
		if len(fields) > 0 {
			for _, field := range specField.Fields {
				if field.TagName != fields[0] {
					continue
				}
				if len(fields) > 1 {
					if field.Type != "struct" {
						break
					}
					return find(fields[1:], append(hitPath, field.Name), field)
				}
				if field.Type != "array:struct" {
					break
				}
				pkeys, _ := field.PrimaryKeys()
				if len(pkeys) == 0 {
					return nil, nil, errors.New("primary-key not setup")
				}
				return field, append(hitPath, field.Name), nil
			}
		}
		return nil, nil, errors.New("no table spec-field found")
	}

	hitField, hitPath, err := find(fields, []string{}, &lynkapi.SpecField{
		Fields: it.spec.Fields,
	})
	if err != nil {
		return err
	}
	it.tables[tableName] = &table{
		name: tableName,
		path: hitPath,
		spec: &lynkapi.TableSpec{
			Name:   tableName,
			Fields: hitField.Fields,
		},
		field: hitField,
	}
	return nil
}

func lowerName(s string) string {
	var (
		b1 = []byte(s)
		b2 []byte
	)
	for i, c := range b1 {
		if c >= 'A' && c <= 'Z' {
			if i > 0 && (b1[i-1] < 'A' || b1[i-1] > 'Z') {
				b2 = append(b2, '-')
			}
			b2 = append(b2, 'a'+(c-'A'))
		} else {
			b2 = append(b2, c)
		}
	}
	return string(b2)
}

func jsonPrint(name string, o any) {
	js, _ := json.Marshal(o)
	fmt.Println(name, string(js))
}
