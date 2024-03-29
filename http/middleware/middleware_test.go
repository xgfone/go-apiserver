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
	"reflect"
	"testing"
)

func TestMiddlewares(t *testing.T) {
	ms := Middlewares(nil).Clone().
		Append(appendPathSuffix("/m1")).
		AppendFunc(appendPathSuffix("/m2")).
		InsertFunc(appendPathSuffix("/m3")).
		Insert(appendPathSuffix("/m4"))

	nothing := func(next http.Handler) http.Handler { return next }
	ms = ms.InsertFunc(nothing)
	ms = ms.InsertFunc(nothing, nothing)
	ms = ms.InsertFunc(nothing, nothing, nothing)
	ms = ms.AppendFunc(nothing)
	ms = ms.AppendFunc(nothing, nothing)
	ms = ms.AppendFunc(nothing, nothing, nothing)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/path", nil)
	h := ms.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	h.ServeHTTP(rec, req)

	expect := "/path/m4/m3/m1/m2"
	if req.URL.Path != expect {
		t.Errorf("expect path '%s', but got '%s'", expect, req.URL.Path)
	}
}

func TestNamedPriorityMiddleware(t *testing.T) {
	ms := Middlewares{
		New("m1", 1, func(next http.Handler) http.Handler { return next }),
		New("m4", 3, func(next http.Handler) http.Handler { return next }),
		New("m2", 2, func(next http.Handler) http.Handler { return next }),
		New("m3", 3, func(next http.Handler) http.Handler { return next }),
	}
	if h := ms[0].Handler(nil); h != nil {
		t.Errorf("expect handler nil, but got %T", h)
	}

	Sort(ms)
	names := make([]string, len(ms))
	for i, m := range ms {
		names[i] = m.(interface{ Name() string }).Name()
	}

	expects := []string{"m1", "m2", "m4", "m3"}
	if !reflect.DeepEqual(names, expects) {
		t.Errorf("expect %v, but got %v", expects, names)
	}
}
