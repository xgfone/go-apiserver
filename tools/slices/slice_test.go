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

package slices

import (
	"fmt"
	"testing"
)

func TestMax(t *testing.T) {
	if v := Max([]int{3, 2, 5, 4, 1}); v != 5 {
		t.Errorf("expect 5, but got %v", v)
	}

	if v := Max([]int{}); v != 0 {
		t.Errorf("expect 0, but got %v", v)
	}
}

func TestMin(t *testing.T) {
	if v := Min([]int{3, 2, 5, 4, 1}); v != 1 {
		t.Errorf("expect 1, but got %v", v)
	}

	if v := Min([]int{}); v != 0 {
		t.Errorf("expect 0, but got %v", v)
	}
}

func ExampleConvert() {
	type Ints []int

	ints1 := []int{1, 2, 3}
	ints2 := Ints{4, 5, 6}
	int64s1 := Convert(ints1, func(v int) int64 { return int64(v) })
	int64s2 := Convert(ints2, func(v int) int64 { return int64(v) })

	fmt.Println(int64s1)
	fmt.Println(int64s2)

	// Output:
	// [1 2 3]
	// [4 5 6]
}

func ExampleMerge() {
	type Slice1 []int64
	type Slice2 []int

	slice1 := []int64{1, 2}
	slice2 := Slice2{1, 2}

	slice1 = Merge(slice1, Slice2{3, 4}, func(v int) int64 { return int64(v) })
	slice2 = Merge(slice2, Slice1{3, 4}, func(v int64) int { return int(v) })

	fmt.Println(slice1)
	fmt.Println(slice2)

	// Output:
	// [1 2 3 4]
	// [1 2 3 4]
}

func ExampleIndex() {
	ints1 := []int{2, 1, 3, 1, 4}
	fmt.Println(Index(ints1, 0))
	fmt.Println(Index(ints1, 1))

	type Ints []int
	ints2 := Ints{2, 1, 3, 1, 4}
	fmt.Println(Index(ints2, 0))
	fmt.Println(Index(ints2, 1))

	// Output:
	// -1
	// 1
	// -1
	// 1
}

func ExampleLastIndex() {
	ints1 := []int{2, 1, 3, 1, 4}
	fmt.Println(LastIndex(ints1, 0))
	fmt.Println(LastIndex(ints1, 1))

	type Ints []int
	ints2 := Ints{2, 1, 3, 1, 4}
	fmt.Println(LastIndex(ints2, 0))
	fmt.Println(LastIndex(ints2, 1))

	// Output:
	// -1
	// 3
	// -1
	// 3
}

func ExampleSetEqual() {
	s1 := []string{"a", "b", "c"}
	s2 := []string{"b", "c", "a"}
	s3 := []string{"a", "b", "b"}
	if SetEqual(s1, s2) {
		fmt.Printf("%v is equal to %v\n", s1, s2)
	} else {
		fmt.Printf("%v is not equal to %v\n", s1, s2)
	}

	if SetEqual(s1, s3) {
		fmt.Printf("%v is equal to %v\n", s1, s3)
	} else {
		fmt.Printf("%v is not equal to %v\n", s1, s3)
	}

	// Output:
	// [a b c] is equal to [b c a]
	// [a b c] is not equal to [a b b]
}

func ExampleContains() {
	fmt.Println(Contains([]int{1, 2, 3}, 0))
	fmt.Println(Contains([]int{1, 2, 3}, 1))
	fmt.Println(Contains([]int{1, 2, 3}, 2))
	fmt.Println(Contains([]int{1, 2, 3}, 3))
	fmt.Println(Contains([]int{1, 2, 3}, 4))

	// Output:
	// false
	// true
	// true
	// true
	// false
}

func ExampleReverse() {
	vs1 := []string{"a", "b", "c", "d"}
	Reverse(vs1)
	fmt.Println(vs1)

	vs2 := []int{1, 2, 3, 4}
	Reverse(vs2)
	fmt.Println(vs2)

	// Output:
	// [d c b a]
	// [4 3 2 1]
}

func ExampleToInterfaces() {
	ss := []string{"a", "b", "c"}
	vs1 := ToInterfaces(ss)
	fmt.Printf("%T: %v\n", vs1, vs1)

	ints := []int{1, 2, 3}
	vs2 := ToInterfaces(ints)
	fmt.Printf("%T: %v\n", vs2, vs2)

	// Output:
	// []interface {}: [a b c]
	// []interface {}: [1 2 3]
}
