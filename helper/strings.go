// Copyright 2021 xgfone
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

// IsInteger reports whether s is the integer or not.
func IsInteger(s string) bool {
	if s == "" {
		return false
	}

	for i, _len := 0, len(s); i < _len; i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}

	return true
}

// InStrings reports whether the string s is in the string slice ss.
func InStrings(s string, ss []string) (yes bool) {
	for _len := len(ss) - 1; _len >= 0; _len-- {
		if ss[_len] == s {
			return true
		}
	}
	return false
}

// StringsEqual reports whether the element set of the two strings are equal.
func StringsEqual(ss1, ss2 []string) bool {
	len1 := len(ss1)
	if len1 != len(ss2) {
		return false
	}

	for i := 0; i < len1; i++ {
		if !InStrings(ss1[i], ss2) {
			return false
		}
	}

	return true
}
