package oneobject

import (
	"errors"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strings"
	"sync"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/lynkdb/lynkapi/go/codec"
	"github.com/lynkdb/lynkapi/go/lynkapi"
)

type Instance struct {
	mu      sync.Mutex
	name    string
	file    string
	spec    *lynkapi.TypeSpec
	object  any
	tables  map[string]*table
	flusher Flusher
}

type Flusher func() error

type table struct {
	path  []string
	name  string
	spec  *lynkapi.TableSpec
	field *lynkapi.FieldSpec
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

func NewInstanceFromFile(name, file string, obj any, args ...any) (*Instance, error) {

	b, err := ioutil.ReadFile(file)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if err == nil {
		if err = codec.Json.Decode(b, obj); err != nil {
			return nil, err
		}
	}

	inst, err := NewInstance(name, obj, args...)
	if err != nil {
		return nil, err
	}
	inst.file = file

	inst.flusher = func() error {
		b, _ := codec.Json.Encode(inst.object, &codec.JsonOptions{
			Width: 120,
		})
		return ioutil.WriteFile(inst.file, b, 0640)
	}

	return inst, nil
}

func NewInstance(name string, obj any, args ...any) (*Instance, error) {
	spec, _, err := lynkapi.NewSpecFromStruct(obj)
	if err != nil {
		return nil, err
	}
	inst := &Instance{
		name:   name,
		spec:   spec,
		object: obj,
		tables: map[string]*table{},
	}

	for _, arg := range args {
		if arg == nil {
			continue
		}
		switch arg.(type) {
		case Flusher:
			inst.flusher = arg.(Flusher)
		}
	}
	return inst, nil
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

	indexFields := map[string]*lynkapi.FieldSpec{}
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

	orderDef := int32(0)
	sortFilter := func(row *lynkapi.DataRow) *lynkapi.DataRow {
		// TODO
		if q.Sort != nil && q.Sort.Field != "" {
			if v, ok := row.Fields[q.Sort.Field]; ok && v.GetNumberValue() > 0 {
				row.InterOrder = int64(v.GetNumberValue())
			} else {
				orderDef += 1
				row.InterOrder = row.InterOrder
			}
		}
		return row
	}

	if q.Filter != nil {
		if len(q.Filter.Field) > 0 {
			tp, ok := indexFields[lowerName(q.Filter.Field)]
			if !ok {
				return nil, errors.New("filter/field not found")
			}
			switch tp.Type {
			case lynkapi.FieldSpec_String:
				filters[tp.Name] = q.Filter.Value
			}
		} else if len(q.Filter.Inner) > 0 {
			for _, fr := range q.Filter.Inner {
				tp, ok := indexFields[lowerName(fr.Field)]
				if !ok {
					return nil, errors.New("filter/field not found")
				}
				switch tp.Type {
				case lynkapi.FieldSpec_String:
					filters[tp.Name] = fr.Value
				}
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

		row := &lynkapi.DataRow{
			Id:     tbl.spec.PrimaryId(fieldValues),
			Fields: fieldValues,
		}

		row = sortFilter(row)

		rs.Rows = append(rs.Rows, row)
	}

	if q.Sort != nil {
		if q.Sort.Type == "desc" {
			sort.Slice(rs.Rows, func(i, j int) bool {
				return rs.Rows[i].InterOrder > rs.Rows[j].InterOrder
			})
		} else {
			sort.Slice(rs.Rows, func(i, j int) bool {
				return rs.Rows[i].InterOrder < rs.Rows[j].InterOrder
			})
		}
	}

	if len(rs.Rows) == 0 {
		rs.Status = lynkapi.NewServiceStatus(lynkapi.StatusCode_NotFound, "")
	} else {
		rs.Status = lynkapi.NewServiceStatusOK()
	}

	return rs, nil
}

const (
	kInsertRaw int = iota + 1
	kInsertIgsert
	kInsertUpsert
)

func (it *Instance) Insert(q *lynkapi.DataInsert) (*lynkapi.DataResult, error) {
	return it.insert(q, kInsertRaw)
}

func (it *Instance) Igsert(q *lynkapi.DataInsert) (*lynkapi.DataResult, error) {
	return it.insert(q, kInsertIgsert)
}

func (it *Instance) Upsert(q *lynkapi.DataInsert) (*lynkapi.DataResult, error) {
	return it.insert(q, kInsertUpsert)
}

func (it *Instance) insert(q *lynkapi.DataInsert, typ int) (*lynkapi.DataResult, error) {

	if len(q.Fields) == 0 || len(q.Fields) != len(q.Values) {
		return nil, errors.New("invalid request (fields != values)")
	}

	it.mu.Lock()
	defer it.mu.Unlock()

	tbl, ok := it.tables[q.TableName]
	if !ok {
		return nil, errors.New("table not found")
	}

	vtbl, err := findValue(tbl.path, reflect.ValueOf(it.object))
	if err != nil {
		return nil, err
	}

	var (
		pks, pkm, ukm = tbl.field.PrimaryKeys()
		data          = map[string]*structpb.Value{}
		pkv           = map[string]*structpb.Value{} // primary-key
		ukv           = map[string]*structpb.Value{} // unique-key
	)

	for i, tagName := range q.Fields {

		data[tagName] = q.Values[i]

		//
		specField, ok := ukm[tagName]
		if !ok {
			continue
		}

		switch {
		case specField.Type == "string":
			if s := strings.TrimSpace(q.Values[i].String()); s == "" {
				if fn := specField.FuncAttr("rand_hex", "object_id"); fn != nil {
					q.Values[i] = structpb.NewStringValue(fn.GenId())
				} else {
					return nil, errors.New("primary-key/unique-key not be null")
				}
			}
			ukv[specField.TagName] = q.Values[i]
			if _, ok = pkm[tagName]; ok {
				pkv[specField.TagName] = q.Values[i]
			}

		default:
			return nil, errors.New("un-impl")
		}
	}

	if len(pkv) < len(pks) {

		for _, specField := range pkm {

			if _, ok := pkv[specField.TagName]; ok {
				continue
			}
			if _, ok := data[specField.TagName]; ok {
				continue
			}

			fa := specField.FuncAttr("rand_hex", "object_id")
			if fa == nil {
				continue
			}

			value := structpb.NewStringValue(fa.GenId())

			data[specField.TagName] = value

			pkv[specField.TagName] = value
			ukv[specField.TagName] = value

			q.Values = append(q.Values, value)
			q.Fields = append(q.Fields, specField.TagName)
		}
	}

	if len(pkv) != len(pks) {
		return nil, errors.New("primary-key not found")
	}

	tp := vtbl.Type().Elem()
	if tp.Kind() == reflect.Pointer {
		tp = tp.Elem()
	}

	reqData := reflect.New(tp)

	js, _ := codec.Json.Encode(data)
	if err := codec.Json.Decode(js, reqData.Interface()); err != nil {
		return nil, err
	}

	rs := &lynkapi.DataResult{}

	for i := 0; i < vtbl.Len(); i++ {
		row := vtbl.Index(i)
		if row.Kind() == reflect.Pointer {
			row = row.Elem()
		}
		if !row.IsValid() || row.Kind() != reflect.Struct {
			continue
		}
		ukhit := 0
		for _, fd := range ukm {
			fv := row.FieldByName(fd.Name)
			if !fv.IsValid() {
				continue
			}
			if uv, ok := ukv[fd.TagName]; ok && uv.GetStringValue() == fv.String() {
				ukhit += 1
			}
		}

		if ukhit == 0 {
			continue
		}

		switch typ {
		case kInsertRaw:
			return rs, lynkapi.NewConflictError("row exist")

		case kInsertUpsert:
			if _, err := tbl.field.DataMerge(vtbl.Index(i).Interface(), reqData.Interface()); err != nil {
				return nil, err
			}

			if err := it.Flush(); err != nil {
				return nil, err
			}
		}

		rs.Status = lynkapi.NewServiceStatusOK()
		return rs, nil
	}

	dst := reflect.New(tp)
	_, err = tbl.field.DataMerge(dst.Interface(), reqData.Interface())
	if err != nil {
		return nil, err
	}
	vtbl.Set(reflect.Append(vtbl, dst))

	if err := it.Flush(); err != nil {
		return nil, err
	}

	rs.Status = lynkapi.NewServiceStatusOK()
	return rs, nil
}

func (it *Instance) Delete(q *lynkapi.DataDelete) (*lynkapi.DataResult, error) {

	it.mu.Lock()
	defer it.mu.Unlock()

	tbl, ok := it.tables[q.TableName]
	if !ok {
		return nil, errors.New("table not found")
	}

	vtbl, err := findValue(tbl.path, reflect.ValueOf(it.object))
	if err != nil {
		return nil, err
	}

	if q.Filter == nil || len(q.Filter.Inner) == 0 {
		return nil, errors.New("filter not found")
	}

	var (
		idx         = map[string]*structpb.Value{}
		pks, pkm, _ = tbl.field.PrimaryKeys()
	)
	for _, fr := range q.Filter.Inner {
		if fr.Value == nil {
			continue
		}
		specField, ok := pkm[fr.Field]
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

	for i := 0; i < vtbl.Len(); i++ {
		v := vtbl.Index(i)
		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}
		if !v.IsValid() || v.Kind() != reflect.Struct {
			continue
		}
		pkhit := 0
		for _, pk := range pkm {
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

		ls := reflect.New(vtbl.Type()).Elem()
		for j := 0; j < i; j++ {
			ls.Set(reflect.Append(ls, vtbl.Index(j)))
		}
		for j := i + 1; j < vtbl.Len(); j++ {
			ls.Set(reflect.Append(ls, vtbl.Index(j)))
		}
		vtbl.Set(ls)

		if err := it.Flush(); err != nil {
			return nil, err
		}

		break
	}

	rs.Status = lynkapi.NewServiceStatusOK()
	return rs, nil
}

func (it *Instance) TableSetup(path string) error {

	var (
		fields = strings.Split(strings.TrimSpace(path), "__")
		find   func(fields, hitPath []string, specField *lynkapi.FieldSpec) (*lynkapi.FieldSpec, []string, error)
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

	find = func(fields, hitPath []string, specField *lynkapi.FieldSpec) (*lynkapi.FieldSpec, []string, error) {
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
				pkeys, _, _ := field.PrimaryKeys()
				if len(pkeys) == 0 {
					return nil, nil, errors.New("primary-key not setup")
				}
				return field, append(hitPath, field.Name), nil
			}
		}
		return nil, nil, errors.New("no table spec-field found")
	}

	hitField, hitPath, err := find(fields, []string{}, &lynkapi.FieldSpec{
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
			Kind:   hitField.Kind,
			Fields: hitField.Fields,
		},
		field: hitField,
	}
	return nil
}

func (it *Instance) Flush() error {
	if it.flusher != nil {
		return it.flusher()
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
