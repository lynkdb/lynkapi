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
	"errors"
	"sync"

	"github.com/hooto/hlog4g/hlog"

	"google.golang.org/protobuf/types/known/structpb"
)

type DataService interface {
	Instance() *DataInstance

	Query(q *DataQuery) (*DataResult, error)
	Upsert(q *DataUpsert) (*DataResult, error)
	Igsert(q *DataIgsert) (*DataResult, error)
	Delete(q *DataDelete) (*DataResult, error)
}

type dataProjectManager struct {
	mu      sync.RWMutex
	project *DataProject

	index map[string]int

	services map[string]DataService
}

func newDataProjectManager() *dataProjectManager {
	return &dataProjectManager{
		project:  &DataProject{},
		index:    map[string]int{},
		services: map[string]DataService{},
	}
}

func (it *dataProjectManager) service(instanceName string) DataService {
	it.mu.Lock()
	defer it.mu.Unlock()

	// idx, ok := it.index[instanceName]
	// if !ok {
	// 	return nil, nil
	// }
	// inst := it.project.Instances[idx]
	// for _, tbl := range inst.Tables {
	// 	if tableName == tbl.Name {
	// 		return inst, tbl
	// 	}
	// }
	// return inst, nil

	ds, ok := it.services[instanceName]
	if !ok {
		return nil
	}
	return ds
}

func (it *dataProjectManager) RegisterService(ds DataService) error {

	inst := ds.Instance()

	if inst == nil || inst.Name == "" {
		return errors.New("name not setup")
	}

	if !NameIdentifier.MatchString(inst.Name) {
		return errors.New("invalid name")
	}

	it.mu.Lock()
	defer it.mu.Unlock()

	idx, ok := it.index[inst.Name]
	if !ok {
		it.index[inst.Name] = len(it.project.Instances)
		it.project.Instances = append(it.project.Instances, inst)
	} else {
		it.project.Instances[idx] = inst
	}

	it.services[inst.Name] = ds

	hlog.Printf("info", "register data service %s, tables %d", inst.Name, len(inst.Spec.Tables))

	return nil
}

func (it *DataInstance) SetName(name string) *DataInstance {
	if NameIdentifier.MatchString(name) {
		it.Name = name
	}
	return it
}

func (it *DataUpsert) SetField(name string, obj any) {
	var value *structpb.Value
	switch obj.(type) {
	case string:
		value = structpb.NewStringValue(obj.(string))
	case map[string]interface{}:
		if st, err := structpb.NewStruct(obj.(map[string]interface{})); err == nil {
			value = structpb.NewStructValue(st)
		}
	}
	if value == nil {
		return
	}
	for i, field := range it.Fields {
		if field == name {
			it.Values[i] = value
			return
		}
	}
	it.Fields = append(it.Fields, name)
	it.Values = append(it.Values, value)
}

func (it *DataIgsert) SetField(name string, obj any) {
	var value *structpb.Value
	switch obj.(type) {
	case string:
		value = structpb.NewStringValue(obj.(string))

	case map[string]interface{}:
		if st, err := structpb.NewStruct(obj.(map[string]interface{})); err == nil {
			value = structpb.NewStructValue(st)
		}
	}
	if value == nil {
		return
	}
	for i, field := range it.Fields {
		if field == name {
			it.Values[i] = value
			return
		}
	}
	it.Fields = append(it.Fields, name)
	it.Values = append(it.Values, value)
}

func (it *DataQuery_Filter) And(field string, obj any) *DataQuery_Filter {
	if obj != nil {
		if v, err := structpb.NewValue(obj); err == nil {
			it.Inner = append(it.Inner, &DataQuery_Filter{
				Field: field,
				Value: v,
			})
		}
	}

	return it
}
