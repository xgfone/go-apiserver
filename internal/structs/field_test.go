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

package structs

import (
	"fmt"
	"reflect"
)

func ExampleGetAllFields() {
	type Uint uint
	var s struct {
		string
		Uint

		str string
		Int int
	}

	fields := GetAllFields(reflect.TypeOf(s))
	for _, field := range fields {
		fmt.Printf("name=%s, isexported=%v, anonymous=%v\n",
			field.Name, field.IsExported(), field.Anonymous)
	}

	// Output:
	// name=string, isexported=false, anonymous=true
	// name=Uint, isexported=true, anonymous=true
	// name=str, isexported=false, anonymous=false
	// name=Int, isexported=true, anonymous=false
}

func ExampleGetAllFieldsWithTag() {
	type Uint uint
	var s struct {
		string
		Uint

		str string
		Int int64 `json:"int64"`

		int    `json:"-"`
		Ignore int `json:"-"`
	}

	fields := GetAllFieldsWithTag(reflect.TypeOf(s), "json")
	for name, field := range fields {
		fmt.Printf("name=%s, isexported=%v, anonymous=%v\n",
			name, field.IsExported(), field.Anonymous)
	}

	// Output:
	// name=string, isexported=false, anonymous=true
	// name=Uint, isexported=true, anonymous=true
	// name=str, isexported=false, anonymous=false
	// name=int64, isexported=true, anonymous=false
}
