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

// Package atomic provides an extra atomic operation.
package atomic

import "sync/atomic"

type valueWrapper struct{ Value interface{} }

// Value is the same as sync/atomic.Value, but requires that the type
// of the stored value is consistent.
//
// Notice: the stored value is allowed to be nil or any interfaces.
type Value struct{ value atomic.Value }

// NewValue returns a new Value with the init value.
func NewValue(initValue interface{}) *Value {
	v := new(Value)
	v.Store(initValue)
	return v
}

// Load refers to sync/atomic.Value#Load.
func (v *Value) Load() interface{} {
	return v.value.Load().(valueWrapper).Value
}

// Store refers to sync/atomic.Value#Store.
func (v *Value) Store(val interface{}) {
	v.value.Store(valueWrapper{val})
}

// CompareAndSwap refers to sync/atomic.Value#CompareAndSwap.
func (v *Value) CompareAndSwap(old, new interface{}) (swapped bool) {
	return v.value.CompareAndSwap(valueWrapper{old}, valueWrapper{new})
}

// Swap refers to sync/atomic.Value#Swap.
func (v *Value) Swap(new interface{}) (old interface{}) {
	return v.value.Swap(valueWrapper{new}).(valueWrapper).Value
}
