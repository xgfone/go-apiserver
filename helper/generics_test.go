// Copyright 2023 xgfone
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

package helper

import "testing"

func TestCompare(t *testing.T) {
	if v := Compare(1, 2); v != -1 {
		t.Errorf("expect -1, but got %v", v)
	}
	if v := Compare(1, 1); v != 0 {
		t.Errorf("expect 0, but got %v", v)
	}
	if v := Compare(2, 1); v != 1 {
		t.Errorf("expect 1, but got %v", v)
	}
}

func TestMax(t *testing.T) {
	if m := Max(1, 2); m != 2 {
		t.Errorf("expect 2, but got %v", m)
	}
	if m := Max(2, 1); m != 2 {
		t.Errorf("expect 2, but got %v", m)
	}
}

func TestMin(t *testing.T) {
	if m := Min(1, 2); m != 1 {
		t.Errorf("expect 1, but got %v", m)
	}
	if m := Min(2, 1); m != 1 {
		t.Errorf("expect 1, but got %v", m)
	}
}
