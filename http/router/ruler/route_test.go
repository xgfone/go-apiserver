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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xgfone/go-apiserver/http/middleware"
)

// MatchFunc is a matching function.
type matchFunc func(*http.Request) bool

// Match implements the interface Matcher.
func (f matchFunc) Match(r *http.Request) bool { return f(r) }

type testroute struct {
	path string
	code int
}

func newTestRoute(p string, c int) testroute                           { return testroute{path: p, code: c} }
func (r testroute) Match(req *http.Request) bool                       { return req.URL.Path == r.path }
func (r testroute) ServeHTTP(w http.ResponseWriter, req *http.Request) { w.WriteHeader(r.code) }
func (r testroute) Route(p int) Route {
	return NewRoute(p, matchFunc(r.Match), http.HandlerFunc(r.ServeHTTP))
}

func TestRoute(t *testing.T) {
	route := newTestRoute("/path", 204).Route(1)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://localhost/path", nil)
	route.ServeHTTP(rec, req)
	if rec.Code != 204 {
		t.Errorf("expect status code %d, but got %d", 204, rec.Code)
	}

	route.Use(middleware.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware", "1")
			next.ServeHTTP(w, r)
		})
	}))

	rec = httptest.NewRecorder()
	route.ServeHTTP(rec, req)
	if m := rec.Header().Get("X-Middleware"); m != "1" {
		t.Errorf("expect got header value '%s', but got '%s'", "1", m)
	}
}
