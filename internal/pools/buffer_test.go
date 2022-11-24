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

import "testing"

func TestBuffer(t *testing.T) {
	pool, buffer := GetBuffer(60)
	if cap := buffer.Cap(); cap != 64 {
		t.Errorf("expect buffer capacity %d, but got %d", 64, cap)
	}
	PutBuffer(pool, buffer)

	pool, buffer = GetBuffer(100)
	if cap := buffer.Cap(); cap != 128 {
		t.Errorf("expect buffer capacity %d, but got %d", 128, cap)
	}
	PutBuffer(pool, buffer)

	pool, buffer = GetBuffer(200)
	if cap := buffer.Cap(); cap != 256 {
		t.Errorf("expect buffer capacity %d, but got %d", 256, cap)
	}
	PutBuffer(pool, buffer)

	pool, buffer = GetBuffer(500)
	if cap := buffer.Cap(); cap != 512 {
		t.Errorf("expect buffer capacity %d, but got %d", 512, cap)
	}
	PutBuffer(pool, buffer)

	pool, buffer = GetBuffer(1000)
	if cap := buffer.Cap(); cap != 1024 {
		t.Errorf("expect buffer capacity %d, but got %d", 1024, cap)
	}
	PutBuffer(pool, buffer)

	pool, buffer = GetBuffer(2000)
	if cap := buffer.Cap(); cap != 2048 {
		t.Errorf("expect buffer capacity %d, but got %d", 2048, cap)
	}
	PutBuffer(pool, buffer)

	pool, buffer = GetBuffer(4000)
	if cap := buffer.Cap(); cap != 4096 {
		t.Errorf("expect buffer capacity %d, but got %d", 4096, cap)
	}
	PutBuffer(pool, buffer)

	pool, buffer = GetBuffer(5000)
	if cap := buffer.Cap(); cap != 8192 {
		t.Errorf("expect buffer capacity %d, but got %d", 8192, cap)
	}
	PutBuffer(pool, buffer)
}
