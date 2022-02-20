// Copyright 2022 xgfone
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

package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func logMiddleware(buf *bytes.Buffer, name string, prio int) Middleware {
	return NewMiddleware(name, prio, func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(buf, "middleware '%s' before\n", name)
			h.ServeHTTP(rw, r)
			fmt.Fprintf(buf, "middleware '%s' after\n", name)
		})
	})
}

func TestRouteMiddleware(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(201)
		buf.WriteString("handler\n")
	})

	mws := NewManager()
	mws.SetHandler(handler)
	mws.Use(logMiddleware(buf, "log1", 1), logMiddleware(buf, "log2", 2))

	middlewares := mws.GetMiddlewares()
	if _len := len(middlewares); _len != 2 {
		t.Errorf("expect %d middlewres, but got %d", 2, _len)
	} else {
		for i, mw := range middlewares {
			if i == 0 && mw.Name() != "log2" {
				t.Errorf("0: expect middleware 'log2', but got '%s'", mw.Name())
			}
			if i == 1 && mw.Name() != "log1" {
				t.Errorf("1: expect middleware 'log1', but got '%s'", mw.Name())
			}
		}
	}

	req, _ := http.NewRequest("GET", "http://127.0.0.1", nil)
	rec := httptest.NewRecorder()
	mws.ServeHTTP(rec, req)

	if rec.Code != 201 {
		t.Fatalf("expect the status code '%d', but got '%d'", 201, rec.Code)
	}

	expects := []string{
		"middleware 'log2' before",
		"middleware 'log1' before",
		"handler",
		"middleware 'log1' after",
		"middleware 'log2' after",
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
	mws.Cancel("log2")

	req, _ = http.NewRequest("GET", "http://127.0.0.1", nil)
	rec = httptest.NewRecorder()
	mws.ServeHTTP(rec, req)

	expects = []string{
		"middleware 'log1' before",
		"handler",
		"middleware 'log1' after",
		"",
	}
	results = strings.Split(buf.String(), "\n")
	if len(results) != len(expects) {
		t.Errorf("expect %d lines, but got %d: %v", len(expects), len(results), results)
	} else {
		for i := 0; i < len(results); i++ {
			if results[i] != expects[i] {
				t.Errorf("expect '%s', but got '%s'", expects[i], results[i])
			}
		}
	}
}
