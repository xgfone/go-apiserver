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

package recover

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xgfone/go-apiserver/http/reqresp"
)

func TestRecover(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	h := Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Errorf("expect status code %d, but got %d", 200, rec.Code)
	} else if p := rec.Header().Get("X-Panic"); p == "1" {
		t.Errorf("unexpect response header '%s'", "X-Panic")
	}

	rec = httptest.NewRecorder()
	h = Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { panic("test") }))
	h.ServeHTTP(rec, req)
	if rec.Code != 500 {
		t.Errorf("expect status code %d, but got %d", 500, rec.Code)
	} else if p := rec.Header().Get("X-Panic"); p != "1" {
		t.Errorf("expect response header '%s', but got not", "X-Panic")
	}

	rec = httptest.NewRecorder()
	c := reqresp.AcquireContext()
	c.Request = req.WithContext(reqresp.SetContext(req.Context(), c))
	c.ResponseWriter = reqresp.AcquireResponseWriter(rec)
	h.ServeHTTP(c.ResponseWriter, c.Request)
	if rec.Code != 500 {
		t.Errorf("expect status code %d, but got %d", 500, rec.Code)
	} else if p := rec.Header().Get("X-Panic"); p != "1" {
		t.Errorf("expect response header '%s', but got not", "X-Panic")
	} else if c.Err == nil {
		t.Errorf("expect an error, but got nil")
	} else if e, ok := c.Err.(panicerror); !ok {
		t.Errorf("expect a panic error, but got %T", c.Err)
	} else if s, _ := e.panicv.(string); s != "test" {
		t.Errorf("expect panic '%s', but got %+v", "test", e.panicv)
	}
}
