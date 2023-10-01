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

// MakeSlice returns a new slice.
//
// If both cap and defaultCap are equal to 0, it is equal to make(S, len).
// If cap is equal to 0, use defaultCap as cap instead, which is equal to
// make(S, len, defaultCap).
func MakeSlice[S ~[]E, E any, I ~int | ~int64](len, cap, defaultCap I) S {
	if cap == 0 {
		if cap = defaultCap; cap == 0 {
			return make(S, len)
		}
	}

	if cap < len {
		panic("slice.Make: the cap is less than len")
	}
	return make(S, len, cap)
}

// MergeSlice merges a set of slices in turn to one slice.
//
// If no arguments, return nil.
// If all the arguments are empty, return a empty slice with cap==0.
func MergeSlice[S ~[]E, E any](ss ...S) S {
	switch _len := len(ss); _len {
	case 0:
		return nil

	case 1:
		return ss[0]

	case 2:
		len1, len2 := len(ss[0]), len(ss[1])
		switch {
		case len1 == 0:
			return ss[1]

		case len2 == 0:
			return ss[0]

		default:
			vs := make(S, len1+len2)
			copy(vs, ss[0])
			copy(vs[len1:], ss[1])
			return vs
		}

	default:
		var tlen int
		var nonil bool
		for i := 0; i < _len; i++ {
			if ss[i] != nil {
				nonil = true
				tlen += len(ss[i])
			}
		}

		if !nonil {
			return nil
		}
		if tlen == 0 {
			return S{}
		}

		vs := make(S, 0, tlen)
		for i := 0; i < _len; i++ {
			vs = append(vs, ss[i]...)
		}
		return vs
	}
}
