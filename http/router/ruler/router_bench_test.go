// Copyright 2022 xgfone
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

	"github.com/xgfone/go-apiserver/internal/test"
)

func BenchmarkExactRouteWithNoopMiddleware4(b *testing.B) {
	benchmarkRouteWithPath(b, "/exact/path")
}

func BenchmarkParamRouteWithNoopMiddleware4(b *testing.B) {
	benchmarkRouteWithPath(b, "/path/:param1/:param2")
}

func benchmarkRouteWithPath(b *testing.B, path string) {
	router := NewRouter()
	router.Path(path).GETFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
	})
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("notfound")
	})

	req := httptest.NewRequest(http.MethodGet, path, nil)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var rw test.ResponseWriter
		router.ServeHTTP(rw, req)
	}
}
