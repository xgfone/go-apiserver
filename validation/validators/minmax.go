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

	"github.com/xgfone/go-apiserver/validation"
)

// Min returns a validator to checks the value is less than i.
//
// Support the types as follow:
//   - Integer, Float: compare the value
//   - String, Array, Slice, Map: compare the length of them
//
func Min(i float64) validation.Validator {
	s := strconv.FormatFloat(i, 'f', -1, 64)
	rule := fmt.Sprintf("min(%s)", s)
	return validation.NewValidator(rule, func(v interface{}) error {
		switch t := v.(type) {
		case int:
			if float64(t) < i {
				return fmt.Errorf("the integer is less than %s", s)
			}
		case int8:
			if float64(t) < i {
				return fmt.Errorf("the integer is less than %s", s)
			}
		case int16:
			if float64(t) < i {
				return fmt.Errorf("the integer is less than %s", s)
			}
		case int32:
			if float64(t) < i {
				return fmt.Errorf("the integer is less than %s", s)
			}
		case int64:
			if float64(t) < i {
				return fmt.Errorf("the integer is less than %s", s)
			}

		case uint:
			if float64(t) < i {
				return fmt.Errorf("the integer is less than %s", s)
			}
		case uint8:
			if float64(t) < i {
				return fmt.Errorf("the integer is less than %s", s)
			}
		case uint16:
			if float64(t) < i {
				return fmt.Errorf("the integer is less than %s", s)
			}
		case uint32:
			if float64(t) < i {
				return fmt.Errorf("the integer is less than %s", s)
			}
		case uint64:
			if float64(t) < i {
				return fmt.Errorf("the integer is less than %s", s)
			}

		case float32:
			if float64(t) < i {
				return fmt.Errorf("the float is less than %s", s)
			}
		case float64:
			if t < i {
				return fmt.Errorf("the float is less than %s", s)
			}

		case string:
			if len(t) < int(i) {
				return fmt.Errorf("the string length is less than %s", s)
			}

		default:
			vf := reflect.ValueOf(v)
			switch vf.Kind() {
			case reflect.Array, reflect.Slice, reflect.Map:
				if vf.Len() < int(i) {
					return fmt.Errorf("the length is less than %s", s)
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
	return validation.NewValidator(rule, func(v interface{}) error {
		switch t := v.(type) {
		case int:
			if float64(t) > i {
				return fmt.Errorf("the integer is greater than %s", s)
			}
		case int8:
			if float64(t) > i {
				return fmt.Errorf("the integer is greater than %s", s)
			}
		case int16:
			if float64(t) > i {
				return fmt.Errorf("the integer is greater than %s", s)
			}
		case int32:
			if float64(t) > i {
				return fmt.Errorf("the integer is greater than %s", s)
			}
		case int64:
			if float64(t) > i {
				return fmt.Errorf("the integer is greater than %s", s)
			}

		case uint:
			if float64(t) > i {
				return fmt.Errorf("the integer is greater than %s", s)
			}
		case uint8:
			if float64(t) > i {
				return fmt.Errorf("the integer is greater than %s", s)
			}
		case uint16:
			if float64(t) > i {
				return fmt.Errorf("the integer is greater than %s", s)
			}
		case uint32:
			if float64(t) > i {
				return fmt.Errorf("the integer is greater than %s", s)
			}
		case uint64:
			if float64(t) > i {
				return fmt.Errorf("the integer is greater than %s", s)
			}

		case float32:
			if float64(t) > i {
				return fmt.Errorf("the float is greater than %s", s)
			}
		case float64:
			if t > i {
				return fmt.Errorf("the float is greater than %s", s)
			}

		case string:
			if len(t) > int(i) {
				return fmt.Errorf("the string length is greater than %s", s)
			}

		default:
			vf := reflect.ValueOf(v)
			switch vf.Kind() {
			case reflect.Array, reflect.Slice, reflect.Map:
				if vf.Len() > int(i) {
					return fmt.Errorf("the length is greater than %s", s)
				}
			default:
				return fmt.Errorf("unsupported type '%T'", v)
			}
		}

		return nil
	})
}
