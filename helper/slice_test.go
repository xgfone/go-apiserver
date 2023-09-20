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

func TestMake(t *testing.T) {
	if _cap := cap(MakeSlice[[]string](0, 0, 0)); _cap != 0 {
		t.Errorf("expect cap %d, but got %d", 0, _cap)
	}

	if _cap := cap(MakeSlice[[]string](0, 0, 1)); _cap != 1 {
		t.Errorf("expect cap %d, but got %d", 1, _cap)
	}
	if _len := len(MakeSlice[[]string](0, 0, 1)); _len != 0 {
		t.Errorf("expect len %d, but got %d", 0, _len)
	}

	if _cap := cap(MakeSlice[[]string](1, 0, 0)); _cap != 1 {
		t.Errorf("expect cap %d, but got %d", 1, _cap)
	}
	if _len := cap(MakeSlice[[]string](1, 0, 0)); _len != 1 {
		t.Errorf("expect len %d, but got %d", 1, _len)
	}
}
