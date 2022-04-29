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

package atomic

import "testing"

func TestValue(t *testing.T) {
	var v Value
	v.Store(111)

	if i, ok := v.Load().(int); !ok || i != 111 {
		t.Errorf("expect int '%d', but got %T '%v'", 111, i, i)
	}

	old := v.Swap(222)
	if i, ok := old.(int); !ok || i != 111 {
		t.Errorf("expect int '%d', but got %T '%v'", 111, i, i)
	}
	if i, ok := v.Load().(int); !ok || i != 222 {
		t.Errorf("expect int '%d', but got %T '%v'", 222, i, i)
	}

	if v.CompareAndSwap(111, 333) {
		t.Error("unexpected CompareAndSwap from 111 to 333")
	} else if !v.CompareAndSwap(222, 333) {
		t.Error("fail to CompareAndSwap from 222 to 333")
	} else if current := v.Load().(int); current != 333 {
		t.Errorf("expect int '%d', but got '%d'", 333, current)
	}
}
