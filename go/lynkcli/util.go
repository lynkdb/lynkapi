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

package lynkcli

import (
	"fmt"
	"strconv"
	"strings"
)

type FlagSet struct {
	path    string
	rawArgs []string
	VarArgs []string
	args    map[string]FlagValue
}

func flagVarParse(s string) []string {
	s = strings.TrimSpace(s)

	var (
		VarArgs []string
		rawArgs = strings.Split(s, " ")
	)

	for _, v := range rawArgs {
		if len(v) == 0 {
			continue
		}
		if v[0] == '-' {
			break
		}
		VarArgs = append(VarArgs, v)
	}

	return VarArgs
}

func flagParse(s string) FlagSet {

	s = strings.TrimSpace(s)

	fset := FlagSet{
		path:    "",
		rawArgs: strings.Split(s, " "),
		VarArgs: []string{},
		args:    map[string]FlagValue{},
	}

	if n := strings.Index(s, " -"); n >= 0 {
		fset.path = s[:n]
	} else {
		fset.path = s
	}

	for i, k := range fset.rawArgs {

		if len(k) < 2 || k[0] != '-' {
			continue
		}

		k = strings.Trim(k, "-")

		if n := strings.Index(k, "="); n > 0 {
			if n+1 < len(k) {
				fset.args[k[:n]] = FlagValue(k[n+1:])
			} else {
				fset.args[k[:n]] = FlagValue("")
			}
			continue
		}

		if len(fset.rawArgs) <= (i+1) || fset.rawArgs[i+1][0] == '-' {
			fset.args[k] = FlagValue([]byte(""))
			continue
		}

		v := fset.rawArgs[i+1]

		fset.args[k] = FlagValue([]byte(v))
	}

	return fset
}

func (it *FlagSet) ValueOK(key string) (FlagValue, bool) {

	if v, ok := it.args[key]; ok {
		return v, ok
	}

	return nil, false
}

func (it *FlagSet) Value(key string) FlagValue {

	if v, ok := it.ValueOK(key); ok {
		return v
	}

	return FlagValue{}
}

func (it *FlagSet) setValue(key, val string) {
	it.args[key] = FlagValue([]byte(val))
}

func (it *FlagSet) Has(key string) bool {
	if _, ok := it.args[key]; ok {
		return true
	}
	return false
}

func (it *FlagSet) Each(fn func(key, val string)) {
	for k, v := range it.args {
		fn(k, v.String())
	}
}

// Universal Bytes
type FlagValue []byte

// String converts the value-bytes to string
func (bx FlagValue) String() string {
	return string(bx)
}

// Bool converts the value-bytes to bool
func (bx FlagValue) Bool() bool {
	if len(bx) > 0 {
		if b, err := strconv.ParseBool(string(bx)); err == nil {
			return b
		}
	}
	return false
}

// Int64 converts the value-bytes to int64
func (bx FlagValue) Int64() int64 {
	if len(bx) > 0 {
		if i64, err := strconv.ParseInt(string(bx), 10, 64); err == nil {
			return i64
		}
	}
	return 0
}

// Float64 converts the value-bytes to float64
func (bx FlagValue) Float64() float64 {
	if len(bx) > 0 {
		if f64, err := strconv.ParseFloat(string(bx), 64); err == nil {
			return f64
		}
	}
	return 0
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

func uptimeFormat(sec int64) string {

	s := ""

	d := (sec / 86400)
	if d > 1 {
		s = fmt.Sprintf("%d days ", d)
	} else if d == 1 {
		s = fmt.Sprintf("%d day ", d)
	}

	sec = sec % 86400
	h := sec / 3600

	sec = sec % 3600
	m := sec / 60

	sec = sec % 60

	s += fmt.Sprintf("%02d:%02d:%02d", h, m, sec)

	return s
}
