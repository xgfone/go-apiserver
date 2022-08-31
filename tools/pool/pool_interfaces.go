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

// Pre-define some interfaces pools with the different capacity.
var (
	Interfaces8   = NewInterfacesPool(8)
	Interfaces16  = NewInterfacesPool(16)
	Interfaces32  = NewInterfacesPool(32)
	Interfaces64  = NewInterfacesPool(64)
	Interfaces128 = NewInterfacesPool(128)
	Interfaces256 = NewInterfacesPool(256)
	Interfaces512 = NewInterfacesPool(512)
	Interfaces1K  = NewInterfacesPool(1024)
)

// GetInterfaces returns an interfaces from the befitting pool,
// which can be released into the original pool by calling the release function.
func GetInterfaces(cap int) *Interfaces {
	if cap <= 8 {
		return Interfaces8.Get()
	} else if cap <= 16 {
		return Interfaces16.Get()
	} else if cap <= 32 {
		return Interfaces32.Get()
	} else if cap <= 64 {
		return Interfaces64.Get()
	} else if cap <= 128 {
		return Interfaces128.Get()
	} else if cap <= 256 {
		return Interfaces256.Get()
	} else if cap <= 512 {
		return Interfaces512.Get()
	} else {
		return Interfaces1K.Get()
	}
}

// Interfaces is used to enclose []interface.
type Interfaces struct {
	Interfaces []interface{}
	pool       *InterfacesPool
}

// Release releases the interfaces into the original pool.
func (i *Interfaces) Release() {
	if i != nil {
		i.pool.Put(i)
	}
}

// InterfacesPool is the pool to allocate the interfaces.
type InterfacesPool sync.Pool

// NewInterfacesPool returns a new interfaces pool.
func NewInterfacesPool(cap int) *InterfacesPool {
	pool := new(InterfacesPool)
	(*sync.Pool)(pool).New = func() interface{} {
		return &Interfaces{pool: pool, Interfaces: make([]interface{}, 0, cap)}
	}
	return pool
}

// Get returns an interfaces from the pool.
func (p *InterfacesPool) Get() *Interfaces {
	i := (*sync.Pool)(p).Get().(*Interfaces)
	i.Interfaces = i.Interfaces[:0]
	return i
}

// Put puts the interfaces back into the pool.
func (p *InterfacesPool) Put(i *Interfaces) { (*sync.Pool)(p).Put(i) }
