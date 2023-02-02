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

package atomic

import "testing"

func TestValue(t *testing.T) {
	value := NewValue(123)
	value.Store(true)
	if v, ok := value.Load().(bool); !ok || !v {
		t.Errorf("expect %v, but got %v", true, v)
	}
}

func TestValueSwap(t *testing.T) {
	value := NewValue(true)
	if v, ok := value.Swap(123).(bool); !ok || !v {
		t.Errorf("expect %v, but got %v", true, v)
	}

	if !value.CompareAndSwap(123, false) {
		t.Errorf("expect compare and swap, but fail")
	}

	if value.CompareAndSwap(true, false) {
		t.Errorf("not expect compare and swap, but success")
	}

	if v, ok := value.Load().(bool); !ok || v {
		t.Errorf("expect %v, but got %v", false, v)
	}
}
