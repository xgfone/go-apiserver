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

package maps

import "fmt"

func ExampleConvert() {
	type Maps map[string]int

	intmap1 := Maps{"a": 1, "b": 2}
	intmap2 := map[string]int{"a": 3, "b": 4}

	int64map1 := Convert(intmap1, func(k string, v int) (string, int64) { return k, int64(v) })
	int64map2 := Convert(intmap2, func(k string, v int) (string, int64) { return k, int64(v) })

	fmt.Printf("%T\n", int64map1)
	fmt.Printf("%T\n", int64map2)
	fmt.Printf("%s=%v\n", "a", int64map1["a"])
	fmt.Printf("%s=%v\n", "b", int64map1["b"])
	fmt.Printf("%s=%v\n", "a", int64map2["a"])
	fmt.Printf("%s=%v\n", "b", int64map2["b"])

	// Output:
	// map[string]int64
	// map[string]int64
	// a=1
	// b=2
	// a=3
	// b=4
}

func ExampleConvertValues() {
	type Maps map[string]int

	intmap1 := Maps{"a": 1, "b": 2}
	intmap2 := map[string]int{"a": 3, "b": 4}

	int64map1 := ConvertValues(intmap1, func(v int) int64 { return int64(v) })
	int64map2 := ConvertValues(intmap2, func(v int) int64 { return int64(v) })

	fmt.Printf("%T\n", int64map1)
	fmt.Printf("%T\n", int64map2)
	fmt.Printf("%s=%v\n", "a", int64map1["a"])
	fmt.Printf("%s=%v\n", "b", int64map1["b"])
	fmt.Printf("%s=%v\n", "a", int64map2["a"])
	fmt.Printf("%s=%v\n", "b", int64map2["b"])

	// Output:
	// map[string]int64
	// map[string]int64
	// a=1
	// b=2
	// a=3
	// b=4
}

func ExampleAddSlice() {
	type Map map[string]int
	type Slice []byte

	maps1 := Map{"a": 1}
	maps2 := map[string]int{"a": 1}
	AddSlice(maps1, []byte{'b'}, func(v byte) (string, int) { return string(v), int(v) })
	AddSlice(maps2, Slice{'b'}, func(v byte) (string, int) { return string(v), int(v) })

	fmt.Printf("%s=%v\n", "a", maps1["a"])
	fmt.Printf("%s=%v\n", "b", maps1["b"])
	fmt.Printf("%s=%v\n", "a", maps2["a"])
	fmt.Printf("%s=%v\n", "b", maps2["b"])

	// Output:
	// a=1
	// b=98
	// a=1
	// b=98
}

func ExampleAddSliceAsValue() {
	type Map map[string]int
	type Slice []int

	maps1 := Map{"a": 1}
	maps2 := map[string]int{"a": 1}
	AddSliceAsValue(maps1, []int{2}, func(v int) string { return "b" })
	AddSliceAsValue(maps2, Slice{2}, func(v int) string { return "b" })

	fmt.Printf("%s=%v\n", "a", maps1["a"])
	fmt.Printf("%s=%v\n", "b", maps1["b"])
	fmt.Printf("%s=%v\n", "a", maps2["a"])
	fmt.Printf("%s=%v\n", "b", maps2["b"])

	// Output:
	// a=1
	// b=2
	// a=1
	// b=2
}

func ExampleDeleteSlice() {
	type Map map[string]int
	type Slice []string

	maps1 := Map{"a": 1, "b": 2, "c": 3}
	maps2 := map[string]int{"a": 1, "b": 2, "c": 3}
	DeleteSlice(maps1, []string{"a", "b"})
	DeleteSlice(maps2, Slice{"a", "b"})

	fmt.Println(maps1)
	fmt.Println(maps2)

	// Output:
	// map[c:3]
	// map[c:3]
}

func ExampleDeleteSliceFunc() {
	type Map map[string]int
	type Slice []byte

	maps1 := Map{"a": 1, "b": 2, "c": 3}
	maps2 := map[string]int{"a": 1, "b": 2, "c": 3}
	DeleteSliceFunc(maps1, []byte{'a', 'b'}, func(b byte) string { return string(b) })
	DeleteSliceFunc(maps2, Slice{'a', 'b'}, func(b byte) string { return string(b) })

	fmt.Println(maps1)
	fmt.Println(maps2)

	// Output:
	// map[c:3]
	// map[c:3]
}
