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

package helper

import "reflect"

// IsPointer reports whether kind, such as reflect.Value or reflect.Type,
// is a pointer.
func IsPointer(kind interface{ Kind() reflect.Kind }) bool {
	return kind.Kind() == KindPointer
}

// Implements reports whether the value has implemented the interface iface.
//
// The type of value and iface may be one of reflect.Type and reflect.Value.
// If iface is not one of the types of reflect.Type and reflect.Value,
// it should be a pointer to the interface.
func Implements(value, iface interface{}) bool {
	var v, i reflect.Type

	switch t := value.(type) {
	case reflect.Type:
		v = t
	case reflect.Value:
		v = t.Type()
	default:
		v = reflect.TypeOf(value)
	}

	switch t := iface.(type) {
	case reflect.Type:
		i = t
	case reflect.Value:
		i = t.Type()
	default:
		i = reflect.TypeOf(iface)
	}

	if IsPointer(i) {
		i = i.Elem()
	}

	return v.Implements(i)
}

// FillNilPtr fills the zero value of its base type if value is a pointer
// and equal to nil. Or, do nothing and return the original value.
func FillNilPtr(value reflect.Value) reflect.Value {
	if IsPointer(value) && value.CanSet() && value.IsNil() {
		value.Set(reflect.New(value.Type().Elem()))
	}
	return value
}

// IndirectValue returns the underlying value of the pointer or interface
// if the input value is a pointer or interface. Or, return the input.
func IndirectValue(value reflect.Value) reflect.Value {
	switch value.Kind() {
	case KindPointer, reflect.Interface:
		if !value.IsNil() {
			value = IndirectValue(value.Elem())
		}
	}
	return value
}
