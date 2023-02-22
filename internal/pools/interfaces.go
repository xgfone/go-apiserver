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

import "sync"

var (
	interfacesPool8   = newInterfacesPool(8)
	interfacesPool16  = newInterfacesPool(16)
	interfacesPool32  = newInterfacesPool(32)
	interfacesPool64  = newInterfacesPool(64)
	interfacesPool128 = newInterfacesPool(128)
	interfacesPool256 = newInterfacesPool(256)
	interfacesPool512 = newInterfacesPool(512)
	interfacesPool1K  = newInterfacesPool(1024)
)

func newInterfacesPool(cap int) *sync.Pool {
	return &sync.Pool{New: func() interface{} {
		return &Interfaces{Interfaces: make([]interface{}, 0, cap)}
	}}
}

// Interfaces is used to wrap the interface slice.
type Interfaces struct {
	Interfaces []interface{}
}

// PutInterfaces puts the interface slice back into the pool.
func PutInterfaces(pool *sync.Pool, interfaces *Interfaces) {
	interfaces.Interfaces = interfaces.Interfaces[:0]
	pool.Put(interfaces)
}

// GetInterfaces returns an interface slice from the befitting pool,
// which can be released into the original pool.
func GetInterfaces(cap int) (*sync.Pool, *Interfaces) {
	if cap <= 8 {
		return interfacesPool8, interfacesPool8.Get().(*Interfaces)
	} else if cap <= 16 {
		return interfacesPool16, interfacesPool16.Get().(*Interfaces)
	} else if cap <= 32 {
		return interfacesPool32, interfacesPool32.Get().(*Interfaces)
	} else if cap <= 64 {
		return interfacesPool64, interfacesPool64.Get().(*Interfaces)
	} else if cap <= 128 {
		return interfacesPool128, interfacesPool128.Get().(*Interfaces)
	} else if cap <= 256 {
		return interfacesPool256, interfacesPool256.Get().(*Interfaces)
	} else if cap <= 512 {
		return interfacesPool512, interfacesPool512.Get().(*Interfaces)
	} else {
		return interfacesPool1K, interfacesPool1K.Get().(*Interfaces)
	}
}
