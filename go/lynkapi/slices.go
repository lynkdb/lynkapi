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

func SlicesSearchFunc[T any](s []*T, cmp func(*T) bool) *T {
	for _, e := range s {
		if cmp(e) {
			return e
		}
	}
	return nil
}

func SlicesDeleteFunc[T any](s []*T, cmp func(*T) bool) ([]*T, bool) {
	for i, e := range s {
		if cmp(e) {
			return append(s[:i], s[i+1:]...), true
		}
	}
	return s, false
}
