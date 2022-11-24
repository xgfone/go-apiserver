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

func TestInterfaces(t *testing.T) {
	pool, ifaces := GetInterfaces(0)
	if cap := cap(ifaces.Interfaces); cap != 8 {
		t.Errorf("expect interfaces capacity %d, but got %d", 8, cap)
	}
	PutInterfaces(pool, ifaces)

	pool, ifaces = GetInterfaces(10)
	if cap := cap(ifaces.Interfaces); cap != 16 {
		t.Errorf("expect interfaces capacity %d, but got %d", 16, cap)
	}
	PutInterfaces(pool, ifaces)

	pool, ifaces = GetInterfaces(30)
	if cap := cap(ifaces.Interfaces); cap != 32 {
		t.Errorf("expect interfaces capacity %d, but got %d", 32, cap)
	}
	PutInterfaces(pool, ifaces)

	pool, ifaces = GetInterfaces(60)
	if cap := cap(ifaces.Interfaces); cap != 64 {
		t.Errorf("expect interfaces capacity %d, but got %d", 64, cap)
	}
	PutInterfaces(pool, ifaces)

	pool, ifaces = GetInterfaces(100)
	if cap := cap(ifaces.Interfaces); cap != 128 {
		t.Errorf("expect interfaces capacity %d, but got %d", 128, cap)
	}
	PutInterfaces(pool, ifaces)

	pool, ifaces = GetInterfaces(200)
	if cap := cap(ifaces.Interfaces); cap != 256 {
		t.Errorf("expect interfaces capacity %d, but got %d", 256, cap)
	}
	PutInterfaces(pool, ifaces)

	pool, ifaces = GetInterfaces(500)
	if cap := cap(ifaces.Interfaces); cap != 512 {
		t.Errorf("expect interfaces capacity %d, but got %d", 512, cap)
	}
	PutInterfaces(pool, ifaces)

	pool, ifaces = GetInterfaces(1000)
	if cap := cap(ifaces.Interfaces); cap != 1024 {
		t.Errorf("expect interfaces capacity %d, but got %d", 1024, cap)
	}
	PutInterfaces(pool, ifaces)
}
