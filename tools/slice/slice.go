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

// Package slice provides some convenient slice functions.
package slice

// Clone clones the slice and returns the new.
func Clone[T ~[]E, E any](slice T) T {
	newslice := make(T, len(slice))
	copy(newslice, slice)
	return newslice
}

// Index returns the index where v is in vs, else -1.
func Index[T ~[]E, E comparable](vs T, v E) int {
	for i, e := range vs {
		if e == v {
			return i
		}
	}
	return -1
}

// Equal reports whether the element set of the two slices are equal.
func Equal[T ~[]E, E comparable](vs1, vs2 T) bool {
	len1 := len(vs1)
	if len1 != len(vs2) {
		return false
	}

	for i := 0; i < len1; i++ {
		if !Contains(vs1, vs2[i]) || !Contains(vs2, vs1[i]) {
			return false
		}
	}

	return true
}

// Contains reports whether the slice vs contains the value v.
func Contains[T ~[]E, E comparable](vs T, v E) bool {
	for i, _len := 0, len(vs); i < _len; i++ {
		if vs[i] == v {
			return true
		}
	}
	return false
}

// Reverse reverses the elements in the slice.
func Reverse[T ~[]E, E any](vs T) {
	_len := len(vs) - 1
	if _len <= 0 {
		return
	}

	for i, j := 0, _len/2; i <= j; i++ {
		k := _len - i
		vs[i], vs[k] = vs[k], vs[i]
	}
}

// ToInterfaces converts []any to []interface{}.
func ToInterfaces[T ~[]E, E any](vs T) []interface{} {
	is := make([]interface{}, len(vs))
	for i, _len := 0, len(vs); i < _len; i++ {
		is[i] = vs[i]
	}
	return is
}
