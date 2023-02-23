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

import "golang.org/x/exp/constraints"

// Max returns the maximum element of the slice.
//
// Return ZERO instead if the slice is empty.
func Max[S ~[]E, E constraints.Ordered](vs S) E {
	switch _len := len(vs); _len {
	case 0:
		var v E
		return v

	case 1:
		return vs[0]

	default:
		v := vs[0]
		for i := 1; i < _len; i++ {
			if v < vs[i] {
				v = vs[i]
			}
		}
		return v
	}
}

// Min returns the minimum element of the slice.
//
// Return ZERO instead if the slice is empty.
func Min[S ~[]E, E constraints.Ordered](vs S) E {
	switch _len := len(vs); _len {
	case 0:
		var v E
		return v

	case 1:
		return vs[0]

	default:
		v := vs[0]
		for i := 1; i < _len; i++ {
			if v > vs[i] {
				v = vs[i]
			}
		}
		return v
	}
}

// Convert converts the slice from []E1 to []E2.
func Convert[S1 ~[]E1, E1, E2 any](vs S1, convert func(E1) E2) []E2 {
	newslice := make([]E2, len(vs))
	for i, e := range vs {
		newslice[i] = convert(e)
	}
	return newslice
}

// Clone clones the slice and returns the new.
//
// NOTICE: it is a shallow clone.
func Clone[S ~[]E, E any](slice S) S {
	if slice == nil {
		return nil
	}

	newslice := make(S, len(slice))
	copy(newslice, slice)
	return newslice
}

// Merge merges the slice from src into dst and returns the new dst.
func Merge[S1 ~[]E1, S2 ~[]E2, E1, E2 any](dst S1, src S2, convert func(E2) E1) S1 {
	for _, e := range src {
		dst = append(dst, convert(e))
	}
	return dst
}

// Index returns the first index where v is in vs, or -1.
func Index[S ~[]E, E comparable](vs S, v E) int {
	for i, e := range vs {
		if e == v {
			return i
		}
	}
	return -1
}

// LastIndex returns the last index where v is in vs, or -1.
func LastIndex[S ~[]E, E comparable](vs S, v E) int {
	for _len := len(vs) - 1; _len >= 0; _len-- {
		if vs[_len] == v {
			return _len
		}
	}
	return -1
}

// LastIndexFunc returns the last index i satisfying equal(vs[i]), or -1.
func LastIndexFunc[S ~[]E, E any](vs S, equal func(E) bool) int {
	for _len := len(vs) - 1; _len >= 0; _len-- {
		if equal(vs[_len]) {
			return _len
		}
	}
	return -1
}

// Equal reports whether the element and order of the two slices are equal.
func Equal[S ~[]E, E comparable](vs1, vs2 S) bool {
	len1 := len(vs1)
	if len1 != len(vs2) {
		return false
	}

	for i := 0; i < len1; i++ {
		if vs1[i] != vs2[i] {
			return false
		}
	}

	return true
}

// SetEqual reports whether the element set of the two slices are equal.
func SetEqual[S ~[]E, E comparable](vs1, vs2 S) bool {
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
func Contains[S ~[]E, E comparable](vs S, v E) bool {
	for i, _len := 0, len(vs); i < _len; i++ {
		if vs[i] == v {
			return true
		}
	}
	return false
}

// Reverse reverses the elements in the slice.
func Reverse[S ~[]E, E any](vs S) {
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
func ToInterfaces[S ~[]E, E any](vs S) []interface{} {
	is := make([]interface{}, len(vs))
	for i, _len := 0, len(vs); i < _len; i++ {
		is[i] = vs[i]
	}
	return is
}
