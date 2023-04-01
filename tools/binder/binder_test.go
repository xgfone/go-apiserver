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

package binder

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// Int is the customized int.
type Int int

// UnmarshalBind implements the interface Unmarshaler.
func (i *Int) UnmarshalBind(src interface{}) (err error) {
	switch v := src.(type) {
	case int:
		*i = Int(v)
	case string:
		var _v int64
		_v, err = strconv.ParseInt(v, 10, 64)
		if err == nil {
			*i = Int(_v)
		}
	default:
		err = fmt.Errorf("unsupport to convert %T to Int", src)
	}
	return
}

func (i Int) String() string {
	return fmt.Sprint(int64(i))
}

// Struct is the customized struct.
type Struct struct {
	Name string
	Age  Int
}

// UnmarshalBind implements the interface Unmarshaler.
func (s *Struct) UnmarshalBind(src interface{}) (err error) {
	if maps, ok := src.(map[string]interface{}); ok {
		s.Name, _ = maps["Name"].(string)
		err = s.Age.UnmarshalBind(maps["Age"])
		return
	}
	return fmt.Errorf("unsupport to convert %T to a struct", src)
}

func (s Struct) String() string {
	return fmt.Sprintf("Name=%s, Age=%d", s.Name, s.Age)
}

func ExampleBindStructFromMap() {
	type Person struct {
		Name  string    `json:"name"`
		Age   *int      `json:"age"`
		Birth time.Time `json:"birth"`
	}

	// For basic types
	fmt.Println("------ Basic ------")
	var basicStruct struct {
		Myself Person `json:"myself"`
		Person

		Other struct {
			Class  int    `json:"class"`
			Ignore string `json:"-"`
		} `json:",squash"` // squash is equal to the anonymous struct.

		Duration time.Duration `json:"duration"`
	}
	basicMap := map[string]interface{}{
		"myself": map[string]interface{}{
			"name":  "Aaron",
			"age":   "18",                   // string => int, it supports to convert between different types.
			"birth": "2023-02-01T00:00:00Z", // string => time.Time
		},

		"name":  "Venus",
		"age":   20,
		"birth": 1672531200, // int(unix timestamp) => time.Time

		"class":  1,
		"Ingore": "abc",
		"ignore": "xyz",

		"duration": "1s", // or 1000(ms)
	}
	err := BindStructFromMap(&basicStruct, "json", basicMap)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Myself.Name=%s, Myself.Age=%d, Myself.Birth=%s\n",
			basicStruct.Myself.Name, *basicStruct.Myself.Age, basicStruct.Myself.Birth.Format(time.RFC3339))

		fmt.Printf("Person.Name=%s, Person.Age=%d, Person.Birth=%s\n",
			basicStruct.Person.Name, *basicStruct.Person.Age, basicStruct.Person.Birth.Format(time.RFC3339))

		fmt.Printf("Other.Class=%d, Other.Ingore=%s\n",
			basicStruct.Other.Class, basicStruct.Other.Ignore)

		fmt.Printf("Duration=%s\n", basicStruct.Duration)
	}

	// For Container types
	fmt.Println()
	fmt.Println("------ Container ------")
	var containerStruct struct {
		Maps    map[string]interface{} `json:"maps"`
		Slices  []string               `json:"slices"`
		Structs []struct {
			Field int        `json:"field"`
			Query url.Values `json:"query"`
		} `json:"structs"`
	}
	containerMap := map[string]interface{}{
		"maps":   map[string]string{"k1": "v1"},
		"slices": []interface{}{"a", "b"},
		"structs": []map[string]interface{}{
			{
				"field": "123",
				"query": map[string][]string{
					"k2": {"v2"},
				},
			},
			{
				"field": 456,
				"query": map[string][]string{
					"k3": {"v3"},
				},
			},
		},
	}
	err = BindStructFromMap(&containerStruct, "json", containerMap)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Maps: %v\n", containerStruct.Maps)
		fmt.Printf("Slices: %v\n", containerStruct.Slices)
		for i, s := range containerStruct.Structs {
			fmt.Printf("Structs[%d]: Field=%d, Query=%v\n", i, s.Field, s.Query)
		}
	}

	// For Interface types
	fmt.Println()
	fmt.Println("------ Interface ------")
	var iface1 Struct
	var iface2 Int
	var ifaceStruct = struct {
		Interface1 Unmarshaler
		Interface2 Unmarshaler
		Interface3 interface{} // Use to store any type value.
		// Unmarshaler         // Do not use the anonymous interface.
	}{
		Interface1: &iface1, // For Unmarshaler, must be set to a pointer
		Interface2: &iface2, //  to an implementation.
	}
	ifaceMap := map[string]interface{}{
		"Interface1": map[string]interface{}{
			"Name": "Aaron",
			"Age":  18,
		},
		"Interface2": "123",
		"Interface3": "any",
	}
	err = BindStructFromMap(&ifaceStruct, "json", ifaceMap)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Interface1: %s\n", ifaceStruct.Interface1)
		fmt.Printf("Interface2: %v\n", ifaceStruct.Interface2)
		fmt.Printf("Interface3: %v\n", ifaceStruct.Interface3)
	}

	// Output:
	// ------ Basic ------
	// Myself.Name=Aaron, Myself.Age=18, Myself.Birth=2023-02-01T00:00:00Z
	// Person.Name=Venus, Person.Age=20, Person.Birth=2023-01-01T00:00:00Z
	// Other.Class=1, Other.Ingore=
	// Duration=1s
	//
	// ------ Container ------
	// Maps: map[k1:v1]
	// Slices: [a b]
	// Structs[0]: Field=123, Query=map[k2:[v2]]
	// Structs[1]: Field=456, Query=map[k3:[v3]]
	//
	// ------ Interface ------
	// Interface1: Name=Aaron, Age=18
	// Interface2: 123
	// Interface3: any
}
