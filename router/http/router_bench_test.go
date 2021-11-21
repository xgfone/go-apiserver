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

package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xgfone/go-apiserver/router/http/matcher"
)

func BenchmarkRoute(b *testing.B) {
	router := NewRouter()
	router.NotFound = http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("notfound")
	})
	handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(200)
		io.WriteString(rw, "OK")
	})

	for _, r := range gplusAPI {
		m := matcher.And(
			matcher.Must(matcher.Path(r.Path)),
			matcher.Must(matcher.Method(r.Method)),
		)

		router.AddRoute(NewRoute(r.Method+r.Path, 0, m, handler))
	}

	benchmarkRoutes(b, router, gplusAPI)
}

func benchmarkRoutes(b *testing.B, router http.Handler, routes []*testRoute) {
	r := httptest.NewRequest("GET", "/", nil)
	u := r.URL
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, route := range routes {
			r.Method = route.Method
			u.Path = route.Path
			router.ServeHTTP(w, r)
		}
	}
}

type testRoute struct {
	Method string
	Path   string
}

var gplusAPI = []*testRoute{
	// People
	{"GET", "/people/:userId"},
	{"GET", "/people"},
	{"GET", "/activities/:activityId/people/:collection"},
	{"GET", "/people/:userId/people/:collection"},
	{"GET", "/people/:userId/openIdConnect"},

	// Activities
	{"GET", "/people/:userId/activities/:collection"},
	{"GET", "/activities/:activityId"},
	{"GET", "/activities"},

	// Comments
	{"GET", "/activities/:activityId/comments"},
	{"GET", "/comments/:commentId"},

	// Moments
	{"POST", "/people/:userId/moments/:collection"},
	{"GET", "/people/:userId/moments/:collection"},
	{"DELETE", "/moments/:id"},
}
