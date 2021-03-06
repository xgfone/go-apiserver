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
	"strings"
	"testing"
)

func BenchmarkRandomString(b *testing.B) {
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			RandomString(8, DefaultCharset)
		}
	})
}

func TestRandomString(t *testing.T) {
	s := RandomString(8, DefaultCharset)
	if len(s) != 8 {
		t.Errorf("expect the string length %d, but got %d", 8, len(s))
	}

	for _, b := range s {
		if !strings.ContainsRune(DefaultCharset, b) {
			t.Errorf("unknown the charset '%s'", string(b))
		}
	}
}

func TestInStrings(t *testing.T) {
	ss := []string{"a", "b", "c"}

	if !InStrings("b", ss) {
		t.Errorf("donot find the string '%s'", "b")
	}

	if InStrings("d", ss) {
		t.Errorf("unexpect to find the string '%s'", "d")
	}
}

func TestStringsEqual(t *testing.T) {
	s1 := []string{"a", "b", "b"}
	s2 := []string{"a", "b", "c"}
	if StringsEqual(s1, s2) {
		t.Error("unexpected strings equal")
	}
	if StringsEqual(s2, s1) {
		t.Error("unexpected strings equal")
	}

	if !StringsEqual(s1, s1) {
		t.Error("expect strings equal, but got not equal")
	}
}

func TestTruncateStringByLen(t *testing.T) {
	if s := TruncateStringByLen("abc", 4); s != "abc" {
		t.Errorf("expect '%s', but got '%s'", "abc", s)
	}

	if s := TruncateStringByLen("abc", 3); s != "abc" {
		t.Errorf("expect '%s', but got '%s'", "abc", s)
	}

	if s := TruncateStringByLen("abc", 2); s != "ab" {
		t.Errorf("expect '%s', but got '%s'", "ab", s)
	}
}

func TestTruncateStringByRuneNum(t *testing.T) {
	if s := TruncateStringByRuneNum("abc", 4); s != "abc" {
		t.Errorf("expect '%s', but got '%s'", "abc", s)
	}

	if s := TruncateStringByRuneNum("abc", 3); s != "abc" {
		t.Errorf("expect '%s', but got '%s'", "abc", s)
	}

	if s := TruncateStringByRuneNum("abc", 2); s != "ab" {
		t.Errorf("expect '%s', but got '%s'", "ab", s)
	}

	if s := TruncateStringByRuneNum("a???c", 4); s != "a???c" {
		t.Errorf("expect '%s', but got '%s'", "a???c", s)
	}

	if s := TruncateStringByRuneNum("a???c", 3); s != "a???c" {
		t.Errorf("expect '%s', but got '%s'", "a???c", s)
	}

	if s := TruncateStringByRuneNum("a???c", 2); s != "a???" {
		t.Errorf("expect '%s', but got '%s'", "a???", s)
	}

	if s := TruncateStringByRuneNum("a\xffc", 4); s != "a\xffc" {
		t.Errorf("expect '%s', but got '%s'", "a\xffc", s)
	}

	if s := TruncateStringByRuneNum("a\xffc", 3); s != "a\xffc" {
		t.Errorf("expect '%s', but got '%s'", "a\xffc", s)
	}

	if s := TruncateStringByRuneNum("a\xffc", 2); s != "a\xff" {
		t.Errorf("expect '%s', but got '%s'", "a\xff", s)
	}
}
