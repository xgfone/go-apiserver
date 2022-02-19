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

package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
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
	return NewMiddleware(name, func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(buf, "middleware '%s' before\n", name)
			h.ServeHTTP(rw, r)
			fmt.Fprintf(buf, "middleware '%s' after\n", name)
		})
	})
}
