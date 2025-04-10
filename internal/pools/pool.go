// Copyright 2025 xgfone
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

package pools

import (
	"bytes"
	"fmt"
	"sync"
)

func newBuffer(cap int) *bytes.Buffer {
	return bytes.NewBuffer(make([]byte, 0, cap))
}

var (
	bufpool256  = sync.Pool{New: func() any { return newBuffer(256) }}
	bufpool64KB = sync.Pool{New: func() any { return newBuffer(64 * 1024) }}
)

func GetBuffer(cap int) (pool *sync.Pool, buf *bytes.Buffer) {
	switch cap {
	case 256:
		pool = &bufpool256

	case 64 * 1024: //64KB
		pool = &bufpool64KB

	default:
		panic(fmt.Errorf("GetBuffer: unsupported cap %d", cap))
	}

	buf = pool.Get().(*bytes.Buffer)
	buf.Reset()
	return
}

func PutBuffer(pool *sync.Pool, buf *bytes.Buffer) {
	buf.Reset()
	pool.Put(buf)
}

/// ----------------------------------------------------------------------- ///

type Bytes struct {
	Buffer []byte
	pool   *sync.Pool
}

var bytespool1k = &sync.Pool{New: func() any { return newBytes(1024) }}

func newBytes(cap int) *Bytes {
	return &Bytes{Buffer: make([]byte, cap)}
}

func GetBytes(len int) (buf *Bytes) {
	switch {
	case len == 1024:
		buf = bytespool1k.Get().(*Bytes)
		buf.pool = bytespool1k

	default:
		panic(fmt.Errorf("GetBytes: unsupported len %d", len))
	}

	return
}

func PutBytes(buf *Bytes) {
	if buf.pool != nil {
		buf.pool.Put(buf)
	}
}
