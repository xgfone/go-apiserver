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

package requestid

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestId(t *testing.T) {
	handler := RequestId(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if rid := req.Header.Get("X-Request-Id"); len(rid) != 24 {
		t.Errorf("expect requests id length %d, but got %d", 24, len(rid))
	}

	req.Header.Set("X-Request-Id", "abc")
	handler.ServeHTTP(rec, req)

	if rid := req.Header.Get("X-Request-Id"); rid != "abc" {
		t.Errorf("expect requests id '%s', but got '%s'", "abc", rid)
	}
}
