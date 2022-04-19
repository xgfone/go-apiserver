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

package validators

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/xgfone/go-apiserver/helper"
	"github.com/xgfone/go-apiserver/validation"
)

var errNilPointer = fmt.Errorf("unexpected empty pointer")

// Min returns a validator to checks the value is less than i.
//
// Support the types as follow:
//   - Integer, Float: compare the value
//   - String, Array, Slice, Map: compare the length of them
//   - Pointer to types above
//
func Min(i float64) validation.Validator {
	s := strconv.FormatFloat(i, 'f', -1, 64)
	rule := fmt.Sprintf("min(%s)", s)

	errFloat := fmt.Errorf("the float is less than %s", s)
	errInteger := fmt.Errorf("the integer is less than %s", s)
	errString := fmt.Errorf("the string length is less than %s", s)
	errContainer := fmt.Errorf("the length is less than %s", s)
	return validation.NewValidator(rule, func(v interface{}) error {
		switch t := helper.Indirect(v).(type) {
		case nil:
			return errNilPointer

		case int:
			if float64(t) < i {
				return errInteger
			}
		case int8:
			if float64(t) < i {
				return errInteger
			}
		case int16:
			if float64(t) < i {
				return errInteger
			}
		case int32:
			if float64(t) < i {
				return errInteger
			}
		case int64:
			if float64(t) < i {
				return errInteger
			}

		case uint:
			if float64(t) < i {
				return errInteger
			}
		case uint8:
			if float64(t) < i {
				return errInteger
			}
		case uint16:
			if float64(t) < i {
				return errInteger
			}
		case uint32:
			if float64(t) < i {
				return errInteger
			}
		case uint64:
			if float64(t) < i {
				return errInteger
			}

		case float32:
			if float64(t) < i {
				return errFloat
			}
		case float64:
			if t < i {
				return errFloat
			}

		case string:
			if len(t) < int(i) {
				return errString
			}

		default:
			switch vf := reflect.ValueOf(t); vf.Kind() {
			case reflect.Array, reflect.Slice, reflect.Map:
				if vf.Len() < int(i) {
					return errContainer
				}
			default:
				return fmt.Errorf("unsupported type '%T'", v)
			}
		}

		return nil
	})
}

// Max returns a validator to checks the value is greater than i.
//
// Support the types as follow:
//   - Integer, Float: compare the value
//   - String, Array, Slice, Map: compare the length of them
//
func Max(i float64) validation.Validator {
	s := strconv.FormatFloat(i, 'f', -1, 64)
	rule := fmt.Sprintf("max(%s)", s)

	errFloat := fmt.Errorf("the float is greater than %s", s)
	errInteger := fmt.Errorf("the integer is greater than %s", s)
	errString := fmt.Errorf("the string length is greater than %s", s)
	errContainer := fmt.Errorf("the length is greater than %s", s)
	return validation.NewValidator(rule, func(v interface{}) error {
		switch t := helper.Indirect(v).(type) {
		case nil:
			return errNilPointer

		case int:
			if float64(t) > i {
				return errInteger
			}
		case int8:
			if float64(t) > i {
				return errInteger
			}
		case int16:
			if float64(t) > i {
				return errInteger
			}
		case int32:
			if float64(t) > i {
				return errInteger
			}
		case int64:
			if float64(t) > i {
				return errInteger
			}

		case uint:
			if float64(t) > i {
				return errInteger
			}
		case uint8:
			if float64(t) > i {
				return errInteger
			}
		case uint16:
			if float64(t) > i {
				return errInteger
			}
		case uint32:
			if float64(t) > i {
				return errInteger
			}
		case uint64:
			if float64(t) > i {
				return errInteger
			}

		case float32:
			if float64(t) > i {
				return errFloat
			}
		case float64:
			if t > i {
				return errFloat
			}

		case string:
			if len(t) > int(i) {
				return errString
			}

		default:
			switch vf := reflect.ValueOf(t); vf.Kind() {
			case reflect.Array, reflect.Slice, reflect.Map:
				if vf.Len() > int(i) {
					return errContainer
				}
			default:
				return fmt.Errorf("unsupported type '%T'", v)
			}
		}

		return nil
	})
}

// Range returns a validator to checks the value is in range [smallest, biggest],
// which is equal to And(Min(smallest), Max(biggest)).
//
// Support the types as follow:
//   - Integer, Float: compare the value
//   - String, Array, Slice, Map: compare the length of them
//
func Range(smallest, biggest float64) validation.Validator {
	left := strconv.FormatFloat(smallest, 'f', -1, 64)
	right := strconv.FormatFloat(biggest, 'f', -1, 64)
	rule := fmt.Sprintf("range(%s, %s)", left, right)

	errFloat := fmt.Errorf("the float is not in range [%s, %s]", left, right)
	errInteger := fmt.Errorf("the integer is not in range [%s, %s]", left, right)
	errString := fmt.Errorf("the string length is not in range [%s, %s]", left, right)
	errContainer := fmt.Errorf("the length is not in range [%s, %s]", left, right)
	return validation.NewValidator(rule, func(v interface{}) error {
		switch t := helper.Indirect(v).(type) {
		case nil:
			return errNilPointer

		case int:
			if inRange(float64(t), smallest, biggest) {
				return errInteger
			}
		case int8:
			if inRange(float64(t), smallest, biggest) {
				return errInteger
			}
		case int16:
			if inRange(float64(t), smallest, biggest) {
				return errInteger
			}
		case int32:
			if inRange(float64(t), smallest, biggest) {
				return errInteger
			}
		case int64:
			if inRange(float64(t), smallest, biggest) {
				return errInteger
			}

		case uint:
			if inRange(float64(t), smallest, biggest) {
				return errInteger
			}
		case uint8:
			if inRange(float64(t), smallest, biggest) {
				return errInteger
			}
		case uint16:
			if inRange(float64(t), smallest, biggest) {
				return errInteger
			}
		case uint32:
			if inRange(float64(t), smallest, biggest) {
				return errInteger
			}
		case uint64:
			if inRange(float64(t), smallest, biggest) {
				return errInteger
			}

		case float32:
			if inRange(float64(t), smallest, biggest) {
				return errFloat
			}
		case float64:
			if inRange(t, smallest, biggest) {
				return errFloat
			}

		case string:
			if inRange(float64(len(t)), smallest, biggest) {
				return errString
			}

		default:
			switch vf := reflect.ValueOf(t); vf.Kind() {
			case reflect.Array, reflect.Slice, reflect.Map:
				if inRange(float64(vf.Len()), smallest, biggest) {
					return errContainer
				}

			default:
				return fmt.Errorf("unsupported type '%T'", v)
			}
		}

		return nil
	})
}

func inRange(v, smallest, biggest float64) bool {
	return smallest <= v && v <= biggest
}
