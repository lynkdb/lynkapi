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
	"google.golang.org/protobuf/types/known/structpb"
)

func NewDataQuery() *DataQuery {
	return &DataQuery{}
}

func (it *DataQuery) AddFilter(field string, obj any) *DataQuery {
	it.addFilter(field, obj)
	return it
}

func (it *DataQuery) AddFuncFilter(fn string, obj any) *DataQuery {
	if fr := it.addFilter(fn, obj); fr != nil {
		fr.Type = "func"
	}
	return it
}

func (it *DataQuery) SetLimit(v int32) *DataQuery {
	it.Limit = v
	return it
}

func (it *DataQuery) SetOffset(v int32) *DataQuery {
	it.Offset = v
	return it
}

func (it *DataQuery) addFilter(field string, obj any) *DataQuery_Filter {

	v, err := structpb.NewValue(obj)
	if err != nil {
		return &DataQuery_Filter{} // TODO
	}

	fr := &DataQuery_Filter{
		Field: field,
		Value: v,
	}

	if it.Filter == nil {
		it.Filter = fr
	} else if it.Filter.Inner == nil {
		it.Filter = &DataQuery_Filter{
			Inner: []*DataQuery_Filter{
				it.Filter,
				fr,
			},
		}
	} else {
		it.Filter.Inner = append(it.Filter.Inner, fr)
	}
	return fr
}

func (it *DataQuery) SearchFilter(field string) *DataQuery_Filter {
	if it.Filter == nil {
		return nil
	}
	if it.Filter.Field == field {
		return it.Filter
	}
	for _, fr := range it.Filter.Inner {
		if fr.Field == field {
			return fr
		}
	}
	return nil
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

func (it *DataQuery_Filter) StringValue() string {
	if it != nil && it.Value != nil {
		return it.Value.GetStringValue()
	}
	return ""
}
