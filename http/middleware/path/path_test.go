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

package path

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespond204(t *testing.T) {
	h := Repsond204("/path1")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "" {
			panic("unreachable")
		}
		w.WriteHeader(201)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/path1", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != 204 {
		t.Errorf("expect status code %d, but got %d", 204, rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/path1/", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != 204 {
		t.Errorf("expect status code %d, but got %d", 204, rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/path2", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != 201 {
		t.Errorf("expect status code %d, but got %d", 201, rec.Code)
	}
}
