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
