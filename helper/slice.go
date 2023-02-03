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

// SliceEqual reports whether the element set of the two slices are equal.
func SliceEqual[E comparable](vs1, vs2 []E) bool {
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
func Contains[E comparable](vs []E, v E) bool {
	for i, _len := 0, len(vs); i < _len; i++ {
		if vs[i] == v {
			return true
		}
	}
	return false
}

// Reverse reverses the elements in the slice.
func Reverse[E any](vs []E) {
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
func ToInterfaces[T any](vs []T) []interface{} {
	is := make([]interface{}, len(vs))
	for i, _len := 0, len(vs); i < _len; i++ {
		is[i] = vs[i]
	}
	return is
}
