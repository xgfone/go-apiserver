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

package cors

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/xgfone/go-toolkit/httpx"
)

func TestCORS(t *testing.T) {
	handler := CORS(Config{}).Handler(httpx.Handler204)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	handler.ServeHTTP(rec, req)
	header := rec.Header()
	if vary := header.Get("Vary"); vary != "Origin" {
		t.Errorf("expect vary '%s', but got '%s'", "Origin", vary)
	}
	if allowOrigin := header.Get("Access-Control-Allow-Origin"); allowOrigin != "*" {
		t.Errorf("expect allow origin '%s', but got '%s'", "*", allowOrigin)
	}

	rec = httptest.NewRecorder()
	req.Method = http.MethodOptions
	handler.ServeHTTP(rec, req)
	header = rec.Header()
	if allowOrigin := header.Get("Access-Control-Allow-Origin"); allowOrigin != "*" {
		t.Errorf("expect allow origin '%s', but got '%s'", "*", allowOrigin)
	}

	allowMethods := header.Get("Access-Control-Allow-Methods")
	if reflect.DeepEqual(allowMethods, DefaultAllowMethods) {
		t.Errorf("expect allow methods %v, but got %v", DefaultAllowMethods, allowMethods)
	}
}
