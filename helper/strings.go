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

// Pre-define some charsets to generate the random string.
var (
	NumCharset      = "0123456789"
	HexCharset      = NumCharset + "abcdefABCDEF"
	HexLowerCharset = NumCharset + "abcdef"
	HexUpperCharset = NumCharset + "ABCDEF"

	AlphaLowerCharset = "abcdefghijklmnopqrstuvwxyz"
	AlphaUpperCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	AlphaNumLowerCharset = NumCharset + AlphaLowerCharset
	AlphaNumUpperCharset = NumCharset + AlphaUpperCharset
	AlphaNumCharset      = NumCharset + AlphaLowerCharset + AlphaNumUpperCharset

	DefaultCharset = AlphaNumLowerCharset
)

// RandomString generates a random string with the length from the given charsets.
func RandomString(n int, charset string) string {
	buf := make([]byte, n)
	Random(buf, charset)
	return string(buf) // TODO: use unsafe.String??
}

// Random generates a random string with the length from the given charsets into buf.
func Random(buf []byte, charset string) {
	nlen := len(charset)
	for i, _len := 0, len(buf); i < _len; i++ {
		buf[i] = charset[rand.Intn(nlen)]
	}
}

// IsDigitString reports whether the string s only contains the characters [0-9].
func IsDigitString(s string) bool {
	if s == "" {
		return false
	}

	for i, _len := 0, len(s); i < _len; i++ {
		switch s[i] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		default:
			return false
		}
	}

	return true
}

// IsIntegerString reports whether the string s is the integer or not.
func IsIntegerString(s string) bool {
	if s == "" {
		return false
	}

	var i int
	switch s[0] {
	case '+', '-':
		i = 1
	}

	for _len := len(s); i < _len; i++ {
		switch s[i] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		default:
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
