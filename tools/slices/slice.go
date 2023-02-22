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

// Package slices provides some convenient slice functions.
package slices

// Convert s
func Convert[T1 ~[]E1, E1, E2 any](vs T1, convert func(E1) E2) []E2 {
	newslice := make([]E2, len(vs))
	for i, e := range vs {
		newslice[i] = convert(e)
	}
	return newslice
}

// Clone clones the slice and returns the new.
//
// NOTICE: it is a shallow clone.
func Clone[T ~[]E, E any](slice T) T {
	if slice == nil {
		return nil
	}

	newslice := make(T, len(slice))
	copy(newslice, slice)
	return newslice
}

// Index returns the first index where v is in vs, or -1.
func Index[T ~[]E, E comparable](vs T, v E) int {
	for i, e := range vs {
		if e == v {
			return i
		}
	}
	return -1
}

// LastIndex returns the last index where v is in vs, or -1.
func LastIndex[T ~[]E, E comparable](vs T, v E) int {
	for _len := len(vs) - 1; _len >= 0; _len-- {
		if vs[_len] == v {
			return _len
		}
	}
	return -1
}

// LastIndexFunc returns the last index i satisfying equal(vs[i]), or -1.
func LastIndexFunc[T ~[]E, E any](vs T, equal func(E) bool) int {
	for _len := len(vs) - 1; _len >= 0; _len-- {
		if equal(vs[_len]) {
			return _len
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
