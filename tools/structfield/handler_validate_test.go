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

package structfield

import (
	"fmt"

	_ "github.com/xgfone/go-apiserver/validation/validators/defaults"
)

func ExampleNewValidatorHandler() {
	type S0 struct {
		F4 string `validate:"oneof(\"a\", \"b\")"`
		F5 *[]int `validate:"array(min(1))"`
	}

	type S struct {
		F1 string `validate:"zero || max(8)"`    // General Type
		F2 *int64 `validate:"min(1) && max==10"` // Pointer Type
		F3 []S0
	}

	var v S
	v.F2 = new(int64)
	*v.F2 = 5
	v.F3 = []S0{{F4: "a", F5: &[]int{1, 2}}}
	fmt.Println(Reflect(nil, &v))

	v.F1 = "abc"
	fmt.Println(Reflect(nil, &v))

	v.F1 = "abcdefgxyz"
	fmt.Println(Reflect(nil, &v))

	v.F1 = ""
	*v.F2 = 100
	fmt.Println(Reflect(nil, &v))

	*v.F2 = 1
	v.F3[0].F4 = "c"
	fmt.Println(Reflect(nil, &v))

	v.F3[0].F4 = "b"
	(*v.F3[0].F5)[0] = 0
	fmt.Println(Reflect(nil, &v))

	// Output:
	// <nil>
	// <nil>
	// F1: the string length is greater than 8
	// F2: the integer is greater than 10
	// F4: the string 'c' is not one of [a b]
	// F5: 0th element is invalid: the integer is less than 1
}
