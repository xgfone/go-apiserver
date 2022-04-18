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

import "reflect"

// Indirect returns the underlying value of the pointer or interface
// if the input value is a pointer or interface. Or, return the input.
//
// Return nil if the input value is a pointer(nil), or interface(nil).
func Indirect(value interface{}) interface{} {
	if value == nil {
		return nil
	}

	switch vf := reflect.ValueOf(value); vf.Kind() {
	case reflect.Ptr, reflect.Interface:
		if vf.IsNil() {
			return nil
		}
		return Indirect(vf.Elem().Interface())

	default:
		return value
	}
}
