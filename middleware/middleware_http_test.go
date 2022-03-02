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

	"github.com/xgfone/go-apiserver/internal/test"
)

var _ http.Handler = testHTTPHandler{}

type testHTTPHandler struct {
	buf     *bytes.Buffer
	name    string
	handler http.Handler
}

func (h testHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(h.buf, "middleware '%s' before\n", h.name)
	h.handler.ServeHTTP(w, r)
	fmt.Fprintf(h.buf, "middleware '%s' after\n", h.name)
}

func httpMiddleware(name string, priority int, buf *bytes.Buffer) Middleware {
	return NewMiddleware(name, priority, func(i interface{}) interface{} {
		return testHTTPHandler{buf: buf, name: name, handler: i.(http.Handler)}
	})
}

func TestMiddlewareManagerHTTP(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		buf.WriteString("handler\n")
	})

	manager := NewManager(nil)
	manager.Use(httpMiddleware("mw2", 2, buf), httpMiddleware("mw1", 1, buf))

	expects := []string{
		"middleware 'mw1' before",
		"middleware 'mw2' before",
		"handler",
		"middleware 'mw2' after",
		"middleware 'mw1' after",
		"",
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
	manager.WrapHandler(httpHandler).(http.Handler).ServeHTTP(rec, req)
	if rec.Code != 201 {
		t.Errorf("expect status code '%d', but got '%d'", 201, rec.Code)
	}
	test.CheckStrings(t, "MiddlewareManagerHTTP", strings.Split(buf.String(), "\n"), expects)

	buf.Reset()
	rec = httptest.NewRecorder()
	manager.SetHandler(httpHandler)
	manager.ServeHTTP(rec, req)
	if rec.Code != 201 {
		t.Errorf("expect status code '%d', but got '%d'", 201, rec.Code)
	}
	test.CheckStrings(t, "MiddlewareManagerHTTP", strings.Split(buf.String(), "\n"), expects)
}
