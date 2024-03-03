// Copyright 2024 xgfone
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
	"fmt"
	"slices"
	"sort"
	"testing"
)

func TestMapKeys(t *testing.T) {
	expectints := []int{1, 2}
	intmap := map[int]int{1: 11, 2: 22}
	ints := MapKeys(intmap)
	sort.Ints(ints)
	if !slices.Equal(expectints, ints) {
		t.Errorf("expect %v, but got %v", expectints, ints)
	}

	expectstrs := []string{"a", "b"}
	strmap := map[string]string{"a": "aa", "b": "bb"}
	strs := MapKeys(strmap)
	sort.Strings(strs)
	if !slices.Equal(expectstrs, strs) {
		t.Errorf("expect %v, but got %v", expectstrs, strs)
	}
}

func TestMapValues(t *testing.T) {
	expectints := []int{11, 22}
	intmap := map[int]int{1: 11, 2: 22}
	ints := MapValues(intmap)
	sort.Ints(ints)
	if !slices.Equal(expectints, ints) {
		t.Errorf("expect %v, but got %v", expectints, ints)
	}

	expectstrs := []string{"aa", "bb"}
	strmap := map[string]string{"a": "aa", "b": "bb"}
	strs := MapValues(strmap)
	sort.Strings(strs)
	if !slices.Equal(expectstrs, strs) {
		t.Errorf("expect %v, but got %v", expectstrs, strs)
	}
}

func ExampleMapKeysFunc() {
	type Key struct {
		K string
		V int32
	}
	maps := map[Key]bool{
		{K: "a", V: 1}: true,
		{K: "b", V: 2}: true,
		{K: "c", V: 3}: true,
	}

	keys := MapKeysFunc(maps, func(k Key) string { return k.K })
	slices.Sort(keys)
	fmt.Println(keys)

	// Output:
	// [a b c]
}

func ExampleMapValuesFunc() {
	type Value struct {
		V int
	}
	maps := map[string]Value{
		"a": {V: 1},
		"b": {V: 2},
		"c": {V: 3},
	}

	values := MapValuesFunc(maps, func(v Value) int { return v.V })
	slices.Sort(values)
	fmt.Println(values)

	// Output:
	// [1 2 3]
}

func ExampleToSetMap() {
	setmap := ToSetMap([]string{"a", "b", "c"})
	fmt.Println(setmap)

	// Output:
	// map[a:{} b:{} c:{}]
}

func ExampleToBoolMap() {
	boolmap := ToBoolMap([]string{"a", "b", "c"})
	fmt.Println(boolmap)

	// Output:
	// map[a:true b:true c:true]
}

func ExampleToSetMapFunc() {
	type S struct {
		K string
		V int32
	}

	values := []S{{K: "a", V: 1}, {K: "b", V: 2}, {K: "c", V: 3}}
	setmap := ToSetMapFunc(values, func(s S) string { return s.K })
	fmt.Println(setmap)

	// Output:
	// map[a:{} b:{} c:{}]
}

func ExampleToBoolMapFunc() {
	type S struct {
		K string
		V int32
	}

	values := []S{{K: "a", V: 1}, {K: "b", V: 2}, {K: "c", V: 3}}
	setmap := ToBoolMapFunc(values, func(s S) string { return s.K })
	fmt.Println(setmap)

	// Output:
	// map[a:true b:true c:true]
}
