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

package pool

import (
	"bytes"
	"sync"
)

// Pre-define some buffer pools with the different capacity.
var (
	Buffer64  = NewBufferPool(64)
	Buffer128 = NewBufferPool(128)
	Buffer256 = NewBufferPool(256)
	Buffer512 = NewBufferPool(512)
	Buffer1K  = NewBufferPool(1024)
	Buffer2K  = NewBufferPool(2048)
	Buffer4K  = NewBufferPool(4096)
	Buffer8K  = NewBufferPool(8192)
)

// GetBuffer returns a buffer from the befitting pool, which can be released
// into the original pool by calling the release function.
func GetBuffer(cap int) *Buffer {
	if cap <= 64 {
		return Buffer64.Get()
	} else if cap <= 128 {
		return Buffer128.Get()
	} else if cap <= 256 {
		return Buffer256.Get()
	} else if cap <= 512 {
		return Buffer512.Get()
	} else if cap <= 1024 {
		return Buffer1K.Get()
	} else if cap <= 2048 {
		return Buffer2K.Get()
	} else if cap <= 4096 {
		return Buffer4K.Get()
	} else {
		return Buffer8K.Get()
	}
}

// Buffer is used to enclose bytes.Buffer.
type Buffer struct {
	*bytes.Buffer

	pool *BufferPool
}

// Release releases the buffer into the original pool.
func (b *Buffer) Release() {
	if b != nil {
		b.pool.Put(b)
	}
}

// BufferPool is the pool to allocate the Buffer.
type BufferPool sync.Pool

// NewBufferPool returns a new buffer pool.
func NewBufferPool(cap int) *BufferPool {
	pool := new(BufferPool)
	(*sync.Pool)(pool).New = func() interface{} {
		return &Buffer{pool: pool, Buffer: bytes.NewBuffer(make([]byte, 0, cap))}
	}
	return pool
}

// Get returns a buffer from the pool.
func (p *BufferPool) Get() *Buffer {
	buf := (*sync.Pool)(p).Get().(*Buffer)
	buf.Reset()
	return buf
}

// Put puts the buffer back into the pool.
func (p *BufferPool) Put(b *Buffer) { (*sync.Pool)(p).Put(b) }
