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

package header

import (
	"net/http"
	"testing"
)

func TestContentType(t *testing.T) {
	header := make(http.Header)
	if ct := ContentType(header); ct != "" {
		t.Errorf("unexpect Content-Type '%s'", ct)
	}

	header.Set("Content-Type", "application/json")
	if ct := ContentType(header); ct != "application/json" {
		t.Errorf("unexpect Content-Type '%s'", ct)
	}

	header.Set("Content-Type", "; charset=UTF-8")
	if ct := ContentType(header); ct != "" {
		t.Errorf("unexpect Content-Type '%s'", ct)
	}

	header.Set("Content-Type", "application/json; charset=UTF-8")
	if ct := ContentType(header); ct != "application/json" {
		t.Errorf("expect Content-Type '%s', but got '%s'", "application/json", ct)
	}
}

func TestCharset(t *testing.T) {
	header := make(http.Header)
	if charset := Charset(header); charset != "" {
		t.Errorf("unexpect charset '%s'", charset)
	}

	header.Set("Content-Type", "application/json")
	if charset := Charset(header); charset != "" {
		t.Errorf("unexpect charset '%s'", charset)
	}

	header.Set("Content-Type", "charset=UTF-8")
	if charset := Charset(header); charset != "UTF-8" {
		t.Errorf("expect charset '%s', but got '%s'", "UTF-8", charset)
	}

	header.Set("Content-Type", "; charset=UTF-8")
	if charset := Charset(header); charset != "UTF-8" {
		t.Errorf("expect charset '%s', but got '%s'", "UTF-8", charset)
	}

	header.Set("Content-Type", "application/json; charset=UTF-8")
	if charset := Charset(header); charset != "UTF-8" {
		t.Errorf("expect charset '%s', but got '%s'", "UTF-8", charset)
	}

	header.Set("Content-Type", "application/json; version=1; charset=UTF-8")
	if charset := Charset(header); charset != "UTF-8" {
		t.Errorf("expect charset '%s', but got '%s'", "UTF-8", charset)
	}
}

func TestAccept(t *testing.T) {
	expects := []string{
		"text/html",
		"image/webp",
		"application/",
		"",
	}

	header := make(http.Header)
	header.Set(HeaderAccept, "text/html, application/*;q=0.9, image/webp, */*;q=0.8")
	accepts := Accept(header)

	if len(expects) != len(accepts) {
		t.Errorf("expect %d accepts, but got %d", len(expects), len(accepts))
	} else {
		for i := range expects {
			if expects[i] != accepts[i] {
				t.Errorf("%d: expect '%s', got '%s'", i, expects[i], accepts[i])
			}
		}
	}
}
