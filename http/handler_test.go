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

package http

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testNameHandler string

func (h testNameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(201)
	w.Write([]byte(h))
}

func wrapHandlerFunc(h http.Handler, w http.ResponseWriter, r *http.Request) {
	h.ServeHTTP(w, r)
}

func TestWrappedHandler(t *testing.T) {
	handler1 := WrapHandler(testNameHandler("handler"), wrapHandlerFunc)
	handler2 := WrapHandler(handler1, wrapHandlerFunc)
	handler3 := WrapHandler(handler2, wrapHandlerFunc)
	handler4 := NewSwitchHandler(handler3)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
	handler4.ServeHTTP(rec, req)
	if rec.Code != 201 {
		t.Errorf("expect status code '%d', but got '%d'", 201, rec.Code)
	} else if body := rec.Body.String(); body != "handler" {
		t.Errorf("expect response body '%s', but got '%s'", "handler", body)
	}

	handler := UnwrapHandler(handler4)
	if name, ok := handler.(testNameHandler); !ok {
		t.Errorf("expect handler type 'testNameHandler', but got '%T'", handler)
	} else if name != "handler" {
		t.Errorf("expect handler name '%s', but got '%s'", "handler", name)
	}
}

func logMiddleware(buf *bytes.Buffer, name string) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(buf, "middleware '%s' before\n", name)
			h.ServeHTTP(rw, r)
			fmt.Fprintf(buf, "middleware '%s' after\n", name)
		})
	}
}

func TestMiddlewareHandler(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	h := func(name string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(buf, name)
		})
	}

	mh := NewMiddlewareHandler(h("h1"), logMiddleware(buf, "mw1"), logMiddleware(buf, "mw2"))
	mh.Use(logMiddleware(buf, "mw3"), logMiddleware(buf, "mw4"))

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
	mh.ServeHTTP(rec, req)
	expects := []string{
		"middleware 'mw1' before",
		"middleware 'mw2' before",
		"middleware 'mw3' before",
		"middleware 'mw4' before",
		"h1",
		"middleware 'mw4' after",
		"middleware 'mw3' after",
		"middleware 'mw2' after",
		"middleware 'mw1' after",
		"",
	}
	results := strings.Split(buf.String(), "\n")
	if len(results) != len(expects) {
		t.Errorf("expect %d lines, but got %d", len(expects), len(results))
	} else {
		for i := 0; i < len(results); i++ {
			if results[i] != expects[i] {
				t.Errorf("expect '%s', but got '%s'", expects[i], results[i])
			}
		}
	}

	buf.Reset()
	mh.Swap(h("h2"))
	mh.ServeHTTP(rec, req)
	expects = []string{
		"middleware 'mw1' before",
		"middleware 'mw2' before",
		"middleware 'mw3' before",
		"middleware 'mw4' before",
		"h2",
		"middleware 'mw4' after",
		"middleware 'mw3' after",
		"middleware 'mw2' after",
		"middleware 'mw1' after",
		"",
	}
	results = strings.Split(buf.String(), "\n")
	if len(results) != len(expects) {
		t.Errorf("expect %d lines, but got %d", len(expects), len(results))
	} else {
		for i := 0; i < len(results); i++ {
			if results[i] != expects[i] {
				t.Errorf("expect '%s', but got '%s'", expects[i], results[i])
			}
		}
	}
}
