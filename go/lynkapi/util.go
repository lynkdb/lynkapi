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
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"math/rand"
	"time"

	xxhash "github.com/cespare/xxhash/v2"
)

func jsonPrint(o any) {
	js, _ := json.MarshalIndent(o, "", "  ")
	fmt.Println(string(js))
}

func jsonEncode(o any) []byte {
	js, _ := json.MarshalIndent(o, "", "  ")
	return js
}

func RandObjectId(length int) string {
	if length < 8 {
		length = 8
	} else if length > 32 {
		length = 32
	} else if (length % 2) != 0 {
		length += 1
	}
	length = length / 2

	b := make([]byte, length)
	binary.BigEndian.PutUint32(b[:4], uint32(time.Now().Unix()))

	if _, err := rand.Read(b[2:]); err != nil {
		for i := 2; i < length; i++ {
			b[i] = uint8(rand.Intn(256))
		}
	}

	return hex.EncodeToString(b)
}

func RandHexString(length int) string {
	return hex.EncodeToString(RandBytes(length / 2))
}

func RandBytes(size int) []byte {

	const maxBytes = 1024

	if size < 1 {
		size = 1
	} else if size > maxBytes {
		size = maxBytes
	}

	bs := make([]byte, size)
	if _, err := rand.Read(bs); err != nil {
		for i := range bs {
			bs[i] = uint8(rand.Intn(256))
		}
	}

	return bs
}

func Uint32HexString(v uint32) string {
	u := make([]byte, 4)
	binary.BigEndian.PutUint32(u, v)
	return hex.EncodeToString(u)
}

func Uint64HexString(v uint64) string {
	u := make([]byte, 8)
	binary.BigEndian.PutUint64(u, v)
	return hex.EncodeToString(u)
}

func Crc32Hash(b []byte) uint32 {
	return crc32.ChecksumIEEE(b)
}

func Crc32HashHexString(b []byte) string {
	return Uint32HexString(Crc32Hash(b))
}

func XxHash(b []byte) uint64 {
	return xxhash.Sum64(b)
}

func XxHashHexString(b []byte) string {
	return Uint64HexString(XxHash(b))
}
