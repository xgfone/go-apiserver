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
	"testing"

	"github.com/xgfone/go-apiserver/validation"
	"github.com/xgfone/go-apiserver/validation/validators"
	"github.com/xgfone/go-apiserver/validation/validators/defaults"
)

type testT2 struct {
	F1 string `validate:"required"`
	F2 int    `validate:"zero || min==5"`
}

func (t testT2) Validate() error {
	if t.F2 > 0 && t.F2 != len(t.F1) {
		return fmt.Errorf("F2 is not equal to the length of F1")
	}
	return nil
}

func TestBuilderValidateStruct(t *testing.T) {
	defaults.RegisterDefaults(validation.DefaultBuilder)

	var s struct {
		F testT2
	}

	s.F.F1 = "abc"
	if err := validation.ValidateStruct(s); err != nil {
		t.Error(err)
	}

	s.F.F2 = 3
	if err := validation.ValidateStruct(s); err == nil {
		t.Errorf("expect an error, but got nil")
	} else if s := "F.F2: the integer is less than 5"; err.Error() != s {
		t.Errorf("expect the error '%s', but got '%s'", s, err.Error())
	}

	s.F.F2 = 5
	if err := validation.ValidateStruct(s); err == nil {
		t.Errorf("expect an error, but got nil")
	} else if s := "F: F2 is not equal to the length of F1"; err.Error() != s {
		t.Errorf("expect the error '%s', but got '%s'", s, err.Error())
	}

	s.F.F1 = "abcde"
	s.F.F2 = len(s.F.F1)
	if err := validation.ValidateStruct(s); err != nil {
		t.Errorf("unexpected error '%s'", err.Error())
	}
}

func ExampleBuilder() {
	// Register the builder functions.
	validation.RegisterFunction(validation.NewFunctionWithoutArgs("zero", validators.Zero))
	validation.RegisterFunction(validation.NewFunctionWithOneFloat("min", validators.Min))
	validation.RegisterFunction(validation.NewFunctionWithOneFloat("max", validators.Max))
	validation.RegisterFunction(validation.NewFunctionWithStrings("oneof", validators.OneOf))
	validation.RegisterFunction(validation.NewFunctionWithValidators("array", validation.Array))
	validation.RegisterFunction(validation.NewFunctionWithValidators("mapk", validation.MapK))
	validation.RegisterFunction(validation.NewFunctionWithValidators("mapv", validation.MapV))
	validation.RegisterFunction(validation.NewFunctionWithValidators("mapkv", validation.MapKV))

	// Add the global symbols.
	validation.RegisterSymbol("v1", "a")
	validation.RegisterSymbol("v2", "b")

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

	// Example 2: Identifier+operator mode
	fmt.Println("\n--- Identifier+Operator Mode ---")

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
	fmt.Println(validation.BuildValidator(rule4))
	fmt.Println(validation.Validate(map[string]int16{"a": 10}, rule4))
	fmt.Println(validation.Validate(map[string]int32{"abcd": 123}, rule4))

	// Exampe 6: Validate the struct
	fmt.Println("\n--- Struct ---")
	type s struct {
		F1 string `validate:"zero || max(8)"`    // General Type
		F2 *int64 `validate:"min(1) && max==10"` // Pointer Type

		F3 struct { // Embedded Anonymous Struct
			F4 string `validate:"oneof(\"a\", \"b\")"`
			F5 *[]int `validate:"array(min(1))"`
		}
	}
	var v s
	v.F2 = new(int64)
	v.F3.F5 = &[]int{1, 2}
	v.F3.F4 = "a"

	*v.F2 = 1
	fmt.Println(validation.ValidateStruct(v))

	v.F1 = "abc"
	fmt.Println(validation.ValidateStruct(v))

	v.F1 = "abcdefgxyz"
	fmt.Println(validation.ValidateStruct(v))

	v.F1 = ""
	v.F3.F4 = "c"
	fmt.Println(validation.ValidateStruct(v))

	v.F3.F4 = "a"
	(*v.F3.F5)[0] = 0
	fmt.Println(validation.ValidateStruct(v))

	// Example 7: Others
	fmt.Println("\n--- Others ---")
	const oneof = `oneof(v1, v2, "c")`
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
	// --- Identifier+Operator Mode ---
	// Rule: (zero || (min(3) && max(10)))
	// <nil>
	// the string length is less than 3
	// <nil>
	// the string length is greater than 10
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
	// mapv(min(10) && max(100)) <nil>
	// <nil>
	// map value '123' is invalid: the integer is greater than 100
	//
	// --- Struct ---
	// <nil>
	// <nil>
	// F1: the string length is greater than 8
	// F3.F4: the string 'c' is not one of [a b]
	// F3.F5: 0th element is invalid: the integer is less than 1
	//
	// --- Others ---
	// <nil>
	// the string 'x' is not one of [a b c]
}
