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

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestManager(t *testing.T) {
	m := NewManager(nil)
	m.Append()
	m.Append(MiddlewareFunc(func(next http.Handler) http.Handler { return next }))

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expect a panic, but got not")
			}
		}()

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
		m.ServeHTTP(rec, req)
	}()

	m.Reset()
	m.SetHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	m.ServeHTTP(rec, req)
	if rec.Code != 204 {
		t.Errorf("expect status code %d, but got %d", 204, rec.Code)
	}
}
