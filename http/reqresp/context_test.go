// Copyright 2023~2025 xgfone
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

package reqresp

import (
	"bytes"
	"net/http"
	"net/url"
	"testing"

	"github.com/xgfone/go-apiserver/http/header"
)

func TestContextBinder(t *testing.T) {
	var req struct {
		Int    int    `default:"222"`
		Uint   int    `default:"333"`
		String string `default:"abc"`
	}
	req.Int = 111

	c := AcquireContext()
	body := bytes.NewBufferString(`{"Uint":444}`)
	c.Request, _ = http.NewRequest("GET", "http://localhost", body)
	c.Request.Header.Set(header.HeaderContentType, header.MIMEApplicationJSON)

	err := c.BindBody(&req)
	if err != nil {
		t.Error(err)
	}

	if req.Int != 111 {
		t.Errorf("expect Int is equal to %d, but got %d", 111, req.Int)
	}
	if req.Uint != 444 {
		t.Errorf("expect Uint is equal to %d, but got %d", 444, req.Uint)
	}
	if req.String != "abc" {
		t.Errorf("expect String is equal to '%s', but '%s'", "abc", req.String)
	}
}

func TestContextGetQueryInt64(t *testing.T) {
	c := Context{Query: make(url.Values, 2)}
	c.Query.Set("key", "123")

	if v, err := c.GetQueryInt64("key", true); err != nil {
		t.Error(err)
	} else if v != 123 {
		t.Errorf("expect %d, but got %d", 123, v)
	}

	if v, err := c.GetQueryInt64("abc", false); err != nil {
		t.Error(err)
	} else if v != 0 {
		t.Errorf("expect %d, but got %d", 0, v)
	}

	if _, err := c.GetQueryInt64("abc", true); err == nil {
		t.Errorf("expect an error, but got nil")
	} else if s := err.Error(); s != "missing abc" {
		t.Errorf("expect error '%s', but got '%s'", "missing abc", s)
	}
}

func TestContextTranslate(t *testing.T) {
	c := Context{Request: &http.Request{}}
	if s := c.Translate("hello %s", "xgfone"); s != "hello xgfone" {
		t.Errorf("expect '%s', but got '%s'", "hello xgfone", s)
	}

	c.Translator = DefaultTranslate
	if s := c.Translate("hello %s", "xgfone"); s != "hello xgfone" {
		t.Errorf("expect '%s', but got '%s'", "hello xgfone", s)
	}
}
