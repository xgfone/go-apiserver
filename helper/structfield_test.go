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

import (
	"fmt"
	"reflect"
)

func ExampleGetStructFieldByName() {
	print := func(v interface{}, name string) {
		if vf, ok := GetStructFieldByName(v, name); ok {
			fmt.Printf("%s: %v\n", name, vf.Interface())
		} else {
			fmt.Printf("no the field named \"%s\"\n", name)
		}
	}

	var v struct {
		Field1 int
		Field2 string
		Field3 struct {
			Field4 int
			Field5 string
			Field6 struct {
				Field7 int
				Field8 string
			}
		}
	}
	v.Field1 = 123
	v.Field2 = "abc"
	v.Field3.Field4 = 456
	v.Field3.Field5 = "rst"
	v.Field3.Field6.Field7 = 789
	v.Field3.Field6.Field8 = "xyz"

	fmt.Println("--- Struct ---")
	print(v, ".Field1")
	print(v, "Field2")
	print(v, "Field4")
	print(v, "Field3.Field4")
	print(v, "Field3.Field7")
	print(v, "Field3.Field6.Field7")
	print(v, "Field3.Field6.Field9")

	fmt.Println()
	fmt.Println("--- Pointer to Struct ---")
	print(&v, ".Field1")
	print(&v, "Field2")
	print(&v, "Field4")
	print(&v, "Field3.Field4")
	print(&v, "Field3.Field7")
	print(&v, "Field3.Field6.Field7")
	print(&v, "Field3.Field6.Field9")

	fmt.Println()
	fmt.Println("--- reflect.Value of Struct ---")
	print(reflect.ValueOf(v), ".Field1")
	print(reflect.ValueOf(v), "Field2")
	print(reflect.ValueOf(v), "Field4")
	print(reflect.ValueOf(v), "Field3.Field4")
	print(reflect.ValueOf(v), "Field3.Field7")
	print(reflect.ValueOf(v), "Field3.Field6.Field7")
	print(reflect.ValueOf(v), "Field3.Field6.Field9")

	fmt.Println()
	fmt.Println("--- reflect.Value of Pointer to Struct ---")
	print(reflect.ValueOf(&v), ".Field1")
	print(reflect.ValueOf(&v), "Field2")
	print(reflect.ValueOf(&v), "Field4")
	print(reflect.ValueOf(&v), "Field3.Field4")
	print(reflect.ValueOf(&v), "Field3.Field7")
	print(reflect.ValueOf(&v), "Field3.Field6.Field7")
	print(reflect.ValueOf(&v), "Field3.Field6.Field9")

	// Output:
	// --- Struct ---
	// .Field1: 123
	// Field2: abc
	// no the field named "Field4"
	// Field3.Field4: 456
	// no the field named "Field3.Field7"
	// Field3.Field6.Field7: 789
	// no the field named "Field3.Field6.Field9"
	//
	// --- Pointer to Struct ---
	// .Field1: 123
	// Field2: abc
	// no the field named "Field4"
	// Field3.Field4: 456
	// no the field named "Field3.Field7"
	// Field3.Field6.Field7: 789
	// no the field named "Field3.Field6.Field9"
	//
	// --- reflect.Value of Struct ---
	// .Field1: 123
	// Field2: abc
	// no the field named "Field4"
	// Field3.Field4: 456
	// no the field named "Field3.Field7"
	// Field3.Field6.Field7: 789
	// no the field named "Field3.Field6.Field9"
	//
	// --- reflect.Value of Pointer to Struct ---
	// .Field1: 123
	// Field2: abc
	// no the field named "Field4"
	// Field3.Field4: 456
	// no the field named "Field3.Field7"
	// Field3.Field6.Field7: 789
	// no the field named "Field3.Field6.Field9"
}
