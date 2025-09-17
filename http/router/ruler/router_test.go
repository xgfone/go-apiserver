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

package ruler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xgfone/go-toolkit/httpx"
)

func TestRouter(t *testing.T) {
	r := NewRouter()
	group := r.Group("/prefix")
	group.Path("/path").Host("localhost").GET(httpx.Handler204)
	group.PathPrefix("/path").Host("localhost").GET(httpx.Handler400)
	group.Host("localhost").Path("/path").GET(httpx.Handler500)

	expect1 := "(Host(`localhost`) && Path(`/prefix/path`) && Method(`GET`))"
	expect2 := "(Host(`localhost`) && PathPrefix(`/prefix/path`) && Method(`GET`))"
	for i, r := range r.Routes() {
		desc := r.Matcher.(fmt.Stringer).String()
		switch desc {
		case expect1, expect2:
		default:
			t.Errorf("%d: unexpected matcher: %s", i, desc)
		}
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://localhost/prefix/path", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 204 {
		t.Errorf("expect status code %d, but got %d", 204, rec.Code)
	}
}
