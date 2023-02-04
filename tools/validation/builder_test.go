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

package validation

import (
	"fmt"
	"strings"
	"testing"

	"github.com/xgfone/go-apiserver/tools/validation/validator"
)

type testT2 struct {
	F1 string `validate:"required"`
	F2 int    `validate:"zero || min==5"`
}

func (t testT2) Validate(interface{}) error {
	if t.F2 > 0 && t.F2 != len(t.F1) {
		return fmt.Errorf("F2 is not equal to the length of F1")
	}
	return nil
}

func TestRuleRanger(t *testing.T) {
	expectErrMsg := "the integer is not in range [1, 10]"

	if err := Validate(nil, 0, "ranger(1,10)"); err == nil {
		t.Errorf("expect the error, but got nil")
	} else if err.Error() != expectErrMsg {
		t.Errorf("expect the error '%s', but got '%s'", expectErrMsg, err.Error())
	}

	if err := Validate(nil, 1, "ranger(1,10)"); err != nil {
		t.Errorf("unexpect the error: %s", err.Error())
	}

	if err := Validate(nil, 10, "ranger(1,10)"); err != nil {
		t.Errorf("unexpect the error: %s", err.Error())
	}

	if err := Validate(nil, 11, "ranger(1,10)"); err == nil {
		t.Errorf("expect the error, but got nil")
	} else if err.Error() != expectErrMsg {
		t.Errorf("expect the error '%s', but got '%s'", expectErrMsg, err.Error())
	}
}

func TestRuleTimeDuration(t *testing.T) {
	if err := Validate(nil, "1a", `duration`); err == nil {
		t.Errorf("expect an error, but got nil")
	}

	if err := Validate(nil, "1s", `duration`); err != nil {
		t.Errorf("expect nil, but got '%s'", err.Error())
	}

	if err := Validate(nil, "2022-08-07", `timeformat`); err == nil {
		t.Errorf("expect an error, but got nil")
	}

	if err := Validate(nil, "2022-08-07", `dateformat`); err != nil {
		t.Errorf("expect nil, but got '%s'", err.Error())
	}

	if err := Validate(nil, "01:02:03", `timeformat`); err != nil {
		t.Errorf("expect nil, but got '%s'", err.Error())
	}

	if err := Validate(nil, "2022-08-07 01:02:03", `datetimeformat`); err != nil {
		t.Errorf("expect nil, but got '%s'", err.Error())
	}

}

func ExampleValidatorFunction() {
	// New a validator "oneof".
	ss := []string{"one", "two", "three"}
	desc := fmt.Sprintf(`oneof("%s")`, strings.Join(ss, `", "`))
	oneof := validator.NewValidator(desc, func(_, i interface{}) error {
		if s, ok := i.(string); ok {
			for _, _s := range ss {
				if _s == s {
					return nil
				}
			}
			return fmt.Errorf("the string '%s' is not one of %v", s, ss)
		}
		return fmt.Errorf("unsupported type '%T'", i)
	})

	// Register the "oneof" validator as a Function.
	rule := "oneof"
	builder := NewBuilder()
	builder.RegisterFunction(ValidatorFunction(rule, oneof))

	// Print the validator description.
	fmt.Println(oneof.String())

	// Validate the value and print the result.
	fmt.Println(builder.Validate(nil, "one", rule))
	fmt.Println(builder.Validate(nil, "two", rule))
	fmt.Println(builder.Validate(nil, "three", rule))
	fmt.Println(builder.Validate(nil, "four", rule))

	// Output:
	// oneof("one", "two", "three")
	// <nil>
	// <nil>
	// <nil>
	// the string 'four' is not one of [one two three]
}

func ExampleBuilder_RegisterValidatorOneof() {
	const rule = "isnumber"
	numbers := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}

	builder := NewBuilder()
	builder.RegisterValidatorOneof(rule, numbers...)

	// Validate the value and print the result.
	fmt.Println(builder.Validate(nil, "0", rule))
	fmt.Println(builder.Validate(nil, "9", rule))
	fmt.Println(builder.Validate(nil, "a", rule))

	// Output:
	// <nil>
	// <nil>
	// the string 'a' is not one of [0 1 2 3 4 5 6 7 8 9]
}

func ExampleBuilder() {
	////// Register the validator building functions.
	// RegisterFunction(NewFunctionWithOneFloat("min", validator.Min))
	// RegisterFunction(NewFunctionWithOneFloat("max", validator.Max))
	// RegisterFunction(NewFunctionWithStrings("oneof", validator.OneOf))
	// RegisterFunction(NewFunctionWithValidators("array", validator.Array))
	// RegisterFunction(NewFunctionWithValidators("mapk", validator.MapK))
	// RegisterFunction(NewFunctionWithValidators("mapv", validator.MapV))
	// RegisterFunction(NewFunctionWithValidators("mapkv", validator.MapKV))

	////// Register the validator building function based on the bool validation.
	// isZero := func(i interface{}) bool { return reflect.ValueOf(i).IsZero() }
	// RegisterValidatorFuncBool("zero", isZero, fmt.Errorf("the value is expected to be zero"))
	// RegisterValidatorFunc("structure", ValidateStruct)

	// Add the global symbols.
	RegisterSymbol("v1", "a")
	RegisterSymbol("v2", "b")

	// Example 1: function mode
	fmt.Println("\n--- Function Mode ---")

	c := NewContext()
	err := Build(c, "min(1) && max(10)")
	if err != nil {
		fmt.Println(err)
		return
	}

	validator := c.Validator()
	fmt.Printf("Rule: %s\n", validator.String())
	fmt.Println(validator.Validate(nil, 0))
	fmt.Println(validator.Validate(nil, 1))
	fmt.Println(validator.Validate(nil, 5))
	fmt.Println(validator.Validate(nil, 10))
	fmt.Println(validator.Validate(nil, 11))

	// Example 2: Identifier+operator mode
	fmt.Println("\n--- Identifier+Operator Mode ---")

	c = NewContext()
	err = Build(c, "zero || (min==3 && max==10)")
	if err != nil {
		fmt.Println(err)
		return
	}

	validator = c.Validator()
	fmt.Printf("Rule: %s\n", validator.String())
	fmt.Println(validator.Validate(nil, ""))
	fmt.Println(validator.Validate(nil, "a"))
	fmt.Println(validator.Validate(nil, "abc"))
	fmt.Println(validator.Validate(nil, "abcdefghijklmn"))

	// Example 3: The simpler validation way
	const rule1 = "zero || (min==3 && max==10)"
	fmt.Println(Validate(nil, "", rule1))
	fmt.Println(Validate(nil, "a", rule1))
	fmt.Println(Validate(nil, "abc", rule1))
	fmt.Println(Validate(nil, "abcdefghijklmn", rule1))

	// Example 4: Validate the array
	fmt.Println("\n--- Array ---")
	const rule2 = "zero || array(min(1), max(10))"
	fmt.Println(Validate(nil, []int{1, 2, 3}, rule2))
	fmt.Println(Validate(nil, []string{"a", "bc", "def"}, rule2))
	fmt.Println(Validate(nil, []int{}, rule2))
	fmt.Println(Validate(nil, []int{0, 1, 2}, rule2))
	fmt.Println(Validate(nil, []string{"a", "bc", ""}, rule2))

	// Example 5: Valiate the map
	fmt.Println("\n--- Map ---")
	const rule3 = `mapk(min(1) && max(3))`
	fmt.Println(Validate(nil, map[string]int{"a": 123}, rule3))
	fmt.Println(Validate(nil, map[string]int8{"abcd": 123}, rule3))

	const rule4 = `mapv(min==10 && max==100)`
	fmt.Println(BuildValidator(rule4))
	fmt.Println(Validate(nil, map[string]int16{"a": 10}, rule4))
	fmt.Println(Validate(nil, map[string]int32{"abcd": 123}, rule4))

	// Example 6: Others
	fmt.Println("\n--- Others ---")
	const oneof = `oneof(v1, v2, "c")`
	fmt.Println(Validate(nil, "a", oneof))
	fmt.Println(Validate(nil, "x", oneof))

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
	// --- Others ---
	// <nil>
	// the string 'x' is not one of [a b c]
}
