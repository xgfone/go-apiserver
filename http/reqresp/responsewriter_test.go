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

package reqresp

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func wrapHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		next(NewResponseWriter(rw), r)
	}
}

func TestResponseWriter(t *testing.T) {
	handler := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		if code := rw.(ResponseWriter).StatusCode(); code != 200 {
			t.Errorf("unexpected status code %d", code)
		}

		rw.WriteHeader(201)
		fmt.Fprintf(rw, "%d", rw.(ResponseWriter).StatusCode())
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:8001", nil)
	wrapHandler(handler).ServeHTTP(rec, req)

	if rec.Code != 201 {
		t.Errorf("expect the status code '%d', but got '%d'", 201, rec.Code)
	}
	if body := rec.Body.String(); body != "201" {
		t.Errorf("expect the response body '%s', but got '%s'", "201", body)
	}

	if w, ok := NewResponseWriter(rec).(WrappedResponseWriter); !ok {
		t.Error("expect WrappedResponseWriter")
	} else if r, ok := w.Unwrap().(*httptest.ResponseRecorder); !ok {
		t.Error("expect httptest.ResponseRecorder")
	} else if r != rec {
		t.Error("unexpected httptest.ResponseRecorder")
	}
}

func BenchmarkResponseWriter(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:8001", nil)
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, ok := w.(http.Flusher); !ok {
			panic("not http.Flusher")
		}
		if _, ok := w.(http.CloseNotifier); !ok {
			panic("not http.CloseNotifier")
		}
		if _, ok := w.(http.Hijacker); ok {
			panic("not expected http.Hijacker")
		}
		if _, ok := w.(http.Pusher); ok {
			panic("not expected http.Pusher")
		}
		w.WriteHeader(200)
	})

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			var rw http.ResponseWriter = testResponseWriter{}
			rw = NewResponseWriter(rw)
			handler.ServeHTTP(rw, req)
		}
	})
}

type testResponseWriter struct{}

func (w testResponseWriter) Write([]byte) (int, error) { return 0, nil }
func (w testResponseWriter) Header() http.Header       { return nil }
func (w testResponseWriter) WriteHeader(int)           {}
func (w testResponseWriter) Flush()                    {}
func (w testResponseWriter) CloseNotify() <-chan bool  { return nil }
