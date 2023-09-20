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
