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

import (
	"unsafe"

	"math/rand"
)

var intn = rand.Intn

// Pre-define some charsets to generate the random string.
var (
	NumCharset      = "0123456789"
	HexCharset      = NumCharset + "abcdefABCDEF"
	HexLowerCharset = NumCharset + "abcdef"
	HexUpperCharset = NumCharset + "ABCDEF"

	AlphaLowerCharset = "abcdefghijklmnopqrstuvwxyz"
	AlphaUpperCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	AlphaCharset      = AlphaLowerCharset + AlphaUpperCharset

	AlphaNumLowerCharset = NumCharset + AlphaLowerCharset
	AlphaNumUpperCharset = NumCharset + AlphaUpperCharset
	AlphaNumCharset      = NumCharset + AlphaCharset

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
		buf[i] = charset[intn(nlen)]
	}
}

// String converts the value b from []byte to string.
//
// NOTICE: b must not be modified.
func String(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// Bytes converts the value s from string to []byte.
//
// NOTICE: s must not be modified.
func Bytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
