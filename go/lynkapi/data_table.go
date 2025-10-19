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
	"fmt"
	"strings"
)

const (
	TableSpec_Index_FullTextSearch string = "fts"
)

var tableSpec_Index_Types = map[string]string{
	TableSpec_Index_FullTextSearch: "Full Text Search Index",
}

func (it *TableSpec) SetField(tagName, typ string) (*FieldSpec, error) {

	if !NameIdentifier.MatchString(tagName) {
		return nil, fmt.Errorf("invalid field name (%s)", tagName)
	}

	if FieldSpec_Types[typ] == "" {
		return nil, fmt.Errorf("type (%s) not support")
	}

	var fs *FieldSpec

	for _, field := range it.Fields {
		if field.TagName == tagName ||
			field.Name == tagName {
			fs = field
			break
		}
	}
	if fs == nil {
		fs = &FieldSpec{
			Name:    strings.ToLower(tagName),
			TagName: tagName,
		}
		it.Fields = append(it.Fields, fs)
	}
	if typ != fs.Type {
		fs.Type = typ
	}
	return fs, nil
}

func (it *TableSpec) Field(tagName string) (*FieldSpec, int) {

	if it != nil {
		for i, field := range it.Fields {
			if field.TagName == tagName ||
				field.Name == tagName {
				return field, i
			}
		}
	}
	return nil, -1
}

func (it *TableSpec) SetIndex(fields, typ string) error {

	ifds, err := it.indexFields(fields)
	if err != nil {
		return err
	}

	var (
		fs   = strings.Join(ifds, ",")
		fidx *TableSpec_Index
	)
	for _, idx := range it.Indexes {
		if idx.Fields == fs {
			fidx = idx
			break
		}
	}
	if fidx == nil {
		fidx = &TableSpec_Index{
			Fields: fs,
		}
		it.Indexes = append(it.Indexes, fidx)
	}

	if tableSpec_Index_Types[typ] != "" {
		fidx.Type = typ
	}

	return nil
}

func (it *TableSpec) indexFields(fields string) ([]string, error) {
	var (
		fds  = strings.Split(fields, ",")
		ifds = []string{}
	)
	for _, v := range fds {

		if !NameIdentifier.MatchString(v) {
			return nil, fmt.Errorf("invalid field name (%s)", v)
		}

		fd, _ := it.Field(v)
		if fd == nil {
			return nil, fmt.Errorf("field (%s) not found", v)
		}
		ifds = append(ifds, fd.Name)
	}
	if len(ifds) == 0 {
		return nil, fmt.Errorf("no field setup")
	}
	return ifds, nil
}
