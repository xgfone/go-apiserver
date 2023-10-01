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

import (
	"reflect"
	"testing"
)

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

func TestMergeSlice(t *testing.T) {
	s1 := []int{1, 2}
	s2 := []int{3, 4}

	if v := MergeSlice[[]int](); v != nil {
		t.Errorf("expect a nil, but got %T", v)
	}
	if v := MergeSlice[[]int](nil); v != nil {
		t.Errorf("expect a nil, but got %T", v)
	}

	if v := MergeSlice[[]int](nil, nil); v != nil {
		t.Errorf("expect a nil, but got %T", v)
	} else if cap := cap(v); cap != 0 {
		t.Errorf("expect cap==%d, but got %d", 0, cap)
	}
	if v := MergeSlice[[]int](nil, nil, nil); v != nil {
		t.Errorf("expect a nil, but got %T", v)
	} else if cap := cap(v); cap != 0 {
		t.Errorf("expect cap==%d, but got %d", 0, cap)
	}
	if v := MergeSlice[[]int](nil, nil, []int{}); v == nil {
		t.Errorf("got unexpected nil")
	} else if cap := cap(v); cap != 0 {
		t.Errorf("expect cap==%d, but got %d", 0, cap)
	}

	if v := MergeSlice[[]int](s1); !reflect.DeepEqual(s1, v) {
		t.Errorf("expect %v, but got %v", s1, v)
	}
	if v := MergeSlice[[]int](s1, nil); !reflect.DeepEqual(s1, v) {
		t.Errorf("expect %v, but got %v", s1, v)
	}
	if v := MergeSlice[[]int](nil, s1); !reflect.DeepEqual(s1, v) {
		t.Errorf("expect %v, but got %v", s1, v)
	}

	expect := []int{1, 2, 3, 4}
	if v := MergeSlice[[]int](s1, s2); !reflect.DeepEqual(expect, v) {
		t.Errorf("expect %v, but got %v", expect, v)
	}

	expect = []int{3, 4, 1, 2}
	if v := MergeSlice[[]int](s2, s1); !reflect.DeepEqual(expect, v) {
		t.Errorf("expect %v, but got %v", expect, v)
	}

	expect = []int{3, 4, 1, 2, 5, 6}
	if v := MergeSlice[[]int](s2, s1, []int{5, 6}); !reflect.DeepEqual(expect, v) {
		t.Errorf("expect %v, but got %v", expect, v)
	}
}
