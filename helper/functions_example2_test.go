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

type (
	getter  string
	wrapper string
)

func (g getter) Get() string     { return string(g) }
func (w wrapper) Unwrap() string { return string(w) }

func ExampleUnwrap() {
	s, ok := Unwrap[string](getter("a"))
	fmt.Println(s, ok)

	s, ok = Unwrap[string](wrapper("b"))
	fmt.Println(s, ok)

	s, ok = Unwrap[string]("c")
	fmt.Println(s, ok)

	func() {
		defer func() { fmt.Println("panic:", recover()) }()
		Unwrap[string](123)
	}()

	// Output:
	// a true
	// b true
	// c false
	// panic: interface conversion: interface {} is int, not string
}
