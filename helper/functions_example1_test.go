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

package helper

import "fmt"

func ExampleUnwrapAll() {
	err1 := fmt.Errorf("err1")
	err2 := fmt.Errorf("err2: %w", err1)
	err3 := fmt.Errorf("err3: %w", err2)
	err4 := fmt.Errorf("err4: %w", err3)

	err := UnwrapAll[error](err4)
	fmt.Println(err)

	// Output:
	// err1
}

func ExampleIndirect() {
	fmt.Println(Indirect(123))   // 123
	fmt.Println(Indirect("abc")) // abc

	var p *int
	fmt.Println(Indirect(p)) // <nil>

	p = new(int)
	fmt.Println(Indirect(p)) // 0

	*p = 123
	fmt.Println(Indirect(p)) // 123

	var pp **int
	fmt.Println(Indirect(pp)) // <nil>

	pp = new(*int)
	fmt.Println(Indirect(pp)) // <nil>

	pt := new(int)
	*pt = 456
	*pp = pt
	fmt.Println(Indirect(pp)) // 456

	**pp = *p
	fmt.Println(Indirect(pp)) // 123
	fmt.Println(Indirect(pt)) // 123

	var i interface{}
	fmt.Println(Indirect(i)) // <nil>

	i = p
	fmt.Println(Indirect(i))  // 123
	fmt.Println(Indirect(&i)) // 123

	// Output:
	// 123
	// abc
	// <nil>
	// 0
	// 123
	// <nil>
	// <nil>
	// 456
	// 123
	// 123
	// <nil>
	// 123
	// 123
}
