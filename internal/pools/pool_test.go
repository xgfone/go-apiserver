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

import "testing"

func TestGetBuffer(t *testing.T) {
	p, b := GetBuffer(256)
	if p == nil || b == nil {
		t.Errorf("failed to get the Buffer")
	} else if b.Cap() != 256 {
		t.Errorf("expect cap %d, but got %d", 256, b.Cap())
	} else if b.Len() != 0 {
		t.Errorf("expect len 0, but got %d", b.Len())
	}
	PutBuffer(p, b)

	p, b = GetBuffer(64 * 1024)

	if p == nil || b == nil {
		t.Errorf("failed to get the Buffer")
	} else if b.Cap() != 64*1024 {
		t.Errorf("expect cap %d, but got %d", 64*1024, b.Cap())
	} else if b.Len() != 0 {
		t.Errorf("expect len 0, but got %d", b.Len())
	}
	PutBuffer(p, b)

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("exepct an error, but got nil")
			}
		}()

		p, b := GetBuffer(123)
		PutBuffer(p, b)
	}()
}
