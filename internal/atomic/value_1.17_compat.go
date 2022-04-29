// Copyright 2021 xgfone
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

//go:build !go1.17
// +build !go1.17

package atomic

import (
	"reflect"
	"sync"
	"sync/atomic"
)

// Value is used to update the value atomically.
type Value struct {
	atomic.Value
	sync.Mutex // We only guarantee the Swap and CompareAndSwap methods.
}

func (v *Value) CompareAndSwap(old, new interface{}) (swapped bool) {
	v.Lock()
	current := v.Value.Load()
	if swapped = reflect.DeepEqual(old, current); swapped {
		v.Value.Store(new)
	}
	v.Unlock()
	return
}

func (v *Value) Swap(new interface{}) (old interface{}) {
	v.Lock()
	old = v.Value.Load()
	v.Value.Store(new)
	v.Unlock()
	return
}
