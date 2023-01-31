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
	"fmt"
	"net"
	"reflect"
)

func ExampleImplements() {
	var v net.AddrError

	if Implements(v, (*error)(nil)) {
		fmt.Println("net.AddrError has implemented the interface error")
	} else {
		fmt.Println("net.AddrError has not implemented the interface error")
	}

	if Implements(&v, (*error)(nil)) {
		fmt.Println("*net.AddrError has implemented the interface error")
	} else {
		fmt.Println("*net.AddrError has not implemented the interface error")
	}

	// Output:
	// net.AddrError has not implemented the interface error
	// *net.AddrError has implemented the interface error
}

func ExampleFillNilPtr() {
	var (
		iv int      // For the base type
		ip *int     // For the pointer to the base type
		s  struct { // For the struct & field
			StrV string  // For the struct field value
			StrP *string // For the pointer to the struct field value
		}
	)

	fill := func(v reflect.Value) reflect.Value {
		if v = FillNilPtr(v); IsPointer(v) {
			v = v.Elem()
		}
		return v
	}

	structvalue := reflect.ValueOf(&s).Elem()
	fill(structvalue.Field(0)).SetString("abc")
	fill(structvalue.Field(1)).SetString("xyz")
	fill(reflect.ValueOf(&iv).Elem()).SetInt(123)
	fill(reflect.ValueOf(&ip).Elem()).SetInt(456)

	fmt.Println(iv)
	fmt.Println(*ip)
	fmt.Println(s.StrV)
	fmt.Println(*s.StrP)

	// Output:
	// 123
	// 456
	// abc
	// xyz
}
