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

func ExampleGetFieldTag() {
	type T struct {
		Int   int
		Int8  int8  `json:""`
		Int16 int16 `json:"int16"`
		Int32 int32 `json:",arg"`
		Int64 int64 `json:"int64,arg"`
		Bool  bool  `json:"-"`
	}

	stype := reflect.TypeOf(T{})
	for i := 0; i < stype.NumField(); i++ {
		field := stype.Field(i)
		fieldName, tagName, tagArg := GetFieldTag(field, "json")
		fmt.Printf("field=%s, tag=%s, arg=%s\n", fieldName, tagName, tagArg)
	}

	// Output:
	// field=Int, tag=, arg=
	// field=Int8, tag=, arg=
	// field=int16, tag=int16, arg=
	// field=Int32, tag=, arg=arg
	// field=int64, tag=int64, arg=arg
	// field=, tag=, arg=
	//
}
