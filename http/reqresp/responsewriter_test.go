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

	"github.com/xgfone/go-apiserver/internal/test"
)

func wrapHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw := AcquireResponseWriter(w)
		defer ReleaseResponseWriter(rw)
		next(rw, r)
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

	rw := AcquireResponseWriter(rec)
	if w, ok := rw.(interface{ Unwrap() http.ResponseWriter }); !ok {
		t.Error("expect wrapped ResponseWriter")
	} else if r, ok := w.Unwrap().(*httptest.ResponseRecorder); !ok {
		t.Error("expect httptest.ResponseRecorder")
	} else if r != rec {
		t.Error("unexpected httptest.ResponseRecorder")
	}
}

func BenchmarkResponseWriter(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:8001", nil)
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
	})

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			var rw test.ResponseWriter
			w := AcquireResponseWriter(rw)
			handler.ServeHTTP(w, req)
			ReleaseResponseWriter(w)
		}
	})
}
