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

package action

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xgfone/go-apiserver/result"
)

func BenchmarkRouter(b *testing.B) {
	router := NewRouter()
	router.HandleResponse = func(*Context, result.Response) error { return nil }
	router.NotFound = http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("notfound")
	})

	router.RegisterContextFunc("Test1", func(c *Context) {})
	router.RegisterContextFunc("Test2", func(c *Context) {})

	req := httptest.NewRequest("GET", "http://127.0.0.1", nil)
	req.Header.Set(HeaderAction, "Test1")
	rec := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(rec, req)
	}
}
