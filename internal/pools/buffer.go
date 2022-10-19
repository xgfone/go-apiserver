// Copyright 2022 xgfone
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
	"sync"
)

var (
	bufferPool64  = newBufferPool(64)
	bufferPool128 = newBufferPool(128)
	bufferPool256 = newBufferPool(256)
	bufferPool512 = newBufferPool(512)
	bufferPool1K  = newBufferPool(1024)
	bufferPool2K  = newBufferPool(2048)
	bufferPool4K  = newBufferPool(4096)
	bufferPool8K  = newBufferPool(8192)
)

func newBufferPool(cap int) *sync.Pool {
	return &sync.Pool{New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, cap))
	}}
}

// PutBuffer puts the buffer back into the pool.
func PutBuffer(pool *sync.Pool, buf *bytes.Buffer) {
	buf.Reset()
	pool.Put(buf)
}

// GetBuffer returns a buffer from the befitting pool,
// which can be released into the original pool.
func GetBuffer(cap int) (*sync.Pool, *bytes.Buffer) {
	if cap <= 64 {
		return bufferPool64, bufferPool64.Get().(*bytes.Buffer)
	} else if cap <= 128 {
		return bufferPool128, bufferPool128.Get().(*bytes.Buffer)
	} else if cap <= 256 {
		return bufferPool256, bufferPool256.Get().(*bytes.Buffer)
	} else if cap <= 512 {
		return bufferPool512, bufferPool512.Get().(*bytes.Buffer)
	} else if cap <= 1024 {
		return bufferPool1K, bufferPool1K.Get().(*bytes.Buffer)
	} else if cap <= 2048 {
		return bufferPool2K, bufferPool2K.Get().(*bytes.Buffer)
	} else if cap <= 4096 {
		return bufferPool4K, bufferPool4K.Get().(*bytes.Buffer)
	} else {
		return bufferPool8K, bufferPool8K.Get().(*bytes.Buffer)
	}
}
