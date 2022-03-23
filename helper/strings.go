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

import (
	"math/rand"
	"unicode/utf8"
)

// DefaultCharset is the default charset to generate the random string.
var DefaultCharset = "0123456789abcdefghijklmnopqrstuvwxyz"

// RandomString generates a random string with the length from the given charsets.
func RandomString(n int, charset string) string {
	nlen := len(charset)
	buf := make([]byte, n)
	for i := 0; i < n; i++ {
		buf[i] = charset[rand.Intn(nlen)]
	}
	return string(buf)
}

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
		if !InStrings(ss1[i], ss2) || !InStrings(ss2[i], ss1) {
			return false
		}
	}

	return true
}

// TruncateStringByLen truncates the length of the string s to maxLen
// if exceeding maxLen. Or, returns the original s.
func TruncateStringByLen(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) > maxLen {
		return s[:maxLen]
	}
	return s
}

// TruncateStringByRuneNum truncates the number of runes of the string s
// to maxNum if exceeding maxNum. Or, returns the original s.
func TruncateStringByRuneNum(s string, maxNum int) string {
	if maxNum <= 0 {
		return ""
	}

	var count, index int
	for _s := s; len(_s) > 0; {
		_, n := utf8.DecodeRuneInString(_s)
		index += n
		if count++; count == maxNum {
			return s[:index]
		}
		_s = _s[n:]
	}

	return s
}
