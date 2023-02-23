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

import (
	"reflect"
	"time"

	"golang.org/x/exp/constraints"
)

// Now is used to customize the time Now.
var Now = time.Now

// ScannerFunc is a scanner function.
type ScannerFunc func(src interface{}) error

// Scan implements the interface sql.Scanner.
func (f ScannerFunc) Scan(src interface{}) error { return f(src) }

// Indirect returns the underlying value of the pointer or interface
// if the input value is a pointer or interface. Or, return the input.
//
// Return nil if the input value is a pointer(nil), or interface(nil).
func Indirect(value interface{}) interface{} {
	if value == nil {
		return nil
	}

	switch vf := reflect.ValueOf(value); vf.Kind() {
	case reflect.Pointer, reflect.Interface:
		if vf.IsNil() {
			return nil
		}
		return Indirect(vf.Elem().Interface())

	default:
		return value
	}
}

// Unwrap unwraps the inner value of v with ok==true if v has implemented
// the interface { Unwrap() T } or { Get() T }.
// Or, assert v to T and return it with ok==false instead.
func Unwrap[T any](v interface{}) (inner T, ok bool) {
	switch _v := v.(type) {
	case interface{ Get() T }:
		return _v.Get(), true

	case interface{ Unwrap() T }:
		return _v.Unwrap(), true

	default:
		return v.(T), false
	}
}

// UnwrapAll is the same as Unwrap, but unwraps the innest value of v.
func UnwrapAll[T any](v interface{}) T {
	for {
		if t, ok := Unwrap[T](v); ok {
			v = t
		} else {
			return t
		}
	}
}

// Compare compares left and right and returns
//
//	 0 if left == right
//	-1 if left <  right
//	+1 if left >  right
func Compare[T constraints.Ordered](left, right T) int {
	switch {
	case left < right:
		return -1
	case left == right:
		return 0
	default:
		return 1
	}
}

// Max returns the maximum of left and right.
func Max[T constraints.Ordered](left, right T) T {
	if left < right {
		return right
	}
	return left
}

// Min returns the minimum of left and right.
func Min[T constraints.Ordered](left, right T) T {
	if right < left {
		return right
	}
	return left
}
