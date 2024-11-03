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

package codec

import (
	"encoding/json"

	"github.com/tidwall/pretty"
)

var (
	Json Codec = jsonCodec{}
)

type jsonCodec struct{}

func (jsonCodec) Name() string {
	return "json"
}

func (jsonCodec) Encode(v any, args ...any) ([]byte, error) {

	if len(args) > 0 {
		for _, arg := range args {
			if arg == nil {
				continue
			}
			switch arg.(type) {
			case *JsonOptions:
				b, err := json.Marshal(v)
				if err != nil {
					return nil, err
				}
				opts := arg.(*JsonOptions).reset()
				return pretty.PrettyOptions(b, &pretty.Options{
					Width:    opts.Width,
					Prefix:   "",
					Indent:   opts.Indent,
					SortKeys: false,
				}), nil
			}
		}
		return json.MarshalIndent(v, "", "  ")
	}

	return json.Marshal(v)
}

func (jsonCodec) Decode(b []byte, v any) error {
	return json.Unmarshal(b, v)
}

type JsonOptions struct {
	Width  int
	Indent string
}

func (it *JsonOptions) reset() *JsonOptions {
	if it.Width == 0 {
		it.Width = 80
	} else if it.Width < 60 {
		it.Width = 60
	} else if it.Width > 1000 {
		it.Width = 1000
	}
	if it.Indent == "" {
		it.Indent = "  "
	}
	return it
}
