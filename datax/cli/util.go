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

package cli

import (
	"fmt"
	"strconv"
	"strings"
)

type flagSet struct {
	path    string
	rawArgs []string
	varArgs []string
	args    map[string]flagValue
}

func flagVarParse(s string) []string {
	s = strings.TrimSpace(s)

	var (
		varArgs []string
		rawArgs = strings.Split(s, " ")
	)

	for _, v := range rawArgs {
		if len(v) == 0 {
			continue
		}
		if v[0] == '-' {
			break
		}
		varArgs = append(varArgs, v)
	}

	return varArgs
}

func flagParse(s string) flagSet {

	s = strings.TrimSpace(s)

	fset := flagSet{
		path:    "",
		rawArgs: strings.Split(s, " "),
		varArgs: []string{},
		args:    map[string]flagValue{},
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
				fset.args[k[:n]] = flagValue(k[n+1:])
			} else {
				fset.args[k[:n]] = flagValue("")
			}
			continue
		}

		if len(fset.rawArgs) <= (i+1) || fset.rawArgs[i+1][0] == '-' {
			fset.args[k] = flagValue([]byte(""))
			continue
		}

		v := fset.rawArgs[i+1]

		fset.args[k] = flagValue([]byte(v))
	}

	return fset
}

func (it *flagSet) ValueOK(key string) (flagValue, bool) {

	if v, ok := it.args[key]; ok {
		return v, ok
	}

	return nil, false
}

func (it *flagSet) Value(key string) flagValue {

	if v, ok := it.ValueOK(key); ok {
		return v
	}

	return flagValue{}
}

func (it *flagSet) Has(key string) bool {
	if _, ok := it.args[key]; ok {
		return true
	}
	return false
}

func (it *flagSet) Each(fn func(key, val string)) {
	for k, v := range it.args {
		fn(k, v.String())
	}
}

// Universal Bytes
type flagValue []byte

// String converts the value-bytes to string
func (bx flagValue) String() string {
	return string(bx)
}

// Bool converts the value-bytes to bool
func (bx flagValue) Bool() bool {
	if len(bx) > 0 {
		if b, err := strconv.ParseBool(string(bx)); err == nil {
			return b
		}
	}
	return false
}

// Int64 converts the value-bytes to int64
func (bx flagValue) Int64() int64 {
	if len(bx) > 0 {
		if i64, err := strconv.ParseInt(string(bx), 10, 64); err == nil {
			return i64
		}
	}
	return 0
}

// Float64 converts the value-bytes to float64
func (bx flagValue) Float64() float64 {
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
