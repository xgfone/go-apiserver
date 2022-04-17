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

package validation_test

import (
	"fmt"

	"github.com/xgfone/go-apiserver/validation"
	"github.com/xgfone/go-apiserver/validation/helper"
	"github.com/xgfone/go-apiserver/validation/validators"
)

func ExampleBuilder() {
	helper.RegisterFuncNoArg(validation.DefaultBuilder, "zero", validators.Zero)
	helper.RegisterFuncOneFloat(validation.DefaultBuilder, "min", validators.Min)
	helper.RegisterFuncOneFloat(validation.DefaultBuilder, "max", validators.Max)
	helper.RegisterFuncStrings(validation.DefaultBuilder, "oneof", validators.OneOf)
	helper.RegisterFuncValidators(validation.DefaultBuilder, "array", validation.Array)
	helper.RegisterFuncValidators(validation.DefaultBuilder, "mapk", validation.MapK)
	helper.RegisterFuncValidators(validation.DefaultBuilder, "mapv", validation.MapV)
	helper.RegisterFuncValidators(validation.DefaultBuilder, "mapkv", validation.MapKV)

	// Example 1: function mode
	fmt.Println("\n--- Function Mode ---")

	c := validation.NewContext()
	err := validation.Build(c, "min(1) && max(10)")
	if err != nil {
		fmt.Println(err)
		return
	}

	validator := c.Validator()
	fmt.Printf("Rule: %s\n", validator.String())
	fmt.Println(validator.Validate(0))
	fmt.Println(validator.Validate(1))
	fmt.Println(validator.Validate(5))
	fmt.Println(validator.Validate(10))
	fmt.Println(validator.Validate(11))

	// Example 2: identity+operator mode
	fmt.Println("\n--- Identity+Operator Mode ---")

	c = validation.NewContext()
	err = validation.Build(c, "zero || (min==3 && max==10)")
	if err != nil {
		fmt.Println(err)
		return
	}

	validator = c.Validator()
	fmt.Printf("Rule: %s\n", validator.String())
	fmt.Println(validator.Validate(""))
	fmt.Println(validator.Validate("a"))
	fmt.Println(validator.Validate("abc"))
	fmt.Println(validator.Validate("abcdefghijklmn"))

	// Example 3: The simpler validation way
	const rule1 = "zero || (min==3 && max==10)"
	fmt.Println()
	fmt.Println(validation.Validate("", rule1))
	fmt.Println(validation.Validate("a", rule1))
	fmt.Println(validation.Validate("abc", rule1))
	fmt.Println(validation.Validate("abcdefghijklmn", rule1))

	// Example 4: Validate the array
	fmt.Println("\n--- Array ---")
	const rule2 = "zero || array(min(1), max(10))"
	fmt.Println(validation.Validate([]int{1, 2, 3}, rule2))
	fmt.Println(validation.Validate([]string{"a", "bc", "def"}, rule2))
	fmt.Println(validation.Validate([]int{}, rule2))
	fmt.Println(validation.Validate([]int{0, 1, 2}, rule2))
	fmt.Println(validation.Validate([]string{"a", "bc", ""}, rule2))

	// Example 5: Valiate the map
	fmt.Println("\n--- Map ---")
	const rule3 = `mapk(min(1) && max(3))`
	fmt.Println(validation.Validate(map[string]int{"a": 123}, rule3))
	fmt.Println(validation.Validate(map[string]int8{"abcd": 123}, rule3))

	const rule4 = `mapv(min==10 && max==100)`
	fmt.Println(validation.Validate(map[string]int16{"a": 10}, rule4))
	fmt.Println(validation.Validate(map[string]int32{"abcd": 123}, rule4))

	// Exampe 6: Validate the struct
	fmt.Println("\n--- Struct ---")
	type s struct {
		F1 string `validate:"zero || max(8)"`
		F2 int    `validate:"min(1) && max==10"`
	}
	fmt.Println(validation.ValidateStruct(s{F1: "", F2: 1}))
	fmt.Println(validation.ValidateStruct(s{F1: "abc", F2: 2}))
	fmt.Println(validation.ValidateStruct(s{F1: "abcdefgxyz", F2: 3}))

	// Only return the error of F1 because And validator uses the short circuit.
	fmt.Println(validation.ValidateStruct(s{F1: "abcdefgxyz", F2: 0}))

	// Example 7: Others
	fmt.Println("\n--- Others ---")
	const oneof = `oneof("a", "b", "c")`
	fmt.Println(validation.Validate("a", oneof))
	fmt.Println(validation.Validate("x", oneof))

	// Output:
	//
	// --- Function Mode ---
	// Rule: (min(1) && max(10))
	// the integer is less than 1
	// <nil>
	// <nil>
	// <nil>
	// the integer is greater than 10
	//
	// --- Identity+Operator Mode ---
	// Rule: (zero || (min(3) && max(10)))
	// <nil>
	// the string length is less than 3
	// <nil>
	// the string length is greater than 10
	//
	// <nil>
	// the string length is less than 3
	// <nil>
	// the string length is greater than 10
	//
	// --- Array ---
	// <nil>
	// <nil>
	// <nil>
	// 0th element is invalid: the integer is less than 1
	// 2th element is invalid: the string length is less than 1
	//
	// --- Map ---
	// <nil>
	// map key 'abcd' is invalid: the string length is greater than 3
	// <nil>
	// map value '123' is invalid: the integer is greater than 100
	//
	// --- Struct ---
	// <nil>
	// <nil>
	// F1: the string length is greater than 8
	// F1: the string length is greater than 8
	//
	// --- Others ---
	// <nil>
	// the string 'x' is not in [a b c]
}
