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

package router

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xgfone/go-apiserver/http/handler"
	"github.com/xgfone/go-apiserver/http/matcher"
)

func BenchmarkExactRouteWithNoopMiddleware4(b *testing.B) {
	benchmarkRouteWithPath(b, "/exact/path")
}

func BenchmarkParamRouteWithNoopMiddleware4(b *testing.B) {
	benchmarkRouteWithPath(b, "/path/:param1/:param2")
}

func benchmarkRouteWithPath(b *testing.B, path string) {
	newMiddleware := handler.NewMiddleware

	router := NewRouter()
	router.Use(newMiddleware("md1", noopMd), newMiddleware("md2", noopMd))
	router.Global(newMiddleware("md1", noopMd), newMiddleware("md2", noopMd))
	router.SetNotFoundFunc(func(http.ResponseWriter, *http.Request) { panic("notfound") })
	handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(200)
		io.WriteString(rw, "OK")
	})

	const method = http.MethodGet
	m := matcher.And(matcher.Must(matcher.Path(path)), matcher.Must(matcher.Method(method)))
	router.AddRoute(NewRoute(method+path, 0, m, handler))

	rec := httptest.NewRecorder()
	rec.Body.WriteString("OK")
	req := httptest.NewRequest(method, path, nil)
	if req.Method != method || req.URL.Path != path {
		panic(fmt.Sprintf("method=%s, path=%s", req.Method, req.URL.Path))
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rec.Body.Reset()
		router.ServeHTTP(rec, req)
	}
}

func noopMd(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(rw, r)
	})
}
