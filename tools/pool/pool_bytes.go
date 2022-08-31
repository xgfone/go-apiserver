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

import "sync"

// Pre-define some bytes pools with the different capacity.
var (
	Bytes64  = NewBytesPool(64)
	Bytes128 = NewBytesPool(128)
	Bytes256 = NewBytesPool(256)
	Bytes512 = NewBytesPool(512)
	Bytes1K  = NewBytesPool(1024)
	Bytes2K  = NewBytesPool(2048)
	Bytes4K  = NewBytesPool(4096)
	Bytes8K  = NewBytesPool(8192)
)

// GetBytes returns a bytes from the befitting pool, which can be released
// into the original pool by calling the release function.
func GetBytes(cap int) *Bytes {
	if cap <= 64 {
		return Bytes64.Get()
	} else if cap <= 128 {
		return Bytes128.Get()
	} else if cap <= 256 {
		return Bytes256.Get()
	} else if cap <= 512 {
		return Bytes512.Get()
	} else if cap <= 1024 {
		return Bytes1K.Get()
	} else if cap <= 2048 {
		return Bytes2K.Get()
	} else if cap <= 4096 {
		return Bytes4K.Get()
	} else {
		return Bytes8K.Get()
	}
}

// Bytes is used to enclose the byte slice []byte.
type Bytes struct {
	Bytes []byte
	pool  *BytesPool
}

// Release releases the bytes into the original pool.
func (b *Bytes) Release() {
	if b != nil {
		b.pool.Put(b)
	}
}

// BytesPool is the pool to allocate the bytes.
type BytesPool sync.Pool

// NewBytesPool returns a new bytes pool.
func NewBytesPool(cap int) *BytesPool {
	pool := new(BytesPool)
	(*sync.Pool)(pool).New = func() interface{} {
		return &Bytes{pool: pool, Bytes: make([]byte, 0, cap)}
	}
	return pool
}

// Get returns a bytes from the pool.
func (p *BytesPool) Get() *Bytes {
	b := (*sync.Pool)(p).Get().(*Bytes)
	b.Bytes = b.Bytes[:0]
	return b
}

// Put puts the bytes back into the pool.
func (p *BytesPool) Put(b *Bytes) { (*sync.Pool)(p).Put(b) }
