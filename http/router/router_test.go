// Copyright 2022~2023 xgfone
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

package router

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/xgfone/go-apiserver/http/middleware"
	"github.com/xgfone/go-apiserver/internal/test"
)

func logMiddleware(buf *bytes.Buffer, name string) middleware.Middleware {
	return middleware.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(buf, "middleware '%s' before\n", name)
			next.ServeHTTP(rw, r)
			fmt.Fprintf(buf, "middleware '%s' after\n", name)
		})
	})
}

func TestRouter(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	router := NewRouter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/path1" {
			w.WriteHeader(201)
			fmt.Fprintln(buf, "handler")
		} else {
			w.WriteHeader(204)
			fmt.Fprintln(buf, "notfound")
		}
	}))

	router.Use(logMiddleware(buf, "log1"), logMiddleware(buf, "log2"))
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/path1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != 201 {
		t.Errorf("expect status code '%d', but got '%d'", 201, rec.Code)
	} else {
		test.CheckStrings(t, "TestRouter", strings.Split(buf.String(), "\n"), []string{
			"middleware 'log1' before",
			"middleware 'log2' before",
			"handler",
			"middleware 'log2' after",
			"middleware 'log1' after",
			"",
		})
	}

	buf.Reset()
	req = httptest.NewRequest(http.MethodGet, "http://127.0.0.1/path2", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != 204 {
		t.Errorf("expect status code '%d', but got '%d'", 201, rec.Code)
	} else {
		test.CheckStrings(t, "TestRouter", strings.Split(buf.String(), "\n"), []string{
			"middleware 'log1' before",
			"middleware 'log2' before",
			"notfound",
			"middleware 'log2' after",
			"middleware 'log1' after",
			"",
		})
	}
}
