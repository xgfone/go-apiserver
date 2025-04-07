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
