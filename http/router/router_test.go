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

package router

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/xgfone/go-apiserver/internal/test"
	"github.com/xgfone/go-apiserver/middleware"
)

type rmFunc func(http.ResponseWriter, *http.Request, http.Handler)

func (f rmFunc) ServeHTTP(http.ResponseWriter, *http.Request) {}
func (f rmFunc) Route(w http.ResponseWriter, r *http.Request, no http.Handler) {
	f(w, r, no)
}

func logMiddleware(buf *bytes.Buffer, name string, prio int) middleware.Middleware {
	return middleware.NewMiddleware(name, prio, func(h interface{}) interface{} {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(buf, "middleware '%s' before\n", name)
			h.(http.Handler).ServeHTTP(rw, r)
			fmt.Fprintf(buf, "middleware '%s' after\n", name)
		})
	})
}

func TestRouter(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	router := NewRouter(rmFunc(func(w http.ResponseWriter, r *http.Request, no http.Handler) {
		if r.Method == http.MethodGet && r.URL.Path == "/path1" {
			w.WriteHeader(201)
			fmt.Fprintln(buf, "handler")
		} else {
			no.ServeHTTP(w, r)
		}
	}))

	router.Middlewares.Use(logMiddleware(buf, "log1", 1), logMiddleware(buf, "log2", 2))
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
		fmt.Fprintln(buf, "notfound")
	})

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
